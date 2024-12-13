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
	"github.com/rilldata/rill/runtime/drivers"
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

	return &metricsViewCacheKeyResolver{
		runtime:    opts.Runtime,
		instanceID: opts.InstanceID,
		mvName:     tr.MetricsView,
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

func (r *metricsViewCacheKeyResolver) Key() string {
	var sb strings.Builder
	sb.WriteString(runtime.ResourceKindMetricsView)
	sb.WriteString(":")
	sb.WriteString(r.mvName)
	sb.WriteString(":")
	sb.WriteString("cache_key")
	if r.mv.Cache.KeyTtlSeconds != 0 {
		sb.WriteString(":")
		sb.WriteString(truncateTime(time.Now(), r.mv.Cache.KeyTtlSeconds).Format(time.RFC3339))
	}
	hash, err := hashstructure.Hash(sb.String(), hashstructure.FormatV2, nil)
	if err != nil {
		panic(err)
	}
	return strconv.FormatUint(hash, 16)
}

func (r *metricsViewCacheKeyResolver) Refs() []*runtimev1.ResourceName {
	return []*runtimev1.ResourceName{}
}

func (r *metricsViewCacheKeyResolver) Validate(ctx context.Context) error {
	return nil
}

func (r *metricsViewCacheKeyResolver) ResolveInteractive(ctx context.Context) (runtime.ResolverResult, error) {
	olap, release, err := r.runtime.OLAP(ctx, r.instanceID, r.mv.Connector)
	if err != nil {
		return nil, err
	}
	defer release()

	if !*r.mv.Cache.Enabled { // not enabled, ideally should not reach here
		return nil, errCachingDisabled
	}

	res, err := olap.Execute(ctx, &drivers.Statement{
		Query:    r.mv.Cache.KeySql,
		Priority: r.args.Priority,
	})
	if err != nil {
		return nil, err
	}
	res.SetCap(1)
	return runtime.NewDriverResolverResult(res), nil
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
