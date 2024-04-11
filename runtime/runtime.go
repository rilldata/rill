package runtime

import (
	"context"
	"errors"
	"fmt"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/pkg/conncache"
	"github.com/rilldata/rill/runtime/pkg/email"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
)

var tracer = otel.Tracer("github.com/rilldata/rill/runtime")

type Options struct {
	MetastoreConnector           string
	SystemConnectors             []*runtimev1.Connector
	ConnectionCacheSize          int
	QueryCacheSizeBytes          int64
	SecurityEngineCacheSize      int
	ControllerLogBufferCapacity  int
	ControllerLogBufferSizeBytes int64
	AllowHostAccess              bool
	EmitInstanceHeartBeat        bool
}

type Runtime struct {
	Email                    *email.Client
	opts                     *Options
	logger                   *zap.Logger
	activity                 *activity.Client
	metastore                drivers.Handle
	registryCache            *registryCache
	connCache                conncache.Cache
	queryCache               *queryCache
	securityEngine           *securityEngine
	instanceHeartbeatEmitter *instanceHeartbeatEmitter
}

func New(ctx context.Context, opts *Options, logger *zap.Logger, ac *activity.Client, emailClient *email.Client) (*Runtime, error) {
	if emailClient == nil {
		emailClient = email.New(email.NewNoopSender())
	}

	rt := &Runtime{
		Email:          emailClient,
		opts:           opts,
		logger:         logger,
		activity:       ac,
		queryCache:     newQueryCache(opts.QueryCacheSizeBytes),
		securityEngine: newSecurityEngine(opts.SecurityEngineCacheSize, logger),
	}

	rt.connCache = rt.newConnectionCache()

	store, _, err := rt.AcquireSystemHandle(ctx, opts.MetastoreConnector)
	if err != nil {
		return nil, err
	}
	rt.metastore = store
	reg, ok := rt.metastore.AsRegistry()
	if !ok {
		return nil, fmt.Errorf("metastore must be a valid registry")
	}

	rt.registryCache = newRegistryCache(rt, reg, logger, ac)
	err = rt.registryCache.init(ctx)
	if err != nil {
		return nil, err
	}
	if rt.opts.EmitInstanceHeartBeat {
		rt.instanceHeartbeatEmitter = newInstanceHeartbeatEmitter(rt, time.Minute)
		rt.instanceHeartbeatEmitter.emit()
	}

	return rt, nil
}

func (r *Runtime) AllowHostAccess() bool {
	return r.opts.AllowHostAccess
}

func (r *Runtime) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	r.registryCache.close(ctx)
	err1 := r.queryCache.close()
	err2 := r.connCache.Close(ctx) // Also closes metastore // TODO: Propagate ctx cancellation
	if r.instanceHeartbeatEmitter != nil {
		r.instanceHeartbeatEmitter.close()
	}
	return errors.Join(err1, err2)
}

func (r *Runtime) ResolveMetricsViewSecurity(attributes map[string]any, instanceID string, mv *runtimev1.MetricsViewSpec, lastUpdatedOn time.Time) (*ResolvedMetricsViewSecurity, error) {
	inst, err := r.Instance(context.Background(), instanceID)
	if err != nil {
		return nil, err
	}
	return r.securityEngine.resolveMetricsViewSecurity(instanceID, inst.Environment, mv, lastUpdatedOn, attributes)
}

// GetInstanceAttributes fetches an instance and converts its annotations to attributes
// nil is returned if an error occurred or instance was not found
func (r *Runtime) GetInstanceAttributes(ctx context.Context, instanceID string) []attribute.KeyValue {
	instance, err := r.Instance(ctx, instanceID)
	if err != nil {
		return nil
	}

	return instanceAnnotationsToAttribs(instance)
}

func instanceAnnotationsToAttribs(instance *drivers.Instance) []attribute.KeyValue {
	attrs := make([]attribute.KeyValue, 0, len(instance.Annotations)+1)
	attrs = append(attrs, attribute.String("instance_id", instance.ID))
	for k, v := range instance.Annotations {
		attrs = append(attrs, attribute.String(k, v))
	}
	return attrs
}
