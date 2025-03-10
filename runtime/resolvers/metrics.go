package resolvers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/metricsview"
	"github.com/rilldata/rill/runtime/pkg/mapstructureutil"
)

func init() {
	runtime.RegisterResolverInitializer("metrics", newMetrics)
}

type metricsResolver struct {
	runtime        *runtime.Runtime
	instanceID     string
	executor       *metricsview.Executor
	query          *metricsview.Query
	args           *metricsResolverArgs
	claims         *runtime.SecurityClaims
	metricsHasTime bool
}

type metricsResolverArgs struct {
	Priority      int        `mapstructure:"priority"`
	ExecutionTime *time.Time `mapstructure:"execution_time"`
}

func newMetrics(ctx context.Context, opts *runtime.ResolverOptions) (runtime.Resolver, error) {
	qry := &metricsview.Query{}
	if err := mapstructureutil.WeakDecode(opts.Properties, qry); err != nil {
		return nil, err
	}

	args := &metricsResolverArgs{}
	if err := mapstructureutil.WeakDecode(opts.Args, args); err != nil {
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

	security, err := opts.Runtime.ResolveSecurity(opts.InstanceID, opts.Claims, res)
	if err != nil {
		return nil, err
	}

	if !security.CanAccess() {
		return nil, runtime.ErrForbidden
	}

	executor, err := metricsview.NewExecutor(ctx, opts.Runtime, opts.InstanceID, mv, res.GetMetricsView().State.Streaming, security, args.Priority)
	if err != nil {
		return nil, err
	}

	return &metricsResolver{
		runtime:        opts.Runtime,
		instanceID:     opts.InstanceID,
		executor:       executor,
		query:          qry,
		args:           args,
		claims:         opts.Claims,
		metricsHasTime: mv.TimeDimension != "",
	}, nil
}

func (r *metricsResolver) Close() error {
	r.executor.Close()
	return nil
}

func (r *metricsResolver) CacheKey(ctx context.Context) ([]byte, bool, error) {
	// get the underlying executor's cache key
	key, ok, err := cacheKeyForMetricsView(ctx, r.runtime, r.instanceID, r.query.MetricsView, r.args.Priority)
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

func (r *metricsResolver) Refs() []*runtimev1.ResourceName {
	return []*runtimev1.ResourceName{{Kind: runtime.ResourceKindMetricsView, Name: r.query.MetricsView}}
}

func (r *metricsResolver) Validate(ctx context.Context) error {
	return r.executor.ValidateQuery(r.query)
}

func (r *metricsResolver) ResolveInteractive(ctx context.Context) (runtime.ResolverResult, error) {
	if r.metricsHasTime {
		tsRes, err := resolveTimestampResult(ctx, r.runtime, r.instanceID, r.query.MetricsView, r.claims, r.args.Priority)
		if err != nil {
			return nil, err
		}

		err = r.executor.BindQuery(ctx, r.query, tsRes)
		if err != nil {
			return nil, err
		}
	}

	res, err := r.executor.Query(ctx, r.query, r.args.ExecutionTime)
	if err != nil {
		return nil, err
	}
	return runtime.NewDriverResolverResult(res), nil
}

func (r *metricsResolver) ResolveExport(ctx context.Context, w io.Writer, opts *runtime.ResolverExportOptions) error {
	return errors.New("not implemented")
}
