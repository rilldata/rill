package runtime

import (
	"context"
	"errors"
	"fmt"
	"sync"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.uber.org/zap"
)

// Built-in resource kinds
const (
	ResourceKindProjectParser  string = "rill.runtime.v1.ProjectParser"
	ResourceKindSource         string = "rill.runtime.v1.SourceV2"
	ResourceKindModel          string = "rill.runtime.v1.ModelV2"
	ResourceKindMetricsView    string = "rill.runtime.v1.MetricsViewV2"
	ResourceKindMigration      string = "rill.runtime.v1.Migration"
	ResourceKindPullTrigger    string = "rill.runtime.v1.PullTrigger"
	ResourceKindRefreshTrigger string = "rill.runtime.v1.RefreshTrigger"
	ResourceKindBucketPlanner  string = "rill.runtime.v1.BucketPlanner"
)

// GlobalProjectParserName is the name of the instance-global project parser resource that is created for each new instance.
var GlobalProjectParserName = &runtimev1.ResourceName{Kind: ResourceKindProjectParser, Name: "project_parser"}

// Instances returns all instances managed by the runtime.
func (r *Runtime) Instances(ctx context.Context) ([]*drivers.Instance, error) {
	return r.registryCache.list()
}

// Instance looks up an instance by ID. Instances are cached in-memory, so this is a cheap operation.
func (r *Runtime) Instance(ctx context.Context, instanceID string) (*drivers.Instance, error) {
	return r.registryCache.get(instanceID)
}

// Controller returns the controller for the given instance. If the controller stopped with a fatal error, that error will be returned here until it's restarted.
func (r *Runtime) Controller(instanceID string) (*Controller, error) {
	return r.registryCache.getController(instanceID)
}

// WaitUntilIdle waits until the instance's controller is idle (not reconciling any resources).
func (r *Runtime) WaitUntilIdle(ctx context.Context, instanceID string) error {
	ctrl, err := r.Controller(instanceID)
	if err != nil {
		return err
	}
	return ctrl.WaitUntilIdle(ctx)
}

// CreateInstance creates a new instance and starts a controller for it.
func (r *Runtime) CreateInstance(ctx context.Context, inst *drivers.Instance) error {
	return r.registryCache.create(ctx, inst)
}

// EditInstance edits an existing instance.
// If restartController is true, the instance's controller will be re-opened and all cached connections for the instance will be evicted.
// Until the controller and connections have been closed and re-opened, calls related to the instance may return transient errors.
func (r *Runtime) EditInstance(ctx context.Context, inst *drivers.Instance, restartController bool) error {
	return r.registryCache.edit(ctx, inst, restartController)
}

// DeleteInstance deletes an instance and stops its controller.
func (r *Runtime) DeleteInstance(ctx context.Context, instanceID string, dropDB bool) error {
	inst, err := r.registryCache.get(instanceID)
	if err != nil {
		if errors.Is(err, drivers.ErrNotFound) {
			return nil
		}
		return err
	}

	// For idempotency, it's ok for some steps to fail

	// Get OLAP info for dropDB
	olapDriver, olapCfg, err := r.connectorConfig(ctx, instanceID, inst.OLAPConnector)
	if err != nil {
		r.logger.Error("delete instance: error getting config", zap.Error(err), zap.String("instance_id", instanceID), observability.ZapCtx(ctx))
	}

	// Delete the instance
	completed, err := r.registryCache.delete(ctx, instanceID)
	if err != nil {
		r.logger.Error("delete instance: error deleting from registry", zap.Error(err), zap.String("instance_id", instanceID), observability.ZapCtx(ctx))
	}

	// Wait for the controller to stop and the connection cache to be evicted
	<-completed

	// Can now drop the OLAP
	if dropDB {
		err = drivers.Drop(olapDriver, olapCfg, r.logger)
		if err != nil {
			r.logger.Error("could not drop database", zap.Error(err), zap.String("instance_id", instanceID), observability.ZapCtx(ctx))
		}
	}

	// If catalog is not embedded, catalog data is in the metastore, and should be cleaned up
	if !inst.EmbedCatalog {
		catalog, ok := r.metastore.AsCatalogStore(instanceID)
		if ok {
			err = catalog.DeleteResources(ctx)
			if err != nil {
				r.logger.Error("delete instance: error deleting catalog", zap.Error(err), zap.String("instance_id", instanceID), observability.ZapCtx(ctx))
			}
		}
	}

	return nil
}

// registryCache caches all the runtime's instances and manages the life-cycle of their controllers.
// It ensures that a controller is started for every instance, and that a controller is completely stopped before getting restarted when edited.
type registryCache struct {
	logger        *zap.Logger
	activity      activity.Client
	rt            *Runtime
	store         drivers.RegistryStore
	mu            sync.RWMutex
	instances     map[string]*instanceWithController
	baseCtx       context.Context
	baseCtxCancel context.CancelFunc
}

type instanceWithController struct {
	instance      *drivers.Instance
	controller    *Controller
	controllerErr error

	// State for managing controller execution
	ctx    context.Context
	cancel context.CancelFunc
	reopen bool
	closed chan struct{}
}

func newRegistryCache(ctx context.Context, rt *Runtime, registry drivers.RegistryStore, logger *zap.Logger, ac activity.Client) (*registryCache, error) {
	baseCtx, baseCtxCancel := context.WithCancel(context.Background())

	r := &registryCache{
		logger:        logger,
		activity:      ac,
		rt:            rt,
		store:         registry,
		instances:     make(map[string]*instanceWithController),
		baseCtx:       baseCtx,
		baseCtxCancel: baseCtxCancel,
	}

	insts, err := r.store.FindInstances(ctx)
	if err != nil {
		return nil, err
	}

	for _, inst := range insts {
		r.add(inst)
	}

	for _, inst := range r.instances {
		if inst.controller == nil {
			continue
		}
		err := inst.controller.WaitUntilReady(ctx)
		if err != nil {
			return nil, err
		}
		r.ensureProjectParser(ctx, inst.instance.ID)
	}

	return r, nil
}

func (r *registryCache) close(ctx context.Context) error {
	wg := sync.WaitGroup{}

	r.mu.Lock()
	for _, inst := range r.instances {
		inst := inst
		wg.Add(1)
		go func() {
			select {
			case <-inst.closed:
			case <-ctx.Done():
			}
			wg.Done()
		}()
	}
	r.mu.Unlock()

	r.baseCtxCancel()
	wg.Wait()
	return nil
}

func (r *registryCache) list() ([]*drivers.Instance, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	res := make([]*drivers.Instance, 0, len(r.instances))
	for _, iwc := range r.instances {
		res = append(res, iwc.instance)
	}

	return res, nil
}

func (r *registryCache) get(instanceID string) (*drivers.Instance, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	iwc, ok := r.instances[instanceID]
	if !ok {
		return nil, drivers.ErrNotFound
	}

	return iwc.instance, nil
}

func (r *registryCache) getController(instanceID string) (*Controller, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	iwc, ok := r.instances[instanceID]
	if !ok {
		return nil, drivers.ErrNotFound
	}

	if iwc.controllerErr != nil {
		return nil, iwc.controllerErr
	}

	return iwc.controller, nil
}

func (r *registryCache) create(ctx context.Context, inst *drivers.Instance) error {
	err := r.store.CreateInstance(ctx, inst)
	if err != nil {
		return err
	}

	r.add(inst)
	r.ensureProjectParser(ctx, inst.ID)

	return nil
}

func (r *registryCache) add(inst *drivers.Instance) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.instances[inst.ID]; ok {
		panic(fmt.Errorf("instance %q already open", inst.ID))
	}

	iwc := &instanceWithController{instance: inst}
	r.instances[inst.ID] = iwc
	r.restartController(iwc)
}

func (r *registryCache) edit(ctx context.Context, inst *drivers.Instance, restartController bool) error {
	err := r.store.EditInstance(ctx, inst)
	if err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	iwc, ok := r.instances[inst.ID]
	if !ok {
		panic(fmt.Errorf("instance %q not found", inst.ID))
	}

	iwc.instance = inst
	if restartController {
		r.restartController(iwc)
	}

	return nil
}

func (r *registryCache) delete(ctx context.Context, instanceID string) (chan struct{}, error) {
	err := r.store.DeleteInstance(ctx, instanceID)
	if err != nil {
		return nil, err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	iwc, ok := r.instances[instanceID]
	if !ok {
		panic(fmt.Errorf("instance %q not found", instanceID))
	}
	delete(r.instances, instanceID)

	iwc.cancel()

	return iwc.closed, nil
}

func (r *registryCache) ensureProjectParser(ctx context.Context, instanceID string) {
	r.mu.RLock()
	iwc := r.instances[instanceID]
	r.mu.RUnlock()
	if iwc.controller == nil {
		return
	}
	err := iwc.controller.WaitUntilReady(ctx)
	if err != nil {
		return
	}

	_, err = iwc.controller.Get(ctx, GlobalProjectParserName, false)
	if err == nil {
		return
	}
	if !errors.Is(err, drivers.ErrResourceNotFound) {
		r.logger.Error("could not get project parser", zap.Error(err), zap.String("instance_id", instanceID), observability.ZapCtx(ctx))
		return
	}

	err = iwc.controller.Create(ctx, GlobalProjectParserName, nil, nil, nil, &runtimev1.Resource{
		Resource: &runtimev1.Resource_ProjectParser{
			ProjectParser: &runtimev1.ProjectParser{Spec: &runtimev1.ProjectParserSpec{}},
		},
	})
	if err != nil {
		r.logger.Error("could not create project parser", zap.Error(err), zap.String("instance_id", instanceID), observability.ZapCtx(ctx))
	}
}

func (r *registryCache) restartController(iwc *instanceWithController) {
	// If controller isn't nil, it's already running. Have the already running goroutine stop and restart it.
	// (Easiest way to ensure only one controller is running at a time, without blocking the caller until the currently running one is done.)
	if iwc.controller != nil {
		iwc.reopen = true
		iwc.cancel()
		return
	}

	// Reset execution state
	// NOTE: Duplicating this here and in the goroutine to ensure there's no moment where the mutex is unlocked and neither of controller or controllerErr is set.
	iwc.controller, iwc.controllerErr = NewController(r.rt, iwc.instance.ID, r.logger, r.activity)
	iwc.ctx, iwc.cancel = context.WithCancel(r.baseCtx)
	iwc.reopen = false
	iwc.closed = make(chan struct{})
	if iwc.controllerErr != nil {
		// If NewController errored, no need to start a goroutine
		iwc.cancel()
		close(iwc.closed)
		return
	}

	// Start goroutine that runs the controller
	go func() {
		// Loop in case reopen gets set
		for {
			err := iwc.controller.Run(iwc.ctx)
			iwc.cancel() // Always ensure cleanup

			// When an instance is edited, connector config may have changed.
			// So we want to evict all open connections for that instance, but it's unsafe to do so while the controller is running.
			// So this is the only place where we can do it safely.
			if r.baseCtx.Err() == nil {
				r.rt.connCache.evictAll(r.baseCtx, iwc.instance.ID)
			}

			// If not reopening, exit
			if !iwc.reopen {
				r.mu.Lock()
				iwc.controller = nil
				iwc.controllerErr = err
				close(iwc.closed)
				r.mu.Unlock()
				return
			}

			// Reopening – reset execution state and proceed to next loop iteration
			// NOTE: Not resetting iwc.closed (keeping it open)
			r.mu.Lock()
			iwc.controller, iwc.controllerErr = NewController(r.rt, iwc.instance.ID, r.logger, r.activity)
			iwc.ctx, iwc.cancel = context.WithCancel(r.baseCtx)
			iwc.reopen = false
			if iwc.controllerErr != nil {
				// If NewController errored, no need to start a goroutine
				iwc.cancel()
				close(iwc.closed)
				return
			}
			r.mu.Unlock()
		}
	}()
}
