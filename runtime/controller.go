package runtime

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"go.uber.org/zap"
	"go.uber.org/zap/exp/zapslog"
	"golang.org/x/exp/slog"
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

// ControllerOptions provides additional configuration of a controller
type ControllerOptions struct{}

// Controller manages the catalog for a single instance and runs reconcilers to migrate the catalog (and related resources in external databases) into the desired state.
type Controller struct {
	Runtime     *Runtime
	InstanceID  string
	Logger      *slog.Logger
	opts        *ControllerOptions
	reconcilers map[string]Reconciler
	catalog     catalogCache
	running     atomic.Bool
}

// NewController creates a new Controller
func NewController(ctx context.Context, rt *Runtime, instanceID string, logger *zap.Logger, opts *ControllerOptions) (*Controller, error) {
	c := &Controller{
		Runtime:     rt,
		InstanceID:  instanceID,
		opts:        opts,
		reconcilers: make(map[string]Reconciler),
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

type taskEntry struct {
	cancel context.CancelFunc
}

// Run starts and runs the controller's event loop. It returns when ctx is cancelled or an unrecoverable error occurs.
func (c *Controller) Run(ctx context.Context) error {
	if c.running.Swap(true) {
		panic("controller is already running")
	}

	// PRINCIPLE: No transitive dependencies can run at the same time
	// PRINCIPLE: Deps that would create cycles are not added. On each deletion or ref update, all such candidates are re-assessed.

	// Make DAG
	// Enqueue all roots for reconciliation
	// In catalog, enqueue updated items for reconciliation

	triggers := make(chan *runtimev1.ResourceName, 100)
	completed := make(chan *runtimev1.ResourceName, 100)

	// var schedule []runtimev1.ResourceName + time
	timer := time.NewTimer(time.Hour)
	timer.Stop()

	for {
		select {
		// Reconcile should be started
		case n := <-triggers:
			// If marked running, send a restart signal to the same task (should never call finished).

			// If marked pending, ignore.

			// Check if an ancestor is running. If so, mark pending and skip.

			// If children running, mark pending and schedule for when children finish.

			// If deleted and children running, cancel all children, mark pending and schedule for when children finish.

			// Mark all descendants as pending.

		case n := <-completed:
			// Reconcile finished.
			// If the item was deleted, remove it permanently.
			// For each child in DAG, check that parents are not pending. If so, enqueue for reconciliation.
			//
		case <-timer.C:
			// Time has arrived to pop something from the schedule, and put it on trigger queue

		case <-ctx.Done():
			// Cancel all running tasks and exit

		}
	}

	// Make DAG of initial refs

	// Start loop

	// Keep map of tasks (consists of name, ctx with name, cancel func)

	// Keep channel of names to trigger

	// Keep

	// TODO: Load all resources and build DAG of refs

	// TODO: Run event loop: build schedule, linearize and trigger reconcilers on changes

	// Semantics: Run one reconcile per resource name at a time. On new trigger, cancel the old one then invoke new call.
	// - Only needs catalog locking for non-self resources
	// - If A is running
	// 	- And edits itself: no retrigger
	// 	- And is edited by other: retrigger
	// 	- And ref is updated: retrigger
	// - Must handle panics in reconcilers

	// Catalog Notes:
	// - Cache resources for catalog calls in-memory – make workload write-heavy.
	// - Allow configurable disk reads/flushes. Enables lightweight/ephemeral state updates, like progress info.
	// - Reads/writes to DB should transactionally check versions of BOTH the controller and resource. Return ErrInconsistentVersion on failure. (Indicates split-brain.)

	// TODO: Add handling for rapid deletion and creation?

	return c.close(context.Background())
}

func (c *Controller) close(ctx context.Context) error {
	var errs []error
	for _, r := range c.reconcilers {
		errs = append(errs, r.Close(ctx))
	}

	c.catalog.flush()

	return errors.Join(errs...)
}

func (c *Controller) Lock() {
	// Dont reconcile changes until unlocked
	panic("not implemented")
}

func (c *Controller) Unlock() {
	// Unlock
	panic("not implemented")
}

func (c *Controller) Get(ctx context.Context, name *runtimev1.ResourceName) (*runtimev1.Resource, error) {
	// Don't return ones that are deleted
	panic("not implemented")
}

func (c *Controller) List(ctx context.Context) ([]*runtimev1.Resource, error) {
	// Don't return ones that are deleted
	panic("not implemented")
}

func (c *Controller) Create(ctx context.Context, name *runtimev1.ResourceName, refs []*runtimev1.ResourceName, owner *runtimev1.ResourceName, paths []string, r *runtimev1.Resource) error {
	panic("not implemented")
}

func (c *Controller) UpdateMeta(ctx context.Context, name *runtimev1.ResourceName, refs []*runtimev1.ResourceName, owner *runtimev1.ResourceName, paths []string, newName *runtimev1.ResourceName) error {
	panic("not implemented")
}

func (c *Controller) UpdateSpec(ctx context.Context, name *runtimev1.ResourceName, refs []*runtimev1.ResourceName, owner *runtimev1.ResourceName, paths []string, r *runtimev1.Resource) error {
	panic("not implemented")
}

func (c *Controller) UpdateState(ctx context.Context, name *runtimev1.ResourceName, r *runtimev1.Resource) error {
	// Need to accept a cancelled ctx if it's tagged for the resource
	panic("not implemented")
}

func (c *Controller) UpdateError(ctx context.Context, name *runtimev1.ResourceName, err error) error {
	panic("not implemented")
}

func (c *Controller) Delete(ctx context.Context, names ...*runtimev1.ResourceName) error {
	panic("not implemented")
}

func (c *Controller) Flush(ctx context.Context) error {
	panic("not implemented")
}

func (c *Controller) Retrigger(ctx context.Context, name *runtimev1.ResourceName, t time.Time) error {
	panic("not implemented")
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

// reconciler gets or lazily initializes a reconciler
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
