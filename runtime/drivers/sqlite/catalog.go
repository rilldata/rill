package sqlite

import (
	"context"
	"fmt"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"google.golang.org/protobuf/proto"
)

func (c *connection) FindEntries(ctx context.Context, instanceID string, typ drivers.ObjectType) []*drivers.CatalogEntry {
	if typ == drivers.ObjectTypeUnspecified {
		return c.findEntries(ctx, "WHERE instance_id = ?", instanceID)
	}
	return c.findEntries(ctx, "WHERE instance_id = ? AND type = ?", instanceID, typ)
}

func (c *connection) FindEntry(ctx context.Context, instanceID, name string) (*drivers.CatalogEntry, bool) {
	// Names are stored with case everywhere, but the checks should be case-insensitive. Hence, the translation to lower case here.
	es := c.findEntries(ctx, "WHERE instance_id = ? AND LOWER(name) = LOWER(?)", instanceID, name)
	if len(es) == 0 {
		return nil, false
	}
	return es[0], true
}

func (c *connection) findEntries(ctx context.Context, whereClause string, args ...any) []*drivers.CatalogEntry {
	sql := fmt.Sprintf("SELECT name, type, object, path, size_in_bytes, created_on, updated_on, refreshed_on FROM catalog %s ORDER BY lower(name)", whereClause)

	rows, err := c.db.QueryxContext(ctx, sql, args...)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var res []*drivers.CatalogEntry
	for rows.Next() {
		var objBlob []byte
		e := &drivers.CatalogEntry{}

		err := rows.Scan(&e.Name, &e.Type, &objBlob, &e.Path, &e.SizeInBytes, &e.CreatedOn, &e.UpdatedOn, &e.RefreshedOn)
		if err != nil {
			panic(err)
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

	return res
}

func (c *connection) CreateEntry(ctx context.Context, instanceID string, e *drivers.CatalogEntry) error {
	// Serialize object
	obj, err := proto.Marshal(e.Object)
	if err != nil {
		return err
	}

	now := time.Now()
	_, err = c.db.ExecContext(
		ctx,
		"INSERT INTO catalog(instance_id, name, type, object, path, size_in_bytes, created_on, updated_on, refreshed_on) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		instanceID,
		e.Name,
		e.Type,
		obj,
		e.Path,
		e.SizeInBytes,
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

func (c *connection) UpdateEntry(ctx context.Context, instanceID string, e *drivers.CatalogEntry) error {
	// Serialize object
	obj, err := proto.Marshal(e.Object)
	if err != nil {
		return err
	}

	now := time.Now()
	_, err = c.db.ExecContext(
		ctx,
		"UPDATE catalog SET type = ?, object = ?, path = ?, size_in_bytes = ?, updated_on = ?, refreshed_on = ? WHERE instance_id = ? AND name = ?",
		e.Type,
		obj,
		e.Path,
		e.SizeInBytes,
		now,
		e.RefreshedOn,
		instanceID,
		e.Name,
	)
	if err != nil {
		return err
	}

	e.UpdatedOn = now
	return nil
}

func (c *connection) DeleteEntry(ctx context.Context, instanceID, name string) error {
	_, err := c.db.ExecContext(ctx, "DELETE FROM catalog WHERE instance_id = ? AND LOWER(name) = LOWER(?)", instanceID, name)
	return err
}
