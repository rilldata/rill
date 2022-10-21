package duckdb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/rilldata/rill/runtime/drivers"
)

func (c *connection) FindObjects(ctx context.Context, instanceID string) []*drivers.CatalogObject {
	var res []*drivers.CatalogObject
	err := c.db.Select(&res, "SELECT * FROM rill.catalog ORDER BY name")
	if err != nil {
		panic(err)
	}
	return res
}

func (c *connection) FindObject(ctx context.Context, instanceID string, name string) (*drivers.CatalogObject, bool) {
	res := &drivers.CatalogObject{}
	// Names are stored with case everywhere but the checks should be case-insensitive.
	// Hence, the translation to lower case here
	err := c.db.QueryRowxContext(ctx, "SELECT * FROM rill.catalog WHERE LOWER(name) = ?", strings.ToLower(name)).StructScan(res)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, false
		}
		panic(err)
	}
	return res, true
}

func (c *connection) CreateObject(ctx context.Context, instanceID string, obj *drivers.CatalogObject) error {
	// safeguard to make sure duplicates are not created
	_, found := c.FindObject(ctx, instanceID, obj.Name)
	if found {
		return errors.New(fmt.Sprintf("duplicate key : %s", obj.Name))
	}

	now := time.Now()
	_, err := c.db.ExecContext(
		ctx,
		"INSERT INTO rill.catalog(name, type, sql, refreshed_on, created_on, updated_on) VALUES (?, ?, ?, ?, ?, ?)",
		obj.Name,
		obj.Type,
		obj.SQL,
		now,
		now,
		now,
	)
	if err != nil {
		return err
	}
	// We assign manually instead of using RETURNING because it doesn't work for timestamps in SQLite
	obj.RefreshedOn = now
	obj.CreatedOn = now
	obj.UpdatedOn = now
	return nil
}

func (c *connection) UpdateObject(ctx context.Context, instanceID string, obj *drivers.CatalogObject) error {
	now := time.Now()
	_, err := c.db.ExecContext(
		ctx,
		"UPDATE rill.catalog SET type = ?, sql = ?, refreshed_on = ?, updated_on = ? WHERE name = ?",
		obj.Type,
		obj.SQL,
		obj.RefreshedOn,
		now,
		obj.Name,
	)
	if err != nil {
		return err
	}
	// We assign manually instead of using RETURNING because it doesn't work for timestamps in SQLite
	obj.UpdatedOn = now
	return nil
}

func (c *connection) DeleteObject(ctx context.Context, instanceID string, name string) error {
	_, err := c.db.ExecContext(ctx, "DELETE FROM rill.catalog WHERE name = ?", name)
	return err
}
