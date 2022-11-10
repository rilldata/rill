package sqlite

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/rilldata/rill/runtime/drivers"
)

// FindInstances implements drivers.RegistryStore
func (c *connection) FindInstances(ctx context.Context) []*drivers.Instance {
	var res []*drivers.Instance
	err := c.db.Select(&res, "SELECT * FROM instances ORDER BY id")
	if err != nil {
		panic(err)
	}
	return res
}

// FindInstance implements drivers.RegistryStore
func (c *connection) FindInstance(ctx context.Context, id string) (*drivers.Instance, bool) {
	res := &drivers.Instance{}
	err := c.db.QueryRowxContext(ctx, "SELECT * FROM instances WHERE id = $1", id).StructScan(res)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, false
		}
		panic(err)
	}
	return res, true
}

// CreateInstance implements drivers.RegistryStore
func (c *connection) CreateInstance(ctx context.Context, inst *drivers.Instance) error {
	if inst.ID == "" {
		inst.ID = uuid.NewString()
	}

	now := time.Now()
	_, err := c.db.ExecContext(
		ctx,
		"INSERT INTO instances(id, driver, dsn, object_prefix, exposed, embed_catalog, created_on, updated_on) "+
			"VALUES ($1, $2, $3, $4, $5, $6, $7, $7)",
		inst.ID,
		inst.Driver,
		inst.DSN,
		inst.ObjectPrefix,
		inst.Exposed,
		inst.EmbedCatalog,
		now,
	)
	if err != nil {
		return err
	}
	// We assign manually instead of using RETURNING because it doesn't work for timestamps in SQLite
	inst.CreatedOn = now
	inst.UpdatedOn = now
	return nil
}

// DeleteInstance implements drivers.RegistryStore
func (c *connection) DeleteInstance(ctx context.Context, id string) error {
	_, err := c.db.ExecContext(ctx, "DELETE FROM instances WHERE id=$1", id)
	return err
}

// FindRepos implements drivers.RegistryStore
func (c *connection) FindRepos(ctx context.Context) []*drivers.Repo {
	var res []*drivers.Repo
	err := c.db.Select(&res, "SELECT * FROM repos ORDER BY id")
	if err != nil {
		panic(err)
	}
	return res
}

// FindRepo implements drivers.RegistryStore
func (c *connection) FindRepo(ctx context.Context, id string) (*drivers.Repo, bool) {
	res := &drivers.Repo{}
	err := c.db.QueryRowxContext(ctx, "SELECT * FROM repos WHERE id = $1", id).StructScan(res)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, false
		}
		panic(err)
	}
	return res, true
}

// CreateRepo implements drivers.RegistryStore
func (c *connection) CreateRepo(ctx context.Context, repo *drivers.Repo) error {
	id := uuid.NewString()
	now := time.Now()
	_, err := c.db.ExecContext(
		ctx,
		"INSERT INTO repos(id, driver, dsn, created_on, updated_on) VALUES ($1, $2, $3, $4, $4)",
		id,
		repo.Driver,
		repo.DSN,
		now,
	)
	if err != nil {
		return err
	}
	// We assign manually instead of using RETURNING because it doesn't work for timestamps in SQLite
	repo.ID = id
	repo.CreatedOn = now
	repo.UpdatedOn = now
	return nil
}

// DeleteRepo implements drivers.RegistryStore
func (c *connection) DeleteRepo(ctx context.Context, id string) error {
	_, err := c.db.ExecContext(ctx, "DELETE FROM repos WHERE id=$1", id)
	return err
}
