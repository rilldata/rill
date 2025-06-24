package sqlite

import (
	"context"
	"strings"
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

	_, err = c.db.ExecContext(ctx, "DELETE FROM catalogv2 WHERE instance_id=? AND kind=? AND lower(name)=lower(?)", c.instanceID, k, n)
	if err != nil {
		return err
	}

	return nil
}

func (c *catalogStore) DeleteResources(ctx context.Context) error {
	_, err := c.db.ExecContext(ctx, "DELETE FROM catalogv2 WHERE instance_id=?", c.instanceID)
	if err != nil {
		return err
	}

	return nil
}

func (c *catalogStore) FindModelPartitions(ctx context.Context, opts *drivers.FindModelPartitionsOptions) ([]drivers.ModelPartition, error) {
	var qry strings.Builder
	var args []any

	qry.WriteString("SELECT key, data_json, idx, watermark, executed_on, error, elapsed_ms FROM model_partitions WHERE instance_id=? AND model_id=?")
	args = append(args, c.instanceID, opts.ModelID)

	if opts.WhereErrored {
		qry.WriteString(" AND error != ''")
	}

	if opts.WherePending {
		qry.WriteString(" AND executed_on IS NULL")
	}

	if opts.AfterIndex != 0 || opts.AfterKey != "" {
		qry.WriteString(" AND (idx > ? OR (idx = ? AND key > ?))")
		args = append(args, opts.AfterIndex, opts.AfterIndex, opts.AfterKey)
	}

	qry.WriteString(" ORDER BY idx, key")

	if opts.Limit != 0 {
		qry.WriteString(" LIMIT ?")
		args = append(args, opts.Limit)
	}

	rows, err := c.db.QueryContext(ctx, qry.String(), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []drivers.ModelPartition
	for rows.Next() {
		var elapsedMs int64
		r := drivers.ModelPartition{}
		err := rows.Scan(&r.Key, &r.DataJSON, &r.Index, &r.Watermark, &r.ExecutedOn, &r.Error, &elapsedMs)
		if err != nil {
			return nil, err
		}
		r.Elapsed = time.Duration(elapsedMs) * time.Millisecond
		res = append(res, r)
	}

	if rows.Err() != nil {
		return nil, err
	}

	return res, nil
}

func (c *catalogStore) FindModelPartitionsByKeys(ctx context.Context, modelID string, keys []string) ([]drivers.ModelPartition, error) {
	// We can't pass a []string as a bound parameter, so we have to build a query with a corresponding number of placeholders.
	var qry strings.Builder
	var args []any
	qry.WriteString("SELECT key, data_json, idx, watermark, executed_on, error, elapsed_ms FROM model_partitions WHERE instance_id=? AND model_id=? AND key IN (")
	args = append(args, c.instanceID, modelID)

	qry.Grow(len(keys)*2 + 14) // Makes room for one ",?" per key plus the ORDER BY clause
	for i, k := range keys {
		if i == 0 {
			qry.WriteString("?")
		} else {
			qry.WriteString(",?")
		}
		args = append(args, k)
	}
	qry.WriteString(") ORDER BY key")

	rows, err := c.db.QueryxContext(ctx, qry.String(), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []drivers.ModelPartition
	for rows.Next() {
		var elapsedMs int64
		r := drivers.ModelPartition{}
		err := rows.Scan(&r.Key, &r.DataJSON, &r.Index, &r.Watermark, &r.ExecutedOn, &r.Error, &elapsedMs)
		if err != nil {
			return nil, err
		}
		r.Elapsed = time.Duration(elapsedMs) * time.Millisecond
		res = append(res, r)
	}

	if rows.Err() != nil {
		return nil, err
	}

	return res, nil
}

func (c *catalogStore) CheckModelPartitionsHaveErrors(ctx context.Context, modelID string) (bool, error) {
	rows, err := c.db.QueryContext(
		ctx,
		"SELECT 1 FROM model_partitions WHERE instance_id=? AND model_id=? AND error != '' LIMIT 1",
		c.instanceID,
		modelID,
	)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	var hasErrors bool
	if rows.Next() {
		hasErrors = true
	}

	if rows.Err() != nil {
		return false, err
	}

	return hasErrors, nil
}

func (c *catalogStore) InsertModelPartition(ctx context.Context, modelID string, partition drivers.ModelPartition) error {
	_, err := c.db.ExecContext(
		ctx,
		"INSERT INTO model_partitions(instance_id, model_id, key, data_json, idx, watermark, executed_on, error, elapsed_ms) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		c.instanceID,
		modelID,
		partition.Key,
		partition.DataJSON,
		partition.Index,
		partition.Watermark,
		partition.ExecutedOn,
		partition.Error,
		partition.Elapsed.Milliseconds(),
	)
	if err != nil {
		return err
	}

	return nil
}

func (c *catalogStore) UpdateModelPartition(ctx context.Context, modelID string, partition drivers.ModelPartition) error {
	_, err := c.db.ExecContext(
		ctx,
		"UPDATE model_partitions SET data_json=?, idx=?, watermark=?, executed_on=?, error=?, elapsed_ms=? WHERE instance_id=? AND model_id=? AND key=?",
		partition.DataJSON,
		partition.Index,
		partition.Watermark,
		partition.ExecutedOn,
		partition.Error,
		partition.Elapsed.Milliseconds(),
		c.instanceID,
		modelID,
		partition.Key,
	)
	if err != nil {
		return err
	}

	return nil
}

func (c *catalogStore) UpdateModelPartitionPending(ctx context.Context, modelID, partitionKey string) error {
	_, err := c.db.ExecContext(
		ctx,
		"UPDATE model_partitions SET executed_on=NULL WHERE instance_id=? AND model_id=? AND key=?",
		c.instanceID,
		modelID,
		partitionKey,
	)
	if err != nil {
		return err
	}

	return nil
}

func (c *catalogStore) UpdateModelPartitionsPendingIfError(ctx context.Context, modelID string) error {
	_, err := c.db.ExecContext(
		ctx,
		"UPDATE model_partitions SET executed_on=NULL WHERE instance_id=? AND model_id=? AND error != ''",
		c.instanceID,
		modelID,
	)
	if err != nil {
		return err
	}

	return nil
}

func (c *catalogStore) DeleteModelPartitions(ctx context.Context, modelID string) error {
	_, err := c.db.ExecContext(ctx, "DELETE FROM model_partitions WHERE instance_id=? AND model_id=?", c.instanceID, modelID)
	if err != nil {
		return err
	}

	return nil
}

func (c *catalogStore) FindInstanceHealth(ctx context.Context, instanceID string) (*drivers.InstanceHealth, error) {
	var h drivers.InstanceHealth
	err := c.db.QueryRowContext(ctx, "SELECT health_json, updated_on FROM instance_health WHERE instance_id=?", instanceID).Scan(&h.HealthJSON, &h.UpdatedOn)
	if err != nil {
		return nil, err
	}

	return &h, nil
}

func (c *catalogStore) UpsertInstanceHealth(ctx context.Context, h *drivers.InstanceHealth) error {
	_, err := c.db.ExecContext(ctx, `INSERT INTO instance_health(instance_id, health_json, updated_on) Values (?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(instance_id) DO UPDATE SET health_json=excluded.health_json, updated_on=excluded.updated_on;
	`, h.InstanceID, h.HealthJSON)
	return err
}
