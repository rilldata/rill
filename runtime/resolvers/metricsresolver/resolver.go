package metricsresolver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"strconv"
	"time"

	"github.com/mitchellh/hashstructure/v2"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/duration"
	"github.com/rilldata/rill/runtime/pkg/mapstructureutil"
	"github.com/rilldata/rill/runtime/pkg/timeutil"
	"github.com/rilldata/rill/runtime/queries"
)

type Resolver struct {
	query               *Query
	metricsView         *runtimev1.MetricsViewSpec
	security            *runtime.ResolvedMetricsViewSecurity
	olap                drivers.OLAPStore
	olapRelease         func()
	dialect             drivers.Dialect
	exporting           bool
	priority            int
	executionTime       *time.Time
	interactiveRowLimit int64

	// Cache of the resolved time anchor for relative time ranges
	timeAnchor time.Time
}

type resolverArgs struct {
	Priority      int        `mapstructure:"priority"`
	ExecutionTime *time.Time `mapstructure:"execution_time"`
}

func New(ctx context.Context, opts *runtime.ResolverOptions) (runtime.Resolver, error) {
	qry := &Query{}
	if err := mapstructureutil.WeakDecode(opts.Properties, qry); err != nil {
		return nil, err
	}

	args := &resolverArgs{}
	if err := mapstructureutil.WeakDecode(opts.Args, args); err != nil {
		return nil, err
	}

	inst, err := opts.Runtime.Instance(ctx, opts.InstanceID)
	if err != nil {
		return nil, err
	}

	cfg, err := inst.Config()
	if err != nil {
		return nil, err
	}

	ctrl, err := opts.Runtime.Controller(ctx, opts.InstanceID)
	if err != nil {
		return nil, err
	}

	res, err := ctrl.Get(ctx, &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: qry.MetricsView}, false)
	if err != nil {
		return nil, err
	}

	mv := res.GetMetricsView().State.ValidSpec
	if mv == nil {
		return nil, fmt.Errorf("metrics view %q is invalid", res.Meta.Name.Name)
	}

	security, err := opts.Runtime.ResolveMetricsViewSecurity(opts.UserAttributes, opts.InstanceID, mv, res.Meta.StateUpdatedOn.AsTime())
	if err != nil {
		return nil, err
	}

	if security != nil && !security.Access {
		return nil, queries.ErrForbidden
	}

	olap, release, err := opts.Runtime.OLAP(ctx, opts.InstanceID, mv.Connector)
	if err != nil {
		return nil, err
	}

	return &Resolver{
		query:               qry,
		metricsView:         mv,
		security:            security,
		olap:                olap,
		olapRelease:         release,
		dialect:             olap.Dialect(),
		exporting:           opts.ForExport,
		priority:            args.Priority,
		executionTime:       args.ExecutionTime,
		interactiveRowLimit: cfg.InteractiveSQLRowLimit,
	}, nil
}

func (r *Resolver) Close() error {
	r.olapRelease()
	return nil
}

func (r *Resolver) Key() string {
	hash, err := hashstructure.Hash(r.query, hashstructure.FormatV2, nil)
	if err != nil {
		panic(err)
	}
	return strconv.FormatUint(hash, 16)
}

func (r *Resolver) Refs() []*runtimev1.ResourceName {
	return []*runtimev1.ResourceName{{Kind: runtime.ResourceKindMetricsView, Name: r.query.MetricsView}}
}

func (r *Resolver) Validate(ctx context.Context) error {
	// TODO: E.g. check dims/measures exist?
	return nil
}

func (r *Resolver) ResolveInteractive(ctx context.Context) (*runtime.ResolverResult, error) {
	err := r.rewriteQueryTimeRanges(ctx)
	if err != nil {
		return nil, err
	}

	ast, err := BuildAST(r.metricsView, r.security, r.query, r.dialect)
	if err != nil {
		return nil, err
	}

	sql, args, err := ast.SQL()
	if err != nil {
		return nil, err
	}

	log.Printf("SQL: %s", sql)

	res, err := r.olap.Execute(ctx, &drivers.Statement{
		Query:    sql,
		Args:     args,
		Priority: r.priority,
	})
	if err != nil {
		return nil, err
	}
	defer res.Close()

	out := []map[string]any{}
	for res.Rows.Next() {
		if int64(len(out)) >= r.interactiveRowLimit {
			return nil, fmt.Errorf("sql resolver: interactive query limit exceeded: returned more than %d rows", r.interactiveRowLimit)
		}

		row := make(map[string]any)
		err = res.Rows.MapScan(row)
		if err != nil {
			return nil, err
		}
		out = append(out, row)
	}

	data, err := json.Marshal(out)
	if err != nil {
		return nil, err
	}

	// This is a little hacky, but for now we only cache results from DuckDB queries
	var cache bool
	if r.olap.Dialect() == drivers.DialectDuckDB {
		cache = true
	}

	return &runtime.ResolverResult{
		Data:   data,
		Schema: res.Schema,
		Cache:  cache,
	}, nil
}

func (r *Resolver) ResolveExport(ctx context.Context, w io.Writer, opts *runtime.ResolverExportOptions) error {
	return errors.New("not implemented")
}

// rewriteQueryTimeRanges rewrites the time ranges in the query to fixed start/end timestamps.
func (r *Resolver) rewriteQueryTimeRanges(ctx context.Context) error {
	tz := time.UTC
	if r.query.TimeZone != nil {
		var err error
		tz, err = time.LoadLocation(*r.query.TimeZone)
		if err != nil {
			return fmt.Errorf("invalid time zone %q: %w", *r.query.TimeZone, err)
		}
	}

	err := r.resolveTimeRange(ctx, r.query.TimeRange, tz)
	if err != nil {
		return fmt.Errorf("failed to resolve time range: %w", err)
	}

	err = r.resolveTimeRange(ctx, r.query.ComparisonTimeRange, tz)
	if err != nil {
		return fmt.Errorf("failed to resolve comparison time range: %w", err)
	}

	return nil
}

// resolveTimeRange resolves the given time range, ensuring only its Start and End properties are populated.
func (r *Resolver) resolveTimeRange(ctx context.Context, tr *TimeRange, tz *time.Location) error {
	if tr == nil || tr.IsZero() {
		return nil
	}

	if tr.Start.IsZero() && tr.End.IsZero() {
		t, err := r.resolveTimeAnchor(ctx)
		if err != nil {
			return err
		}
		tr.End = t
	}

	var isISO bool
	if tr.IsoDuration != "" {
		d, err := duration.ParseISO8601(tr.IsoDuration)
		if err != nil {
			return fmt.Errorf("invalid iso_duration %q: %w", tr.IsoDuration, err)
		}

		if !tr.Start.IsZero() && !tr.End.IsZero() {
			return errors.New(`cannot resolve "iso_duration" for a time range with fixed "start" and "end" timestamps`)
		} else if !tr.Start.IsZero() {
			tr.End = d.Add(tr.Start)
		} else if !tr.End.IsZero() {
			tr.Start = d.Sub(tr.End)
		} else {
			// In practice, this shouldn't happen since we resolve a time anchor dynamically if both start and end are zero.
			return errors.New(`cannot resolve "iso_duration" for a time range without "start" and "end" timestamps`)
		}

		isISO = true
	}

	if tr.IsoOffset != "" {
		d, err := duration.ParseISO8601(tr.IsoOffset)
		if err != nil {
			return fmt.Errorf("invalid iso_offset %q: %w", tr.IsoOffset, err)
		}

		if !tr.Start.IsZero() {
			tr.Start = d.Sub(tr.Start)
		}
		if !tr.End.IsZero() {
			tr.End = d.Sub(tr.End)
		}

		isISO = true
	}

	// Only modify the start and end if ISO duration or offset was sent.
	// This is to maintain backwards compatibility for calls from the UI.
	if isISO {
		fdow := int(r.metricsView.FirstDayOfWeek)
		if fdow > 7 || fdow <= 0 {
			fdow = 1
		}
		fmoy := int(r.metricsView.FirstMonthOfYear)
		if fmoy > 12 || fmoy <= 0 {
			fmoy = 1
		}
		if !tr.RoundToGrain.Valid() {
			return fmt.Errorf("invalid time grain %q", tr.RoundToGrain)
		}
		if tr.RoundToGrain != TimeGrainUnspecified {
			if !tr.Start.IsZero() {
				tr.Start = timeutil.TruncateTime(tr.Start, tr.RoundToGrain.ToTimeutil(), tz, fdow, fmoy)
			}
			if !tr.End.IsZero() {
				tr.End = timeutil.TruncateTime(tr.End, tr.RoundToGrain.ToTimeutil(), tz, fdow, fmoy)
			}
		}
	}

	// Clear all other fields than Start and End
	tr.IsoDuration = ""
	tr.IsoOffset = ""
	tr.RoundToGrain = TimeGrainUnspecified

	return nil
}

// resolveTimeAnchor resolves a time anchor based on the metric view's watermark expression.
// If the resolved time anchor is zero, it defaults to the current time.
func (r *Resolver) resolveTimeAnchor(ctx context.Context) (time.Time, error) {
	if !r.timeAnchor.IsZero() {
		return r.timeAnchor, nil
	}

	if r.executionTime != nil {
		return *r.executionTime, nil
	}

	var expr string
	if r.metricsView.WatermarkExpression != "" {
		expr = r.metricsView.WatermarkExpression
	} else if r.metricsView.TimeDimension != "" {
		expr = fmt.Sprintf("MAX(%s)", r.dialect.EscapeIdentifier(r.metricsView.TimeDimension))
	} else {
		return time.Time{}, errors.New("cannot determine time anchor for relative time range")
	}

	sql := fmt.Sprintf("SELECT %s FROM %s", expr, r.dialect.EscapeTable(r.metricsView.Database, r.metricsView.DatabaseSchema, r.metricsView.Table))

	res, err := r.olap.Execute(ctx, &drivers.Statement{
		Query:    sql,
		Priority: r.priority,
	})
	if err != nil {
		return time.Time{}, err
	}
	defer res.Close()

	var t time.Time
	for res.Next() {
		if err := res.Scan(&t); err != nil {
			return time.Time{}, fmt.Errorf("failed to scan time anchor: %w", err)
		}
	}

	if t.IsZero() {
		return time.Now(), nil
	}

	r.timeAnchor = t
	return t, nil
}
