package resolvers

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/metricsview"
	"github.com/rilldata/rill/runtime/metricsview/executor"
	"github.com/rilldata/rill/runtime/pkg/mapstructureutil"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const defaultTimestampsCacheTTL = 5 * time.Minute

func init() {
	runtime.RegisterResolverInitializer("metrics_time_range", newMetricsViewTimeRangeResolver)
}

type metricsViewTimeRangeResolver struct {
	runtime    *runtime.Runtime
	instanceID string
	mvName     string
	mv         *runtimev1.MetricsViewSpec
	executor   *executor.Executor
	args       *metricsViewTimeRangeResolverArgs
}

type metricsViewTimeRangeResolverArgs struct {
	Priority      int    `mapstructure:"priority"`
	TimeDimension string `mapstructure:"time_dimension"` // if empty, the default time dimension in mv is used
}

type metricsViewTimeRange struct {
	MetricsView string `mapstructure:"metrics_view"`
}

func newMetricsViewTimeRangeResolver(ctx context.Context, opts *runtime.ResolverOptions) (runtime.Resolver, error) {
	tr := &metricsViewTimeRange{}
	if err := mapstructureutil.WeakDecode(opts.Properties, tr); err != nil {
		return nil, err
	}

	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		span.SetAttributes(attribute.String("metrics_view", tr.MetricsView))
	}

	args := &metricsViewTimeRangeResolverArgs{}
	if err := mapstructureutil.WeakDecode(opts.Args, args); err != nil {
		return nil, err
	}

	ctrl, err := opts.Runtime.Controller(ctx, opts.InstanceID)
	if err != nil {
		return nil, err
	}

	res, err := ctrl.Get(ctx, &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: tr.MetricsView}, false)
	if err != nil {
		return nil, err
	}

	mv := res.GetMetricsView().State.ValidSpec
	if mv == nil {
		return nil, fmt.Errorf("metrics view %q is invalid", res.Meta.Name.Name)
	}

	if mv.TimeDimension == "" && args.TimeDimension == "" {
		return nil, fmt.Errorf("no time dimension specified for metrics view %q", tr.MetricsView)
	}

	security, err := opts.Runtime.ResolveSecurity(ctx, opts.InstanceID, opts.Claims, res)
	if err != nil {
		return nil, err
	}
	if !security.CanAccess() {
		return nil, runtime.ErrForbidden
	}
	var userAttrs map[string]any
	if opts.Claims != nil {
		userAttrs = opts.Claims.UserAttributes
	}

	ex, err := executor.New(ctx, opts.Runtime, opts.InstanceID, mv, false, security, args.Priority, userAttrs)
	if err != nil {
		return nil, err
	}

	return &metricsViewTimeRangeResolver{
		runtime:    opts.Runtime,
		instanceID: opts.InstanceID,
		mvName:     tr.MetricsView,
		mv:         mv,
		executor:   ex,
		args:       args,
	}, nil
}

func (r *metricsViewTimeRangeResolver) Close() error {
	r.executor.Close()
	return nil
}

func (r *metricsViewTimeRangeResolver) CacheKey(ctx context.Context) ([]byte, bool, error) {
	key, ok, err := cacheKeyForMetricsView(ctx, r.runtime, r.instanceID, r.mvName, r.args.Priority)
	if err != nil {
		return nil, false, err
	}

	// When spec has rollups and MV-level caching is disabled, use TTL-based caching to ensure rollup timestamp queries are cached
	if !ok && len(r.mv.Rollups) > 0 {
		ttl := defaultTimestampsCacheTTL
		if r.mv.CacheTimestampsTtlSeconds > 0 {
			ttl = time.Duration(r.mv.CacheTimestampsTtlSeconds) * time.Second
		}
		bucket := time.Now().Truncate(ttl).Unix()
		key = []byte(fmt.Sprintf("ts:%d", bucket))
		ok = true
	}

	key = append(key, []byte(r.args.TimeDimension)...)
	return key, ok, nil
}

func (r *metricsViewTimeRangeResolver) Refs() []*runtimev1.ResourceName {
	return []*runtimev1.ResourceName{{Kind: runtime.ResourceKindMetricsView, Name: r.mvName}}
}

func (r *metricsViewTimeRangeResolver) Validate(ctx context.Context) error {
	return nil
}

func (r *metricsViewTimeRangeResolver) ResolveInteractive(ctx context.Context) (runtime.ResolverResult, error) {
	ts, err := r.executor.Timestamps(ctx, r.args.TimeDimension)
	if err != nil {
		return nil, err
	}

	// Build base row
	baseRow := map[string]any{"table": ""}
	if !ts.Min.IsZero() {
		baseRow["min"] = ts.Min
		baseRow["max"] = ts.Max
		baseRow["watermark"] = ts.Watermark
	} else {
		baseRow["min"] = nil
		baseRow["max"] = nil
		baseRow["watermark"] = nil
	}

	rows := []map[string]any{baseRow}

	// add rollups to subsequent rows
	for table, rts := range ts.Rollups {
		row := map[string]any{"table": table}
		if !rts.Min.IsZero() {
			row["min"] = rts.Min
			row["max"] = rts.Max
		} else {
			row["min"] = nil
			row["max"] = nil
		}
		rows = append(rows, row)
	}

	schema := &runtimev1.StructType{
		Fields: []*runtimev1.StructType_Field{
			{Name: "table", Type: &runtimev1.Type{Code: runtimev1.Type_CODE_STRING}},
			{Name: "min", Type: &runtimev1.Type{Code: runtimev1.Type_CODE_TIMESTAMP, Nullable: true}},
			{Name: "max", Type: &runtimev1.Type{Code: runtimev1.Type_CODE_TIMESTAMP, Nullable: true}},
			{Name: "watermark", Type: &runtimev1.Type{Code: runtimev1.Type_CODE_TIMESTAMP, Nullable: true}},
		},
	}
	return runtime.NewMapsResolverResult(rows, schema), nil
}

func (r *metricsViewTimeRangeResolver) ResolveExport(ctx context.Context, w io.Writer, opts *runtime.ResolverExportOptions) error {
	return errors.New("not implemented")
}

func (r *metricsViewTimeRangeResolver) InferRequiredSecurityRules() ([]*runtimev1.SecurityRule, error) {
	return nil, errors.New("security rule inference not implemented")
}

// resolveTimestampResult resolves timestamps for a metrics view including rollup data.
// It parses all rows: the first row (table="") populates the base TimestampsResult, and subsequent rows (table=name) populate TimestampsResult.Rollups.
func resolveTimestampResult(ctx context.Context, rt *runtime.Runtime, instanceID, metricsViewName, timeDimension string, security *runtime.SecurityClaims, priority int) (metricsview.TimestampsResult, error) {
	res, _, err := rt.Resolve(ctx, &runtime.ResolveOptions{
		InstanceID: instanceID,
		Resolver:   "metrics_time_range",
		ResolverProperties: map[string]any{
			"metrics_view": metricsViewName,
		},
		Args: map[string]any{
			"priority":       priority,
			"time_dimension": timeDimension,
		},
		Claims: security,
	})
	if err != nil {
		return metricsview.TimestampsResult{}, err
	}
	defer res.Close()

	var tsRes metricsview.TimestampsResult
	hasBase := false
	for {
		row, err := res.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return metricsview.TimestampsResult{}, err
		}

		table, _ := row["table"].(string)
		mn, err := anyToTime(row["min"])
		if err != nil {
			return metricsview.TimestampsResult{}, err
		}
		mx, err := anyToTime(row["max"])
		if err != nil {
			return metricsview.TimestampsResult{}, err
		}

		if table == "" {
			hasBase = true
			tsRes.Min = mn
			tsRes.Max = mx
			tsRes.Watermark, err = anyToTime(row["watermark"])
			if err != nil {
				return metricsview.TimestampsResult{}, err
			}
		} else {
			if tsRes.Rollups == nil {
				tsRes.Rollups = make(map[string]metricsview.TimestampsResult)
			}
			tsRes.Rollups[table] = metricsview.TimestampsResult{Min: mn, Max: mx}
		}
	}

	if !hasBase {
		return metricsview.TimestampsResult{}, errors.New("time range query returned no results")
	}

	return tsRes, nil
}

func anyToTime(tm any) (time.Time, error) {
	if tm == nil {
		return time.Time{}, nil
	}

	tmStr, ok := tm.(string)
	if !ok {
		t, ok := tm.(time.Time)
		if !ok {
			return time.Time{}, fmt.Errorf("unable to convert type %T to Time", tm)
		}
		return t, nil
	}
	return time.Parse(time.RFC3339Nano, tmStr)
}
