package clickhouse

import (
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"strings"
	"time"

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

func (c *Connection) createTableAsSelect(ctx context.Context, name, sql string, outputProps *ModelOutputProperties) (*tableWriteMetrics, error) {
	ctx = contextWithQueryID(ctx)

	var onClusterClause string
	if c.config.Cluster != "" {
		onClusterClause = "ON CLUSTER " + safeSQLName(c.config.Cluster)
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
		return &tableWriteMetrics{duration: time.Since(t)}, nil
	} else if outputProps.Typ == "DICTIONARY" {
		err := c.createDictionary(ctx, name, sql, outputProps)
		if err != nil {
			return nil, err
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
	return &tableWriteMetrics{duration: time.Since(t)}, nil
}

type InsertTableOptions struct {
	Strategy drivers.IncrementalStrategy
}

func (c *Connection) insertTableAsSelect(ctx context.Context, name, sql string, opts *InsertTableOptions, outputProps *ModelOutputProperties) (*tableWriteMetrics, error) {
	ctx = contextWithQueryID(ctx)
	start := time.Now()

	if opts.Strategy == drivers.IncrementalStrategyAppend {
		err := c.Exec(ctx, &drivers.Statement{
			Query:    fmt.Sprintf("INSERT INTO %s %s", safeSQLName(name), sql),
			Priority: 1,
		})
		if err != nil {
			return nil, err
		}
		return &tableWriteMetrics{duration: time.Since(start)}, nil
	}

	if opts.Strategy == drivers.IncrementalStrategyPartitionOverwrite {
		_, onCluster, err := c.entityType(ctx, c.config.Database, name)
		if err != nil {
			return nil, err
		}
		onClusterClause := ""
		if onCluster {
			onClusterClause = "ON CLUSTER " + safeSQLName(c.config.Cluster)
		}
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

		// run 'OPTIMIZE' before partition replacement if configured
		if c.config.OptimizeTemporaryTablesBeforePartitionReplace {
			err = c.optimizeTable(ctx, tempName)
			if err != nil {
				c.logger.Warn("clickhouse: failed to optimize temporary table", zap.String("name", tempName), zap.Error(err), observability.ZapCtx(ctx))
				// Don't fail the entire operation if optimize fails - just log and continue
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
		return &tableWriteMetrics{duration: time.Since(start)}, nil
	}

	if opts.Strategy == drivers.IncrementalStrategyMerge {
		_, onCluster, err := c.entityType(ctx, c.config.Database, name)
		if err != nil {
			return nil, err
		}
		onClusterClause := ""
		if onCluster {
			onClusterClause = "ON CLUSTER " + safeSQLName(c.config.Cluster)
		}
		// get the engine info of the given table
		engine, err := c.getTableEngine(ctx, name)
		if err != nil {
			return nil, err
		}
		if !strings.Contains(engine, "ReplacingMergeTree") {
			return nil, fmt.Errorf("clickhouse: merge strategy requires ReplacingMergeTree engine")
		}

		// insert into table using the merge strategy
		err = c.Exec(ctx, &drivers.Statement{
			Query:    fmt.Sprintf("INSERT INTO %s %s %s", safeSQLName(name), onClusterClause, sql),
			Priority: 1,
		})
		if err != nil {
			return nil, err
		}
		return &tableWriteMetrics{duration: time.Since(start)}, nil
	}

	return nil, fmt.Errorf("incremental insert strategy %q not supported", opts.Strategy)
}

func (c *Connection) dropTable(ctx context.Context, name string) error {
	ctx = contextWithQueryID(ctx)
	typ, onCluster, err := c.entityType(ctx, c.config.Database, name)
	if err != nil {
		return err
	}
	var onClusterClause string
	if onCluster {
		onClusterClause = "ON CLUSTER " + safeSQLName(c.config.Cluster)
	}
	switch typ {
	case "VIEW":
		return c.Exec(ctx, &drivers.Statement{
			Query:    fmt.Sprintf("DROP VIEW %s %s", safeSQLName(name), onClusterClause),
			Priority: 100,
		})
	case "DICTIONARY":
		// first drop the dictionary
		err := c.Exec(ctx, &drivers.Statement{
			Query:    fmt.Sprintf("DROP DICTIONARY %s %s", safeSQLName(name), onClusterClause),
			Priority: 100,
		})
		// then drop the temp table
		_ = c.Exec(ctx, &drivers.Statement{
			Query:    fmt.Sprintf("DROP TABLE %s %s", safeSQLName(tempTableForDictionary(name)), onClusterClause),
			Priority: 100,
		})
		return err
	case "TABLE":
		// drop the main table
		err := c.Exec(ctx, &drivers.Statement{
			Query:    fmt.Sprintf("DROP TABLE %s %s", safeSQLName(name), onClusterClause),
			Priority: 100,
		})
		if err != nil {
			return err
		}
		// then drop the local table in case of cluster
		if onCluster && !strings.HasSuffix(name, "_local") {
			return c.Exec(ctx, &drivers.Statement{
				Query:    fmt.Sprintf("DROP TABLE %s %s", safeSQLName(localTableName(name)), onClusterClause),
				Priority: 100,
			})
		}
		return nil
	default:
		return fmt.Errorf("clickhouse: unknown entity type %q", typ)
	}
}

func (c *Connection) renameEntity(ctx context.Context, oldName, newName string) error {
	ctx = contextWithQueryID(ctx)
	typ, onCluster, err := c.entityType(ctx, c.config.Database, oldName)
	if err != nil {
		return err
	}
	var onClusterClause string
	if onCluster {
		onClusterClause = "ON CLUSTER " + safeSQLName(c.config.Cluster)
	}

	switch typ {
	case "VIEW":
		return c.renameView(ctx, oldName, newName, onClusterClause)
	case "DICTIONARY":
		return c.renameTable(ctx, oldName, newName, onClusterClause)
	case "TABLE":
		if !onCluster {
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
	err := c.writeDB.QueryRowContext(ctx, fmt.Sprintf("EXISTS %s", safeSQLName(newName))).Scan(&exists)
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
	var onClusterClause string
	if c.config.Cluster != "" {
		onClusterClause = "ON CLUSTER " + safeSQLName(c.config.Cluster)
	}
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
	onClusterClause := "ON CLUSTER " + safeSQLName(c.config.Cluster)

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
	var onClusterClause string
	if c.config.Cluster != "" {
		onClusterClause = "ON CLUSTER " + safeSQLName(c.config.Cluster)
	}
	if sql == "" {
		if outputProps.Columns == "" {
			return fmt.Errorf("clickhouse: no columns specified for dictionary %q", name)
		}
		return c.Exec(ctx, &drivers.Statement{
			Query:    fmt.Sprintf("CREATE OR REPLACE DICTIONARY %s %s %s %s", safeSQLName(name), onClusterClause, outputProps.Columns, outputProps.EngineFull),
			Priority: 100,
		})
	}

	// create a temp table first
	// NOTE :: this can only be dropped when the dictionary is dropped
	tempTable := tempTableForDictionary(name)
	err := c.createTable(ctx, tempTable, sql, outputProps)
	if err != nil {
		return err
	}
	err = c.Exec(ctx, &drivers.Statement{
		Query:    fmt.Sprintf("INSERT INTO %s %s", safeSQLName(tempTable), sql),
		Priority: 100,
	})
	if err != nil {
		return err
	}

	if outputProps.Columns == "" {
		// infer columns
		outputProps.Columns, err = c.columnClause(ctx, tempTable)
		if err != nil {
			return err
		}
	}

	if outputProps.PrimaryKey == "" {
		return fmt.Errorf("clickhouse: no primary key specified for dictionary %q", name)
	}

	srcTbl := fmt.Sprintf("CLICKHOUSE(TABLE %s)", c.Dialect().EscapeStringValue(tempTable))
	if outputProps.DictionarySourceUser != "" {
		if outputProps.DictionarySourcePassword == "" {
			return fmt.Errorf("clickhouse: no password specified for dictionary user")
		}
		srcTbl = fmt.Sprintf("CLICKHOUSE(TABLE %s USER %s PASSWORD %s)", c.Dialect().EscapeStringValue(tempTable), safeSQLString(outputProps.DictionarySourceUser), safeSQLString(outputProps.DictionarySourcePassword))
	}

	// create dictionary
	return c.Exec(ctx, &drivers.Statement{
		Query:    fmt.Sprintf(`CREATE OR REPLACE DICTIONARY %s %s %s PRIMARY KEY %s SOURCE(%s) LAYOUT(HASHED()) LIFETIME(0)`, safeSQLName(name), onClusterClause, outputProps.Columns, outputProps.PrimaryKey, srcTbl),
		Priority: 100,
	})
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
	if res.Next() {
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
		tbl = fmt.Sprint("cluster(", safeSQLString(c.config.Cluster), ", system.parts)")
		name = localTableName(name)
	}
	res, err := c.Query(ctx, &drivers.Statement{
		Query:    fmt.Sprintf("SELECT DISTINCT partition FROM %s WHERE table = ?", tbl),
		Args:     []any{name},
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
	var onClusterClause string
	if c.config.Cluster != "" {
		onClusterClause = "ON CLUSTER " + safeSQLName(c.config.Cluster)
		dest = localTableName(dest)
		src = localTableName(src)
	}
	return c.Exec(ctx, &drivers.Statement{
		Query:    fmt.Sprintf("ALTER TABLE %s %s REPLACE PARTITION ? FROM %s", safeSQLName(dest), onClusterClause, safeSQLName(src)),
		Args:     []any{part},
		Priority: 1,
	})
}

func (c *Connection) optimizeTable(ctx context.Context, tableName string) error {
	var onClusterClause string
	if c.config.Cluster != "" {
		onClusterClause = "ON CLUSTER " + safeSQLName(c.config.Cluster)
		// For clustered tables, optimize the local table
		tableName = localTableName(tableName)
	}

	return c.Exec(ctx, &drivers.Statement{
		Query:    fmt.Sprintf("OPTIMIZE TABLE %s %s", safeSQLName(tableName), onClusterClause),
		Priority: 1,
	})
}

func localTableName(name string) string {
	return name + "_local"
}

func tempTableForDictionary(name string) string {
	return name + "_dict_temp_"
}

func safeSQLString(name string) string {
	return drivers.DialectClickHouse.EscapeStringValue(name)
}
