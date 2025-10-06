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

func init() {
	runtime.RegisterResolverInitializer("metrics_time_range", newMetricsViewTimeRangeResolver)
}

type metricsViewTimeRangeResolver struct {
	runtime    *runtime.Runtime
	instanceID string
	mvName     string
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

	ex, err := executor.New(ctx, opts.Runtime, opts.InstanceID, mv, false, security, args.Priority)
	if err != nil {
		return nil, err
	}

	return &metricsViewTimeRangeResolver{
		runtime:    opts.Runtime,
		instanceID: opts.InstanceID,
		mvName:     tr.MetricsView,
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

	row := map[string]any{}
	if !ts.Min.IsZero() {
		row["min"] = ts.Min
		row["max"] = ts.Max
		row["watermark"] = ts.Watermark
	} else {
		row["min"] = nil
		row["max"] = nil
		row["watermark"] = nil
	}
	schema := &runtimev1.StructType{
		Fields: []*runtimev1.StructType_Field{
			{Name: "min", Type: &runtimev1.Type{Code: runtimev1.Type_CODE_TIMESTAMP, Nullable: true}},
			{Name: "max", Type: &runtimev1.Type{Code: runtimev1.Type_CODE_TIMESTAMP, Nullable: true}},
			{Name: "watermark", Type: &runtimev1.Type{Code: runtimev1.Type_CODE_TIMESTAMP, Nullable: true}},
		},
	}
	return runtime.NewMapsResolverResult([]map[string]any{row}, schema), nil
}

func (r *metricsViewTimeRangeResolver) ResolveExport(ctx context.Context, w io.Writer, opts *runtime.ResolverExportOptions) error {
	return errors.New("not implemented")
}

func (r *metricsViewTimeRangeResolver) InferRequiredSecurityRules() ([]*runtimev1.SecurityRule, error) {
	return nil, errors.New("security rule inference not implemented")
}

func resolveTimestampResult(ctx context.Context, rt *runtime.Runtime, instanceID, metricsViewName, timeDimension string, security *runtime.SecurityClaims, priority int) (metricsview.TimestampsResult, error) {
	res, err := rt.Resolve(ctx, &runtime.ResolveOptions{
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

	row, err := res.Next()
	if err != nil {
		if errors.Is(err, io.EOF) {
			return metricsview.TimestampsResult{}, errors.New("time range query returned no results")
		}
		return metricsview.TimestampsResult{}, err
	}

	tsRes := metricsview.TimestampsResult{}

	tsRes.Min, err = anyToTime(row["min"])
	if err != nil {
		return tsRes, err
	}
	tsRes.Max, err = anyToTime(row["max"])
	if err != nil {
		return tsRes, err
	}
	tsRes.Watermark, err = anyToTime(row["watermark"])
	if err != nil {
		return tsRes, err
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
