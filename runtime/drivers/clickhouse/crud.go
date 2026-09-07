package clickhouse

import (
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/graceful"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.uber.org/zap"
)

// tableWriteMetrics reports metrics for an execution that mutates table data.
type tableWriteMetrics struct {
	// duration is the time taken to run user queries only.
	duration time.Duration
}

func (c *Connection) createTableAsSelect(ctx context.Context, name, sql string, outputProps *ModelOutputProperties, beforeCreate, afterCreate string) (*tableWriteMetrics, error) {
	onClusterClause := c.onClusterClause()

	// Execute beforeCreate query
	if beforeCreate != "" {
		if err := c.Exec(ctx, &drivers.Statement{Query: beforeCreate, Priority: 100}); err != nil {
			return nil, fmt.Errorf("failed to execute pre_exec: %w", err)
		}
	}

	t := time.Now()
	if outputProps.Typ == "VIEW" {
		err := c.Exec(ctx, &drivers.Statement{
			Query:    fmt.Sprintf("CREATE OR REPLACE VIEW %s %s AS %s", safeSQLName(name), onClusterClause, sql),
			Priority: 100,
		})
		if err != nil {
			return nil, err
		}
		if afterCreate != "" {
			if err := c.Exec(ctx, &drivers.Statement{Query: afterCreate, Priority: 100}); err != nil {
				return nil, fmt.Errorf("failed to execute post_exec: %w", err)
			}
		}
		return &tableWriteMetrics{duration: time.Since(t)}, nil
	} else if outputProps.Typ == "DICTIONARY" {
		err := c.createDictionary(ctx, name, sql, outputProps)
		if err != nil {
			return nil, err
		}
		if afterCreate != "" {
			if err := c.Exec(ctx, &drivers.Statement{Query: afterCreate, Priority: 100}); err != nil {
				return nil, fmt.Errorf("failed to execute post_exec: %w", err)
			}
		}
		return &tableWriteMetrics{duration: time.Since(t)}, nil
	}
	// on replicated databases `create table t as select * from ...` is prohibited
	// so we need to create a table first and then insert data into it
	if err := c.createTable(ctx, name, sql, outputProps); err != nil {
		return nil, err
	}
	// insert into table
	err := c.Exec(ctx, &drivers.Statement{
		Query:    fmt.Sprintf("INSERT INTO %s %s", safeSQLName(name), sql),
		Priority: 100,
	})
	if err != nil {
		return nil, err
	}

	// Execute afterCreate query
	if afterCreate != "" {
		if err := c.Exec(ctx, &drivers.Statement{Query: afterCreate, Priority: 100}); err != nil {
			return nil, fmt.Errorf("failed to execute post_exec: %w", err)
		}
	}
	return &tableWriteMetrics{duration: time.Since(t)}, nil
}

type InsertTableOptions struct {
	Strategy     drivers.IncrementalStrategy
	BeforeInsert string
	AfterInsert  string
}

func (c *Connection) insertTableAsSelect(ctx context.Context, name, sql string, opts *InsertTableOptions, outputProps *ModelOutputProperties) (*tableWriteMetrics, error) {
	// Execute BeforeInsert query
	if opts.BeforeInsert != "" {
		if err := c.Exec(ctx, &drivers.Statement{Query: opts.BeforeInsert, Priority: 100}); err != nil {
			return nil, fmt.Errorf("failed to execute pre_exec: %w", err)
		}
	}

	start := time.Now()

	if opts.Strategy == drivers.IncrementalStrategyAppend {
		err := c.Exec(ctx, &drivers.Statement{
			Query:    fmt.Sprintf("INSERT INTO %s %s", safeSQLName(name), sql),
			Priority: 1,
		})
		if err != nil {
			return nil, err
		}
		if opts.AfterInsert != "" {
			if err := c.Exec(ctx, &drivers.Statement{Query: opts.AfterInsert, Priority: 100}); err != nil {
				return nil, fmt.Errorf("failed to execute post_exec: %w", err)
			}
		}
		return &tableWriteMetrics{duration: time.Since(start)}, nil
	}

	if opts.Strategy == drivers.IncrementalStrategyPartitionOverwrite {
		onClusterClause := c.onClusterClause()
		// Get the engine info of the given table
		engine, err := c.getTableEngine(ctx, name)
		if err != nil {
			return nil, err
		}
		// create temp table with the same schema using a deterministic name
		tempName := fmt.Sprintf("__rill_temp_%s_%x", name, md5.Sum([]byte(sql)))
		// clean up the temp table
		defer func() {
			// cleanup using a different ctx to prevent cleanups being impacted by the main ctx cancellation
			// this is a best effort cleanup and query can still timeout and we don't want to wait forever due to blocked calls
			// this is triggered before the table is even created to handle situations
			// where before the client can trigger query cancel the query succeeds and the view is created but the driver stil reports query cancelled
			ctx, cancel := graceful.WithMinimumDuration(ctx, 15*time.Second)
			defer cancel()

			err = c.dropTable(ctx, tempName)
			if err != nil && !errors.Is(err, drivers.ErrNotFound) {
				c.logger.Warn("clickhouse: failed to drop temp table", zap.String("name", tempName), zap.Error(err), observability.ZapCtx(ctx))
			}
		}()
		// create temp table
		if engine == "Distributed" {
			// create a local table first
			err = c.Exec(ctx, &drivers.Statement{
				Query:    fmt.Sprintf("CREATE OR REPLACE TABLE %s %s AS %s", safeSQLName(localTableName(tempName)), onClusterClause, safeSQLName(localTableName(name))),
				Priority: 1,
			})
			if err != nil {
				return nil, err
			}
			// then create the distributed table
			err = c.createDistributedTable(ctx, tempName, outputProps)
			if err != nil {
				return nil, err
			}
		} else {
			err = c.Exec(ctx, &drivers.Statement{
				Query:    fmt.Sprintf("CREATE OR REPLACE TABLE %s %s AS %s", safeSQLName(tempName), onClusterClause, safeSQLName(name)),
				Priority: 1,
			})
			if err != nil {
				return nil, err
			}
		}

		// insert into temp table
		err = c.Exec(ctx, &drivers.Statement{
			Query:    fmt.Sprintf("INSERT INTO %s %s", safeSQLName(tempName), sql),
			Priority: 1,
		})
		if err != nil {
			return nil, err
		}

		// sync the replica before partition replacement so all inserted parts are visible
		// SYSTEM SYNC REPLICA only works on replicated engines, so skip it for non-replicated ones
		if isReplicatedEngine(outputProps.Engine) || isReplicatedEngine(outputProps.EngineFull) {
			err = c.syncReplica(ctx, tempName)
			if err != nil {
				return nil, err
			}
		}

		// list partitions from the temp table
		partitions, err := c.getTablePartitions(ctx, tempName)
		if err != nil {
			return nil, err
		}
		// iterate over partitions and replace them in the main table
		for _, part := range partitions {
			// alter the main table to replace the partition
			err = c.replacePartition(ctx, tempName, name, part)
			if err != nil {
				return nil, err
			}
		}
		if opts.AfterInsert != "" {
			if err := c.Exec(ctx, &drivers.Statement{Query: opts.AfterInsert, Priority: 100}); err != nil {
				return nil, fmt.Errorf("failed to execute post_exec: %w", err)
			}
		}
		return &tableWriteMetrics{duration: time.Since(start)}, nil
	}

	if opts.Strategy == drivers.IncrementalStrategyMerge {
		// get the engine info of the given table - local table for distributed tables
		var n string
		if c.config.Cluster != "" {
			n = localTableName(name)
		} else {
			n = name
		}
		engine, err := c.getTableEngine(ctx, n)
		if err != nil {
			return nil, err
		}
		if !strings.Contains(engine, "ReplacingMergeTree") {
			return nil, fmt.Errorf("clickhouse: merge strategy requires ReplacingMergeTree engine")
		}

		// insert into table using the merge strategy
		err = c.Exec(ctx, &drivers.Statement{
			Query:    fmt.Sprintf("INSERT INTO %s %s", safeSQLName(name), sql),
			Priority: 1,
		})
		if err != nil {
			return nil, err
		}
		if opts.AfterInsert != "" {
			if err := c.Exec(ctx, &drivers.Statement{Query: opts.AfterInsert, Priority: 100}); err != nil {
				return nil, fmt.Errorf("failed to execute post_exec: %w", err)
			}
		}
		return &tableWriteMetrics{duration: time.Since(start)}, nil
	}

	return nil, fmt.Errorf("incremental insert strategy %q not supported", opts.Strategy)
}

func (c *Connection) dropTable(ctx context.Context, name string) error {
	typ, err := c.entityType(ctx, c.config.Database, name)
	if err != nil {
		return err
	}

	onClusterClause := c.onClusterClause()

	switch typ {
	case "VIEW":
		return c.Exec(ctx, &drivers.Statement{
			Query:    fmt.Sprintf("DROP VIEW IF EXISTS %s %s", safeSQLName(name), onClusterClause),
			Priority: 100,
		})
	case "DICTIONARY":
		// Resolve the source table up front, since the dependency that identifies it disappears with the dictionary.
		srcTable, err := c.dictionarySourceTable(ctx, name)
		if err != nil {
			return err
		}
		// first drop the dictionary
		err = c.Exec(ctx, &drivers.Statement{
			Query:    fmt.Sprintf("DROP DICTIONARY IF EXISTS %s %s", safeSQLName(name), onClusterClause),
			Priority: 100,
		})
		if err != nil {
			return err
		}
		// then drop the table it sourced from, which is now unreferenced
		if srcTable != "" {
			if dropErr := c.dropTable(ctx, srcTable); dropErr != nil && !errors.Is(dropErr, drivers.ErrNotFound) {
				c.logger.Warn("clickhouse: failed to drop dictionary source table", zap.String("name", srcTable), zap.Error(dropErr), observability.ZapCtx(ctx))
			}
		}
		return nil
	case "TABLE":
		// drop the main table
		// use IF EXISTS so drops succeed in cluster mode even for tables that don't exist on every node,
		// e.g. tables created before the cluster was configured or without a `_local` counterpart
		err := c.Exec(ctx, &drivers.Statement{
			Query:    fmt.Sprintf("DROP TABLE IF EXISTS %s %s", safeSQLName(name), onClusterClause),
			Priority: 100,
		})
		if err != nil {
			return err
		}
		// then drop the local table in case of cluster
		if c.config.Cluster != "" && !strings.HasSuffix(name, "_local") {
			return c.Exec(ctx, &drivers.Statement{
				Query:    fmt.Sprintf("DROP TABLE IF EXISTS %s %s", safeSQLName(localTableName(name)), onClusterClause),
				Priority: 100,
			})
		}
		return nil
	default:
		return fmt.Errorf("clickhouse: unknown entity type %q", typ)
	}
}

func (c *Connection) renameEntity(ctx context.Context, oldName, newName string) error {
	typ, err := c.entityType(ctx, c.config.Database, oldName)
	if err != nil {
		return err
	}
	onClusterClause := c.onClusterClause()

	switch typ {
	case "VIEW":
		return c.renameView(ctx, oldName, newName, onClusterClause)
	case "DICTIONARY":
		return c.renameTable(ctx, oldName, newName, onClusterClause)
	case "TABLE":
		if c.config.Cluster == "" {
			return c.renameTable(ctx, oldName, newName, onClusterClause)
		}
		// capture the full engine of the old distributed table
		args := []any{c.config.Database, oldName}
		if c.config.Database == "" {
			args = []any{nil, oldName}
		}
		var engineFull string
		res, err := c.Query(ctx, &drivers.Statement{
			Query:    "SELECT engine_full FROM system.tables WHERE database = coalesce(?, currentDatabase()) AND name = ?",
			Args:     args,
			Priority: 100,
		})
		if err != nil {
			return err
		}

		for res.Next() {
			if err := res.Scan(&engineFull); err != nil {
				res.Close()
				return err
			}
		}
		err = res.Err()
		if err != nil {
			return err
		}
		res.Close()
		engineFull = strings.ReplaceAll(engineFull, localTableName(oldName), localTableName(newName))

		// build the column type clause
		columnClause, err := c.columnClause(ctx, oldName)
		if err != nil {
			return err
		}

		// rename the local table
		err = c.renameTable(ctx, localTableName(oldName), localTableName(newName), onClusterClause)
		if err != nil {
			return err
		}

		// recreate the distributed table
		err = c.Exec(ctx, &drivers.Statement{
			Query:    fmt.Sprintf("CREATE OR REPLACE TABLE %s %s %s Engine = %s", safeSQLName(newName), onClusterClause, columnClause, engineFull),
			Priority: 100,
		})
		if err != nil {
			return err
		}

		// drop the old table
		return c.Exec(ctx, &drivers.Statement{
			Query:    fmt.Sprintf("DROP TABLE %s %s", safeSQLName(oldName), onClusterClause),
			Priority: 100,
		})
	default:
		return fmt.Errorf("clickhouse: unknown entity type %q", typ)
	}
}

func (c *Connection) renameView(ctx context.Context, oldName, newName, onCluster string) error {
	// clickhouse does not support renaming views so we capture the OLD view's select statement and use it to create new view
	args := []any{c.config.Database, oldName}
	if c.config.Database == "" {
		args = []any{nil, oldName}
	}
	res, err := c.Query(ctx, &drivers.Statement{
		Query:    "SELECT as_select FROM system.tables WHERE database = coalesce(?, currentDatabase()) AND name = ?",
		Args:     args,
		Priority: 100,
	})
	if err != nil {
		return err
	}

	var sql string
	if res.Next() {
		if err := res.Scan(&sql); err != nil {
			res.Close()
			return err
		}
	}
	err = res.Err()
	if err != nil {
		return err
	}
	res.Close()

	// create new view
	err = c.Exec(ctx, &drivers.Statement{
		Query:    fmt.Sprintf("CREATE OR REPLACE VIEW %s %s AS %s", safeSQLName(newName), onCluster, sql),
		Priority: 100,
	})
	if err != nil {
		return err
	}

	// drop old view
	err = c.Exec(ctx, &drivers.Statement{
		Query:    fmt.Sprintf("DROP VIEW %s %s", safeSQLName(oldName), onCluster),
		Priority: 100,
	})
	if err != nil {
		c.logger.Error("clickhouse: failed to drop old view", zap.String("name", oldName), zap.Error(err), observability.ZapCtx(ctx))
	}
	return nil
}

func (c *Connection) renameTable(ctx context.Context, oldName, newName, onCluster string) error {
	var exists bool
	err := c.writeDB.QueryRowContext(contextWithQueryID(ctx), fmt.Sprintf("EXISTS %s", safeSQLName(newName))).Scan(&exists)
	if err != nil {
		return err
	}
	if !exists {
		return c.Exec(ctx, &drivers.Statement{
			Query:    fmt.Sprintf("RENAME TABLE %s TO %s %s", safeSQLName(oldName), safeSQLName(newName), onCluster),
			Priority: 100,
		})
	}
	err = c.Exec(ctx, &drivers.Statement{
		Query:    fmt.Sprintf("EXCHANGE TABLES %s AND %s %s", safeSQLName(oldName), safeSQLName(newName), onCluster),
		Priority: 100,
	})
	if err != nil {
		return err
	}
	// drop the old table
	return c.dropTable(context.Background(), oldName)
}

func (c *Connection) createTable(ctx context.Context, name, sql string, outputProps *ModelOutputProperties) error {
	onClusterClause := c.onClusterClause()
	var create strings.Builder
	create.WriteString("CREATE OR REPLACE TABLE ")
	if c.config.Cluster != "" {
		// need to create a local table on the cluster first
		fmt.Fprintf(&create, "%s %s", safeSQLName(localTableName(name)), onClusterClause)
	} else {
		create.WriteString(safeSQLName(name))
	}

	if outputProps.Columns == "" {
		if sql == "" {
			return fmt.Errorf("clickhouse: no columns specified for table %q", name)
		}
		// infer columns
		v := safeSQLName(fmt.Sprintf("__rill_temp_%s_%x", name, md5.Sum([]byte(sql))))
		defer func() {
			// cleanup using a different ctx to prevent cleanups being impacted by the main ctx cancellation
			// this is a best effort cleanup and query can still timeout and we don't want to wait forever due to blocked calls
			// this is triggered before the view is even created to handle situations
			// where before the client can trigger query cancel the query succeeds and the view is created but the driver stil reports query cancelled
			ctx, cancel := graceful.WithMinimumDuration(ctx, 15*time.Second)
			defer cancel()
			_ = c.Exec(ctx, &drivers.Statement{Query: fmt.Sprintf("DROP VIEW IF EXISTS %s %s", v, onClusterClause)})
		}()
		err := c.Exec(ctx, &drivers.Statement{Query: fmt.Sprintf("CREATE OR REPLACE VIEW %s %s AS %s", v, onClusterClause, sql)})
		if err != nil {
			return err
		}
		// create table with same schema as view
		fmt.Fprintf(&create, " AS %s ", v)
	} else {
		fmt.Fprintf(&create, " %s ", outputProps.Columns)
	}

	tableConfig := outputProps.tblConfig()
	create.WriteString(tableConfig)

	// validate incremental strategy
	if outputProps.IncrementalStrategy == drivers.IncrementalStrategyPartitionOverwrite &&
		!strings.Contains(strings.ToUpper(tableConfig), "PARTITION BY") {
		return fmt.Errorf("clickhouse: incremental strategy partition_overwrite requires a partition key")
	}

	// create table
	err := c.Exec(ctx, &drivers.Statement{Query: create.String(), Priority: 100})
	if err != nil {
		return err
	}

	if c.config.Cluster == "" {
		return nil
	}
	// create the distributed table
	return c.createDistributedTable(ctx, name, outputProps)
}

// createDistributedTable creates a distributed table by name assuming that a table with `name`_local already exists
func (c *Connection) createDistributedTable(ctx context.Context, name string, outputProps *ModelOutputProperties) error {
	if c.config.Cluster == "" {
		return fmt.Errorf("clickhouse: cannot create distributed table without a cluster")
	}
	onClusterClause := c.onClusterClause()

	var distributed strings.Builder
	database := "currentDatabase()"
	if c.config.Database != "" {
		database = safeSQLString(c.config.Database)
	}
	fmt.Fprintf(&distributed, "CREATE OR REPLACE TABLE %s %s AS %s", safeSQLName(name), onClusterClause, safeSQLName(localTableName(name)))
	fmt.Fprintf(&distributed, " ENGINE = Distributed(%s, %s, %s", safeSQLString(c.config.Cluster), database, safeSQLString(localTableName(name)))
	if outputProps.DistributedShardingKey != "" {
		fmt.Fprintf(&distributed, ", %s", outputProps.DistributedShardingKey)
	} else {
		fmt.Fprintf(&distributed, ", rand()")
	}
	distributed.WriteString(")")
	if outputProps.DistributedSettings != "" {
		fmt.Fprintf(&distributed, " SETTINGS %s", outputProps.DistributedSettings)
	}
	return c.Exec(ctx, &drivers.Statement{Query: distributed.String(), Priority: 100})
}

func (c *Connection) createDictionary(ctx context.Context, name, sql string, outputProps *ModelOutputProperties) error {
	onClusterClause := c.onClusterClause()
	if sql == "" {
		if outputProps.Columns == "" {
			return fmt.Errorf("clickhouse: no columns specified for dictionary %q", name)
		}
		return c.Exec(ctx, &drivers.Statement{
			Query:    fmt.Sprintf("CREATE OR REPLACE DICTIONARY %s %s %s %s", safeSQLName(name), onClusterClause, outputProps.Columns, outputProps.EngineFull),
			Priority: 100,
		})
	}

	if outputProps.PrimaryKey == "" {
		return fmt.Errorf("clickhouse: no primary key specified for dictionary %q", name)
	}

	// Note the table the dictionary currently sources from, so it can be dropped once the dictionary is repointed.
	oldSrcTable, err := c.dictionarySourceTable(ctx, name)
	if err != nil {
		return err
	}

	// Write the new data to its own table instead of reusing oldSrcTable, which ClickHouse would not let us replace
	// while the dictionary depends on it. It also means the dictionary keeps serving until the new data is ready.
	srcTable := newDictionarySourceTable(name)
	var repointed bool
	defer func() {
		if repointed {
			return
		}
		ctx, cancel := graceful.WithMinimumDuration(ctx, 15*time.Second)
		defer cancel()

		err := c.dropTable(ctx, srcTable)
		if err != nil && !errors.Is(err, drivers.ErrNotFound) {
			c.logger.Warn("clickhouse: failed to drop dictionary source table", zap.String("name", srcTable), zap.Error(err), observability.ZapCtx(ctx))
		}
	}()

	err = c.createTable(ctx, srcTable, sql, outputProps)
	if err != nil {
		return err
	}
	err = c.Exec(ctx, &drivers.Statement{
		Query:    fmt.Sprintf("INSERT INTO %s %s", safeSQLName(srcTable), sql),
		Priority: 100,
	})
	if err != nil {
		return err
	}

	if outputProps.Columns == "" {
		// infer columns
		outputProps.Columns, err = c.columnClause(ctx, srcTable)
		if err != nil {
			return err
		}
	}

	srcTbl := fmt.Sprintf("CLICKHOUSE(TABLE %s)", drivers.EscapeStringValue(srcTable))
	if outputProps.DictionarySourceUser != "" {
		if outputProps.DictionarySourcePassword == "" {
			return fmt.Errorf("clickhouse: no password specified for dictionary user")
		}
		srcTbl = fmt.Sprintf("CLICKHOUSE(TABLE %s USER %s PASSWORD %s)", drivers.EscapeStringValue(srcTable), safeSQLString(outputProps.DictionarySourceUser), safeSQLString(outputProps.DictionarySourcePassword))
	}

	// create dictionary
	err = c.Exec(ctx, &drivers.Statement{
		Query:    fmt.Sprintf(`CREATE OR REPLACE DICTIONARY %s %s %s PRIMARY KEY %s SOURCE(%s) LAYOUT(HASHED()) LIFETIME(0)`, safeSQLName(name), onClusterClause, outputProps.Columns, outputProps.PrimaryKey, srcTbl),
		Priority: 100,
	})
	if err != nil {
		return err
	}
	repointed = true

	// The dictionary no longer depends on its previous source table, so it can be dropped.
	if oldSrcTable != "" {
		err = c.dropTable(ctx, oldSrcTable)
		if err != nil && !errors.Is(err, drivers.ErrNotFound) {
			c.logger.Warn("clickhouse: failed to drop previous dictionary source table", zap.String("name", oldSrcTable), zap.Error(err), observability.ZapCtx(ctx))
		}
	}
	return nil
}

func (c *Connection) columnClause(ctx context.Context, table string) (string, error) {
	var columnClause strings.Builder
	args := []any{c.config.Database, table}
	if c.config.Database == "" {
		args = []any{nil, table}
	}
	res, err := c.Query(ctx, &drivers.Statement{
		Query:    "SELECT name, type FROM system.columns WHERE database = coalesce(?, currentDatabase()) AND table = ?",
		Args:     args,
		Priority: 100,
	})
	if err != nil {
		return "", err
	}
	defer res.Close()

	columnClause.WriteRune('(')
	var col, typ string
	for res.Next() {
		if err := res.Scan(&col, &typ); err != nil {
			return "", err
		}
		if columnClause.Len() > 1 {
			columnClause.WriteString(", ")
		}
		columnClause.WriteString(safeSQLName(col))
		columnClause.WriteString(" ")
		columnClause.WriteString(typ)
	}
	err = res.Err()
	if err != nil {
		return "", err
	}
	columnClause.WriteRune(')')
	return columnClause.String(), nil
}

func (c *Connection) getTableEngine(ctx context.Context, name string) (string, error) {
	var engine string
	args := []any{c.config.Database, name}
	if c.config.Database == "" {
		args = []any{nil, name}
	}
	res, err := c.Query(ctx, &drivers.Statement{
		Query:    "SELECT engine FROM system.tables WHERE database = coalesce(?, currentDatabase()) AND name = ?",
		Args:     args,
		Priority: 1,
	})
	if err != nil {
		return "", err
	}
	defer res.Close()
	for res.Next() {
		if err := res.Scan(&engine); err != nil {
			return "", err
		}
	}
	err = res.Err()
	if err != nil {
		return "", err
	}
	return engine, nil
}

func (c *Connection) getTablePartitions(ctx context.Context, name string) ([]string, error) {
	var tbl string
	if c.config.Cluster == "" {
		tbl = "system.parts"
	} else {
		// just query all replicas in case data is not fully replicated across the cluster
		tbl = fmt.Sprint("clusterAllReplicas(", safeSQLString(c.config.Cluster), ", system.parts)")
		name = localTableName(name)
	}
	var args []any
	if c.config.Database == "" {
		args = []any{nil, name}
	} else {
		args = []any{c.config.Database, name}
	}
	res, err := c.Query(ctx, &drivers.Statement{
		Query:    fmt.Sprintf("SELECT DISTINCT partition FROM %s WHERE database = coalesce(?, currentDatabase()) AND table = ?", tbl),
		Args:     args,
		Priority: 1,
	})
	if err != nil {
		return nil, err
	}
	defer res.Close()
	// collect partitions
	var partitions []string
	for res.Next() {
		var part string
		if err := res.Scan(&part); err != nil {
			return nil, err
		}
		partitions = append(partitions, part)
	}
	err = res.Err()
	if err != nil {
		return nil, err
	}
	return partitions, nil
}

func (c *Connection) replacePartition(ctx context.Context, src, dest, part string) error {
	onClusterClause := c.onClusterClause()
	if c.config.Cluster != "" {
		dest = localTableName(dest)
		src = localTableName(src)
	}
	return c.Exec(ctx, &drivers.Statement{
		Query:    fmt.Sprintf("ALTER TABLE %s %s REPLACE PARTITION ? FROM %s", safeSQLName(dest), onClusterClause, safeSQLName(src)),
		Args:     []any{part},
		Priority: 1,
	})
}

// syncReplica syncs the local replicated table across the cluster
func (c *Connection) syncReplica(ctx context.Context, tableName string) error {
	if c.config.Cluster == "" || !c.config.SyncReplicas {
		return nil
	}
	// get current database
	database, err := c.currentDatabase(ctx)
	if err != nil {
		return err
	}
	return c.Exec(ctx, &drivers.Statement{
		Query:    fmt.Sprintf("SYSTEM SYNC REPLICA %s %s.%s LIGHTWEIGHT", c.onClusterClause(), safeSQLName(database), safeSQLName(localTableName(tableName))),
		Priority: 1,
	})
}

func (c *Connection) currentDatabase(ctx context.Context) (string, error) {
	if c.config.Database != "" {
		return c.config.Database, nil
	}
	var database string
	rows, err := c.Query(ctx, &drivers.Statement{
		Query:    "SELECT currentDatabase()",
		Priority: 1,
	})
	if err != nil {
		return "", err
	}
	defer rows.Close()
	for rows.Next() {
		if err := rows.Scan(&database); err != nil {
			return "", err
		}
	}
	err = rows.Err()
	if err != nil {
		return "", err
	}
	return database, nil
}

// onClusterClause returns the ON CLUSTER clause for DDL statements, or an empty string if no cluster is configured.
func (c *Connection) onClusterClause() string {
	if c.config.Cluster == "" {
		return ""
	}
	return "ON CLUSTER " + safeSQLName(c.config.Cluster)
}

func isReplicatedEngine(engine string) bool {
	return strings.Contains(strings.ToLower(engine), "replicated")
}

const dictionarySourceTableInfix = "_dict_temp_"

// newDictionarySourceTable returns a unique name for a table for the dictionary `name` to source from.
// The name is unique per call because ClickHouse forbids dropping or renaming a table that a dictionary depends on,
// so refreshing a dictionary has to write to a new table and repoint the dictionary at it.
func newDictionarySourceTable(name string) string {
	// A dictionary is staged by creating it under a temporary name and renaming it into place, but its source table
	// is not renamed along with it. Strip the staging prefix so the source table keeps a readable, stable name.
	name = strings.TrimPrefix(name, stagingTablePrefix)
	return name + dictionarySourceTableInfix + strings.ReplaceAll(uuid.New().String(), "-", "")
}

// dictionarySourceTable returns the table the dictionary `name` currently sources from.
// It returns an empty string if the dictionary does not exist or does not source from a table created by Rill.
func (c *Connection) dictionarySourceTable(ctx context.Context, name string) (string, error) {
	args := []any{c.config.Database, name}
	if c.config.Database == "" {
		args = []any{nil, name}
	}
	res, err := c.Query(ctx, &drivers.Statement{
		Query:    "SELECT loading_dependencies_table FROM system.tables WHERE database = coalesce(?, currentDatabase()) AND name = ?",
		Args:     args,
		Priority: 100,
	})
	if err != nil {
		return "", err
	}
	defer res.Close()

	var deps []string
	for res.Next() {
		if err := res.Scan(&deps); err != nil {
			return "", err
		}
	}
	if err := res.Err(); err != nil {
		return "", err
	}

	for _, dep := range deps {
		if strings.Contains(dep, dictionarySourceTableInfix) {
			return dep, nil
		}
	}
	return "", nil
}

func safeSQLString(name string) string {
	return drivers.EscapeStringValue(name)
}
