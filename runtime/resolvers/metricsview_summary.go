package resolvers

import (
	"context"
	"errors"
	"fmt"
	"io"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/metricsview"
	"github.com/rilldata/rill/runtime/pkg/mapstructureutil"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func init() {
	runtime.RegisterResolverInitializer("metrics_summary", newMetricsViewSummaryResolver)
}

type metricsViewSummaryResolver struct {
	runtime    *runtime.Runtime
	instanceID string
	mvName     string
	executor   *metricsview.Executor
	args       *metricsViewSummaryResolverArgs
}

type metricsViewSummaryResolverArgs struct {
	Priority      int    `mapstructure:"priority"`
	TimeDimension string `mapstructure:"time_dimension"` // if empty, the default time dimension in mv is used
}

type metricsViewSummary struct {
	MetricsView string `mapstructure:"metrics_view"`
}

func newMetricsViewSummaryResolver(ctx context.Context, opts *runtime.ResolverOptions) (runtime.Resolver, error) {
	tr := &metricsViewSummary{}
	if err := mapstructureutil.WeakDecode(opts.Properties, tr); err != nil {
		return nil, err
	}

	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		span.SetAttributes(attribute.String("metrics_view", tr.MetricsView))
	}

	args := &metricsViewSummaryResolverArgs{}
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

	return &metricsViewSummaryResolver{
		runtime:    opts.Runtime,
		instanceID: opts.InstanceID,
		mvName:     tr.MetricsView,
		executor:   ex,
		args:       args,
	}, nil
}

func (r *metricsViewSummaryResolver) Close() error {
	r.executor.Close()
	return nil
}

func (r *metricsViewSummaryResolver) CacheKey(ctx context.Context) ([]byte, bool, error) {
	return cacheKeyForMetricsView(ctx, r.runtime, r.instanceID, r.mvName, r.args.Priority)
}

func (r *metricsViewSummaryResolver) Refs() []*runtimev1.ResourceName {
	return []*runtimev1.ResourceName{{Kind: runtime.ResourceKindMetricsView, Name: r.mvName}}
}

func (r *metricsViewSummaryResolver) Validate(ctx context.Context) error {
	return nil
}

func (r *metricsViewSummaryResolver) ResolveInteractive(ctx context.Context) (runtime.ResolverResult, error) {
	ts, err := r.executor.Timestamps(ctx, r.args.TimeDimension)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch timestamps for metrics view '%s': %w", r.mvName, err)
	}

	// Fetch summary statistics (data type and sample value for each dimension)
	summary, err := r.executor.Summary(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch summary statistics for metrics view '%s': %w", r.mvName, err)
	}

	row := map[string]any{
		"min":       ts.Min,
		"max":       ts.Max,
		"watermark": ts.Watermark,
		"summary":   summary,
	}

	schema := &runtimev1.StructType{
		Fields: []*runtimev1.StructType_Field{
			{Name: "min", Type: &runtimev1.Type{Code: runtimev1.Type_CODE_TIMESTAMP, Nullable: true}},
			{Name: "max", Type: &runtimev1.Type{Code: runtimev1.Type_CODE_TIMESTAMP, Nullable: true}},
			{Name: "watermark", Type: &runtimev1.Type{Code: runtimev1.Type_CODE_TIMESTAMP, Nullable: true}},
			{Name: "summary", Type: &runtimev1.Type{Code: runtimev1.Type_CODE_STRUCT, Nullable: true}},
		},
	}

	return runtime.NewMapsResolverResult([]map[string]any{row}, schema), nil
}

func (r *metricsViewSummaryResolver) ResolveExport(ctx context.Context, w io.Writer, opts *runtime.ResolverExportOptions) error {
	return errors.New("not implemented")
}
