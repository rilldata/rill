package sqlite

import (
	"context"
	"time"

	"github.com/rilldata/rill/runtime/drivers"
)

type catalogStore struct {
	*connection
	instanceID string
}

func (c *catalogStore) NextControllerVersion(ctx context.Context) (int64, error) {
	_, err := c.db.ExecContext(ctx, "INSERT OR IGNORE INTO controller_version(instance_id, version) VALUES (?, 0)", c.instanceID)
	if err != nil {
		return 0, err
	}

	_, err = c.db.ExecContext(ctx, "UPDATE controller_version SET version = version + 1 WHERE instance_id=?", c.instanceID)
	if err != nil {
		return 0, err
	}

	// TODO: Get it transactionally
	var version int64
	err = c.db.QueryRowContext(ctx, "SELECT version FROM controller_version WHERE instance_id=?", c.instanceID).Scan(&version)
	if err != nil {
		return 0, err
	}

	return version, nil
}

func (c *catalogStore) CheckControllerVersion(ctx context.Context, v int64) error {
	var version int64
	err := c.db.QueryRowContext(ctx, "SELECT version FROM controller_version WHERE instance_id=?", c.instanceID).Scan(&version)
	if err != nil {
		return err
	}

	if version != v {
		return drivers.ErrInconsistentControllerVersion
	}

	return nil
}

func (c *catalogStore) FindResources(ctx context.Context) ([]drivers.Resource, error) {
	rows, err := c.db.QueryxContext(ctx, "SELECT kind, name, data FROM catalogv2 WHERE instance_id=? ORDER BY kind, lower(name)", c.instanceID)
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

func (c *catalogStore) CreateResource(ctx context.Context, v int64, r drivers.Resource) error {
	err := c.CheckControllerVersion(ctx, v) // TODO: Do it transactionally
	if err != nil {
		return err
	}

	now := time.Now()
	_, err = c.db.ExecContext(
		ctx,
		"INSERT INTO catalogv2(instance_id, kind, name, data, created_on, updated_on) VALUES (?, ?, ?, ?, ?, ?)",
		c.instanceID,
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

func (c *catalogStore) UpdateResource(ctx context.Context, v int64, r drivers.Resource) error {
	err := c.CheckControllerVersion(ctx, v) // TODO: Do it transactionally
	if err != nil {
		return err
	}

	_, err = c.db.ExecContext(
		ctx,
		"UPDATE catalogv2 SET name=?, data=?, updated_on=? WHERE instance_id=? AND kind=? AND lower(name)=lower(?)",
		r.Name,
		r.Data,
		time.Now(),
		c.instanceID,
		r.Kind,
		r.Name,
	)
	if err != nil {
		return err
	}

	return nil
}

func (c *catalogStore) DeleteResource(ctx context.Context, v int64, k, n string) error {
	err := c.CheckControllerVersion(ctx, v) // TODO: Do it transactionally
	if err != nil {
		return err
	}

	_, err = c.db.ExecContext(ctx, "DELETE FROM catalogv2 WHERE kind=? AND lower(name)=lower(?)", k, n)
	if err != nil {
		return err
	}

	return nil
}

func (c *catalogStore) DeleteResources(ctx context.Context) error {
	_, err := c.db.ExecContext(ctx, "DELETE FROM catalogv2")
	if err != nil {
		return err
	}

	return nil
}
