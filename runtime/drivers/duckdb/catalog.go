package duckdb

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
		return c.findEntries(ctx, "")
	}
	return c.findEntries(ctx, "WHERE type = ?", typ)
}

func (c *connection) FindEntry(ctx context.Context, instanceID, name string) (*drivers.CatalogEntry, bool) {
	// Names are stored with case everywhere, but the checks should be case-insensitive.
	// Hence, the translation to lower case here.
	es := c.findEntries(ctx, "WHERE LOWER(name) = LOWER(?)", name)
	if len(es) == 0 {
		return nil, false
	}
	return es[0], true
}

func (c *connection) findEntries(ctx context.Context, whereClause string, args ...any) []*drivers.CatalogEntry {
	conn, release, err := c.getConn(ctx)
	if err != nil {
		panic(err)
	}
	defer release()

	sql := fmt.Sprintf("SELECT name, type, object, path, created_on, updated_on, refreshed_on FROM rill.catalog %s ORDER BY lower(name)", whereClause)
	rows, err := conn.QueryxContext(ctx, sql, args...)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var res []*drivers.CatalogEntry
	for rows.Next() {
		var objBlob []byte
		e := &drivers.CatalogEntry{}

		err := rows.Scan(&e.Name, &e.Type, &objBlob, &e.Path, &e.CreatedOn, &e.UpdatedOn, &e.RefreshedOn)
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
	conn, release, err := c.getConn(ctx)
	if err != nil {
		return err
	}
	defer release()

	// Serialize object
	obj, err := proto.Marshal(e.Object)
	if err != nil {
		return err
	}

	now := time.Now()
	_, err = conn.ExecContext(
		ctx,
		"INSERT INTO rill.catalog(name, type, object, path, created_on, updated_on, refreshed_on) VALUES (?, ?, ?, ?, ?, ?, ?)",
		e.Name,
		e.Type,
		obj,
		e.Path,
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
	conn, release, err := c.getConn(ctx)
	if err != nil {
		return err
	}
	defer release()

	// Serialize object
	obj, err := proto.Marshal(e.Object)
	if err != nil {
		return err
	}

	_, err = conn.ExecContext(
		ctx,
		"UPDATE rill.catalog SET type = ?, object = ?, path = ?, updated_on = ?, refreshed_on = ? WHERE name = ?",
		e.Type,
		obj,
		e.Path,
		e.UpdatedOn, // TODO: Use time.Now()
		e.RefreshedOn,
		e.Name,
	)
	if err != nil {
		return err
	}

	return nil
}

func (c *connection) DeleteEntry(ctx context.Context, instanceID string, name string) error {
	conn, release, err := c.getConn(ctx)
	if err != nil {
		return err
	}
	defer release()

	_, err = conn.ExecContext(ctx, "DELETE FROM rill.catalog WHERE LOWER(name) = LOWER(?)", name)
	return err
}
