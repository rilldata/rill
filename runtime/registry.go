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
	"go.uber.org/zap/zapcore"
)

// GlobalProjectParserName is the name of the instance-global project parser resource that is created for each new instance.
var GlobalProjectParserName = &runtimev1.ResourceName{Kind: ResourceKindProjectParser, Name: "parser"}

// Instances returns all instances managed by the runtime.
func (r *Runtime) Instances(ctx context.Context) ([]*drivers.Instance, error) {
	return r.registryCache.list()
}

// Instance looks up an instance by ID. Instances are cached in-memory, so this is a cheap operation.
func (r *Runtime) Instance(ctx context.Context, instanceID string) (*drivers.Instance, error) {
	return r.registryCache.get(instanceID)
}

// Controller returns the controller for the given instance.
// If the controller is currently initializing, the call will wait until the controller is ready.
// If the controller has closed with a fatal error, that error will be returned here until it's restarted.
func (r *Runtime) Controller(ctx context.Context, instanceID string) (*Controller, error) {
	return r.registryCache.getController(ctx, instanceID)
}

// WaitUntilIdle waits until the instance's controller is idle (not reconciling any resources).
func (r *Runtime) WaitUntilIdle(ctx context.Context, instanceID string, ignoreHidden bool) error {
	ctrl, err := r.Controller(ctx, instanceID)
	if err != nil {
		return err
	}
	return ctrl.WaitUntilIdle(ctx, ignoreHidden)
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
// It ensures that a controller is started for every instance, and that a controller is completely stoppedCh before getting restarted when edited.
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
	ctx       context.Context
	cancel    context.CancelFunc
	running   bool // Invariant: if false, controllerErr must be set
	stoppedCh chan struct{}
	ready     bool
	readyCh   chan struct{}
	reopen    bool
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

	return r, nil
}

func (r *registryCache) init(ctx context.Context) error {
	// NOTE: Can't be called from newRegistryCache because rt.registry must be initialized before we start controllers

	insts, err := r.store.FindInstances(ctx)
	if err != nil {
		return err
	}

	for _, inst := range insts {
		r.add(inst)
	}

	return nil
}

func (r *registryCache) close(ctx context.Context) error {
	wg := sync.WaitGroup{}

	r.mu.Lock()
	for _, inst := range r.instances {
		inst := inst
		wg.Add(1)
		go func() {
			select {
			case <-inst.stoppedCh:
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

func (r *registryCache) getController(ctx context.Context, instanceID string) (*Controller, error) {
	r.mu.RLock()
	iwc, ok := r.instances[instanceID]

	for ok && iwc.running && !iwc.ready {
		readyCh := iwc.readyCh
		r.mu.RUnlock()
		select {
		case <-readyCh:
			// continue
		case <-ctx.Done():
			return nil, ctx.Err()
		}
		r.mu.RLock()
		iwc, ok = r.instances[instanceID]
	}

	defer r.mu.RUnlock()

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
		return fmt.Errorf("instance %q not found", inst.ID)
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
		return nil, fmt.Errorf("instance %q not found", instanceID)
	}
	delete(r.instances, instanceID)

	iwc.cancel()

	return iwc.stoppedCh, nil
}

func (r *registryCache) restartController(iwc *instanceWithController) {
	// If a goroutine for the controller is currently running, have it stop and restart the controller.
	// (Easiest way to ensure only one controller is running at a time, without blocking the caller until the currently running one is done.)
	if iwc.running {
		iwc.reopen = true
		iwc.cancel()
		return
	}

	// Reset execution state
	iwc.ctx, iwc.cancel = context.WithCancel(r.baseCtx)
	iwc.running = true
	iwc.stoppedCh = make(chan struct{})
	iwc.ready = false
	iwc.readyCh = make(chan struct{})
	iwc.reopen = false

	// Start goroutine that runs the controller
	go func() {
		// Loop in case reopen gets set
		for {
			r.logger.Info("controller starting", zap.String("instance_id", iwc.instance.ID))

			ctrl, err := NewController(iwc.ctx, r.rt, iwc.instance.ID, r.logger, r.activity)
			if err == nil {
				r.mu.Lock()
				iwc.controller = ctrl
				iwc.ready = true
				close(iwc.readyCh)
				r.mu.Unlock()

				r.ensureProjectParser(iwc.ctx, iwc.instance.ID, ctrl)

				err = ctrl.Run(iwc.ctx)
			}

			iwc.cancel() // Always ensure cleanup

			r.mu.Lock()
			attrs := []zapcore.Field{zap.String("instance_id", iwc.instance.ID), zap.Error(err), zap.Bool("reopen", iwc.reopen), zap.Bool("called_run", iwc.ready)}
			r.mu.Unlock()

			if r.baseCtx.Err() != nil {
				r.logger.Info("controller stopped", attrs...)
			} else {
				r.logger.Error("controller failed", attrs...)
			}

			// When an instance is edited, connector config may have changed.
			// So we want to evict all open connections for that instance, but it's unsafe to do so while the controller is running.
			// So this is the only place where we can do it safely.
			if r.baseCtx.Err() == nil {
				r.rt.connCache.evictAll(r.baseCtx, iwc.instance.ID)
			}

			r.mu.Lock()

			// If not reopening, exit
			if !iwc.reopen {
				iwc.controller = nil
				iwc.controllerErr = err
				iwc.running = false
				close(iwc.stoppedCh)
				if !iwc.ready { // Ensure readyCh is always closed, even if it never got ready
					close(iwc.readyCh)
				}
				r.mu.Unlock()
				return
			}

			// Reset execution state
			iwc.ctx, iwc.cancel = context.WithCancel(r.baseCtx)
			if iwc.ready {
				iwc.ready = false
				iwc.readyCh = make(chan struct{})
			}
			iwc.reopen = false
			r.mu.Unlock()
		}
	}()
}

func (r *registryCache) ensureProjectParser(ctx context.Context, instanceID string, ctrl *Controller) {
	_, err := ctrl.Get(ctx, GlobalProjectParserName, false)
	if err == nil {
		return
	}
	if !errors.Is(err, drivers.ErrResourceNotFound) {
		r.logger.Error("could not get project parser", zap.Error(err), zap.String("instance_id", instanceID), observability.ZapCtx(ctx))
		return
	}

	err = ctrl.Create(ctx, GlobalProjectParserName, nil, nil, nil, true, &runtimev1.Resource{
		Resource: &runtimev1.Resource_ProjectParser{
			ProjectParser: &runtimev1.ProjectParser{Spec: &runtimev1.ProjectParserSpec{}},
		},
	})
	if err != nil {
		r.logger.Error("could not create project parser", zap.Error(err), zap.String("instance_id", instanceID), observability.ZapCtx(ctx))
	}
}
