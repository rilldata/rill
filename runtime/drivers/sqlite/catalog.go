package sqlite

import (
	"context"
	"fmt"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"google.golang.org/protobuf/proto"
)

type catalogStore struct {
	*connection
	instanceID string
}

func (c *catalogStore) FindEntries(ctx context.Context, typ drivers.ObjectType) ([]*drivers.CatalogEntry, error) {
	if typ == drivers.ObjectTypeUnspecified {
		return c.findEntries(ctx, "WHERE instance_id = ?", c.instanceID)
	}
	return c.findEntries(ctx, "WHERE instance_id = ? AND type = ?", c.instanceID, typ)
}

func (c *catalogStore) FindEntry(ctx context.Context, name string) (*drivers.CatalogEntry, error) {
	// Names are stored with case everywhere, but the checks should be case-insensitive. Hence, the translation to lower case here.
	es, err := c.findEntries(ctx, "WHERE instance_id = ? AND LOWER(name) = LOWER(?)", c.instanceID, name)
	if err != nil {
		return nil, err
	}
	if len(es) == 0 {
		return nil, drivers.ErrNotFound
	}
	return es[0], nil
}

func (c *catalogStore) findEntries(_ context.Context, whereClause string, args ...any) ([]*drivers.CatalogEntry, error) {
	// Override ctx because sqlite sometimes segfaults on context cancellation
	ctx := context.Background()

	sql := fmt.Sprintf("SELECT name, type, object, path, bytes_ingested, created_on, updated_on, refreshed_on FROM catalog %s ORDER BY lower(name)", whereClause)

	rows, err := c.db.QueryxContext(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []*drivers.CatalogEntry
	for rows.Next() {
		var objBlob []byte
		e := &drivers.CatalogEntry{}

		err := rows.Scan(&e.Name, &e.Type, &objBlob, &e.Path, &e.BytesIngested, &e.CreatedOn, &e.UpdatedOn, &e.RefreshedOn)
		if err != nil {
			return nil, err
		}

		// Parse object protobuf
		if objBlob != nil {
			switch e.Type {
			case drivers.ObjectTypeTable:
				e.Object = &runtimev1.Table{}
			case drivers.ObjectTypeSource:
				e.Object = &runtimev1.Source{}
			case drivers.ObjectTypeModel:
				e.Object = &runtimev1.Model{}
			case drivers.ObjectTypeMetricsView:
				e.Object = &runtimev1.MetricsView{}
			default:
				panic(fmt.Errorf("unexpected object type: %v", e.Type))
			}

			err = proto.Unmarshal(objBlob, e.Object)
			if err != nil {
				panic(err)
			}
		}

		res = append(res, e)
	}

	return res, nil
}

func (c *catalogStore) CreateEntry(_ context.Context, e *drivers.CatalogEntry) error {
	// Override ctx because sqlite sometimes segfaults on context cancellation
	ctx := context.Background()

	// Serialize object
	obj, err := proto.Marshal(e.Object)
	if err != nil {
		return err
	}

	now := time.Now()
	_, err = c.db.ExecContext(
		ctx,
		"INSERT INTO catalog(instance_id, name, type, object, path, bytes_ingested, created_on, updated_on, refreshed_on) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		c.instanceID,
		e.Name,
		e.Type,
		obj,
		e.Path,
		e.BytesIngested,
		now,
		now,
		now,
	)
	if err != nil {
		return err
	}

	e.CreatedOn = now
	e.UpdatedOn = now
	e.RefreshedOn = now
	return nil
}

func (c *catalogStore) UpdateEntry(_ context.Context, e *drivers.CatalogEntry) error {
	// Override ctx because sqlite sometimes segfaults on context cancellation
	ctx := context.Background()

	// Serialize object
	obj, err := proto.Marshal(e.Object)
	if err != nil {
		return err
	}

	now := time.Now()
	_, err = c.db.ExecContext(
		ctx,
		"UPDATE catalog SET type = ?, object = ?, path = ?, bytes_ingested = ?, updated_on = ?, refreshed_on = ? WHERE instance_id = ? AND name = ?",
		e.Type,
		obj,
		e.Path,
		e.BytesIngested,
		now,
		e.RefreshedOn,
		c.instanceID,
		e.Name,
	)
	if err != nil {
		return err
	}

	e.UpdatedOn = now
	return nil
}

func (c *catalogStore) DeleteEntry(_ context.Context, name string) error {
	// Override ctx because sqlite sometimes segfaults on context cancellation
	ctx := context.Background()

	_, err := c.db.ExecContext(ctx, "DELETE FROM catalog WHERE instance_id = ? AND LOWER(name) = LOWER(?)", c.instanceID, name)
	return err
}

func (c *catalogStore) DeleteEntries(_ context.Context) error {
	// Override ctx because sqlite sometimes segfaults on context cancellation
	ctx := context.Background()

	_, err := c.db.ExecContext(ctx, "DELETE FROM catalog WHERE instance_id = ?", c.instanceID)
	return err
}

func (c *connection) NextControllerVersion(ctx context.Context) (int64, error) {
	_, err := c.db.ExecContext(ctx, "UPDATE rill.controller_version SET version = version + 1")
	if err != nil {
		return 0, err
	}

	var version int64
	err = c.db.QueryRowContext(ctx, "SELECT version FROM rill.controller_version").Scan(&version)
	if err != nil {
		return 0, err
	}

	return version, nil
}

func (c *connection) CheckControllerVersion(ctx context.Context, v int64) error {
	var version int64
	err := c.db.QueryRowContext(ctx, "SELECT version FROM rill.controller_version").Scan(&version)
	if err != nil {
		return err
	}

	if version != v {
		return drivers.ErrInconsistentControllerVersion
	}

	return nil
}

func (c *connection) FindResources(ctx context.Context) ([]drivers.Resource, error) {
	rows, err := c.db.QueryxContext(ctx, "SELECT kind, name, data FROM rill.catalogv2 ORDER BY kind, name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []drivers.Resource
	for rows.Next() {
		r := drivers.Resource{}
		err := rows.Scan(&r.Kind, &r.Name, &r.Data)
		if err != nil {
			return nil, err
		}
		res = append(res, r)
	}

	if rows.Err() != nil {
		return nil, err
	}

	return res, nil
}

func (c *connection) CreateResource(ctx context.Context, v int64, r drivers.Resource) error {
	err := c.CheckControllerVersion(ctx, v)
	if err != nil {
		return err
	}

	var exists bool
	if err := c.db.QueryRowContext(ctx, "SELECT EXISTS (SELECT 1 FROM rill.catalogv2 WHERE kind = ? AND name = ?)", r.Kind, r.Name).Scan(&exists); err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("catalog entry for kind=%q, name=%q already exists", r.Kind, r.Name)
	}

	now := time.Now()
	_, err = c.db.ExecContext(
		ctx,
		"INSERT INTO rill.catalogv2(kind, name, data, created_on, updated_on) VALUES (?, ?, ?, ?, ?)",
		r.Kind,
		r.Name,
		r.Data,
		now,
		now,
	)
	if err != nil {
		return err
	}

	return nil
}

func (c *connection) UpdateResource(ctx context.Context, v int64, r drivers.Resource) error {
	err := c.CheckControllerVersion(ctx, v)
	if err != nil {
		return err
	}

	_, err = c.db.ExecContext(
		ctx,
		"UPDATE rill.catalogv2 SET kind=?, name=?, data=?, updated_on=?) VALUES (?, ?, ?, ?)",
		r.Kind,
		r.Name,
		r.Data,
		time.Now(),
	)
	if err != nil {
		return err
	}

	return nil
}

func (c *connection) DeleteResource(ctx context.Context, v int64, k, n string) error {
	err := c.CheckControllerVersion(ctx, v)
	if err != nil {
		return err
	}

	_, err = c.db.ExecContext(ctx, "DELETE FROM rill.catalogv2 WHERE kind=? AND name=?", k, n)
	if err != nil {
		return err
	}

	return nil
}

func (c *connection) DeleteResources(ctx context.Context) error {
	_, err := c.db.ExecContext(ctx, "DELETE FROM rill.catalogv2")
	if err != nil {
		return err
	}

	return nil
}
