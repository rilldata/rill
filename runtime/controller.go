package runtime

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/schedule"
	"go.uber.org/zap"
	"go.uber.org/zap/exp/zapslog"
	"golang.org/x/exp/slog"
	"google.golang.org/protobuf/proto"
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

// errCyclicDependency is set as the error on resources that can't be reconciled due to a cyclic dependency
var errCyclicDependency = errors.New("cannot be reconciled due to cyclic dependency")

// Reconciler implements reconciliation logic for all resources of a specific kind.
// Reconcilers are managed and invoked by a Controller.
type Reconciler interface {
	Close(ctx context.Context) error
	AssignSpec(from, to *runtimev1.Resource) error
	AssignState(from, to *runtimev1.Resource) error
	Reconcile(ctx context.Context, n *runtimev1.ResourceName) ReconcileResult
}

// ReconcileResult propagates results from a reconciler invocation
type ReconcileResult struct {
	Err       error
	Retrigger time.Time
}

// ReconcilerInitializer is a function that initializes a new reconciler for a specific controller
type ReconcilerInitializer func(*Controller) Reconciler

// ReconcilerInitializers is a registry of reconciler initializers for different resource kinds.
// There can be only one reconciler per resource kind.
var ReconcilerInitializers = make(map[string]ReconcilerInitializer)

// RegisterReconciler registers a reconciler initializer for a specific resource kind
func RegisterReconcilerInitializer(resourceKind string, initializer ReconcilerInitializer) {
	if ReconcilerInitializers[resourceKind] != nil {
		panic(fmt.Errorf("reconciler already registered for resource kind %q", resourceKind))
	}
	ReconcilerInitializers[resourceKind] = initializer
}

// Controller manages the catalog for a single instance and runs reconcilers to migrate the catalog (and related resources in external databases) into the desired state.
type Controller struct {
	Runtime     *Runtime
	InstanceID  string
	Logger      *slog.Logger
	mu          sync.RWMutex
	running     atomic.Bool
	reconcilers map[string]Reconciler
	catalog     catalogCache
	// queue contains names waiting to be scheduled.
	// It's not a real queue because we schedule the whole queue on each call to processQueue.
	// The only reason we need the queue is to enable batching of changes under locks (reducing need for reconciler cancellations).
	queue           map[string]*runtimev1.ResourceName
	queueNotEmpty   bool
	queueNotEmptyCh chan struct{}
	// timeline tracks resources to be scheduled in the future.
	timeline *schedule.Schedule[string, *runtimev1.ResourceName]
	// invocations tracks currently running reconciler invocations.
	invocations map[string]*invocation
	// completed receives invocations that have finished running.
	completed chan *invocation
}

// NewController creates a new Controller
func NewController(ctx context.Context, rt *Runtime, instanceID string, logger *zap.Logger) (*Controller, error) {
	c := &Controller{
		Runtime:         rt,
		InstanceID:      instanceID,
		reconcilers:     make(map[string]Reconciler),
		queue:           make(map[string]*runtimev1.ResourceName),
		queueNotEmptyCh: make(chan struct{}),
		timeline:        schedule.New[string, *runtimev1.ResourceName](nameStr),
		invocations:     make(map[string]*invocation),
		completed:       make(chan *invocation),
	}

	// TODO: Setup the logger to duplicate logs to a) the Zap logger, b) an in-memory buffer that exposes the logs over the API
	logger = logger.With(zap.String("instance_id", instanceID))
	c.Logger = slog.New(zapslog.NewHandler(logger.Core()))

	cc, err := newCatalogCache(ctx, c, instanceID)
	if err != nil {
		return nil, err
	}
	c.catalog = cc

	return c, nil
}

// Run starts and runs the controller's event loop. It returns when ctx is cancelled or an unrecoverable error occurs.
func (c *Controller) Run(ctx context.Context) error {
	if c.running.Swap(true) {
		panic("controller is already running")
	}

	// Check we are still the leader
	err := c.catalog.checkLeader(ctx)
	if err != nil {
		return err
	}

	// Initially enqueue all resources
	c.mu.Lock()
	for _, rs := range c.catalog.resources {
		for _, r := range rs {
			c.enqueue(r.Meta.Name)
		}
	}
	c.mu.Unlock()

	// Ticker for periodically flushing catalog changes
	flushTicker := time.NewTicker(10 * time.Second)
	defer flushTicker.Stop()

	// Timer for scheduling resources added to c.timeline.
	// Call resetTimelineTimer whenever the timeline may have been changed (must hold mu).
	timelineTimer := time.NewTimer(time.Second)
	defer timelineTimer.Stop()

	timelineTimer.Stop() // We want it stopped initially
	nextTime := time.Time{}
	resetTimelineTimer := func() {
		_, t := c.timeline.Peek()
		if t == nextTime {
			return
		}
		nextTime = t

		timelineTimer.Stop()
		if t.IsZero() {
			return
		}

		d := time.Until(t)
		if d <= 0 {
			// must be positive
			d = time.Nanosecond
		}

		timelineTimer.Reset(d)
	}

	// Run event loop
	var stop bool
	var loopErr error
	var isFlushErr bool
	for !stop {
		select {
		case <-c.queueNotEmptyCh: // There are resources we should schedule
			c.mu.Lock()
			c.processQueue()
			resetTimelineTimer()
			c.mu.Unlock()
		case inv := <-c.completed: // A reconciler invocation has completed
			c.mu.Lock()
			err = c.processCompletedInvocation(inv)
			if err != nil {
				loopErr = err
				stop = true
			}
			c.mu.Unlock()
		case <-timelineTimer.C: // A previous reconciler invocation asked to be re-scheduled now
			c.mu.Lock()
			for c.timeline.Len() > 0 {
				n, t := c.timeline.Peek()
				if t.After(time.Now()) {
					break
				}
				c.timeline.Pop()
				c.enqueue(n)
			}
			resetTimelineTimer()
			c.mu.Unlock()
		case <-flushTicker.C: // It's time to flush the catalog to persistent storage
			c.mu.RLock()
			err = c.catalog.flush(ctx)
			c.mu.RUnlock()
			if err != nil {
				loopErr = err
				isFlushErr = true
				stop = true
			}
		case <-ctx.Done(): // We've been asked to stop
			stop = true
			break
		}
	}

	// Cleanup time
	var closeErr error
	if loopErr != nil {
		closeErr = fmt.Errorf("controller event loop failed: %w", loopErr)
	}

	// Cancel all running invocations
	c.mu.Lock()
	for _, inv := range c.invocations {
		inv.cancel(false)
	}
	c.mu.Unlock()

	// Allow 10 seconds for closing invocations and reconcilers
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Wait for all invocations to finish
	for _, inv := range c.invocations {
		select {
		case <-inv.done:
			continue
		case <-ctx.Done():
			closeErr = fmt.Errorf("timed out waiting for reconcile to finish for resource %q", nameStr(inv.name))
		}
	}

	// Close all reconcilers
	c.mu.Lock()
	for k, r := range c.reconcilers {
		err := r.Close(ctx)
		if err != nil {
			err = fmt.Errorf("failed to close reconciler for %q: %w", k, err)
			closeErr = errors.Join(closeErr, err)
		}
	}
	c.mu.Unlock()

	// Allow 10 seconds for flushing the catalog
	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Flush catalog cache
	if !isFlushErr {
		c.mu.Lock()
		err = c.catalog.flush(ctx)
		if err != nil {
			err = fmt.Errorf("failed to flush catalog: %w", err)
			closeErr = errors.Join(closeErr, err)
		}
		c.mu.Unlock()
	}

	c.running.Store(false)
	return closeErr
}

// Get returns a resource by name.
// Soft-deleted resources (i.e. resources where DeletedOn != nil) are not returned.
func (c *Controller) Get(ctx context.Context, name *runtimev1.ResourceName) (*runtimev1.Resource, error) {
	c.checkRunning()
	c.lock(ctx, true)
	defer c.unlock(ctx, true)
	return c.catalog.get(name, false)
}

// List returns a list of resources of the specified kind.
// If kind is empty, all resources are returned.
// Soft-deleted resources (i.e. resources where DeletedOn != nil) are not returned.
func (c *Controller) List(ctx context.Context, kind string) ([]*runtimev1.Resource, error) {
	c.checkRunning()
	c.lock(ctx, true)
	defer c.unlock(ctx, true)
	return c.catalog.list(kind, false)
}

// Create creates a resource and enqueues it for reconciliation.
// If a resource with the same name is currently being deleted, the deletion will be cancelled.
func (c *Controller) Create(ctx context.Context, name *runtimev1.ResourceName, refs []*runtimev1.ResourceName, owner *runtimev1.ResourceName, paths []string, r *runtimev1.Resource) error {
	c.checkRunning()
	c.lock(ctx, false)
	defer c.unlock(ctx, false)

	// A deleted resource with the same name may exist and be running. If so, we first cancel it.
	requeued := false
	if inv, ok := c.invocations[nameStr(name)]; ok {
		r, err := c.catalog.get(name, true)
		if err != nil {
			return fmt.Errorf("internal: got catalog error for reconciling resource: %w", err)
		}
		if r.Meta.DeletedOn == nil {
			// If a non-deleted resource exists with the same name, we should return an error instead of cancelling.
			return drivers.ErrResourceAlreadyExists
		}
		inv.cancel(true)
		requeued = true
	}

	err := c.catalog.create(name, refs, owner, paths, r)
	if err != nil {
		return err
	}

	if !requeued {
		c.enqueue(name)
	}
	return nil
}

// UpdateMeta updates a resource's meta fields and enqueues it for reconciliation.
// If called from outside the resource's reconciler and the resource is currently reconciling, the current reconciler will be cancelled first.
func (c *Controller) UpdateMeta(ctx context.Context, name *runtimev1.ResourceName, refs []*runtimev1.ResourceName, owner *runtimev1.ResourceName, paths []string, newName *runtimev1.ResourceName) error {
	c.checkRunning()
	c.lock(ctx, false)
	defer c.unlock(ctx, false)

	if !c.isReconcilerForResource(ctx, name) {
		c.cancelInvocation(name, false)
	}

	if newName != nil {
		err := c.catalog.rename(name, newName)
		if err != nil {
			return err
		}
	}

	err := c.catalog.updateMeta(name, refs, owner, paths)
	if err != nil {
		return err
	}

	c.enqueue(name) // TODO: Not if called by self?

	// We updated refs, so it may have broken previous cyclic references
	ns := c.catalog.retryCyclicRefs()
	for _, n := range ns {
		c.enqueue(n)
	}

	return nil
}

// UpdateSpec updates a resource's spec and enqueues it for reconciliation.
// If called from outside the resource's reconciler and the resource is currently reconciling, the current reconciler will be cancelled first.
func (c *Controller) UpdateSpec(ctx context.Context, name *runtimev1.ResourceName, r *runtimev1.Resource) error {
	c.checkRunning()
	c.lock(ctx, false)
	defer c.unlock(ctx, false)

	if !c.isReconcilerForResource(ctx, name) {
		c.cancelInvocation(name, false)
	}

	err := c.catalog.updateSpec(name, r)
	if err != nil {
		return err
	}

	c.enqueue(name) // TODO: Not if called by self?
	return nil
}

// UpdateState updates a resource's state.
// It can only be called from within the resource's reconciler.
// NOTE: Calls to UpdateState succeed even if ctx is cancelled. This enables cancelled reconcilers to update state before finishing.
func (c *Controller) UpdateState(ctx context.Context, name *runtimev1.ResourceName, r *runtimev1.Resource) error {
	c.checkRunning()
	c.lock(ctx, false)
	defer c.unlock(ctx, false)

	if !c.isReconcilerForResource(ctx, name) {
		return fmt.Errorf("can't update resource state from outside of reconciler")
	}

	return c.catalog.updateState(name, r)
}

// UpdateError updates a resource's error.
// Unlike UpdateMeta and UpdateSpec, it does not cancel or enqueue reconciliation for the resource.
func (c *Controller) UpdateError(ctx context.Context, name *runtimev1.ResourceName, err error) error {
	c.checkRunning()
	c.lock(ctx, false)
	defer c.unlock(ctx, false)

	return c.catalog.updateError(name, err)
}

// Delete soft-deletes a resource and enqueues it for reconciliation (with DeletedOn != nil).
// Once the deleting reconciliation has been completed, the resource will be hard deleted.
// If Delete is called from the resource's own reconciler, the resource will be hard deleted immediately (and the calling reconcile's ctx will be canceled immediately).
func (c *Controller) Delete(ctx context.Context, name *runtimev1.ResourceName) error {
	c.checkRunning()
	c.lock(ctx, false)
	defer c.unlock(ctx, false)

	c.cancelInvocation(name, false)

	if c.isReconcilerForResource(ctx, name) {
		return c.catalog.delete(name)
	}

	err := c.catalog.updateDeleted(name)
	if err != nil {
		return err
	}

	c.enqueue(name)
	return nil
}

// Flush forces a flush of the controller's catalog changes to persistent storage.
func (c *Controller) Flush(ctx context.Context) error {
	c.checkRunning()
	c.lock(ctx, false)
	defer c.unlock(ctx, false)

	return c.catalog.flush(ctx)
}

// Reconcile enqueues a resource for reconciliation.
// If the resource is currently reconciling, the current reconciler will be cancelled first.
func (c *Controller) Reconcile(ctx context.Context, name *runtimev1.ResourceName) error {
	c.checkRunning()
	c.lock(ctx, false)
	defer c.unlock(ctx, false)
	c.cancelInvocation(name, false)
	c.enqueue(name)
	return nil
}

// Cancel cancels the current invocation of a resource's reconciler (if it's running).
// It does not re-enqueue the resource for reconciliation.
func (c *Controller) Cancel(ctx context.Context, name *runtimev1.ResourceName) error {
	c.checkRunning()
	c.lock(ctx, false)
	defer c.unlock(ctx, false)
	c.cancelInvocation(name, false)
	return nil
}

// AcquireOLAP gets a handle for a connector in the controller's instance.
func (c *Controller) AcquireConn(ctx context.Context, connector string) (drivers.Handle, func(), error) {
	return c.Runtime.AcquireHandle(ctx, c.InstanceID, connector)
}

// AcquireOLAP gets an OLAP handle for a connector in the controller's instance.
func (c *Controller) AcquireOLAP(ctx context.Context, connector string) (drivers.OLAPStore, func(), error) {
	conn, release, err := c.AcquireConn(ctx, connector)
	if err != nil {
		return nil, nil, err
	}

	olap, ok := conn.AsOLAP(c.InstanceID)
	if !ok {
		release()
		return nil, nil, fmt.Errorf("connector %q is not an OLAP", connector)
	}

	return olap, release, nil
}

// Lock locks the controller's catalog and delays scheduling of new reconciliations until the lock is released.
// It can only be called from within a reconciler invocation.
// While the lock is held, resources can only be edited by a caller using the ctx passed to Lock.
func (c *Controller) Lock(ctx context.Context) {
	inv := invocationFromContext(ctx)
	if inv == nil {
		panic("Lock called outside of a reconciler invocation")
	}
	if inv.holdsLock {
		panic("Lock called by invocation that already holds the lock")
	}
	inv.holdsLock = true
	c.mu.Lock()
}

// Unlock releases the lock acquired by Lock.
func (c *Controller) Unlock(ctx context.Context) {
	inv := invocationFromContext(ctx)
	if inv == nil {
		panic("Unlock called outside of a reconciler invocation")
	}
	if !inv.holdsLock {
		panic("Unlock called by invocation that does not hold the lock")
	}
	inv.holdsLock = false
	c.mu.Unlock()
}

// reconciler gets or lazily initializes a reconciler.
// reconciler is not thread-safe and must be called while c.mu is held.
func (c *Controller) reconciler(resourceKind string) Reconciler {
	reconciler := c.reconcilers[resourceKind]
	if reconciler != nil {
		return reconciler
	}

	initializer := ReconcilerInitializers[resourceKind]
	if initializer == nil {
		panic(fmt.Errorf("no reconciler registered for resource kind %q", resourceKind))
	}

	reconciler = initializer(c)
	c.reconcilers[resourceKind] = reconciler

	return reconciler
}

// checkRunning panics if called when the Controller is not running.
func (c *Controller) checkRunning() {
	if !c.running.Load() {
		panic("controller is not running")
	}
}

// lock locks the controller's mutex, unless ctx belongs to a reconciler invocation that already holds the lock (by having called c.Lock).
func (c *Controller) lock(ctx context.Context, read bool) {
	inv := invocationFromContext(ctx)
	if inv != nil && inv.holdsLock {
		return
	}
	if read {
		c.mu.RLock()
	} else {
		c.mu.Lock()
	}
}

// unlock unlocks the controller's mutex, unless ctx belongs to a reconciler invocation that holds the lock (by having called c.Lock).
func (c *Controller) unlock(ctx context.Context, read bool) {
	inv := invocationFromContext(ctx)
	if inv != nil && inv.holdsLock {
		return
	}
	if read {
		c.mu.RUnlock()
	} else {
		c.mu.Unlock()
	}
}

// isReconcilerForResource returns true if ctx belongs to a reconciler invocation for the given resource.
func (c *Controller) isReconcilerForResource(ctx context.Context, n *runtimev1.ResourceName) bool {
	inv := invocationFromContext(ctx)
	if inv == nil {
		return false
	}
	return proto.Equal(inv.name, n)
}

// enqueue adds a resource to c.queue and ensures processQueue() will be called at some point.
// It must be called while c.mu is held.
func (c *Controller) enqueue(name *runtimev1.ResourceName) {
	// Add to queue
	c.queue[nameStr(name)] = name

	// Make sure the event loop knows there's stuff in the queue.
	if !c.queueNotEmpty {
		c.queueNotEmpty = true
		c.queueNotEmptyCh <- struct{}{}
	}
}

// processQueue calls schedule() for each resource in c.queue.
// It must be called while c.mu is held.
func (c *Controller) processQueue() {
	// Mark all pending. We do it here instead of in schedule() so that the pending parent checks in schedule() reflect all items in c.queue.
	for _, n := range c.queue {
		err := c.catalog.updateStatus(n, runtimev1.ReconcileStatus_RECONCILE_STATUS_PENDING, time.Time{})
		if err != nil {
			c.Logger.Warn("internal: unexpected update error", slog.Any("error", err))
			// TODO: Return? Weird state.
		}
	}

	// Schedule all pending
	for _, n := range c.queue {
		c.schedule(n)
	}

	// Reset queueNotEmpty
	c.queueNotEmpty = false
}

// schedule a resource for reconciliation.
// It contains the main scheduling logic described in the comment for Controller.
// (Basically: deletions and renames run before regular reconciles, and only one invocation runs at a time between a root and a leaf in the DAG.)
//
// schedule does not necessarily invoke reconciliation immediately (though it may), but will ensure that the resource is eventually reconciled.
// If schedule doesn't invoke reconciliation immediately, it will add the resource to another running invocation's waitlist, ensuring the resource will eventually be reconciled.
//
// After calling schedule, the resource can be removed from c.queue.
// It must be called while c.mu is held.
func (c *Controller) schedule(n *runtimev1.ResourceName) {
	// TODO
}

// invoke starts a goroutine to invoke the reconciler and tracks the invocation in c.invocations.
// It must be called while c.mu is held.
func (c *Controller) invoke(n *runtimev1.ResourceName) {
	ctx, cancel := context.WithCancel(context.Background())
	inv := &invocation{
		name:     n,
		done:     make(chan struct{}),
		cancelFn: cancel,
	}
	c.invocations[nameStr(n)] = inv
	ctx = contextWithInvocation(ctx, inv)
	reconciler := c.reconciler(n.Kind) // fetched outside of goroutine to keep access under mutex
	go func() {
		defer func() {
			// Catch panics and set as error
			if r := recover(); r != nil {
				inv.result = ReconcileResult{Err: fmt.Errorf("panic: %v", r)}
				if inv.holdsLock {
					c.Unlock(ctx)
				}
			}
			// Ensure ctx cancel is called (just for cleanup)
			cancel()
			// Send invocation to event loop for post-processing
			c.completed <- inv
		}()
		// Invoke reconciler
		inv.result = reconciler.Reconcile(ctx, n)
	}()
}

// processCompletedInvocation does post-processing after a reconciler invocation completes.
// It must be called while c.mu is held.
// TODO: All the error conditions here are weird. Log or ignore or exit (should we let catalog return inconsistent?)?
func (c *Controller) processCompletedInvocation(inv *invocation) error {
	r, err := c.catalog.get(inv.name, true)
	if err != nil {
		return err
	}

	if r.Meta.DeletedOn != nil && !inv.reschedule { // TODO: Is reschedule a reliable proxy?
		err = c.catalog.delete(r.Meta.Name)
		if err != nil {
			return err
		}
		if inv.result.Err != nil {
			c.Logger.Error("got error while deleting resource", slog.String("name", nameStr(r.Meta.Name)), slog.Any("error", inv.result.Err))
		}
		return nil
	}

	if r.Meta.RenamedFrom != nil && !inv.reschedule { // TODO: Is reschedule a reliable proxy?
		err = c.catalog.clearRenamedFrom(r.Meta.Name)
		if err != nil {
			return err
		}
	}

	// If retrigger requested before now, we'll just reschedule directly
	if !inv.result.Retrigger.After(time.Now()) {
		inv.reschedule = true
		inv.result.Retrigger = time.Time{}
	}

	// Update status and error
	err = c.catalog.updateStatus(inv.name, runtimev1.ReconcileStatus_RECONCILE_STATUS_IDLE, inv.result.Retrigger)
	if err != nil {
		return err
	}
	err = c.catalog.updateError(inv.name, inv.result.Err)
	if err != nil {
		return err
	}

	// Cleanup
	close(inv.done)
	delete(c.invocations, nameStr(inv.name))

	// Let the dominos fall
	if inv.reschedule {
		c.enqueue(inv.name)
	}
	for _, rn := range inv.waitlist {
		c.enqueue(rn)
	}

	return nil
}

// cancelInvocation cancels a running invocation for the resource.
// It does nothing if no invocation is running for the resource.
// It must be called while c.mu is held.
func (c *Controller) cancelInvocation(n *runtimev1.ResourceName, reschedule bool) {
	inv, ok := c.invocations[nameStr(n)]
	if ok {
		inv.cancel(reschedule)
	}
}

// invocation represents a running reconciler invocation for a specific resource.
type invocation struct {
	name       *runtimev1.ResourceName
	done       chan struct{}
	cancelFn   context.CancelFunc
	cancelled  bool
	reschedule bool
	holdsLock  bool
	waitlist   map[string]*runtimev1.ResourceName
	result     ReconcileResult
}

// cancel cancels the invocation.
// It can be called multiple times with different reschedule values, and will be rescheduled if any of the calls ask for it.
// It's not thread-safe (must be called while the controller's mutex is held).
func (i *invocation) cancel(reschedule bool) {
	if !i.cancelled {
		i.cancelled = true
		i.cancelFn()
	}
	i.reschedule = i.reschedule || reschedule
}

// addToWaitlist adds a resource name to the invocation's waitlist.
// Resources on the waitlist will be scheduled after the invocation completes.
// It's not thread safe (must be called while the controller's mutex is held).
func (i *invocation) addToWaitlist(n *runtimev1.ResourceName) {
	if i.waitlist == nil {
		i.waitlist = make(map[string]*runtimev1.ResourceName)
	}
	i.waitlist[nameStr(n)] = n
}

// invocationCtxKey is used for storing an invocation in a context.
type invocationCtxKey struct{}

// contextWithInvocation returns a wrapped context that contains an invocation.
func contextWithInvocation(ctx context.Context, inv *invocation) context.Context {
	return context.WithValue(ctx, invocationCtxKey{}, inv)
}

// invocationFromContext retrieves an invocation from a context.
// If no invocation is in the context, it returns nil.
func invocationFromContext(ctx context.Context) *invocation {
	inv := ctx.Value(invocationCtxKey{})
	if inv != nil {
		return inv.(*invocation)
	}
	return nil
}
