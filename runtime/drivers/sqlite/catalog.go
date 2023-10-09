package sqlite

import (
	"context"
	"fmt"
	"time"

	"github.com/rilldata/rill/runtime/drivers"
)

type catalogStore struct {
	*connection
	instanceID string
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
