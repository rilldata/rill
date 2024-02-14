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

	sql := fmt.Sprintf(`
		SELECT
			id,
			environment,
			olap_connector,
			repo_connector,
			admin_connector,
			catalog_connector,
			created_on,
			updated_on,
			connectors,
			project_connectors,
			variables,
			project_variables,
			annotations,
			embed_catalog,
			watch_repo,
			stage_changes,
			model_default_materialize,
			model_materialize_delay_seconds
		FROM instances %s ORDER BY id
	`, whereClause)

	rows, err := c.db.QueryxContext(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []*drivers.Instance
	for rows.Next() {
		// sqlite doesn't support maps need to read as bytes and convert to map
		var variables, projectVariables, annotations, connectors, projectConnectors []byte
		i := &drivers.Instance{}
		err := rows.Scan(
			&i.ID,
			&i.Environment,
			&i.OLAPConnector,
			&i.RepoConnector,
			&i.AdminConnector,
			&i.CatalogConnector,
			&i.CreatedOn,
			&i.UpdatedOn,
			&connectors,
			&projectConnectors,
			&variables,
			&projectVariables,
			&annotations,
			&i.EmbedCatalog,
			&i.WatchRepo,
			&i.StageChanges,
			&i.ModelDefaultMaterialize,
			&i.ModelMaterializeDelaySeconds,
		)
		if err != nil {
			return nil, err
		}

		i.Connectors, err = unmarshalConnectors(connectors)
		if err != nil {
			return nil, err
		}

		i.ProjectConnectors, err = unmarshalConnectors(projectConnectors)
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
	connectors, err := json.Marshal(inst.Connectors)
	if err != nil {
		return err
	}

	projectConnectors, err := json.Marshal(inst.ProjectConnectors)
	if err != nil {
		return err
	}

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

	now := time.Now()
	_, err = c.db.ExecContext(
		ctx,
		`
		INSERT INTO instances(
			id,
			environment,
			olap_connector,
			repo_connector,
			admin_connector,
			catalog_connector,
			created_on,
			updated_on,
			connectors,
			project_connectors,
			variables,
			project_variables,
			annotations,
			embed_catalog,
			watch_repo,
			stage_changes,
			model_default_materialize,
			model_materialize_delay_seconds
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
		`,
		inst.ID,
		inst.Environment,
		inst.OLAPConnector,
		inst.RepoConnector,
		inst.AdminConnector,
		inst.CatalogConnector,
		now,
		now,
		connectors,
		projectConnectors,
		variables,
		projectVariables,
		annotations,
		inst.EmbedCatalog,
		inst.WatchRepo,
		inst.StageChanges,
		inst.ModelDefaultMaterialize,
		inst.ModelMaterializeDelaySeconds,
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

	projectConnectors, err := json.Marshal(inst.ProjectConnectors)
	if err != nil {
		return err
	}

	now := time.Now()
	_, err = c.db.ExecContext(
		ctx,
		`
		UPDATE instances SET
			environment = $2,
			olap_connector = $3,
			repo_connector = $4,
			admin_connector = $5,
			catalog_connector = $6,
			updated_on = $7,
			connectors = $8,
			project_connectors = $9,
			variables = $10,
			project_variables = $11,
			annotations = $12,
			embed_catalog = $13,
			watch_repo = $14,
			stage_changes = $15,
			model_default_materialize = $16,
			model_materialize_delay_seconds = $17
		WHERE id = $1
		`,
		inst.ID,
		inst.Environment,
		inst.OLAPConnector,
		inst.RepoConnector,
		inst.AdminConnector,
		inst.CatalogConnector,
		now,
		connectors,
		projectConnectors,
		variables,
		projectVariables,
		annotations,
		inst.EmbedCatalog,
		inst.WatchRepo,
		inst.StageChanges,
		inst.ModelDefaultMaterialize,
		inst.ModelMaterializeDelaySeconds,
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

func unmarshalConnectors(s []byte) ([]*runtimev1.Connector, error) {
	if len(s) == 0 {
		return make([]*runtimev1.Connector, 0), nil
	}
	var defs []*runtimev1.Connector
	err := json.Unmarshal(s, &defs)
	return defs, err
}
