package resolvers

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/mitchellh/hashstructure/v2"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/metricsview"
	"github.com/rilldata/rill/runtime/pkg/mapstructureutil"
)

var errCachingDisabled = errors.New("metrics_cache_key: caching is disabled")

func init() {
	runtime.RegisterResolverInitializer("metrics_cache_key", newMetricsViewCacheKeyResolver)
}

type metricsViewCacheKeyResolver struct {
	runtime    *runtime.Runtime
	instanceID string
	mvName     string
	mv         *runtimev1.MetricsViewSpec
	streaming  bool
	exectuor   *metricsview.Executor
	args       *metricsViewCacheKeyResolverArgs
}

type metricsViewCacheKeyResolverArgs struct {
	Priority int `mapstructure:"priority"`
}

type metricsViewCacheKey struct {
	MetricsView string `mapstructure:"metrics_view"`
}

func newMetricsViewCacheKeyResolver(ctx context.Context, opts *runtime.ResolverOptions) (runtime.Resolver, error) {
	tr := &metricsViewCacheKey{}
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

	return &metricsViewCacheKeyResolver{
		runtime:    opts.Runtime,
		instanceID: opts.InstanceID,
		mvName:     tr.MetricsView,
		streaming:  res.GetMetricsView().State.Streaming,
		exectuor:   executor,
		mv:         mv,
		args:       args,
	}, nil
}

func (r *metricsViewCacheKeyResolver) Close() error {
	return nil
}

func (r *metricsViewCacheKeyResolver) Cacheable() bool {
	return true
}

func (r *metricsViewCacheKeyResolver) CacheKey(ctx context.Context) ([]byte, bool, error) {
	var sb strings.Builder
	sb.WriteString(runtime.ResourceKindMetricsView)
	sb.WriteString(":")
	sb.WriteString(r.mvName)
	sb.WriteString(":")
	sb.WriteString("cahe_key")
	ttlSeconds := r.mv.CacheKeyTtlSeconds
	if ttlSeconds == 0 && r.streaming {
		// If streaming, we need to cache the key for 60 seconds
		// For non streaming metrics view we don't need to expire the key as data itself will be invalidated basis the ref's state version
		ttlSeconds = 60
	}
	if ttlSeconds != 0 {
		sb.WriteString(":")
		sb.WriteString(truncateTime(time.Now(), ttlSeconds).Format(time.RFC3339))
	}
	hash, err := hashstructure.Hash(sb.String(), hashstructure.FormatV2, nil)
	if err != nil {
		return nil, false, err
	}
	return []byte(strconv.FormatUint(hash, 16)), true, nil
}

func (r *metricsViewCacheKeyResolver) Refs() []*runtimev1.ResourceName {
	return []*runtimev1.ResourceName{}
}

func (r *metricsViewCacheKeyResolver) Validate(ctx context.Context) error {
	return nil
}

func (r *metricsViewCacheKeyResolver) ResolveInteractive(ctx context.Context) (runtime.ResolverResult, error) {
	key, ok, err := r.exectuor.CacheKey(ctx)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errCachingDisabled
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

func truncateTime(t time.Time, seconds int64) time.Time {
	// Convert x seconds to a duration
	duration := time.Duration(seconds) * time.Second
	// Truncate the time to the nearest x seconds
	return t.Truncate(duration)
}
