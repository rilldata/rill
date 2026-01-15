package runtime

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"sync"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/parser"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/pkg/logbuffer"
	"github.com/rilldata/rill/runtime/pkg/logutil"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// GlobalProjectParserName is the name of the instance-global project parser resource that is created for each new instance.
var GlobalProjectParserName = &runtimev1.ResourceName{Kind: ResourceKindProjectParser, Name: "parser"}

// instanceHeartbeatInterval is the interval at which instance heartbeat events are emitted.
const instanceHeartbeatInterval = time.Minute

// DefaultInstanceID returns the instance ID for the default instance.
// It returns false on runtimes with none or multiple instances.
func (r *Runtime) DefaultInstanceID() (string, bool) {
	inst, ok := r.registryCache.getDefault()
	if !ok {
		return "", false
	}
	return inst.ID, true
}

// Instances returns all instances managed by the runtime.
func (r *Runtime) Instances(ctx context.Context) ([]*drivers.Instance, error) {
	return r.registryCache.list()
}

// Instance looks up an instance by ID. Instances are cached in-memory, so this is a cheap operation.
func (r *Runtime) Instance(ctx context.Context, instanceID string) (*drivers.Instance, error) {
	return r.registryCache.get(instanceID)
}

// InstanceConfig returns the instance's dynamic configuration.
func (r *Runtime) InstanceConfig(ctx context.Context, instanceID string) (drivers.InstanceConfig, error) {
	inst, err := r.Instance(ctx, instanceID)
	if err != nil {
		return drivers.InstanceConfig{}, err
	}
	return inst.Config()
}

// InstanceLogger returns a logger scoped for the given instance. Logs emitted to the logger will also be available in the instance's log buffer.
func (r *Runtime) InstanceLogger(ctx context.Context, instanceID string) (*zap.Logger, error) {
	return r.registryCache.getLogger(instanceID)
}

// InstanceLogs returns an in-memory buffer of recent logs related to the given instance.
func (r *Runtime) InstanceLogs(ctx context.Context, instanceID string) (*logbuffer.Buffer, error) {
	return r.registryCache.getLogbuffer(instanceID)
}

// Controller returns the controller for the given instance.
// If the controller is currently initializing, the call will wait until the controller is ready.
// If the controller has closed with a fatal error, that error will be returned here until it's restarted.
func (r *Runtime) Controller(ctx context.Context, instanceID string) (*Controller, error) {
	ctx, span := tracer.Start(ctx, "Runtime.Controller", trace.WithAttributes(attribute.String("instance_id", instanceID)))
	defer span.End()
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
func (r *Runtime) DeleteInstance(ctx context.Context, instanceID string) error {
	inst, err := r.registryCache.get(instanceID)
	if err != nil {
		if errors.Is(err, drivers.ErrNotFound) {
			return nil
		}
		return err
	}

	// For idempotency, it's ok for some steps to fail

	// Delete the instance
	completed, err := r.registryCache.delete(ctx, instanceID)
	if err != nil {
		r.Logger.Error("delete instance: error deleting from registry", zap.Error(err), zap.String("instance_id", instanceID), observability.ZapCtx(ctx))
	}

	// Wait for the controller to stop and the connection cache to be evicted
	<-completed

	if err := r.storage.RemovePrefix(ctx, inst.ID); err != nil {
		r.Logger.Error("could not drop instance data directory", zap.Error(err), zap.String("instance_id", instanceID), observability.ZapCtx(ctx))
	}

	// Cleanup catalog data
	catalog, ok := r.metastore.AsCatalogStore(instanceID)
	if ok {
		err = catalog.DeleteResources(ctx)
		if err != nil {
			r.Logger.Error("delete instance: error deleting catalog", zap.Error(err), zap.String("instance_id", instanceID), observability.ZapCtx(ctx))
		}
	}

	return nil
}

// DataDir returns the path to a persistent data directory for the given instance. The directory is created if it doesn't exist.
// Storage usage in the returned directory will be reported in the instance's heartbeat events.
func (r *Runtime) DataDir(instanceID string, elem ...string) (string, error) {
	return r.storage.WithPrefix(instanceID).DataDir(elem...)
}

// TempDir returns the path to a temporary directory for the given instance. The directory is created if it doesn't exist.
// The TempDir is a fixed location. The caller is responsible for cleaning up after use.
// The TempDir may be cleared after restarts.
// Storage usage in the returned directory will be reported in the instance's heartbeat events.
func (r *Runtime) TempDir(instanceID string, elem ...string) (string, error) {
	return r.storage.WithPrefix(instanceID).TempDir(elem...)
}

// registryCache caches all the runtime's instances and manages the life-cycle of their controllers.
// It ensures that a controller is started for every instance, and that a controller is completely stopped before getting restarted when edited.
type registryCache struct {
	logger        *zap.Logger
	activity      *activity.Client
	rt            *Runtime
	store         drivers.RegistryStore
	mu            sync.RWMutex
	instances     map[string]*instanceWithController
	baseCtx       context.Context
	baseCtxCancel context.CancelFunc
}

type instanceWithController struct {
	instanceID    string
	instance      *drivers.Instance
	controller    *Controller
	controllerErr error

	logger    *zap.Logger
	logbuffer *logbuffer.Buffer

	// State for managing controller execution
	ctx       context.Context
	cancel    context.CancelFunc
	running   bool // Invariant: if false, controllerErr must be set
	stoppedCh chan struct{}
	ready     bool
	readyCh   chan struct{}
	reopen    bool
}

func newRegistryCache(rt *Runtime, registry drivers.RegistryStore, logger *zap.Logger, ac *activity.Client) *registryCache {
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

	return r
}

func (r *registryCache) init(ctx context.Context) error {
	// NOTE: Can't be called from newRegistryCache because rt.registry must be initialized before we start controllers

	insts, err := r.store.FindInstances(ctx)
	if err != nil {
		return err
	}

	for _, inst := range insts {
		if err := r.add(inst); err != nil {
			return err
		}
	}

	go r.emitHeartbeats()

	return nil
}

func (r *registryCache) close(ctx context.Context) {
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

func (r *registryCache) getDefault() (*drivers.Instance, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if len(r.instances) == 1 {
		for _, iwc := range r.instances {
			return iwc.instance, true
		}
	}

	return nil, false
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

func (r *registryCache) getLogger(instanceID string) (*zap.Logger, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	iwc := r.instances[instanceID]
	if iwc == nil {
		return nil, drivers.ErrNotFound
	}
	return iwc.logger, nil
}

func (r *registryCache) getLogbuffer(instanceID string) (*logbuffer.Buffer, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	iwc := r.instances[instanceID]
	if iwc == nil {
		return nil, drivers.ErrNotFound
	}
	return iwc.logbuffer, nil
}

func (r *registryCache) create(ctx context.Context, inst *drivers.Instance) error {
	err := r.store.CreateInstance(ctx, inst)
	if err != nil {
		return err
	}

	return r.add(inst)
}

func (r *registryCache) add(inst *drivers.Instance) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.instances[inst.ID]; ok {
		panic(fmt.Errorf("instance %q already open", inst.ID))
	}

	iwc := &instanceWithController{
		instanceID: inst.ID,
		instance:   inst,
	}
	r.instances[inst.ID] = iwc

	// Setup the logger to duplicate logs to a) the Zap logger, b) an in-memory buffer that exposes the logs over the API
	buffer := logbuffer.NewBuffer(r.rt.opts.ControllerLogBufferCapacity, r.rt.opts.ControllerLogBufferSizeBytes)
	bufferCore := logutil.NewBufferedZapCore(buffer)

	baseCore := r.logger.Core() // Only add instance_id to the base core
	baseCore = baseCore.With([]zapcore.Field{zap.String("instance_id", iwc.instanceID)})

	iwc.logger = zap.New(zapcore.NewTee(baseCore, bufferCore))
	iwc.logbuffer = buffer

	r.restartController(iwc)
	return nil
}

func (r *registryCache) edit(ctx context.Context, inst *drivers.Instance, restartController bool) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// call edit instance under lock to ensure that concurrent edits do not end up in different entity in cache and db
	err := r.store.EditInstance(ctx, inst)
	if err != nil {
		return err
	}

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
			// Before starting the controller, pull the repo.
			// Even though the project parser will also do this, it needs to happen before the controller starts.
			// This is necessary for models that reference files in the repo, since they may be triggered before the project parser is triggered.
			iwc.logger.Debug("pulling repo")
			err := r.ensureRepoReady(iwc.ctx, iwc.instanceID)
			if err != nil {
				iwc.logger.Warn("failed to pull repo", zap.Error(err))
				// Even if repo pull failed, we'll start the controller
			} else {
				iwc.logger.Debug("repo pulled")
			}

			// Before starting the controller, update the project config.
			// This avoids the controller immediately cancelling and restarting if the project config has changed.
			if err := r.updateProjectConfig(iwc); err != nil {
				iwc.logger.Warn("failed to parse and update the project config before starting the controller", zap.Error(err))
			}

			// Build activity client.
			ac := r.activity
			inst, err := r.get(iwc.instanceID) // Need to use get since we don't currently hold the lock.
			if err == nil {                    // Defensive handling to avoid a race condition if the instance was deleted.
				ac = ac.With(instanceAnnotationsToAttribs(inst)...)
			}

			// Start controller
			iwc.logger.Debug("controller starting")
			ctrl, err := NewController(iwc.ctx, r.rt, iwc.instanceID, iwc.logger, ac)
			if err == nil {
				r.ensureProjectParser(iwc.ctx, iwc.instanceID, ctrl)

				r.mu.Lock()
				iwc.controller = ctrl
				iwc.ready = true
				close(iwc.readyCh)
				r.mu.Unlock()

				iwc.logger.Debug("controller ready")

				err = ctrl.Run(iwc.ctx)
			}

			iwc.cancel() // Always ensure cleanup

			r.mu.Lock()
			attrs := []zapcore.Field{zap.Error(err), zap.Bool("reopen", iwc.reopen), zap.Bool("called_run", iwc.ready)}
			r.mu.Unlock()

			if errors.Is(err, iwc.ctx.Err()) {
				iwc.logger.Debug("controller stopped", attrs...)
			} else {
				iwc.logger.Error("controller failed", attrs...)
			}

			// When an instance is edited, connector config may have changed.
			// So we want to evict all open connections for that instance, but it's unsafe to do so while the controller is running.
			// So this is the only place where we can do it safely.
			if r.baseCtx.Err() == nil {
				r.rt.evictInstanceConnections(iwc.instanceID)
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

func (r *registryCache) ensureRepoReady(ctx context.Context, instanceID string) error {
	repo, release, err := r.rt.Repo(ctx, instanceID)
	if err != nil {
		return err
	}
	defer release()

	// Pull the latest changes
	// on rill developer do not pull latest changes - all pulls should be user triggered
	return repo.Pull(ctx, &drivers.PullOptions{UserTriggered: !r.rt.AllowHostAccess()})
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

func (r *registryCache) emitHeartbeats() {
	ticker := time.NewTicker(instanceHeartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			instances, err := r.list()
			if err != nil {
				r.logger.Error("failed to send instance heartbeat event, instance listing failed", zap.Error(err))
				continue
			}
			for _, instance := range instances {
				r.emitHeartbeatForInstance(instance)
			}
		case <-r.baseCtx.Done():
			return
		}
	}
}

func (r *registryCache) emitHeartbeatForInstance(inst *drivers.Instance) {
	dataDir, err := r.rt.storage.WithPrefix(inst.ID).DataDir()
	if err != nil {
		r.logger.Error("failed to send instance heartbeat event, could not get data directory", zap.String("instance_id", inst.ID), zap.Error(err))
		return
	}

	// Add instance annotations as attributes to pass organization id, project id, etc.
	attrs := instanceAnnotationsToAttribs(inst)
	r.activity.RecordMetric(context.Background(), "data_dir_size_bytes", float64(sizeOfDir(dataDir)), attrs...)
}

// updateProjectConfig updates the project config for the given instance.
// This does the same operation as ProjectParserReconciler's reconcileProjectConfig and is done before starting the controller
// to ensure that when controller first starts, it doesn’t immediately restart due to changed variables
func (r *registryCache) updateProjectConfig(iwc *instanceWithController) error {
	repo, release, err := r.rt.Repo(iwc.ctx, iwc.instanceID)
	if err != nil {
		return err
	}
	defer release()

	instance, err := r.get(iwc.instanceID)
	if err != nil {
		return err
	}

	p, err := parser.Parse(iwc.ctx, repo, iwc.instanceID, instance.Environment, instance.OLAPConnector)
	if err != nil {
		return err
	}

	if p.RillYAML == nil {
		// Empty project
		return nil
	}

	return r.rt.UpdateInstanceWithRillYAML(iwc.ctx, iwc.instanceID, p, false)
}

func sizeOfDir(path string) int64 {
	var size int64
	_ = fs.WalkDir(os.DirFS(path), ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		f, err := d.Info()
		if err != nil {
			return err
		}
		size += f.Size()
		return nil
	})
	return size
}
