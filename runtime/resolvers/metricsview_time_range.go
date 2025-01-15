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
	"github.com/rilldata/rill/runtime/pkg/mapstructureutil"
)

const (
	hourInDay = 24
)

var microsInDay = hourInDay * time.Hour.Microseconds()

func init() {
	runtime.RegisterResolverInitializer("metrics_time_range", newMetricsViewTimeRangeResolver)
}

type metricsViewTimeRangeResolver struct {
	runtime    *runtime.Runtime
	instanceID string
	mvName     string
	executor   *metricsview.Executor
	args       *metricsViewTimeRangeResolverArgs
}

type metricsViewTimeRangeResolverArgs struct {
	Priority int `mapstructure:"priority"`
}

type metricsViewTimeRange struct {
	MetricsView string `mapstructure:"metrics_view"`
}

func newMetricsViewTimeRangeResolver(ctx context.Context, opts *runtime.ResolverOptions) (runtime.Resolver, error) {
	tr := &metricsViewTimeRange{}
	if err := mapstructureutil.WeakDecode(opts.Properties, tr); err != nil {
		return nil, err
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

	if mv.TimeDimension == "" {
		return nil, fmt.Errorf("metrics view '%s' does not have a time dimension", tr.MetricsView)
	}

	security, err := opts.Runtime.ResolveSecurity(opts.InstanceID, opts.Claims, res)
	if err != nil {
		return nil, err
	}

	if !security.CanAccess() {
		return nil, runtime.ErrForbidden
	}

	ex, err := metricsview.NewExecutor(ctx, opts.Runtime, opts.InstanceID, mv, false, security, args.Priority)
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
	return nil
}

func (r *metricsViewTimeRangeResolver) CacheKey(ctx context.Context) ([]byte, bool, error) {
	return cacheKeyForMetricsView(ctx, r.runtime, r.instanceID, r.mvName, r.args.Priority)
}

func (r *metricsViewTimeRangeResolver) Refs() []*runtimev1.ResourceName {
	return []*runtimev1.ResourceName{{Kind: runtime.ResourceKindMetricsView, Name: r.mvName}}
}

func (r *metricsViewTimeRangeResolver) Validate(ctx context.Context) error {
	return nil
}

func (r *metricsViewTimeRangeResolver) ResolveInteractive(ctx context.Context) (runtime.ResolverResult, error) {
	ts, err := r.executor.Timestamps(ctx, nil)
	if err != nil {
		return nil, err
	}

	row := map[string]any{}
	if !ts.Min.IsZero() {
		row["min"] = ts.Min
		row["max"] = ts.Max
		row["watermark"] = ts.Watermark
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

func durationToInterval(duration time.Duration) map[string]any {
	hours := duration.Hours()
	days := int32(0)
	if hours >= hourInDay {
		days = int32(hours / hourInDay)
	}
	micros := duration.Microseconds() - microsInDay*int64(days)
	return map[string]any{
		"days":   days,
		"months": 0,
		"micros": micros,
	}
}
