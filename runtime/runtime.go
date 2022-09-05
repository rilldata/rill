package runtime

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"

	"github.com/rilldata/rill/runtime/infra"
	"github.com/rilldata/rill/runtime/metadata"
)

// Runtime is a data infra proxy and orchestrator.
// It's a multi-tenant service that can manage many projects. Each project is represented by an Instance.
// It supports scale-out when no local infra drivers are registered (i.e. not DuckDB).
type Runtime struct {
	db        metadata.DB
	logger    zerolog.Logger
	instances map[string]*Instance // Note: temporary hack for local POC
}

// New creates a Runtime
func New(db metadata.DB, logger zerolog.Logger) *Runtime {
	return &Runtime{
		db:        db,
		logger:    logger,
		instances: make(map[string]*Instance),
	}
}

// CreateInstance creates a new instance
func (r *Runtime) CreateInstance(driver string, dsn string) (*Instance, error) {
	d, ok := infra.Drivers[driver]
	if !ok {
		return nil, fmt.Errorf("unknown driver '%s'", driver)
	}

	conn, err := d.Open(dsn)
	if err != nil {
		return nil, err
	}

	inst := &Instance{
		ID:   uuid.New(),
		Conn: conn,
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
	Conn infra.Connection
}

// Load looks for the instance and connects to its infra
func (i *Instance) Load() error {
	if i == nil {
		return fmt.Errorf("instance not found")
	}
	return nil
}

func (r *Instance) Query(ctx context.Context, stmt *infra.Statement) (*sqlx.Rows, error) {
	return r.Conn.Execute(ctx, stmt)
}
