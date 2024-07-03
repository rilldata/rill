package duckdb

import (
	"context"
	"fmt"
	"time"

	"github.com/rilldata/rill/runtime/drivers"
)

// NOTE: Not using RETURNING, indexes, transactions or any other fancy features to avoid DuckDB bugs.
// The DuckDB database lock and acquireMetaConn anyway ensure we don't need to worry about concurrency.

func (c *connection) NextControllerVersion(ctx context.Context) (int64, error) {
	conn, release, err := c.acquireMetaConn(ctx)
	if err != nil {
		return 0, err
	}
	defer func() { _ = release() }()

	_, err = conn.ExecContext(ctx, "UPDATE rill.controller_version SET version = version + 1")
	if err != nil {
		return 0, c.checkErr(err)
	}

	var version int64
	err = conn.QueryRowContext(ctx, "SELECT version FROM rill.controller_version").Scan(&version)
	if err != nil {
		return 0, c.checkErr(err)
	}

	return version, nil
}

func (c *connection) CheckControllerVersion(ctx context.Context, v int64) error {
	conn, release, err := c.acquireMetaConn(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = release() }()

	var version int64
	err = conn.QueryRowContext(ctx, "SELECT version FROM rill.controller_version").Scan(&version)
	if err != nil {
		return c.checkErr(err)
	}

	if version != v {
		return drivers.ErrInconsistentControllerVersion
	}

	return nil
}

func (c *connection) FindResources(ctx context.Context) ([]drivers.Resource, error) {
	conn, release, err := c.acquireMetaConn(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = release() }()

	rows, err := conn.QueryxContext(ctx, "SELECT kind, name, data FROM rill.catalogv2 ORDER BY kind, name")
	if err != nil {
		return nil, c.checkErr(err)
	}
	defer rows.Close()

	var res []drivers.Resource
	for rows.Next() {
		r := drivers.Resource{}
		err := rows.Scan(&r.Kind, &r.Name, &r.Data)
		if err != nil {
			return nil, c.checkErr(err)
		}
		res = append(res, r)
	}

	if rows.Err() != nil {
		return nil, c.checkErr(rows.Err())
	}

	return res, nil
}

func (c *connection) CreateResource(ctx context.Context, v int64, r drivers.Resource) error {
	err := c.CheckControllerVersion(ctx, v)
	if err != nil {
		return err
	}

	conn, release, err := c.acquireMetaConn(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = release() }()

	// Using an application side check instead of unique index becauwse of DuckDB limitations on indexes:
	// https://duckdb.org/docs/sql/indexes#over-eager-unique-constraint-checking
	var exists bool
	if err := conn.QueryRowContext(ctx, "SELECT EXISTS (SELECT 1 FROM rill.catalogv2 WHERE kind = ? AND lower(name) = lower(?))", r.Kind, r.Name).Scan(&exists); err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("catalog entry for type=%q, name=%q already exists", r.Kind, r.Name)
	}

	now := time.Now()
	_, err = conn.ExecContext(
		ctx,
		"INSERT INTO rill.catalogv2(kind, name, data, created_on, updated_on) VALUES (?, ?, ?, ?, ?)",
		r.Kind,
		r.Name,
		r.Data,
		now,
		now,
	)
	if err != nil {
		return c.checkErr(err)
	}

	return nil
}

func (c *connection) UpdateResource(ctx context.Context, v int64, r drivers.Resource) error {
	err := c.CheckControllerVersion(ctx, v)
	if err != nil {
		return err
	}

	conn, release, err := c.acquireMetaConn(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = release() }()

	res, err := conn.ExecContext(
		ctx,
		"UPDATE rill.catalogv2 SET name=?, data=?, updated_on=? WHERE kind=? AND lower(name)=lower(?)",
		r.Name,
		r.Data,
		time.Now(),
		r.Kind,
		r.Name,
	)
	if err != nil {
		return c.checkErr(err)
	}
	if n, err := res.RowsAffected(); err == nil {
		if n != 1 {
			return fmt.Errorf("catalog entry for type=%q, name=%q not found", r.Kind, r.Name)
		}
	}

	return nil
}

func (c *connection) DeleteResource(ctx context.Context, v int64, k, n string) error {
	err := c.CheckControllerVersion(ctx, v)
	if err != nil {
		return err
	}

	conn, release, err := c.acquireMetaConn(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = release() }()

	_, err = conn.ExecContext(ctx, "DELETE FROM rill.catalogv2 WHERE kind=? AND lower(name)=lower(?)", k, n)
	if err != nil {
		return c.checkErr(err)
	}

	return nil
}

func (c *connection) DeleteResources(ctx context.Context) error {
	conn, release, err := c.acquireMetaConn(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = release() }()

	_, err = conn.ExecContext(ctx, "DELETE FROM rill.catalogv2")
	if err != nil {
		return c.checkErr(err)
	}

	return nil
}
