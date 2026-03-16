package duckdb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/rduckdb"
)

type tableWriteMetrics struct {
	duration time.Duration
}

type createTableOptions struct {
	view         bool
	initQueries  []string
	beforeCreate string
	afterCreate  string
}

func (c *connection) createTableAsSelect(ctx context.Context, name, sql string, opts *createTableOptions) (*tableWriteMetrics, error) {
	db, release, err := c.acquireDB()
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = release()
	}()
	var beforeCreateFn, afterCreateFn func(ctx context.Context, conn *sqlx.Conn) error
	if opts.beforeCreate != "" {
		beforeCreateFn = func(ctx context.Context, conn *sqlx.Conn) error {
			_, err := conn.ExecContext(ctx, opts.beforeCreate)
			return err
		}
	}
	if opts.afterCreate != "" {
		afterCreateFn = func(ctx context.Context, conn *sqlx.Conn) error {
			_, err := conn.ExecContext(ctx, opts.afterCreate)
			return err
		}
	}
	res, err := db.CreateTableAsSelect(ctx, name, sql, &rduckdb.CreateTableOptions{
		View:           opts.view,
		InitQueries:    opts.initQueries,
		BeforeCreateFn: beforeCreateFn,
		AfterCreateFn:  afterCreateFn,
	})
	if err != nil {
		return nil, c.checkErr(err)
	}
	return &tableWriteMetrics{
		duration: res.Duration,
	}, nil
}

type InsertTableOptions struct {
	InitQueries  []string
	BeforeInsert string
	AfterInsert  string
	ByName       bool
	Strategy     drivers.IncrementalStrategy
	UniqueKey    []string
	// PartitionBy is a SQL expression to use for dropping/replacing partitions with the partition_overwrite incremental strategy.
	PartitionBy string
	// MergeBatchSize controls how many rows from the temp table are matched per DELETE batch during merge.
	// If 0, defaults to 1000000.
	MergeBatchSize int
}

func (c *connection) insertTableAsSelect(ctx context.Context, name, sql string, opts *InsertTableOptions) (*tableWriteMetrics, error) {
	db, release, err := c.acquireDB()
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = release()
	}()
	var byNameClause string
	if opts.ByName {
		byNameClause = "BY NAME"
	}

	if opts.Strategy == drivers.IncrementalStrategyAppend {
		res, err := db.MutateTable(ctx, name, opts.InitQueries, func(ctx context.Context, conn *sqlx.Conn) (retErr error) {
			// Execute the pre SQL and defer execute the post SQL
			if opts.BeforeInsert != "" {
				_, err := conn.ExecContext(ctx, opts.BeforeInsert)
				if err != nil {
					return err
				}
			}
			if opts.AfterInsert != "" {
				defer func() {
					_, afterInsertErr := conn.ExecContext(ctx, opts.AfterInsert)
					retErr = errors.Join(retErr, afterInsertErr)
				}()
			}

			_, err := conn.ExecContext(ctx, fmt.Sprintf("INSERT INTO %s %s (%s\n)", safeSQLName(name), byNameClause, sql))
			return err
		})
		if err != nil {
			return nil, c.checkErr(err)
		}
		return &tableWriteMetrics{
			duration: res.Duration,
		}, nil
	}

	if opts.Strategy == drivers.IncrementalStrategyMerge {
		res, err := db.MutateTable(ctx, name, opts.InitQueries, func(ctx context.Context, conn *sqlx.Conn) (retErr error) {
			// Execute the pre SQL and defer execute the post SQL
			if opts.BeforeInsert != "" {
				_, err := conn.ExecContext(ctx, opts.BeforeInsert)
				if err != nil {
					return err
				}
			}
			if opts.AfterInsert != "" {
				defer func() {
					_, afterInsertErr := conn.ExecContext(ctx, opts.AfterInsert)
					retErr = errors.Join(retErr, afterInsertErr)
				}()
			}

			// Create a temporary table with the new data
			tmp := fmt.Sprintf("__rill_temp_%s", name)
			_, err := conn.ExecContext(ctx, fmt.Sprintf("CREATE OR REPLACE TABLE %s AS (%s\n)", safeSQLName(tmp), sql))
			if err != nil {
				return err
			}

			// check the count of the new data
			// skip if the count is 0
			// if there was no data in the empty file then the detected schema can be different from the current schema which leads to errors or performance issues
			var count int
			err = conn.QueryRowxContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM %s", safeSQLName(tmp))).Scan(&count)
			if err != nil {
				return err
			}
			if count == 0 {
				return nil
			}

			// Build the WHERE clause for unique key matching
			where := ""
			for i, key := range opts.UniqueKey {
				key = safeSQLName(key)
				if i != 0 {
					where += " AND "
				}
				where += fmt.Sprintf("base.%s IS NOT DISTINCT FROM tmp.%s", key, key)
			}

			// Drop the rows from the target table in batches to limit peak memory usage
			// from the join on the tmp table.
			deleteBatchSize := opts.MergeBatchSize
			if deleteBatchSize <= 0 {
				deleteBatchSize = 1000000
			}
			for num := 0; num <= count; num += deleteBatchSize {
				_, err = conn.ExecContext(ctx, fmt.Sprintf(
					"DELETE FROM %s base WHERE EXISTS (SELECT 1 FROM %s tmp WHERE tmp.__rill_row_num >= %d AND tmp.__rill_row_num < %d AND %s)",
					safeSQLName(name), safeSQLName(tmp), num, num+deleteBatchSize, where,
				))
				if err != nil {
					return err
				}
			}

			// Insert the new data into the target table, excluding the internal row number column.
			_, err = conn.ExecContext(ctx, fmt.Sprintf("INSERT INTO %s %s SELECT * EXCLUDE (__rill_row_num) FROM %s", safeSQLName(name), byNameClause, safeSQLName(tmp)))
			return err
		})
		if err != nil {
			return nil, c.checkErr(err)
		}
		return &tableWriteMetrics{
			duration: res.Duration,
		}, nil
	}

	if opts.Strategy == drivers.IncrementalStrategyPartitionOverwrite {
		res, err := db.MutateTable(ctx, name, opts.InitQueries, func(ctx context.Context, conn *sqlx.Conn) (retErr error) {
			// Execute the pre SQL and defer execute the post SQL
			if opts.BeforeInsert != "" {
				_, err := conn.ExecContext(ctx, opts.BeforeInsert)
				if err != nil {
					return err
				}
			}
			if opts.AfterInsert != "" {
				defer func() {
					_, afterInsertErr := conn.ExecContext(ctx, opts.AfterInsert)
					retErr = errors.Join(retErr, afterInsertErr)
				}()
			}

			// Create a temporary table with the new data
			tmp := fmt.Sprintf("__rill_temp_%s", name)
			_, err := conn.ExecContext(ctx, fmt.Sprintf("CREATE OR REPLACE TABLE %s AS (%s\n)", safeSQLName(tmp), sql))
			if err != nil {
				return err
			}

			// Check the count of the new data
			// Skip if the count is 0
			var empty bool
			err = conn.QueryRowxContext(ctx, fmt.Sprintf("SELECT COUNT(*) == 0 FROM %s", safeSQLName(tmp))).Scan(&empty)
			if err != nil {
				return err
			}
			if empty {
				return nil
			}

			// Drop the rows from the target table where the partition expression overlaps with the temporary table
			_, err = conn.ExecContext(ctx, fmt.Sprintf(
				"DELETE FROM %s WHERE %s IN (SELECT DISTINCT %s FROM %s)",
				safeSQLName(name),
				opts.PartitionBy,
				opts.PartitionBy,
				safeSQLName(tmp),
			))
			if err != nil {
				return fmt.Errorf("failed to delete old partitions: %w", err)
			}

			// Insert the new data into the target table
			_, err = conn.ExecContext(ctx, fmt.Sprintf("INSERT INTO %s %s SELECT * FROM %s", safeSQLName(name), byNameClause, safeSQLName(tmp)))
			return err
		})
		if err != nil {
			return nil, c.checkErr(err)
		}
		return &tableWriteMetrics{
			duration: res.Duration,
		}, nil
	}

	return nil, fmt.Errorf("incremental insert strategy %q not supported", opts.Strategy)
}

func (c *connection) dropTable(ctx context.Context, name string) error {
	db, release, err := c.acquireDB()
	if err != nil {
		return err
	}
	defer func() {
		_ = release()
	}()
	err = db.DropTable(ctx, name)
	return c.checkErr(err)
}

func (c *connection) renameTable(ctx context.Context, oldName, newName string) error {
	db, release, err := c.acquireDB()
	if err != nil {
		return err
	}
	defer func() {
		_ = release()
	}()
	err = db.RenameTable(ctx, oldName, newName)
	return c.checkErr(err)
}

func rowsToSchema(r *sqlx.Rows) (*runtimev1.StructType, error) {
	if r == nil {
		return nil, nil
	}

	cts, err := r.ColumnTypes()
	if err != nil {
		return nil, err
	}

	fields := make([]*runtimev1.StructType_Field, len(cts))
	for i, ct := range cts {
		nullable, ok := ct.Nullable()
		if !ok {
			nullable = true
		}

		t, err := databaseTypeToPB(ct.DatabaseTypeName(), nullable)
		if err != nil {
			return nil, err
		}

		fields[i] = &runtimev1.StructType_Field{
			Name: ct.Name(),
			Type: t,
		}
	}

	return &runtimev1.StructType{Fields: fields}, nil
}

// safeSQLName returns a quoted SQL identifier.
func safeSQLName(name string) string {
	return safeName(name)
}

func safeSQLString(name string) string {
	return drivers.DialectDuckDB.EscapeStringValue(name)
}
