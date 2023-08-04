package sqlite

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
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

	sql := fmt.Sprintf("SELECT id, olap_driver, repo_driver, embed_catalog, created_on, updated_on, variables, project_variables, ingestion_limit_bytes, annotations, connectors FROM instances %s ORDER BY id", whereClause)

	rows, err := c.db.QueryxContext(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []*drivers.Instance
	for rows.Next() {
		// sqlite doesn't support maps need to read as bytes and convert to map
		var variables, projectVariables, annotations, connectors []byte
		i := &drivers.Instance{}
		err := rows.Scan(&i.ID, &i.OLAPDriver, &i.RepoDriver, &i.EmbedCatalog, &i.CreatedOn, &i.UpdatedOn, &variables, &projectVariables, &i.IngestionLimitBytes, &annotations, &connectors)
		if err != nil {
			return nil, err
		}
		i.Variables, err = mapFromJSON(variables)
		if err != nil {
			return nil, err
		}

		i.ProjectVariables, err = mapFromJSON(projectVariables)
		if err != nil {
			return nil, err
		}

		i.Annotations, err = mapFromJSON(annotations)
		if err != nil {
			return nil, err
		}

		i.Connectors, err = unmarshalConnectors(connectors)
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
	variables, err := mapToJSON(inst.Variables)
	if err != nil {
		return err
	}

	projectVariables, err := mapToJSON(inst.ProjectVariables)
	if err != nil {
		return err
	}

	annotations, err := mapToJSON(inst.Annotations)
	if err != nil {
		return err
	}

	connectors, err := json.Marshal(inst.Connectors)
	if err != nil {
		return err
	}

	now := time.Now()
	_, err = c.db.ExecContext(
		ctx,
		"INSERT INTO instances(id, olap_driver, repo_driver, embed_catalog, created_on, updated_on, variables, project_variables, ingestion_limit_bytes, annotations, connectors) "+
			"VALUES ($1, $2, $3, $4, $5, $5, $6, $7, $8, $9, $10)",
		inst.ID,
		inst.OLAPDriver,
		inst.RepoDriver,
		inst.EmbedCatalog,
		now,
		variables,
		projectVariables,
		inst.IngestionLimitBytes,
		annotations,
		connectors,
	)
	if err != nil {
		return err
	}

	// We assign manually instead of using RETURNING because it doesn't work for timestamps in SQLite
	inst.CreatedOn = now
	inst.UpdatedOn = now
	return nil
}

// CreateInstance implements drivers.RegistryStore.
func (c *connection) EditInstance(_ context.Context, inst *drivers.Instance) error {
	// Override ctx because sqlite sometimes segfaults on context cancellation
	ctx := context.Background()

	// sqlite doesn't support maps need to convert to json and write as bytes array
	variables, err := mapToJSON(inst.Variables)
	if err != nil {
		return err
	}

	projVariables, err := mapToJSON(inst.ProjectVariables)
	if err != nil {
		return err
	}

	annotations, err := mapToJSON(inst.Annotations)
	if err != nil {
		return err
	}

	connectors, err := json.Marshal(inst.Connectors)
	if err != nil {
		return err
	}

	now := time.Now()
	_, err = c.db.ExecContext(
		ctx,
		"UPDATE instances SET olap_driver = $2, repo_driver = $3, embed_catalog = $4, variables = $5, project_variables = $6, updated_on = $7, ingestion_limit_bytes = $8, annotations = $9 "+
			"WHERE id = $1",
		inst.ID,
		inst.OLAPDriver,
		inst.RepoDriver,
		inst.EmbedCatalog,
		variables,
		projVariables,
		now,
		inst.IngestionLimitBytes,
		annotations,
		connectors,
	)
	if err != nil {
		return err
	}

	// We assign manually instead of using RETURNING because it doesn't work for timestamps in SQLite
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

func mapToJSON(data map[string]string) ([]byte, error) {
	return json.Marshal(data)
}

func mapFromJSON(data []byte) (map[string]string, error) {
	if len(data) == 0 {
		return map[string]string{}, nil
	}
	var m map[string]string
	err := json.Unmarshal(data, &m)
	return m, err
}

func unmarshalConnectors(s []byte) ([]*runtimev1.ConnectorDef, error) {
	if len(s) == 0 {
		return make([]*runtimev1.ConnectorDef, 0), nil
	}
	var defs []*runtimev1.ConnectorDef
	err := json.Unmarshal([]byte(s), &defs)
	return defs, err
}
