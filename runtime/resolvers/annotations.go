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
)

func init() {
	runtime.RegisterResolverInitializer("annotations", newAnnotationsResolver)
}

type annotationsResolver struct {
	metricsView string
	annotation  *runtimev1.MetricsViewSpec_Annotation
	args        *annotationsResolverArgs
	executor    *metricsview.Executor
}

type annotationsResolverProps struct {
	MetricsView string `mapstructure:"metrics_view"`
	Annotation  string `mapstructure:"annotation"`
}

type annotationsResolverArgs struct {
	Priority  int                    `mapstructure:"priority"`
	TimeRange *metricsview.TimeRange `mapstructure:"time_range"`
	TimeGrain runtimev1.TimeGrain    `mapstructure:"time_grain"`
}

func newAnnotationsResolver(ctx context.Context, opts *runtime.ResolverOptions) (runtime.Resolver, error) {
	props := &annotationsResolverProps{}
	if err := mapstructureutil.WeakDecode(opts.Properties, props); err != nil {
		return nil, err
	}

	args := &annotationsResolverArgs{}
	if err := mapstructureutil.WeakDecode(opts.Args, args); err != nil {
		return nil, err
	}

	ctrl, err := opts.Runtime.Controller(ctx, opts.InstanceID)
	if err != nil {
		return nil, err
	}

	res, err := ctrl.Get(ctx, &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: props.MetricsView}, false)
	if err != nil {
		return nil, err
	}

	mv := res.GetMetricsView().State.ValidSpec
	if mv == nil {
		return nil, fmt.Errorf("metrics view %q is invalid", res.Meta.Name.Name)
	}

	security, err := opts.Runtime.ResolveSecurity(opts.InstanceID, opts.Claims, res)
	if err != nil {
		return nil, err
	}

	ex, err := metricsview.NewExecutor(ctx, opts.Runtime, opts.InstanceID, mv, false, security, args.Priority)
	if err != nil {
		return nil, err
	}

	var annotation *runtimev1.MetricsViewSpec_Annotation
	for _, specAnnotation := range mv.Annotations {
		if specAnnotation.Name == props.Annotation {
			annotation = specAnnotation
			break
		}
	}
	if annotation == nil {
		return nil, fmt.Errorf("annotation %q not found in metrics view %q", props.Annotation, props.MetricsView)
	}

	return &annotationsResolver{
		metricsView: props.MetricsView,
		annotation:  annotation,
		args:        args,
		executor:    ex,
	}, nil
}

func (r *annotationsResolver) Close() error {
	r.executor.Close()
	return nil
}

func (r *annotationsResolver) CacheKey(ctx context.Context) ([]byte, bool, error) {
	return nil, false, nil
}

func (r *annotationsResolver) Refs() []*runtimev1.ResourceName {
	return []*runtimev1.ResourceName{
		{Kind: runtime.ResourceKindMetricsView, Name: r.metricsView},
		{Kind: runtime.ResourceKindModel, Name: r.annotation.Model},
	}
}

func (r *annotationsResolver) Validate(ctx context.Context) error {
	return nil
}

func (r *annotationsResolver) ResolveInteractive(ctx context.Context) (runtime.ResolverResult, error) {
	res, err := r.executor.Annotations(ctx, r.annotation, r.args.TimeRange, r.args.TimeGrain)
	if err != nil {
		return nil, err
	}

	return runtime.NewDriverResolverResult(res, nil), nil
}

func (r *annotationsResolver) ResolveExport(ctx context.Context, w io.Writer, opts *runtime.ResolverExportOptions) error {
	return errors.New("not implemented")
}
