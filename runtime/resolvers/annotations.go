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
	"github.com/rilldata/rill/runtime/metricsview/executor"
	"github.com/rilldata/rill/runtime/pkg/mapstructureutil"
)

func init() {
	runtime.RegisterResolverInitializer("metrics_annotations", newAnnotationsResolver)
}

type metricsAnnotationsResolver struct {
	instanceID string
	query      *metricsview.AnnotationsQuery
	mv         *runtimev1.MetricsViewSpec
	executor   *executor.Executor
	runtime    *runtime.Runtime
	claims     *runtime.SecurityClaims
}

func newAnnotationsResolver(ctx context.Context, opts *runtime.ResolverOptions) (runtime.Resolver, error) {
	qry := &metricsview.AnnotationsQuery{}
	if err := mapstructureutil.WeakDecode(opts.Properties, qry); err != nil {
		return nil, err
	}

	ctrl, err := opts.Runtime.Controller(ctx, opts.InstanceID)
	if err != nil {
		return nil, err
	}

	res, err := ctrl.Get(ctx, &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: qry.MetricsView}, false)
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

	ex, err := executor.New(ctx, opts.Runtime, opts.InstanceID, mv, false, security, qry.Priority)
	if err != nil {
		return nil, err
	}

	return &metricsAnnotationsResolver{
		instanceID: opts.InstanceID,
		query:      qry,
		mv:         mv,
		executor:   ex,
		runtime:    opts.Runtime,
		claims:     opts.Claims,
	}, nil
}

func (r *metricsAnnotationsResolver) Close() error {
	r.executor.Close()
	return nil
}

func (r *metricsAnnotationsResolver) CacheKey(ctx context.Context) ([]byte, bool, error) {
	// get the underlying executor's cache key
	key, ok, err := cacheKeyForMetricsView(ctx, r.runtime, r.instanceID, r.query.MetricsView, r.query.Priority)
	if err != nil {
		return nil, false, err
	}
	if !ok {
		return nil, false, nil
	}

	queryMap, err := r.query.AsMap()
	if err != nil {
		return nil, false, err
	}

	queryMap["mv_cache_key"] = key

	b, err := json.Marshal(queryMap)
	return b, true, err
}

func (r *metricsAnnotationsResolver) Refs() []*runtimev1.ResourceName {
	return []*runtimev1.ResourceName{
		{Kind: runtime.ResourceKindMetricsView, Name: r.query.MetricsView},
	}
}

func (r *metricsAnnotationsResolver) Validate(ctx context.Context) error {
	return nil
}

func (r *metricsAnnotationsResolver) ResolveInteractive(ctx context.Context) (runtime.ResolverResult, error) {
	// Only resolve time stamps if an absolute time range is not specified.
	if r.query.TimeRange == nil || r.query.TimeRange.Start.IsZero() || r.query.TimeRange.End.IsZero() {
		tsRes, err := resolveTimestampResult(ctx, r.runtime, r.instanceID, r.query.MetricsView, r.mv.TimeDimension, r.claims, r.query.Priority)
		if err != nil {
			return nil, err
		}

		if r.query.TimeRange == nil || r.query.TimeRange.IsZero() {
			r.query.TimeRange = &metricsview.TimeRange{
				Start: tsRes.Min,
				End:   tsRes.Max,
			}
		}

		err = r.executor.BindAnnotationsQuery(ctx, r.query, tsRes)
		if err != nil {
			return nil, err
		}
	}

	res, err := r.executor.Annotations(ctx, r.query)
	if err != nil {
		return nil, err
	}

	return runtime.NewMapsResolverResult(res, nil), nil
}

func (r *metricsAnnotationsResolver) ResolveExport(ctx context.Context, w io.Writer, opts *runtime.ResolverExportOptions) error {
	return errors.New("not implemented")
}

func (r *metricsAnnotationsResolver) InferRequiredSecurityRules() ([]*runtimev1.SecurityRule, error) {
	return nil, errors.New("security rule inference not implemented")
}
