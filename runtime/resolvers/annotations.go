package resolvers

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/mapstructureutil"
)

func init() {
	runtime.RegisterResolverInitializer("annotations", newAnnotationsResolver)
}

type annotationsResolver struct {
	metricsView string
	annotation  *runtimev1.MetricsViewSpec_Annotation
	olap        drivers.OLAPStore
	olapCloser  func()
	args        *annotationsResolverArgs
}

type annotationsResolverProps struct {
	MetricsView string `mapstructure:"metrics_view"`
	Annotation  string `mapstructure:"annotation"`
}

type annotationsResolverArgs struct {
	Priority  int                  `mapstructure:"priority"`
	TimeRange *runtimev1.TimeRange `mapstructure:"time_range"`
	TimeGrain runtimev1.TimeGrain  `mapstructure:"time_grain"`
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

	security, err := opts.Runtime.ResolveSecurity(opts.InstanceID, opts.Claims, res)
	if err != nil {
		return nil, err
	}

	if !security.CanAccess() {
		return nil, runtime.ErrForbidden
	}

	accessibleMeasures := 0
	for _, measure := range annotation.Measures {
		if security.CanAccessField(measure) {
			accessibleMeasures++
		}
	}
	// None of the measures are accessible, so annotation is not accessible either
	// TODO: metrics view level annotation
	if accessibleMeasures == 0 && !annotation.Global {
		return nil, runtime.ErrForbidden
	}

	modelRes, err := ctrl.Get(ctx, &runtimev1.ResourceName{Kind: runtime.ResourceKindModel, Name: annotation.Model}, false)
	if err != nil {
		return nil, err
	}
	model := modelRes.GetModel().Spec

	olap, olapCloser, err := opts.Runtime.OLAP(ctx, opts.InstanceID, model.OutputConnector)

	return &annotationsResolver{
		metricsView: props.MetricsView,
		annotation:  annotation,
		olap:        olap,
		olapCloser:  olapCloser,
		args:        args,
	}, nil
}

func (r *annotationsResolver) Close() error {
	r.olapCloser()
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
	columns := "*"
	if r.annotation.HasGrain {
		columns += `,(CASE
  WHEN grain = 'millisecond' THEN 1
  WHEN grain = 'second' THEN 2
  WHEN grain = 'minute' THEN 3
  WHEN grain = 'hour' THEN 4
  WHEN grain = 'day' THEN 5
  WHEN grain = 'week' THEN 6
  WHEN grain = 'month' THEN 7
  WHEN grain = 'quarter' THEN 8
  WHEN grain = 'year' THEN 9
  ELSE 0
END) as time_grain`
	}

	var args []any

	if r.args.TimeRange == nil || r.args.TimeRange.Start == nil || r.args.TimeRange.End == nil {
		return nil, errors.New("time range is required")
	}

	start := r.args.TimeRange.Start.AsTime().UTC().Format(time.RFC3339)
	end := r.args.TimeRange.End.AsTime().UTC().Format(time.RFC3339)

	filter := " TRUE "
	if r.annotation.HasTimeEnd {
		filter += " AND ((time >= ? AND time < ?) OR (time_end >= ? AND time_end < ?))"
		args = append(args, start, end, start, end)
	} else {
		filter += " AND time >= ? AND time < ?"
		args = append(args, start, end)
	}
	if r.annotation.HasGrain && r.args.TimeGrain != runtimev1.TimeGrain_TIME_GRAIN_UNSPECIFIED {
		filter += fmt.Sprintf(" AND (time_grain == 0 OR time_grain >= ?)")
		args = append(args, int(r.args.TimeGrain))
	}

	sql := fmt.Sprintf("SELECT %s FROM %s WHERE %s", columns, r.annotation.Model, filter)

	res, err := r.olap.Query(ctx, &drivers.Statement{
		Query:    sql,
		Args:     args,
		Priority: 0,
	})
	if err != nil {
		return nil, err
	}

	return runtime.NewDriverResolverResult(res, nil), nil
}

func (r *annotationsResolver) ResolveExport(ctx context.Context, w io.Writer, opts *runtime.ResolverExportOptions) error {
	return errors.New("not implemented")
}
