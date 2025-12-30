package runtime

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/rilldata/rill/cli/pkg/version"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/parser"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/pkg/conncache"
	"github.com/rilldata/rill/runtime/pkg/email"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/rilldata/rill/runtime/pkg/singleflight"
	"github.com/rilldata/rill/runtime/storage"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
	"golang.org/x/exp/maps"
	"google.golang.org/protobuf/types/known/structpb"
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
	Version                      version.Version
}

type Runtime struct {
	Email          *email.Client
	opts           *Options
	Logger         *zap.Logger
	storage        *storage.Client
	activity       *activity.Client
	metastore      drivers.Handle
	registryCache  *registryCache
	connCache      conncache.Cache
	queryCache     *queryCache
	securityEngine *securityEngine
	configReloader *configReloader
}

func New(ctx context.Context, opts *Options, logger *zap.Logger, st *storage.Client, ac *activity.Client, emailClient *email.Client) (*Runtime, error) {
	if emailClient == nil {
		emailClient = email.New(email.NewNoopSender())
	}

	rt := &Runtime{
		Email:      emailClient,
		opts:       opts,
		Logger:     logger,
		storage:    st,
		activity:   ac,
		queryCache: newQueryCache(opts.QueryCacheSizeBytes),
	}
	rt.securityEngine = newSecurityEngine(opts.SecurityEngineCacheSize, logger, rt)

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

	// NewConfigReloader also trigger reload of configs for all instances.
	rt.configReloader = NewConfigReloader(ctx, rt)

	return rt, nil
}

func (r *Runtime) AllowHostAccess() bool {
	return r.opts.AllowHostAccess
}

func (r *Runtime) Version() version.Version {
	return r.opts.Version
}

func (r *Runtime) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	r.configReloader.close()
	r.registryCache.close(ctx)
	err1 := r.queryCache.close()
	err2 := r.connCache.Close(ctx) // Also closes metastore // TODO: Propagate ctx cancellation
	return errors.Join(err1, err2)
}

func (r *Runtime) ResolveSecurity(ctx context.Context, instanceID string, claims *SecurityClaims, res *runtimev1.Resource) (*ResolvedSecurity, error) {
	inst, err := r.Instance(ctx, instanceID)
	if err != nil {
		return nil, err
	}
	vars := inst.ResolveVariables(false)
	return r.securityEngine.resolveSecurity(ctx, instanceID, inst.Environment, vars, claims, res)
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

func (r *Runtime) UpdateInstanceWithRillYAML(ctx context.Context, instanceID string, p *parser.Parser, restartController bool) error {
	if p.RillYAML == nil {
		return errors.New("rill.yaml is required to update an instance")
	}

	rillYAML := p.RillYAML
	dotEnv := p.GetDotEnv()

	inst, err := r.Instance(ctx, instanceID)
	if err != nil {
		return err
	}

	// Shallow clone for editing
	tmp := *inst
	inst = &tmp

	inst.ProjectDisplayName = rillYAML.DisplayName
	inst.ProjectOLAPConnector = rillYAML.OLAPConnector

	// Dedupe connectors
	connMap := make(map[string]*runtimev1.Connector)
	for _, c := range rillYAML.Connectors {
		config, err := structpb.NewStruct(c.Defaults)
		if err != nil {
			return err
		}
		connMap[c.Name] = &runtimev1.Connector{
			Type:   c.Type,
			Name:   c.Name,
			Config: config,
		}
	}
	for _, r := range p.Resources {
		if r.ConnectorSpec != nil {
			connMap[r.Name.Name] = &runtimev1.Connector{
				Name:                r.Name.Name,
				Type:                r.ConnectorSpec.Driver,
				Config:              r.ConnectorSpec.Properties,
				TemplatedProperties: r.ConnectorSpec.TemplatedProperties,
				Provision:           r.ConnectorSpec.Provision,
				ProvisionArgs:       r.ConnectorSpec.ProvisionArgs,
			}
		}
	}

	conns := make([]*runtimev1.Connector, 0, len(connMap))
	for _, c := range connMap {
		conns = append(conns, c)
	}
	inst.ProjectConnectors = conns

	vars := make(map[string]string)
	for _, v := range rillYAML.Variables {
		vars[v.Name] = v.Default
	}
	for k, v := range dotEnv {
		vars[k] = v
	}
	inst.ProjectVariables = vars
	inst.FeatureFlags = rillYAML.FeatureFlags
	inst.PublicPaths = rillYAML.PublicPaths
	inst.AIInstructions = rillYAML.AIInstructions
	inst.ProjectAIConnector = rillYAML.AIConnector
	return r.EditInstance(ctx, inst, restartController)
}

// UpdateInstanceConnector upserts or removes a connector from an instance
// If connector is nil, the connector is removed; otherwise, it is upserted
func (r *Runtime) UpdateInstanceConnector(ctx context.Context, instanceID, name string, connector *runtimev1.ConnectorSpec) error {
	inst, err := r.Instance(ctx, instanceID)
	if err != nil {
		return err
	}

	// Copy the existing connectors into a new list except the one being updated.
	projConns := make([]*runtimev1.Connector, 0, len(inst.ProjectConnectors))
	for _, c := range inst.ProjectConnectors {
		if c.Name == name {
			continue
		}
		projConns = append(projConns, c)
	}
	// If not removing, append the new/updated connector.
	if connector != nil {
		projConns = append(projConns, &runtimev1.Connector{
			Name:                name,
			Type:                connector.Driver,
			Config:              connector.Properties,
			TemplatedProperties: connector.TemplatedProperties,
			Provision:           connector.Provision,
			ProvisionArgs:       connector.ProvisionArgs,
		})
	}

	// Clone for editing
	tmp := *inst
	inst = &tmp
	inst.ProjectConnectors = projConns

	return r.EditInstance(ctx, inst, false)
}

func (r *Runtime) ReloadConfig(ctx context.Context, instanceID string) error {
	return r.configReloader.reloadConfig(ctx, instanceID)
}

type configReloader struct {
	rt *Runtime
	// cancel background operations on close
	cancel context.CancelFunc
	// singleflight group to deduplicate reloads for same instance
	group singleflight.Group[string, string]

	// to avoid repo handshake refresh on each reload we track last updatedon of each deployment
	// if the deployment has not changed skip the repo pull
	//
	// this can further be optimised to only check for properties that affect repo like url, branch etc but keeping it simple for now
	updatedOn map[string]time.Time
	mu        sync.Mutex
}

func NewConfigReloader(ctx context.Context, rt *Runtime) *configReloader {
	bgctx, bgcancel := context.WithCancel(context.Background())
	c := &configReloader{
		rt:     rt,
		cancel: bgcancel,
	}

	go c.periodicallyReloadConfigs(bgctx)
	return c
}

func (r *configReloader) reloadConfig(ctx context.Context, instanceID string) error {
	_, err := r.group.Do(ctx, instanceID, func(ctx context.Context) (string, error) {
		inst, err := r.rt.Instance(ctx, instanceID)
		if err != nil {
			return "", err
		}

		admin, release, err := r.rt.Admin(ctx, instanceID)
		if err != nil {
			if errors.Is(err, ErrAdminNotConfigured) {
				return instanceID, nil
			}
			return "", err
		}
		defer release()

		r.rt.Logger.Info("Reloading config for instance", zap.String("instance_id", instanceID))

		cfg, err := admin.GetDeploymentConfig(ctx)
		if err != nil {
			return "", err
		}

		// Clone for editing
		tmp := *inst
		inst = &tmp

		// Update variables
		varsChanged := !maps.Equal(inst.Variables, cfg.Variables)
		if varsChanged {
			inst.Variables = cfg.Variables
		}
		inst.Annotations = cfg.Annotations
		inst.FrontendURL = cfg.FrontendURL

		// Force the repo to refresh its handshake if the deployment has changed
		r.mu.Lock()
		updatedOn, ok := r.updatedOn[instanceID]
		r.mu.Unlock()
		if !ok || cfg.UpdatedOn.After(updatedOn) {
			repo, release, err := r.rt.Repo(ctx, inst.ID)
			if err != nil {
				return "", err
			}
			defer release()

			err = repo.Pull(ctx, false, true)
			if err != nil {
				r.rt.Logger.Error("ReloadConfig: failed to pull repo", zap.String("instance_id", inst.ID), zap.Error(err), observability.ZapCtx(ctx))
			}

			// Update the last updatedOn time
			r.mu.Lock()
			r.updatedOn[instanceID] = cfg.UpdatedOn
			r.mu.Unlock()
		}

		err = r.rt.EditInstance(ctx, inst, varsChanged)
		if err != nil {
			return "", err
		}
		return instanceID, nil
	})

	return err
}

func (r *configReloader) periodicallyReloadConfigs(ctx context.Context) {
	reloadAllInstances := func() {
		instances, err := r.rt.Instances(ctx)
		if err != nil {
			r.rt.Logger.Error("periodicallyReloadConfigs: failed to list instances", zap.Error(err))
			return
		}
		for _, inst := range instances {
			err := r.reloadConfig(ctx, inst.ID)
			if err != nil {
				r.rt.Logger.Error("periodicallyReloadConfigs: failed to reload config", zap.String("instance_id", inst.ID), zap.Error(err))
			}
		}
	}
	// first reload immediately
	reloadAllInstances()

	// then periodically every hour
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			reloadAllInstances()
		case <-ctx.Done():
			return
		}
	}
}

func (r *configReloader) close() {
	r.cancel()
}

func instanceAnnotationsToAttribs(instance *drivers.Instance) []attribute.KeyValue {
	attrs := make([]attribute.KeyValue, 0, len(instance.Annotations)+1)
	attrs = append(attrs, attribute.String("instance_id", instance.ID))
	for k, v := range instance.Annotations {
		attrs = append(attrs, attribute.String(k, v))
	}
	return attrs
}
