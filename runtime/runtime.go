package runtime

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"

	"github.com/rilldata/rill/runtime/drivers"
)

// Runtime is a data infra proxy and orchestrator.
// It's a multi-tenant service that can manage many projects. Each project is represented by an Instance.
// It supports scale-out when no local infra drivers are registered (i.e. not DuckDB).
type Runtime struct {
	metadataDB drivers.Connection
	logger     *zap.Logger
	instances  map[string]*Instance // Note: temporary hack for local POC
}

// New creates a Runtime
func New(metadataDB drivers.Connection, logger *zap.Logger) *Runtime {
	return &Runtime{
		metadataDB: metadataDB,
		logger:     logger,
		instances:  make(map[string]*Instance),
	}
}

// CreateInstance creates a new instance
func (r *Runtime) CreateInstance(driver string, dsn string) (*Instance, error) {
	conn, err := drivers.Open(driver, dsn)
	if err != nil {
		return nil, err
	}

	olap, ok := conn.OLAP()
	if !ok {
		return nil, fmt.Errorf("cannot use driver %s for OLAP", driver)
	}

	inst := &Instance{
		ID:   uuid.New(),
		Conn: conn,
		OLAP: olap,
	}

	r.instances[inst.ID.String()] = inst

	return inst, err
}

// InstanceFromID prepares (but doesn't load) and existing instance.
// Call Load on the instance to connect.
func (r *Runtime) InstanceFromID(id uuid.UUID) *Instance {
	inst, ok := r.instances[id.String()]
	if !ok {
		return nil
	}
	return inst
}

// Instance represents a single Rill project (call it session, release, kernel, environment, ...)
type Instance struct {
	ID   uuid.UUID
	Conn drivers.Connection
	OLAP drivers.OLAP
}

// Load looks for the instance and connects to its infra
func (i *Instance) Load() error {
	if i == nil {
		return fmt.Errorf("instance not found")
	}
	return nil
}

func (r *Instance) Query(ctx context.Context, stmt *drivers.Statement) (*sqlx.Rows, error) {
	return r.OLAP.Execute(ctx, stmt)
}
