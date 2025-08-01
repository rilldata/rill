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
	"github.com/rilldata/rill/runtime/pkg/mapstructureutil"
)

func init() {
	runtime.RegisterResolverInitializer("annotations", newAnnotationsResolver)
}

type annotationsResolver struct {
	instanceID  string
	metricsView string
	annotation  string
	args        *annotationsResolverArgs
	executor    *metricsview.Executor
	runtime     *runtime.Runtime
}

type annotationsResolverProps struct {
	MetricsView string `mapstructure:"metrics_view"`
	Annotation  string `mapstructure:"annotation"`
}

type annotationsResolverArgs struct {
	Priority   int                           `mapstructure:"priority"`
	TimeRange  *metricsview.TimeRange        `mapstructure:"time_range"`
	TimeGrain  runtimev1.TimeGrain           `mapstructure:"time_grain"`
	TimeZone   string                        `mapstructure:"time_zone"`
	Limit      *int64                        `mapstructure:"limit"`
	Offset     *int64                        `mapstructure:"offset"`
	Timestamps *metricsview.TimestampsResult `mapstructure:"timestamps"`
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

	return &annotationsResolver{
		instanceID:  opts.InstanceID,
		metricsView: props.MetricsView,
		annotation:  props.Annotation,
		args:        args,
		executor:    ex,
		runtime:     opts.Runtime,
	}, nil
}

func (r *annotationsResolver) Close() error {
	r.executor.Close()
	return nil
}

func (r *annotationsResolver) CacheKey(ctx context.Context) ([]byte, bool, error) {
	key := annotationResolverKey{
		annotationsResolverArgs:  *r.args,
		annotationsResolverProps: annotationsResolverProps{
			MetricsView: r.metricsView,
			Annotation:  r.annotation,
		},
	}
	kb, err := json.Marshal(&key)
	if err != nil {
		panic(err)
	}
	ks := fmt.Sprintf("MetricsViewAnnotations:%s", string(kb))
	return []byte(ks), true, nil
}

func (r *annotationsResolver) Refs() []*runtimev1.ResourceName {
	return []*runtimev1.ResourceName{
		{Kind: runtime.ResourceKindMetricsView, Name: r.metricsView},
	}
}

func (r *annotationsResolver) Validate(ctx context.Context) error {
	return nil
}

func (r *annotationsResolver) ResolveInteractive(ctx context.Context) (runtime.ResolverResult, error) {
	qry := &metricsview.AnnotationsQuery{
		Annotation: r.annotation,
		TimeRange:  r.args.TimeRange,
		Limit:      r.args.Limit,
		Offset:     r.args.Offset,
		TimeZone:   r.args.TimeZone,
		TimeGrain:  r.args.TimeGrain,
	}
	if r.args.Timestamps != nil {
		err := r.executor.BindAnnotationsQuery(ctx, qry, *r.args.Timestamps)
		if err != nil {
			return nil, err
		}
	}

	res, err := r.executor.Annotations(ctx, qry)
	if err != nil {
		return nil, err
	}

	return runtime.NewDriverResolverResult(res, nil), nil
}

func (r *annotationsResolver) ResolveExport(ctx context.Context, w io.Writer, opts *runtime.ResolverExportOptions) error {
	return errors.New("not implemented")
}

type annotationResolverKey struct {
	annotationsResolverArgs
	annotationsResolverProps
}
