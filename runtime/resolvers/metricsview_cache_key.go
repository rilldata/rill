package resolvers

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/metricsview/executor"
	"github.com/rilldata/rill/runtime/pkg/mapstructureutil"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func init() {
	runtime.RegisterResolverInitializer("metrics_cache_key", newMetricsViewCacheKeyResolver)
}

type metricsViewCacheKeyResolver struct {
	runtime    *runtime.Runtime
	instanceID string
	mvName     string
	mv         *runtimev1.MetricsViewSpec
	streaming  bool
	executor   *executor.Executor
	args       *metricsViewCacheKeyResolverArgs
}

type metricsViewCacheKeyResolverArgs struct {
	Priority int `mapstructure:"priority"`
}

type metricsViewCacheKeyProps struct {
	MetricsView string `mapstructure:"metrics_view"`
}

func newMetricsViewCacheKeyResolver(ctx context.Context, opts *runtime.ResolverOptions) (runtime.Resolver, error) {
	tr := &metricsViewCacheKeyProps{}
	if err := mapstructureutil.WeakDecode(opts.Properties, tr); err != nil {
		return nil, err
	}

	args := &metricsViewCacheKeyResolverArgs{}
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

	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		span.SetAttributes(
			attribute.String("metrics_view", tr.MetricsView),
			attribute.Bool("streaming", res.GetMetricsView().State.Streaming),
		)
	}

	security, err := opts.Runtime.ResolveSecurity(ctx, opts.InstanceID, opts.Claims, res)
	if err != nil {
		return nil, err
	}

	if !security.CanAccess() {
		return nil, runtime.ErrForbidden
	}

	executor, err := executor.New(ctx, opts.Runtime, opts.InstanceID, mv, res.GetMetricsView().State.Streaming, security, args.Priority)
	if err != nil {
		return nil, err
	}

	return &metricsViewCacheKeyResolver{
		runtime:    opts.Runtime,
		instanceID: opts.InstanceID,
		mvName:     tr.MetricsView,
		streaming:  res.GetMetricsView().State.Streaming,
		executor:   executor,
		mv:         mv,
		args:       args,
	}, nil
}

func (r *metricsViewCacheKeyResolver) Close() error {
	r.executor.Close()
	return nil
}

func (r *metricsViewCacheKeyResolver) CacheKey(ctx context.Context) ([]byte, bool, error) {
	var key string
	ttl := time.Duration(r.mv.CacheKeyTtlSeconds) * time.Second
	if ttl == 0 && r.streaming {
		// If streaming, we need to cache the key for 60 seconds
		// For non streaming metrics view we don't need to expire the key as it will be invalidated basis the ref's state version
		ttl = time.Minute
	}
	if ttl != 0 {
		key = time.Now().Truncate(ttl).Format(time.RFC3339)
	}
	return []byte(key), true, nil
}

func (r *metricsViewCacheKeyResolver) Refs() []*runtimev1.ResourceName {
	return []*runtimev1.ResourceName{{Kind: runtime.ResourceKindMetricsView, Name: r.mvName}}
}

func (r *metricsViewCacheKeyResolver) Validate(ctx context.Context) error {
	return nil
}

func (r *metricsViewCacheKeyResolver) ResolveInteractive(ctx context.Context) (runtime.ResolverResult, error) {
	key, ok, err := r.executor.CacheKey(ctx)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, runtime.ErrMetricsViewCachingDisabled
	}
	schema := &runtimev1.StructType{
		Fields: []*runtimev1.StructType_Field{
			{Name: "key", Type: &runtimev1.Type{Code: runtimev1.Type_CODE_STRING}},
		},
	}
	return runtime.NewMapsResolverResult([]map[string]interface{}{{"key": key}}, schema), nil
}

func (r *metricsViewCacheKeyResolver) ResolveExport(ctx context.Context, w io.Writer, opts *runtime.ResolverExportOptions) error {
	return errors.New("not implemented")
}

func (r *metricsViewCacheKeyResolver) InferRequiredSecurityRules() ([]*runtimev1.SecurityRule, error) {
	return nil, errors.New("security rule inference not implemented")
}
