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
			project_olap_connector,
			repo_connector,
			admin_connector,
			ai_connector,
			catalog_connector,
			created_on,
			updated_on,
			connectors,
			project_connectors,
			variables,
			project_variables,
			feature_flags,
			annotations,
			watch_repo,
			public_paths
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
		var variables, projectVariables, featureFlags, annotations, connectors, projectConnectors, publicPaths []byte
		i := &drivers.Instance{}
		err := rows.Scan(
			&i.ID,
			&i.Environment,
			&i.OLAPConnector,
			&i.ProjectOLAPConnector,
			&i.RepoConnector,
			&i.AdminConnector,
			&i.AIConnector,
			&i.CatalogConnector,
			&i.CreatedOn,
			&i.UpdatedOn,
			&connectors,
			&projectConnectors,
			&variables,
			&projectVariables,
			&featureFlags,
			&annotations,
			&i.WatchRepo,
			&publicPaths,
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

		i.Variables, err = mapFromJSON[string](variables)
		if err != nil {
			return nil, err
		}

		i.ProjectVariables, err = mapFromJSON[string](projectVariables)
		if err != nil {
			return nil, err
		}

		i.FeatureFlags, err = mapFromJSON[bool](featureFlags)
		if err != nil {
			return nil, err
		}

		i.Annotations, err = mapFromJSON[string](annotations)
		if err != nil {
			return nil, err
		}

		i.PublicPaths, err = arrayFromJSON[string](publicPaths)
		if err != nil {
			return nil, err
		}

		res = append(res, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
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

	featureFlags, err := mapToJSON(inst.FeatureFlags)
	if err != nil {
		return err
	}

	annotations, err := mapToJSON(inst.Annotations)
	if err != nil {
		return err
	}

	publicPaths, err := arrayToJSON(inst.PublicPaths)
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
			project_olap_connector,
			repo_connector,
			admin_connector,
			ai_connector,
			catalog_connector,
			created_on,
			updated_on,
			connectors,
			project_connectors,
			variables,
			project_variables,
			feature_flags,
			annotations,
			watch_repo,
			public_paths
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
		`,
		inst.ID,
		inst.Environment,
		inst.OLAPConnector,
		inst.ProjectOLAPConnector,
		inst.RepoConnector,
		inst.AdminConnector,
		inst.AIConnector,
		inst.CatalogConnector,
		now,
		now,
		connectors,
		projectConnectors,
		variables,
		projectVariables,
		featureFlags,
		annotations,
		inst.WatchRepo,
		publicPaths,
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

	featureFlags, err := mapToJSON(inst.FeatureFlags)
	if err != nil {
		return err
	}

	annotations, err := mapToJSON(inst.Annotations)
	if err != nil {
		return err
	}

	publicPaths, err := arrayToJSON(inst.PublicPaths)
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
			project_olap_connector = $4,
			repo_connector = $5,
			admin_connector = $6,
			ai_connector = $7,
			catalog_connector = $8,
			updated_on = $9,
			connectors = $10,
			project_connectors = $11,
			variables = $12,
			project_variables = $13,
			feature_flags = $14,
			annotations = $15,
			watch_repo = $16,
			public_paths = $17
		WHERE id = $1
		`,
		inst.ID,
		inst.Environment,
		inst.OLAPConnector,
		inst.ProjectOLAPConnector,
		inst.RepoConnector,
		inst.AdminConnector,
		inst.AIConnector,
		inst.CatalogConnector,
		now,
		connectors,
		projectConnectors,
		variables,
		projectVariables,
		featureFlags,
		annotations,
		inst.WatchRepo,
		publicPaths,
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

func mapToJSON[T any](data map[string]T) ([]byte, error) {
	return json.Marshal(data)
}

func mapFromJSON[T any](data []byte) (map[string]T, error) {
	if len(data) == 0 {
		return map[string]T{}, nil
	}
	var m map[string]T
	err := json.Unmarshal(data, &m)
	return m, err
}

func arrayToJSON[T any](data []T) ([]byte, error) {
	return json.Marshal(data)
}

func arrayFromJSON[T any](data []byte) ([]T, error) {
	if len(data) == 0 {
		return []T{}, nil
	}
	var a []T
	err := json.Unmarshal(data, &a)
	return a, err
}

func unmarshalConnectors(s []byte) ([]*runtimev1.Connector, error) {
	if len(s) == 0 {
		return make([]*runtimev1.Connector, 0), nil
	}
	var defs []*runtimev1.Connector
	err := json.Unmarshal(s, &defs)
	return defs, err
}
