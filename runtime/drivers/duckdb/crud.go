package duckdb

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/graceful"
	"github.com/rilldata/rill/runtime/pkg/rduckdb"
	"go.uber.org/zap"
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
	// ByName only applies to the append strategy.
	// The merge and partition_overwrite strategies always insert by name since they reconcile the schema first.
	ByName    bool
	Strategy  drivers.IncrementalStrategy
	UniqueKey []string
	// PartitionBy is a SQL expression to use for dropping/replacing partitions with the partition_overwrite incremental strategy.
	PartitionBy string
	// OnSchemaChange controls how schema differences are handled for merge and partition_overwrite inserts.
	OnSchemaChange drivers.OnSchemaChange
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

	if opts.Strategy == drivers.IncrementalStrategyMerge || opts.Strategy == drivers.IncrementalStrategyPartitionOverwrite {
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
			defer func() {
				bgctx, cancel := graceful.WithMinimumDuration(ctx, time.Second*10)
				defer cancel()
				_, err := conn.ExecContext(bgctx, fmt.Sprintf("DROP TABLE %s", safeSQLName(tmp)))
				if err != nil {
					c.logger.Warn("failed to drop temporary table", zap.Error(err))
				}
			}()

			// Check the count of the new data.
			// Skip schema handling and insertion if the count is 0 because empty files can have an incorrectly detected schema.
			var empty bool
			err = conn.QueryRowxContext(ctx, fmt.Sprintf("SELECT COUNT(*) == 0 FROM %s", safeSQLName(tmp))).Scan(&empty)
			if err != nil {
				return err
			}
			if empty {
				return nil
			}

			targetColumns, err := readDuckDBColumns(ctx, conn, name)
			if err != nil {
				return fmt.Errorf("failed to read schema of %q: %w", name, err)
			}
			sourceColumns, err := readDuckDBColumns(ctx, conn, tmp)
			if err != nil {
				return fmt.Errorf("failed to read schema of %q: %w", tmp, err)
			}

			// Validate the key columns before applying any schema change, so a failed run leaves the target table untouched.
			switch opts.Strategy {
			case drivers.IncrementalStrategyMerge:
				for _, key := range opts.UniqueKey {
					if !containsColumn(sourceColumns, key) {
						return fmt.Errorf("the new data does not contain the %q column from %q", key, "unique_key")
					}
					// Adding a key column would give every existing row a NULL key, which merge treats as a single key.
					if !containsColumn(targetColumns, key) {
						return fmt.Errorf("table %q does not contain the %q column from %q; run a full refresh to change the key", name, key, "unique_key")
					}
				}
			case drivers.IncrementalStrategyPartitionOverwrite:
				// Check that the partition expression resolves against the new data on its own.
				// The DELETE below references it unqualified in a subquery, so a column that is missing from the new data
				// would instead bind to the target table as a correlated reference and match every row.
				_, err = conn.ExecContext(ctx, fmt.Sprintf("SELECT %s FROM %s LIMIT 0", opts.PartitionBy, safeSQLName(tmp)))
				if err != nil {
					return fmt.Errorf("failed to resolve the %q expression %q against the new data: %w", "partition_by", opts.PartitionBy, err)
				}
			}

			cols, err := applySchemaChange(ctx, conn, name, targetColumns, sourceColumns, opts.OnSchemaChange, c.logger)
			if err != nil {
				return err
			}

			switch opts.Strategy {
			case drivers.IncrementalStrategyMerge:
				// Drop the rows from the target table where the unique key is present in the temporary table.
				where := ""
				for i, key := range opts.UniqueKey {
					key = safeSQLName(key)
					if i != 0 {
						where += " AND "
					}
					where += fmt.Sprintf("base.%s IS NOT DISTINCT FROM tmp.%s", key, key)
				}
				_, err = conn.ExecContext(ctx, fmt.Sprintf("DELETE FROM %s base WHERE EXISTS (SELECT 1 FROM %s tmp WHERE %s)", safeSQLName(name), safeSQLName(tmp), where))
				if err != nil {
					return err
				}
			case drivers.IncrementalStrategyPartitionOverwrite:
				// Drop the rows from the target table where the partition expression overlaps with the temporary table.
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
			}

			insertInto := make([]string, len(cols.target))
			selectFrom := make([]string, len(cols.source))
			for i := range cols.target {
				insertInto[i] = safeSQLName(cols.target[i])
				selectFrom[i] = safeSQLName(cols.source[i])
			}
			_, err = conn.ExecContext(ctx, fmt.Sprintf(
				"INSERT INTO %s (%s) SELECT %s FROM %s",
				safeSQLName(name),
				strings.Join(insertInto, ", "),
				strings.Join(selectFrom, ", "),
				safeSQLName(tmp),
			))
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

func (c *connection) mutateTable(ctx context.Context, name, preExec, postExec string) error {
	db, release, err := c.acquireDB()
	if err != nil {
		return err
	}
	defer func() {
		_ = release()
	}()
	_, err = db.MutateTable(ctx, name, nil, func(ctx context.Context, conn *sqlx.Conn) error {
		if preExec != "" {
			if _, err := conn.ExecContext(ctx, preExec); err != nil {
				return err
			}
		}
		_, err := conn.ExecContext(ctx, postExec)
		return err
	})
	return c.checkErr(err)
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
	return drivers.EscapeStringValue(name)
}
