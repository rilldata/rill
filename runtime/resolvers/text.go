package resolvers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/metricsview"
	"github.com/rilldata/rill/runtime/parser"
	"github.com/rilldata/rill/runtime/pkg/mapstructureutil"
)

func init() {
	runtime.RegisterResolverInitializer("text", newText)
}

// textProps contains the static properties for the text resolver.
type textProps struct {
	Text                         string                             `mapstructure:"text"`
	UseFormatTokens              bool                               `mapstructure:"use_format_tokens"`
	TimeZone                     string                             `mapstructure:"time_zone"`
	AdditionalWhereByMetricsView map[string]*metricsview.Expression `mapstructure:"additional_where_by_metrics_view"`
	AdditionalTimeRange          *metricsview.TimeRange             `mapstructure:"additional_time_range"`
}

func newText(ctx context.Context, opts *runtime.ResolverOptions) (runtime.Resolver, error) {
	props := &textProps{}
	if err := mapstructureutil.WeakDecode(opts.Properties, props); err != nil {
		return nil, err
	}

	r := &textResolver{
		rt:         opts.Runtime,
		instanceID: opts.InstanceID,
		claims:     opts.Claims,
		props:      props,
	}

	if err := r.analyzeAndPopulateRefs(ctx); err != nil {
		return nil, err
	}

	return r, nil
}

type textResolver struct {
	rt         *runtime.Runtime
	instanceID string
	claims     *runtime.SecurityClaims
	props      *textProps
	refs       []*runtimev1.ResourceName
}

var _ runtime.Resolver = &textResolver{}

func (r *textResolver) Close() error {
	return nil
}

func (r *textResolver) CacheKey(ctx context.Context) ([]byte, bool, error) {
	return nil, false, nil
}

func (r *textResolver) Refs() []*runtimev1.ResourceName {
	return r.refs
}

func (r *textResolver) Validate(ctx context.Context) error {
	return nil
}

func (r *textResolver) ResolveInteractive(ctx context.Context) (runtime.ResolverResult, error) {
	inst, err := r.rt.Instance(ctx, r.instanceID)
	if err != nil {
		return nil, err
	}

	// Cache for looking up whether a field is a measure.
	measuresCache := make(map[string]map[string]bool)
	isMeasure := func(metricsView, field string) (bool, error) {
		if metricsView == "" {
			return false, nil
		}
		measures, ok := measuresCache[metricsView]
		if !ok {
			_, mv, err := lookupMetricsView(ctx, r.rt, r.instanceID, metricsView)
			if err != nil {
				return false, err
			}
			measures = make(map[string]bool)
			for _, m := range mv.ValidSpec.Measures {
				measures[m.Name] = true
			}
			measuresCache[metricsView] = measures
		}
		_, ok = measures[field]
		return ok, nil
	}

	// Utility function for resolving metrics SQL.
	resolveMetricsSQL := func(sql string, unary bool) ([]map[string]any, error) {
		resolveRes, _, err := r.rt.Resolve(ctx, &runtime.ResolveOptions{
			InstanceID: r.instanceID,
			Resolver:   "metrics_sql",
			ResolverProperties: map[string]any{
				"sql":                              sql,
				"time_zone":                        r.props.TimeZone,
				"additional_where_by_metrics_view": r.props.AdditionalWhereByMetricsView,
				"additional_time_range":            r.props.AdditionalTimeRange,
			},
			Claims: r.claims,
		})
		if err != nil {
			return nil, err
		}
		defer resolveRes.Close()

		var rows []map[string]any
		for {
			row, err := resolveRes.Next()
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
				return nil, fmt.Errorf("failed to get result: %w", err)
			}
			if len(rows) > 0 && unary {
				return nil, fmt.Errorf("metrics_sql in templating must return one row, but the query returned multiple")
			}
			rows = append(rows, row)
		}

		if unary {
			if len(rows) != 1 {
				return nil, fmt.Errorf("metrics_sql in templating must return one row, got none")
			}
			if len(rows[0]) != 1 {
				return nil, fmt.Errorf("metrics_sql in templating only allows one result field, got %d", len(rows[0]))
			}
		}

		// When using format tokens, wrap each measure value with a format token.
		if r.props.UseFormatTokens {
			var mv string
			if meta := resolveRes.Meta(); meta != nil {
				mv, _ = meta["metrics_view"].(string)
			}

			for _, row := range rows {
				for field, val := range row {
					ok, err := isMeasure(mv, field)
					if err != nil {
						return nil, err
					}
					if !ok {
						continue
					}

					data, err := json.Marshal(textFormatToken{MetricsView: mv, Field: field, Value: val})
					if err != nil {
						return nil, fmt.Errorf("failed to marshal measure value %v as JSON: %w", val, err)
					}
					row[field] = fmt.Sprintf("__RILL__FORMAT__(%s)", string(data))
				}
			}
		}

		return rows, nil
	}

	templateData := parser.TemplateData{
		User:      r.claims.UserAttributes,
		Variables: inst.ResolveVariables(false),
		State:     make(map[string]any),
		Resolve: func(ref parser.ResourceName) (string, error) {
			return ref.Name, nil
		},
		ExtraFuncs: map[string]any{
			"metrics_sql": func(sql string) (string, error) {
				rows, err := resolveMetricsSQL(sql, true)
				if err != nil {
					return "", err
				}

				if len(rows) > 0 {
					for _, val := range rows[0] {
						if val, ok := val.(string); ok {
							return val, nil
						}
						return fmt.Sprintf("%v", val), nil
					}
				}
				return "", fmt.Errorf("unreachable: no value in single-column single-row result")
			},
			"metrics_sql_rows": func(sql string) (any, error) {
				return resolveMetricsSQL(sql, false)
			},
		},
	}

	text, err := parser.ResolveTemplate(r.props.Text, templateData, false)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve template: %w", err)
	}

	rows := []map[string]any{{"text": text}}
	schema := &runtimev1.StructType{
		Fields: []*runtimev1.StructType_Field{
			{Name: "text", Type: &runtimev1.Type{Code: runtimev1.Type_CODE_STRING}},
		},
	}
	return runtime.NewMapsResolverResult(rows, schema), nil
}

func (r *textResolver) ResolveExport(ctx context.Context, w io.Writer, opts *runtime.ResolverExportOptions) error {
	return fmt.Errorf("text resolver does not support export")
}

func (r *textResolver) InferRequiredSecurityRules() ([]*runtimev1.SecurityRule, error) {
	return nil, nil
}

// analyzeAndPopulateRefs runs a stub template resolution to discover metrics_sql calls,
// then initializes metrics_sql resolvers to discover the metrics view refs they depend on.
func (r *textResolver) analyzeAndPopulateRefs(ctx context.Context) error {
	// Collect SQL strings by running the template with stub functions.
	var sqls []string
	stubs := map[string]any{
		"metrics_sql": func(sql string) (string, error) {
			sqls = append(sqls, sql)
			return "", nil
		},
		"metrics_sql_rows": func(sql string) (any, error) {
			sqls = append(sqls, sql)
			return []map[string]any{}, nil
		},
	}

	inst, err := r.rt.Instance(ctx, r.instanceID)
	if err != nil {
		return err
	}

	td := parser.TemplateData{
		User:      r.claims.UserAttributes,
		Variables: inst.ResolveVariables(false),
		State:     make(map[string]any),
		Resolve: func(ref parser.ResourceName) (string, error) {
			return ref.Name, nil
		},
		ExtraFuncs: stubs,
	}

	// Ignore errors; we only care about collecting SQL strings.
	_, _ = parser.ResolveTemplate(r.props.Text, td, false)

	// For each collected SQL, initialize a metrics_sql resolver to discover refs.
	initializer, ok := runtime.ResolverInitializers["metrics_sql"]
	if !ok {
		return nil
	}

	seen := make(map[string]bool)
	for _, sql := range sqls {
		res, err := initializer(ctx, &runtime.ResolverOptions{
			Runtime:    r.rt,
			InstanceID: r.instanceID,
			Properties: map[string]any{"sql": sql},
			Claims: &runtime.SecurityClaims{
				UserID:         r.claims.UserID,
				UserAttributes: r.claims.UserAttributes,
				SkipChecks:     true,
			},
		})
		if err != nil {
			continue // Skip SQL strings that fail to parse
		}
		for _, ref := range res.Refs() {
			if ref.Kind == runtime.ResourceKindMetricsView {
				seen[ref.Name] = true
			}
		}
		res.Close()
	}

	for name := range seen {
		r.refs = append(r.refs, &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: name})
	}
	return nil
}

// textFormatToken is the payload inside a __RILL__FORMAT__(...) token generated by the text resolver.
type textFormatToken struct {
	MetricsView string `json:"metrics_view"`
	Field       string `json:"field"`
	Value       any    `json:"value"`
}
