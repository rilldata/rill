package runtime

import (
	"context"
	"errors"
	"fmt"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"go.uber.org/zap"
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

// ErrInconsistentControllerVersion is returned from Controller when an unexpected controller version is observed in the DB.
// An unexpected controller version will only be observed if multiple controllers are running simultanesouly (split brain).
var ErrInconsistentControllerVersion = errors.New("controller: inconsistent version")

// ErrInconsistentResourceVersion is returned from catalog update functions when an unexpected resource version is observed in the DB.
// An unexpected resource version will only be observed if multiple controllers are running simultanesouly (split brain).
var ErrInconsistentResourceVersion = errors.New("controller: inconsistent version")

// ErrResourceNotFound is returned from catalog functions when a referenced resource does not exist.
var ErrResourceNotFound = errors.New("controller: resource not found")

// Reconciler implements reconciliation logic for all resources of a specific kind.
// Reconcilers are managed and invoked by a Controller.
type Reconciler interface {
	Reconcile(ctx context.Context, s *Signal) ReconcileResult
	Close(ctx context.Context) error
}

// ReconcileResult propagates results from a reconciler invocation
type ReconcileResult struct {
	Err       error
	Retrigger time.Time
}

// SignalCode enumerates signals that can trigger a reconciler
type SignalCode int

const (
	SignalCodeUnspecified SignalCode = iota
	SignalCodeTrigger
	SignalCodeRestart
	SignalCodeCreated
	SignalCodeUpdated
	SignalCodeRefAdded
	SignalCodeRefUpdated
	SignalCodeDescendentsIdle
)

// Signal provides a reconciler with context about why it was triggered
type Signal struct {
	// Name is the resource that should be reconciled
	Name *runtimev1.ResourceName
	// Codes describe the signals that caused the reconciler to be triggered.
	// Since triggers are linearized, multiple signals may accummulate before the reconciler is triggered.
	Codes []SignalCode
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
type ControllerOptions struct {
	EmbedCatalogInConnector         string
	PersistDeletedResourcesDuration time.Duration
}

// Controller manages the catalog for a single instance and runs reconcilers to migrate the catalog (and related resources in external databases) into the desired state.
type Controller struct {
	Runtime     *Runtime
	InstanceID  string
	Logger      *slog.Logger
	opts        *ControllerOptions
	reconcilers map[string]Reconciler
	// TODO: Add version for tracking controller version
	// TODO: Add catalog cache
	// TODO: Add in-memory log buffer
}

// NewController creates a new Controller
func NewController(ctx context.Context, rt *Runtime, instanceID string, logger *zap.Logger, opts *ControllerOptions) *Controller {
	c := &Controller{
		Runtime:     rt,
		InstanceID:  instanceID,
		opts:        opts,
		reconcilers: make(map[string]Reconciler),
	}

	// TODO: Create c.Logger that duplicates logs to a) the Zap logger, b) an in-memory buffer for exposing logs over the API

	return c
}

// Run starts and runs the controller's event loop. It returns when ctx is cancelled or an unrecoverable error occurs.
func (c *Controller) Run(ctx context.Context) error {
	// TODO: Increment controller version in DB and store in c.Version

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

	// TODO: Close all reconcilers
	// defer func() {
	// 	var err error
	// 	for _, reconciler := range c.reconcilers {
	// 		err = errors.Join(err, reconciler.Close(ctx))
	// 	}
	// }()

	panic("not implemented")
}

func (c *Controller) Lock() {
	panic("not implemented")
}

func (c *Controller) Unlock() {
	panic("not implemented")
}

func (c *Controller) Get(ctx context.Context, name *runtimev1.ResourceName) (*runtimev1.Resource, error) {
	panic("not implemented")
}

func (c *Controller) List(ctx context.Context) ([]*runtimev1.Resource, error) {
	panic("not implemented")
}

func (c *Controller) Create(ctx context.Context, name *runtimev1.ResourceName, refs []*runtimev1.ResourceName, owner *runtimev1.ResourceName, paths []string, r *runtimev1.Resource) error {
	panic("not implemented")
}

type UpdateMetaOptions struct {
	NewName *runtimev1.ResourceName
	Refs    []*runtimev1.ResourceName
	Owner   *runtimev1.ResourceName
	Paths   []string
}

func (c *Controller) UpdateMeta(ctx context.Context, name *runtimev1.ResourceName, opts *UpdateMetaOptions) error {
	panic("not implemented")
}

func (c *Controller) UpdateSpec(ctx context.Context, name *runtimev1.ResourceName, refs []*runtimev1.ResourceName, owner *runtimev1.ResourceName, paths []string, r *runtimev1.Resource) error {
	panic("not implemented")
}

func (c *Controller) UpdateState(ctx context.Context, name *runtimev1.ResourceName, r *runtimev1.Resource) error {
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

func (c *Controller) AcquireConn(ctx context.Context, connector string) (drivers.Connection, func(), error) {
	panic("not implemented")
}

func (c *Controller) AcquireOLAP(ctx context.Context, connector string) (drivers.OLAPStore, func(), error) {
	conn, release, err := c.AcquireConn(ctx, connector)
	if err != nil {
		return nil, nil, err
	}

	olap, ok := conn.AsOLAP()
	if !ok {
		release()
		return nil, nil, fmt.Errorf("connector %q is not an OLAP", connector)
	}

	return olap, release, nil
}

// reconciler gets or lazily initializes a reconciler
// func (c *Controller) reconciler(resourceKind string) Reconciler {
// 	reconciler := c.reconcilers[resourceKind]
// 	if reconciler != nil {
// 		return reconciler
// 	}

// 	initializer := ReconcilerInitializers[resourceKind]
// 	if initializer == nil {
// 		panic(fmt.Errorf("no reconciler registered for resource kind %q", resourceKind))
// 	}

// 	reconciler = initializer(c)
// 	c.reconcilers[resourceKind] = reconciler

// 	return reconciler
// }
