package clickhouse

import (
	"context"
	"crypto/md5"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/graceful"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.uber.org/zap"
)

const _modelQueryPriority = 100

// tableWriteMetrics reports metrics for an execution that mutates table data.
type tableWriteMetrics struct {
	// duration is the time taken to run user queries only.
	duration time.Duration
}

// createEntity creates a resource (view, table, dictionary) in the database.
func (c *Connection) createEntity(ctx context.Context, name string, inputProps *ModelInputProperties, outputProps *ModelOutputProperties) (*tableWriteMetrics, error) {
	t := time.Now()
	switch outputProps.Typ {
	case "VIEW":
		err := c.createView(ctx, name, inputProps)
		if err != nil {
			return nil, err
		}
	case "TABLE":
		err := c.createTable(ctx, name, inputProps, outputProps)
		if err != nil {
			return nil, err
		}
	case "DICTIONARY":
		err := c.createDictionary(ctx, name, inputProps, outputProps)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("clickhouse: unknown table type %q", outputProps.Typ)
	}
	return &tableWriteMetrics{duration: time.Since(t)}, nil
}

func (c *Connection) createView(ctx context.Context, name string, inputProps *ModelInputProperties) error {
	if inputProps.SQL == "" {
		return fmt.Errorf("clickhouse: no SQL specified for view %q", name)
	}

	var onClusterClause string
	if c.config.Cluster != "" {
		onClusterClause = "ON CLUSTER " + safeSQLName(c.config.Cluster)
	}
	return c.Exec(ctx, &drivers.Statement{
		Query:    fmt.Sprintf("CREATE OR REPLACE VIEW %s %s AS %s", safeSQLName(name), onClusterClause, inputProps.SQL),
		Priority: _modelQueryPriority,
	})
}

func (c *Connection) createTable(ctx context.Context, name string, inputProps *ModelInputProperties, outputProps *ModelOutputProperties) error {
	// On replicated databases `create table t as select * from ...` is prohibited.
	// So we need to create a table first and then insert data into it separately.
	var err error
	if c.config.Cluster == "" {
		err = c.createNonDistributedTable(ctx, name, inputProps, outputProps)
	} else {
		err = c.createDistributedTable(ctx, name, inputProps, outputProps)
	}
	if err != nil {
		return err
	}
	_, err = c.insertTable(ctx, name, inputProps, drivers.IncrementalStrategyAppend)
	return err
}

func (c *Connection) createDistributedTable(ctx context.Context, name string, inputProps *ModelInputProperties, outputProps *ModelOutputProperties) error {
	// Create the underlying local table.
	err := c.createNonDistributedTable(ctx, localTableName(name), inputProps, outputProps)
	if err != nil {
		return err
	}

	// Create the distributed table.
	dbClause := "currentDatabase()"
	if c.config.Database != "" {
		dbClause = safeSQLString(c.config.Database)
	}
	shardingKeyClause := outputProps.DistributedShardingKey
	if shardingKeyClause == "" {
		shardingKeyClause = "rand()"
	}
	var settingsClause string
	if outputProps.DistributedSettings != "" {
		settingsClause = fmt.Sprintf("SETTINGS %s", outputProps.DistributedSettings)
	}
	return c.Exec(ctx, &drivers.Statement{
		Query: fmt.Sprintf(
			"CREATE OR REPLACE TABLE %s %s AS %s ENGINE = Distributed(%s, %s, %s, %s) %s",
			safeSQLName(name),
			c.onClusterClause(),
			safeSQLName(localTableName(name)),
			safeSQLString(c.config.Cluster),
			dbClause,
			safeSQLString(localTableName(name)),
			shardingKeyClause,
			settingsClause,
		),
		Priority: _modelQueryPriority,
	})
}

func (c *Connection) createNonDistributedTable(ctx context.Context, name string, inputProps *ModelInputProperties, outputProps *ModelOutputProperties) error {
	// If an explicit schema is not provided, we attempt to infer it from the SQL query.
	columnsClause := outputProps.Columns
	if columnsClause == "" {
		if inputProps.SQL == "" {
			return fmt.Errorf("clickhouse: no 'sql' or 'output.columns' specified for table %q", name)
		}

		// Create a temporary view for the SQL
		viewName := safeSQLName(fmt.Sprintf("__rill_temp_%s_%x", name, md5.Sum([]byte(inputProps.SQL))))
		defer func() {
			// cleanup using a different ctx to prevent cleanups being impacted by the main ctx cancellation
			// this is a best effort cleanup and query can still timeout and we don't want to wait forever due to blocked calls
			// this is triggered before the view is even created to handle situations
			// where before the client can trigger query cancel the query succeeds and the view is created but the driver stil reports query cancelled
			ctx, cancel := graceful.WithMinimumDuration(ctx, 15*time.Second)
			defer cancel()
			_ = c.Exec(ctx, &drivers.Statement{Query: fmt.Sprintf("DROP VIEW IF EXISTS %s %s", viewName, c.onClusterClause())})
		}()
		err := c.Exec(ctx, &drivers.Statement{Query: fmt.Sprintf("CREATE OR REPLACE VIEW %s %s AS %s", viewName, c.onClusterClause(), inputProps.SQL)})
		if err != nil {
			return err
		}

		// Use "AS <tmp view>" for the columns (we don't need to actually repeat the schema here).
		columnsClause = fmt.Sprintf("AS %s", viewName)
	}

	// Create the table.
	return c.Exec(ctx, &drivers.Statement{
		Query: fmt.Sprintf(
			"CREATE OR REPLACE %s %s %s %s",
			safeSQLName(name),
			c.onClusterClause(),
			columnsClause,
			outputProps.tblConfig(),
		),
		Priority: _modelQueryPriority,
	})
}

func (c *Connection) createDictionary(ctx context.Context, name string, inputProps *ModelInputProperties, outputProps *ModelOutputProperties) error {
	// Handle dictionaries that are NOT based in a SQL query.
	// This enables users to create dictionaries with a custom data source.
	if inputProps.SQL == "" && len(inputProps.InsertSQLs) == 0 {
		if outputProps.EngineFull == "" {
			return fmt.Errorf("clickhouse: no 'sql' query and no 'output.engine_full' specified for dictionary %q", name)
		}
		if outputProps.Columns == "" {
			return fmt.Errorf("clickhouse: missing 'output.columns' config for dictionary %q (columns are required for dictionaries that are not based on a SELECT statement)", name)
		}
		return c.Exec(ctx, &drivers.Statement{
			Query:    fmt.Sprintf("CREATE OR REPLACE DICTIONARY %s %s %s %s", safeSQLName(name), c.onClusterClause(), outputProps.Columns, outputProps.EngineFull),
			Priority: _modelQueryPriority,
		})
	}

	// Dictionaries must have a primary key.
	if outputProps.PrimaryKey == "" {
		return fmt.Errorf("clickhouse: no primary key specified for dictionary %q", name)
	}

	// We need to create an underlying table to store the data for the dictionary.
	err := c.createTable(ctx, dictionaryTableName(name), inputProps, outputProps)
	if err != nil {
		return err
	}

	// Get or infer the schema for the dictionary.
	columns := outputProps.DictionaryColumns
	if columns == "" {
		columns = outputProps.Columns
	}
	if columns == "" {
		columns, err = c.tableColumnsClause(ctx, dictionaryTableName(name))
		if err != nil {
			return err
		}
	}

	// Generate optional username/password clause for the dictionary source.
	var userPasswordClause string
	if outputProps.DictionarySourceUser != "" {
		if outputProps.DictionarySourcePassword == "" {
			return fmt.Errorf("clickhouse: no password specified for dictionary user")
		}
		userPasswordClause = fmt.Sprintf("USER %s PASSWORD %s", c.Dialect().EscapeStringValue(outputProps.DictionarySourceUser), c.Dialect().EscapeStringValue(outputProps.DictionarySourcePassword))
	}

	// Create the dictionary.
	return c.Exec(ctx, &drivers.Statement{
		Query: fmt.Sprintf(
			`CREATE OR REPLACE DICTIONARY %s %s %s PRIMARY KEY %s SOURCE(CLICKHOUSE(TABLE %s %s)) LAYOUT(HASHED()) LIFETIME(0)`,
			safeSQLName(name),
			c.onClusterClause(),
			columns,
			outputProps.PrimaryKey,
			c.Dialect().EscapeStringValue(dictionaryTableName(name)),
			userPasswordClause,
		),
		Priority: _modelQueryPriority,
	})
}

// insertTable inserts data into a table. The table must already exist.
func (c *Connection) insertTable(ctx context.Context, name string, inputProps *ModelInputProperties, strategy drivers.IncrementalStrategy) (*tableWriteMetrics, error) {
	start := time.Now()

	if strategy == drivers.IncrementalStrategyUnspecified || strategy == drivers.IncrementalStrategyAppend {
		var clauses []string
		if inputProps.SQL != "" {
			clauses = append(clauses, inputProps.SQL)
		} else if inputProps.InsertSQLs != nil {
			clauses = append(clauses, inputProps.InsertSQLs...)
		} else {
			return nil, fmt.Errorf("clickhouse: no SQL specified for insert")
		}

		for _, clause := range clauses {
			err := c.Exec(ctx, &drivers.Statement{
				Query:    fmt.Sprintf("INSERT INTO %s %s", safeSQLName(name), clause),
				Priority: _modelQueryPriority,
			})
			if err != nil {
				return nil, err
			}
		}
		return &tableWriteMetrics{duration: time.Since(start)}, nil
	}

	if strategy == drivers.IncrementalStrategyPartitionOverwrite {
		sql := inputProps.SQL
		if sql == "" || len(inputProps.InsertSQLs) != 0 {
			return nil, fmt.Errorf("clickhouse: partition overwrite inserts require a single SQL query")
		}

		// Distributed tables cannot be altered directly, so if it's distributed, we need to alter the local table instead.
		engine, _, err := c.tableEngine(ctx, name)
		if err != nil {
			return nil, err
		}
		if strings.EqualFold(engine, "Distributed") {
			name = localTableName(name)
		}

		// Create a temp table with the same schema using a deterministic name.
		tempName := fmt.Sprintf("__rill_temp_%s_%x", name, md5.Sum([]byte(sql)))
		defer func() {
			// cleanup using a different ctx to prevent cleanups being impacted by the main ctx cancellation
			// this is a best effort cleanup and query can still timeout and we don't want to wait forever due to blocked calls
			// this is triggered before the table is even created to handle situations
			// where before the client can trigger query cancel the query succeeds and the view is created but the driver stil reports query cancelled
			ctx, cancel := graceful.WithMinimumDuration(ctx, 15*time.Second)
			defer cancel()
			err = c.Exec(ctx, &drivers.Statement{
				Query:    fmt.Sprintf("DROP TABLE IF EXISTS %s %s", safeSQLName(tempName), c.onClusterClause()),
				Priority: _modelQueryPriority,
			})
			if err != nil {
				c.logger.Warn("clickhouse: failed to drop temp table", zap.String("name", tempName), zap.Error(err), observability.ZapCtx(ctx))
			}
		}()
		err = c.Exec(ctx, &drivers.Statement{
			Query:    fmt.Sprintf("CREATE OR REPLACE TABLE %s %s AS %s", safeSQLName(tempName), c.onClusterClause(), name),
			Priority: _modelQueryPriority,
		})
		if err != nil {
			return nil, err
		}

		// Insert the partition into the temporary table
		err = c.Exec(ctx, &drivers.Statement{
			Query:    fmt.Sprintf("INSERT INTO %s %s", safeSQLName(tempName), sql),
			Priority: _modelQueryPriority,
		})
		if err != nil {
			return nil, err
		}

		// Iterate over the partitions in the temporary table and replace them in the main table.
		partitions, err := c.tablePartitions(ctx, tempName)
		if err != nil {
			return nil, err
		}
		for _, p := range partitions {
			err = c.Exec(ctx, &drivers.Statement{
				Query:    fmt.Sprintf("ALTER TABLE %s %s REPLACE PARTITION ? FROM %s", safeSQLName(name), c.onClusterClause(), safeSQLName(tempName)),
				Args:     []any{p},
				Priority: _modelQueryPriority,
			})
			if err != nil {
				return nil, err
			}
		}
		return &tableWriteMetrics{duration: time.Since(start)}, nil
	}

	return nil, fmt.Errorf("incremental insert strategy %q not supported", strategy)
}

func (c *Connection) dropEntity(ctx context.Context, name string) error {
	typ, onCluster, err := c.entityType(ctx, name)
	if err != nil {
		return err
	}

	switch typ {
	case "VIEW":
		return c.Exec(ctx, &drivers.Statement{
			Query:    fmt.Sprintf("DROP VIEW IF EXISTS %s %s", safeSQLName(name), c.onClusterClause()),
			Priority: _modelQueryPriority,
		})
	case "TABLE":
		// drop the main table
		err := c.Exec(ctx, &drivers.Statement{
			Query:    fmt.Sprintf("DROP TABLE IF EXISTS %s %s", safeSQLName(name), c.onClusterClause()),
			Priority: _modelQueryPriority,
		})
		if err != nil {
			return err
		}
		// then drop the local table in case of cluster
		if onCluster && !strings.HasSuffix(name, "_local") {
			return c.Exec(ctx, &drivers.Statement{
				Query:    fmt.Sprintf("DROP TABLE %s %s", safeSQLName(localTableName(name)), c.onClusterClause()),
				Priority: _modelQueryPriority,
			})
		}
		return nil
	case "DICTIONARY":
		// first drop the dictionary
		err := c.Exec(ctx, &drivers.Statement{
			Query:    fmt.Sprintf("DROP DICTIONARY IF EXISTS %s %s", safeSQLName(name), c.onClusterClause()),
			Priority: _modelQueryPriority,
		})
		// then drop the underlying table (may not exist for some dictionaries)
		_ = c.Exec(ctx, &drivers.Statement{
			Query:    fmt.Sprintf("DROP TABLE IF EXISTS %s %s", safeSQLName(dictionaryTableName(name)), c.onClusterClause()),
			Priority: _modelQueryPriority,
		})
		return err
	default:
		return fmt.Errorf("clickhouse: unknown entity type %q", typ)
	}
}

func (c *Connection) renameEntity(ctx context.Context, oldName, newName string) error {
	typ, _, err := c.entityType(ctx, oldName)
	if err != nil {
		return err
	}
	switch typ {
	case "VIEW":
		return c.renameView(ctx, oldName, newName)
	case "TABLE":
		return c.renameTable(ctx, oldName, newName)
	case "DICTIONARY":
		return c.renameDictionary(ctx, oldName, newName)
	default:
		return fmt.Errorf("clickhouse: unknown entity type %q", typ)
	}
}

func (c *Connection) renameView(ctx context.Context, oldName, newName string) error {
	// Clickhouse does not support renaming views so we capture the old view's SELECT and use it to create the new view.
	sql, err := c.sqlForView(ctx, oldName)
	if err != nil {
		return err
	}

	// Create the new view.
	err = c.Exec(ctx, &drivers.Statement{
		Query:    fmt.Sprintf("CREATE OR REPLACE VIEW %s %s AS %s", safeSQLName(newName), c.onClusterClause(), sql),
		Priority: _modelQueryPriority,
	})
	if err != nil {
		return err
	}

	// Drop the old view.
	err = c.Exec(ctx, &drivers.Statement{
		Query:    fmt.Sprintf("DROP VIEW %s %s", safeSQLName(oldName), c.onClusterClause()),
		Priority: _modelQueryPriority,
	})
	if err != nil {
		c.logger.Error("clickhouse: failed to drop old view during rename", zap.String("view", oldName), zap.Error(err), observability.ZapCtx(ctx))
	}
	return nil
}

func (c *Connection) renameTable(ctx context.Context, oldName, newName string) error {
	// If the table is not distributed, use renameNonDistributedTable directly.
	engine, engineFull, err := c.tableEngine(ctx, oldName)
	if err != nil {
		return err
	}
	if !strings.EqualFold(engine, "Distributed") {
		return c.renameNonDistributedTable(ctx, oldName, newName)
	}

	// The table is distributed, which means there's an underlying local table that we need to rename as well.

	// Rename the local table.
	err = c.renameNonDistributedTable(ctx, localTableName(oldName), localTableName(newName))
	if err != nil {
		return err
	}

	// Recreate the distributed table for the new local table name.
	// Somewhat hackily, we just replace the old local table name with the new one in the engine expression.
	engineFull = strings.ReplaceAll(engineFull, localTableName(oldName), localTableName(newName)) // TODO: safeSQLName it?
	err = c.Exec(ctx, &drivers.Statement{
		Query:    fmt.Sprintf("CREATE OR REPLACE TABLE %s %s AS %s Engine = %s", safeSQLName(newName), c.onClusterClause(), safeSQLName(localTableName(newName)), engineFull),
		Priority: _modelQueryPriority,
	})
	if err != nil {
		return err
	}

	// Drop the old table
	return c.Exec(ctx, &drivers.Statement{
		Query:    fmt.Sprintf("DROP TABLE %s %s", safeSQLName(oldName), c.onClusterClause()),
		Priority: _modelQueryPriority,
	})
}

func (c *Connection) renameNonDistributedTable(ctx context.Context, oldName, newName string) error {
	var exists bool
	err := c.db.QueryRowContext(ctx, fmt.Sprintf("EXISTS %s", safeSQLName(newName))).Scan(&exists)
	if err != nil {
		return err
	}
	if !exists {
		return c.Exec(ctx, &drivers.Statement{
			Query:    fmt.Sprintf("RENAME TABLE %s TO %s %s", safeSQLName(oldName), safeSQLName(newName), c.onClusterClause()),
			Priority: _modelQueryPriority,
		})
	}
	err = c.Exec(ctx, &drivers.Statement{
		Query:    fmt.Sprintf("EXCHANGE TABLES %s AND %s %s", safeSQLName(oldName), safeSQLName(newName), c.onClusterClause()),
		Priority: _modelQueryPriority,
	})
	if err != nil {
		return err
	}
	// Drop the old table
	err = c.dropEntity(ctx, oldName)
	if err != nil {
		c.logger.Error("clickhouse: failed to drop old table during rename", zap.String("table", oldName), zap.Error(err), observability.ZapCtx(ctx))
	}
	return nil
}

func (c *Connection) renameDictionary(ctx context.Context, oldName, newName string) error {
	// Rename the dictionary's underlying table.
	err := c.renameTable(ctx, dictionaryTableName(oldName), dictionaryTableName(newName))
	if err != nil && !errors.Is(err, drivers.ErrNotFound) {
		return err
	}

	// Rename the dictionary itself.
	// TODO: Tricky, we can't easily swap the table name in the source.
	return nil
}

func (c *Connection) tableColumnsClause(ctx context.Context, table string) (string, error) {
	var db any
	if c.config.Database != "" {
		db = c.config.Database
	}

	res, err := c.Query(ctx, &drivers.Statement{
		Query:    "SELECT name, type FROM system.columns WHERE database = coalesce(?, currentDatabase()) AND table = ?",
		Args:     []any{db, table},
		Priority: _modelQueryPriority,
	})
	if err != nil {
		return "", err
	}
	defer res.Close()

	var clause strings.Builder
	clause.WriteRune('(')
	for res.Next() {
		var col, typ string
		if err := res.Scan(&col, &typ); err != nil {
			return "", err
		}
		if clause.Len() > 1 {
			clause.WriteString(", ")
		}
		clause.WriteString(safeSQLName(col))
		clause.WriteString(" ")
		clause.WriteString(typ)
	}
	err = res.Err()
	if err != nil {
		return "", err
	}
	clause.WriteRune(')')
	return clause.String(), nil
}

func (c *Connection) tableEngine(ctx context.Context, name string) (string, string, error) {
	var db any
	if c.config.Database != "" {
		db = c.config.Database
	}

	res, err := c.Query(ctx, &drivers.Statement{
		Query:    "SELECT engine, engine_full FROM system.tables WHERE database = coalesce(?, currentDatabase()) AND name = ?",
		Args:     []any{db, name},
		Priority: _modelQueryPriority,
	})
	if err != nil {
		return "", "", err
	}
	defer res.Close()

	var found bool
	var engine, engineFull string
	if res.Next() {
		if err := res.Scan(&engine, &engineFull); err != nil {
			return "", "", err
		}
		found = true
	}
	err = res.Err()
	if err != nil {
		return "", "", err
	}
	if !found {
		return "", "", drivers.ErrNotFound
	}
	return engine, engineFull, nil
}

func (c *Connection) tablePartitions(ctx context.Context, name string) ([]string, error) {
	res, err := c.Query(ctx, &drivers.Statement{
		Query:    "SELECT DISTINCT partition FROM system.parts WHERE table = ?",
		Args:     []any{name},
		Priority: _modelQueryPriority,
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

func (c *Connection) entityType(ctx context.Context, name string) (typ string, onCluster bool, err error) {
	conn, release, err := c.acquireMetaConn(ctx)
	if err != nil {
		return "", false, err
	}
	defer func() { _ = release() }()

	var q string
	if c.config.Cluster == "" {
		q = `SELECT
    			multiIf(engine IN ('MaterializedView', 'View'), 'VIEW', engine = 'Dictionary', 'DICTIONARY', 'TABLE') AS type,
    			0 AS is_on_cluster
			FROM system.tables AS t
			JOIN system.databases AS db ON t.database = db.name
			WHERE t.database = coalesce(?, currentDatabase()) AND t.name = ?`
	} else {
		q = `SELECT
    			multiIf(engine IN ('MaterializedView', 'View'), 'VIEW', engine = 'Dictionary', 'DICTIONARY', 'TABLE') AS type,
    			countDistinct(_shard_num) > 1 AS is_on_cluster
			FROM clusterAllReplicas(` + safeSQLName(c.config.Cluster) + `, system.tables) AS t
			JOIN system.databases AS db ON t.database = db.name
			WHERE t.database = coalesce(?, currentDatabase()) AND t.name = ?
			GROUP BY engine, t.name`
	}

	var args []any
	if c.config.Database == "" {
		args = []any{nil, name}
	} else {
		args = []any{c.config.Database, name}
	}

	ctx = contextWithQueryID(ctx)
	row := conn.QueryRowxContext(ctx, q, args...)
	err = row.Scan(&typ, &onCluster)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", false, drivers.ErrNotFound
		}
		return "", false, err
	}

	switch typ {
	case "VIEW", "TABLE", "DICTIONARY":
		// Valid
	default:
		return "", false, fmt.Errorf("clickhouse: unknown entity type %q", typ)
	}

	return typ, onCluster, nil
}

func (c *Connection) sqlForView(ctx context.Context, name string) (string, error) {
	var db any
	if c.config.Database != "" {
		db = c.config.Database
	}

	res, err := c.Query(ctx, &drivers.Statement{
		Query:    "SELECT as_select FROM system.tables WHERE database = coalesce(?, currentDatabase()) AND name = ?",
		Args:     []any{db, name},
		Priority: _modelQueryPriority,
	})
	if err != nil {
		return "", err
	}
	defer res.Close()

	var sql string
	if res.Next() {
		if err := res.Scan(&sql); err != nil {
			return "", err
		}
	}
	err = res.Err()
	if err != nil {
		return "", err
	}

	if sql == "" {
		return "", fmt.Errorf("clickhouse: no SQL found for view %q", name)
	}

	return sql, nil
}

func (c *Connection) onClusterClause() string {
	if c.config.Cluster != "" {
		return "ON CLUSTER " + safeSQLName(c.config.Cluster)
	}
	return ""
}

// localTableName returns the underlying local table name for distributed tables.
func localTableName(name string) string {
	return name + "_local"
}

// dictionaryTableName returns the underlying table name for dictionaries created from SQL queries.
func dictionaryTableName(name string) string {
	return name + "_dict_temp_"
}

func safeSQLString(name string) string {
	return drivers.DialectClickHouse.EscapeStringValue(name)
}

func safeSQLName(name string) string {
	return drivers.DialectClickHouse.EscapeIdentifier(name)
}
