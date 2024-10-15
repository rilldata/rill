package resolvers

import (
	"context"
	"errors"
	"fmt"
	"io"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/pkg/mapstructureutil"
	"github.com/rilldata/rill/runtime/queries"
)

func init() {
	runtime.RegisterResolverInitializer("metrics_time_range", newMetricsViewTimeRangeResolver)
}

type metricsViewTimeRangeResolver struct {
	runtime    *runtime.Runtime
	instanceID string
	query      *queries.MetricsViewTimeRange
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

	security, err := opts.Runtime.ResolveSecurity(opts.InstanceID, opts.Claims, res)
	if err != nil {
		return nil, err
	}

	if !security.CanAccess() {
		return nil, runtime.ErrForbidden
	}

	query := &queries.MetricsViewTimeRange{
		MetricsViewName:    tr.MetricsView,
		MetricsView:        mv,
		ResolvedMVSecurity: security,
	}

	return &metricsViewTimeRangeResolver{
		runtime:    opts.Runtime,
		instanceID: opts.InstanceID,
		query:      query,
		args:       args,
	}, nil
}

func (r *metricsViewTimeRangeResolver) Close() error {
	return nil
}

func (r *metricsViewTimeRangeResolver) Cacheable() bool {
	return true
}

func (r *metricsViewTimeRangeResolver) Key() string {
	return r.query.Key()
}

func (r *metricsViewTimeRangeResolver) Refs() []*runtimev1.ResourceName {
	return r.query.Deps()
}

func (r *metricsViewTimeRangeResolver) Validate(ctx context.Context) error {
	return nil
}

func (r *metricsViewTimeRangeResolver) ResolveInteractive(ctx context.Context) (runtime.ResolverResult, error) {
	err := r.runtime.Query(ctx, r.instanceID, r.query, r.args.Priority)
	if err != nil {
		return nil, err
	}

	// TODO :: Also add interval
	tr := r.query.Result.TimeRangeSummary
	row := map[string]any{}
	if tr.Min != nil {
		row["min"] = tr.Min.AsTime()
	}
	if tr.Max != nil {
		row["max"] = tr.Max.AsTime()
	}
	schema := &runtimev1.StructType{
		Fields: []*runtimev1.StructType_Field{
			{Name: "min", Type: &runtimev1.Type{Code: runtimev1.Type_CODE_TIMESTAMP}},
			{Name: "max", Type: &runtimev1.Type{Code: runtimev1.Type_CODE_TIMESTAMP}},
		},
	}
	return runtime.NewMapsResolverResult([]map[string]any{row}, schema), nil
}

func (r *metricsViewTimeRangeResolver) ResolveExport(ctx context.Context, w io.Writer, opts *runtime.ResolverExportOptions) error {
	return errors.New("not implemented")
}
