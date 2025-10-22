package resolvers

import (
	"context"
	"errors"
	"fmt"
	"io"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/metricsview/executor"
	"github.com/rilldata/rill/runtime/pkg/mapstructureutil"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func init() {
	runtime.RegisterResolverInitializer("metrics_summary", newMetricsSummaryResolver)
}

type metricsSummaryResolver struct {
	runtime    *runtime.Runtime
	instanceID string
	mvName     string
	executor   *executor.Executor
	args       *metricsSummaryArgs
}

type metricsSummaryArgs struct {
	Priority int `mapstructure:"priority"`
}

type metricsSummaryProps struct {
	MetricsView string `mapstructure:"metrics_view"`
}

func newMetricsSummaryResolver(ctx context.Context, opts *runtime.ResolverOptions) (runtime.Resolver, error) {
	tr := &metricsSummaryProps{}
	if err := mapstructureutil.WeakDecode(opts.Properties, tr); err != nil {
		return nil, err
	}

	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		span.SetAttributes(attribute.String("metrics_view", tr.MetricsView))
	}

	args := &metricsSummaryArgs{}
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

	return &metricsSummaryResolver{
		runtime:    opts.Runtime,
		instanceID: opts.InstanceID,
		mvName:     tr.MetricsView,
		executor:   ex,
		args:       args,
	}, nil
}

func (r *metricsSummaryResolver) Close() error {
	r.executor.Close()
	return nil
}

func (r *metricsSummaryResolver) CacheKey(ctx context.Context) ([]byte, bool, error) {
	return cacheKeyForMetricsView(ctx, r.runtime, r.instanceID, r.mvName, r.args.Priority)
}

func (r *metricsSummaryResolver) Refs() []*runtimev1.ResourceName {
	return []*runtimev1.ResourceName{{Kind: runtime.ResourceKindMetricsView, Name: r.mvName}}
}

func (r *metricsSummaryResolver) Validate(ctx context.Context) error {
	return nil
}

func (r *metricsSummaryResolver) ResolveInteractive(ctx context.Context) (runtime.ResolverResult, error) {
	summary, err := r.executor.Summary(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch summary for metrics view '%s': %w", r.mvName, err)
	}

	row := map[string]any{
		"dimensions": summary.Dimensions,
		"time_range": summary.DefaultTimeDimension,
	}

	schema := &runtimev1.StructType{
		Fields: []*runtimev1.StructType_Field{
			{Name: "dimensions", Type: &runtimev1.Type{Code: runtimev1.Type_CODE_ARRAY, Nullable: true}},
			{Name: "time_range", Type: &runtimev1.Type{Code: runtimev1.Type_CODE_STRUCT, Nullable: true}},
		},
	}

	return runtime.NewMapsResolverResult([]map[string]any{row}, schema), nil
}

func (r *metricsSummaryResolver) ResolveExport(ctx context.Context, w io.Writer, opts *runtime.ResolverExportOptions) error {
	return errors.New("not implemented")
}

func (r *metricsSummaryResolver) InferRequiredSecurityRules() ([]*runtimev1.SecurityRule, error) {
	return nil, errors.New("security rule inference not implemented")
}
