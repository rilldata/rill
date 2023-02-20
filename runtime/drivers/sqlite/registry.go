package sqlite

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rilldata/rill/runtime/drivers"
)

// FindInstances implements drivers.RegistryStore.
func (c *connection) FindInstances(ctx context.Context) ([]*drivers.Instance, error) {
	return c.findInstances(ctx, "")
}

// FindInstance implements drivers.RegistryStore.
func (c *connection) FindInstance(ctx context.Context, id string) (*drivers.Instance, error) {
	is, err := c.findInstances(ctx, "WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	if len(is) == 0 {
		return nil, drivers.ErrNotFound
	}
	return is[0], nil
}

func (c *connection) findInstances(_ context.Context, whereClause string, args ...any) ([]*drivers.Instance, error) {
	// Override ctx because sqlite sometimes segfaults on context cancellation
	ctx := context.Background()

	sql := fmt.Sprintf("SELECT id, olap_driver, olap_dsn, repo_driver, repo_dsn, embed_catalog, created_on, updated_on, env, project_env FROM instances %s ORDER BY id", whereClause)

	rows, err := c.db.QueryxContext(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// sqlite doesn't support maps need to read as bytes and convert to map
	var env, projectEnv []byte
	var res []*drivers.Instance
	for rows.Next() {
		i := &drivers.Instance{}
		err := rows.Scan(&i.ID, &i.OLAPDriver, &i.OLAPDSN, &i.RepoDriver, &i.RepoDSN, &i.EmbedCatalog, &i.CreatedOn, &i.UpdatedOn, &env, &projectEnv)
		if err != nil {
			return nil, err
		}
		i.Env, err = fromJSON(env)
		if err != nil {
			return nil, err
		}

		i.ProjectEnv, err = fromJSON(projectEnv)
		if err != nil {
			return nil, err
		}

		res = append(res, i)
	}

	return res, nil
}

// CreateInstance implements drivers.RegistryStore.
func (c *connection) CreateInstance(_ context.Context, inst *drivers.Instance) error {
	// Override ctx because sqlite sometimes segfaults on context cancellation
	ctx := context.Background()

	if inst.ID == "" {
		inst.ID = uuid.NewString()
	}

	// sqlite doesn't support maps need to convert to json and write as bytes array
	env, err := toJSON(inst.Env)
	if err != nil {
		return err
	}

	projectEnv, err := toJSON(inst.ProjectEnv)
	if err != nil {
		return err
	}

	now := time.Now()
	_, err = c.db.ExecContext(
		ctx,
		"INSERT INTO instances(id, olap_driver, olap_dsn, repo_driver, repo_dsn, embed_catalog, created_on, updated_on, env, project_env) "+
			"VALUES ($1, $2, $3, $4, $5, $6, $7, $7, $8, $9)",
		inst.ID,
		inst.OLAPDriver,
		inst.OLAPDSN,
		inst.RepoDriver,
		inst.RepoDSN,
		inst.EmbedCatalog,
		now,
		env,
		projectEnv,
	)
	if err != nil {
		return err
	}

	// We assign manually instead of using RETURNING because it doesn't work for timestamps in SQLite
	inst.CreatedOn = now
	inst.UpdatedOn = now
	return nil
}

// DeleteInstance implements drivers.RegistryStore.
func (c *connection) DeleteInstance(_ context.Context, id string) error {
	// Override ctx because sqlite sometimes segfaults on context cancellation
	ctx := context.Background()

	_, err := c.db.ExecContext(ctx, "DELETE FROM instances WHERE id=$1", id)
	return err
}

func toJSON(data map[string]string) ([]byte, error) {
	return json.Marshal(data)
}

func fromJSON(data []byte) (map[string]string, error) {
	m := make(map[string]string)
	err := json.Unmarshal(data, &m)
	return m, err
}
