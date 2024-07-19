package resolvers

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/mitchellh/hashstructure/v2"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/metricsview"
	"github.com/rilldata/rill/runtime/pkg/mapstructureutil"
)

func init() {
	runtime.RegisterResolverInitializer("metrics", newMetrics)
}

type metricsResolver struct {
	runtime    *runtime.Runtime
	instanceID string
	executor   *metricsview.Executor
	query      *metricsview.Query
	args       *metricsResolverArgs
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

	executor, err := metricsview.NewExecutor(ctx, opts.Runtime, opts.InstanceID, mv, security, args.Priority)
	if err != nil {
		return nil, err
	}
	defer executor.Close()

	return &metricsResolver{
		runtime:    opts.Runtime,
		instanceID: opts.InstanceID,
		executor:   executor,
		query:      qry,
		args:       args,
	}, nil
}

func (r *metricsResolver) Close() error {
	r.executor.Close()
	return nil
}

func (r *metricsResolver) Cacheable() bool {
	return r.executor.Cacheable(r.query)
}

func (r *metricsResolver) Key() string {
	hash, err := hashstructure.Hash(r.query, hashstructure.FormatV2, nil)
	if err != nil {
		panic(err)
	}
	return strconv.FormatUint(hash, 16)
}

func (r *metricsResolver) Refs() []*runtimev1.ResourceName {
	return []*runtimev1.ResourceName{{Kind: runtime.ResourceKindMetricsView, Name: r.query.MetricsView}}
}

func (r *metricsResolver) Validate(ctx context.Context) error {
	return r.executor.ValidateQuery(r.query)
}

func (r *metricsResolver) ResolveInteractive(ctx context.Context) (runtime.ResolverResult, error) {
	res, err := r.executor.Query(ctx, r.query, r.args.ExecutionTime)
	if err != nil {
		return nil, err
	}
	return runtime.NewDriverResolverResult(res), nil
}

func (r *metricsResolver) ResolveExport(ctx context.Context, w io.Writer, opts *runtime.ResolverExportOptions) error {
	return errors.New("not implemented")
}
