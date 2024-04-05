package runtime

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"sync"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"golang.org/x/sync/errgroup"
)

const validateConcurrencyLimit = 10

type ValidateMetricsViewResult struct {
	TimeDimensionErr error
	DimensionErrs    []IndexErr
	MeasureErrs      []IndexErr
	OtherErrs        []error
}

type IndexErr struct {
	Idx int
	Err error
}

func (r *ValidateMetricsViewResult) IsZero() bool {
	return r.TimeDimensionErr == nil && len(r.DimensionErrs) == 0 && len(r.MeasureErrs) == 0 && len(r.OtherErrs) == 0
}

// Error returns a single error containing all validation errors.
// If there are no errors, it returns nil.
func (r *ValidateMetricsViewResult) Error() error {
	var errs []error
	errs = append(errs, r.TimeDimensionErr)
	for _, e := range r.DimensionErrs {
		errs = append(errs, e.Err)
	}
	for _, e := range r.MeasureErrs {
		errs = append(errs, e.Err)
	}
	errs = append(errs, r.OtherErrs...)

	// NOTE: errors.Join returns nil if all input errs are nil.
	return errors.Join(errs...)
}

// ValidateMetricsView validates a metrics view spec.
// NOTE: If we need validation for more resources, we should consider moving it to the queries (or a dedicated validation package).
func (r *Runtime) ValidateMetricsView(ctx context.Context, instanceID string, mv *runtimev1.MetricsViewSpec) (*ValidateMetricsViewResult, error) {
	ctrl, err := r.Controller(ctx, instanceID)
	if err != nil {
		return nil, err
	}

	olap, release, err := ctrl.AcquireOLAP(ctx, mv.Connector)
	if err != nil {
		return nil, err
	}
	defer release()

	// Create the result
	res := &ValidateMetricsViewResult{}

	// Check underlying table exists
	t, err := olap.InformationSchema().Lookup(ctx, mv.Table)
	if err != nil {
		if errors.Is(err, drivers.ErrNotFound) {
			res.OtherErrs = append(res.OtherErrs, fmt.Errorf("table %q does not exist", mv.Table))
			return res, nil
		}
		return nil, fmt.Errorf("could not find table %q: %w", mv.Table, err)
	}

	fields := make(map[string]*runtimev1.StructType_Field, len(t.Schema.Fields))
	for _, f := range t.Schema.Fields {
		fields[strings.ToLower(f.Name)] = f
	}

	// Check time dimension exists
	if mv.TimeDimension != "" {
		f, ok := fields[strings.ToLower(mv.TimeDimension)]
		if !ok {
			res.TimeDimensionErr = fmt.Errorf("timeseries %q is not a column in table %q", mv.TimeDimension, mv.Table)
		} else if f.Type.Code != runtimev1.Type_CODE_TIMESTAMP && f.Type.Code != runtimev1.Type_CODE_DATE {
			res.TimeDimensionErr = fmt.Errorf("timeseries %q is not a TIMESTAMP column", mv.TimeDimension)
		}
	}

	// For performance, attempt to validate all dimensions and measures at once
	err = validateAllDimensionsAndMeasures(ctx, olap, t, mv)
	if err != nil {
		// One or more dimension/measure expressions failed to validate. We need to check each one individually to provide useful errors.
		validateIndividualDimensionsAndMeasures(ctx, olap, t, mv, fields, res)
	}

	// Check the default theme exists
	if mv.DefaultTheme != "" {
		_, err := ctrl.Get(ctx, &runtimev1.ResourceName{Kind: ResourceKindTheme, Name: mv.DefaultTheme}, false)
		if err != nil {
			if errors.Is(err, drivers.ErrNotFound) {
				res.OtherErrs = append(res.OtherErrs, fmt.Errorf("theme %q does not exist", mv.DefaultTheme))
			}
			return nil, fmt.Errorf("could not find theme %q: %w", mv.DefaultTheme, err)
		}
	}

	return res, nil
}

// validateAllDimensionsAndMeasures validates all dimensions and measures with one query. It returns an error if any of the expressions are invalid.
func validateAllDimensionsAndMeasures(ctx context.Context, olap drivers.OLAPStore, t *drivers.Table, mv *runtimev1.MetricsViewSpec) error {
	var dimExprs []string
	var groupIndexes []string
	for idx, d := range mv.Dimensions {
		if d.Column != "" {
			dimExprs = append(dimExprs, olap.Dialect().EscapeIdentifier(d.Column))
		} else {
			dimExprs = append(dimExprs, "("+d.Expression+")")
		}
		groupIndexes = append(groupIndexes, strconv.Itoa(idx+1))
	}
	var metricExprs []string
	for _, m := range mv.Measures {
		metricExprs = append(metricExprs, "("+m.Expression+")")
	}
	var query string
	if len(dimExprs) == 0 && len(metricExprs) == 0 {
		// No metric and dimension, nothing to check
		return nil
	}
	if len(dimExprs) == 0 {
		// Only metrics
		query = fmt.Sprintf("SELECT 1, %s FROM %s GROUP BY 1", strings.Join(metricExprs, ","), olap.Dialect().EscapeIdentifier(t.Name))
	} else if len(metricExprs) == 0 {
		// No metrics
		query = fmt.Sprintf("SELECT %s FROM %s GROUP BY %s", strings.Join(dimExprs, ","), olap.Dialect().EscapeIdentifier(t.Name), strings.Join(groupIndexes, ","))
	} else {
		query = fmt.Sprintf("SELECT %s, %s FROM %s GROUP BY %s", strings.Join(dimExprs, ","), strings.Join(metricExprs, ","), olap.Dialect().EscapeIdentifier(t.Name), strings.Join(groupIndexes, ","))
	}
	err := olap.Exec(ctx, &drivers.Statement{
		Query:  query,
		DryRun: true,
	})
	if err != nil {
		return fmt.Errorf("failed to validate dims and metrics: %w", err)
	}
	return nil
}

// validateIndividualDimensionsAndMeasures validates each dimension and measure individually.
// It adds validation errors to the provided res.
func validateIndividualDimensionsAndMeasures(ctx context.Context, olap drivers.OLAPStore, t *drivers.Table, mv *runtimev1.MetricsViewSpec, fields map[string]*runtimev1.StructType_Field, res *ValidateMetricsViewResult) {
	// Validate dimensions and measures concurrently with a limit of 10 concurrent validations
	var mu sync.Mutex
	var grp errgroup.Group
	grp.SetLimit(validateConcurrencyLimit)

	// Check dimension expressions are valid
	for idx, d := range mv.Dimensions {
		idx := idx
		grp.Go(func() error {
			err := validateDimension(ctx, olap, t, d, fields)
			if err != nil {
				mu.Lock()
				defer mu.Unlock()

				res.DimensionErrs = append(res.DimensionErrs, IndexErr{
					Idx: idx,
					Err: err,
				})
			}
			return nil
		})
	}

	// Check measure expressions are valid
	for idx, m := range mv.Measures {
		idx := idx
		grp.Go(func() error {
			err := validateMeasure(ctx, olap, t, m)
			if err != nil {
				mu.Lock()
				defer mu.Unlock()

				res.MeasureErrs = append(res.MeasureErrs, IndexErr{
					Idx: idx,
					Err: fmt.Errorf("invalid expression for measure %q: %w", m.Name, err),
				})
			}
			return nil
		})
	}

	// Wait for all validations to complete
	_ = grp.Wait()

	// Sort errors by index (for stable output)
	slices.SortFunc(res.DimensionErrs, func(a, b IndexErr) int { return a.Idx - b.Idx })
	slices.SortFunc(res.MeasureErrs, func(a, b IndexErr) int { return a.Idx - b.Idx })
}

// validateDimension validates a metrics view dimension.
func validateDimension(ctx context.Context, olap drivers.OLAPStore, t *drivers.Table, d *runtimev1.MetricsViewSpec_DimensionV2, fields map[string]*runtimev1.StructType_Field) error {
	if d.Column != "" {
		if _, isColumn := fields[strings.ToLower(d.Column)]; !isColumn {
			return fmt.Errorf("failed to validate dimension %q: column %q not found in table", d.Name, d.Column)
		}
		return nil
	}

	err := olap.Exec(ctx, &drivers.Statement{
		Query:  fmt.Sprintf("SELECT (%s) FROM %s GROUP BY 1", d.Expression, olap.Dialect().EscapeIdentifier(t.Name)),
		DryRun: true,
	})
	if err != nil {
		return fmt.Errorf("failed to validate expression for dimension %q: %w", d.Name, err)
	}
	return nil
}

// validateMeasure validates a metrics view measure.
func validateMeasure(ctx context.Context, olap drivers.OLAPStore, t *drivers.Table, m *runtimev1.MetricsViewSpec_MeasureV2) error {
	err := olap.Exec(ctx, &drivers.Statement{
		Query:  fmt.Sprintf("SELECT 1, (%s) FROM %s GROUP BY 1", m.Expression, olap.Dialect().EscapeIdentifier(t.Name)),
		DryRun: true,
	})
	return err
}
