package starrocks

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
)

// selfToSelfExecutor executes models where both input and output are StarRocks.
type selfToSelfExecutor struct {
	c *connection
}

var _ drivers.ModelExecutor = &selfToSelfExecutor{}

// Concurrency returns the recommended concurrency for model execution.
func (e *selfToSelfExecutor) Concurrency(desired int) (int, bool) {
	if desired > 1 {
		return desired, true
	}
	return 1, true
}

// Execute runs the model SQL and materializes results in StarRocks.
func (e *selfToSelfExecutor) Execute(ctx context.Context, opts *drivers.ModelExecuteOptions) (*drivers.ModelResult, error) {
	// Parse input and output properties
	inputProps := &ModelInputProperties{}
	if err := mapstructure.WeakDecode(opts.InputProperties, inputProps); err != nil {
		return nil, fmt.Errorf("failed to parse input properties: %w", err)
	}
	outputProps := &ModelOutputProperties{}
	if err := mapstructure.WeakDecode(opts.OutputProperties, outputProps); err != nil {
		return nil, fmt.Errorf("failed to parse output properties: %w", err)
	}

	// Use model name as table name if not specified
	usedModelName := false
	if outputProps.Table == "" {
		outputProps.Table = opts.ModelName
		usedModelName = true
	}

	tableName := outputProps.Table
	asView := strings.EqualFold(outputProps.Materialize, "VIEW")

	start := time.Now()

	if !opts.IncrementalRun {
		// Full refresh: drop and recreate
		stagingTableName := tableName
		if opts.Env.StageChanges {
			stagingTableName = stagingTableNameFor(tableName)
		}

		// Drop staging table/view if exists
		_ = e.c.dropTable(ctx, stagingTableName, asView)

		// Create table/view
		err := e.c.createTableAsSelect(ctx, stagingTableName, inputProps.SQL, asView, outputProps)
		if err != nil {
			_ = e.c.dropTable(ctx, stagingTableName, asView)
			return nil, fmt.Errorf("failed to create model: %w", err)
		}

		// Rename staging table to final table
		if stagingTableName != tableName {
			err = e.c.renameTable(ctx, stagingTableName, tableName, asView)
			if err != nil {
				return nil, fmt.Errorf("failed to rename staged model: %w", err)
			}
		}
	} else {
		// Incremental: insert into existing table
		err := e.c.insertIntoTable(ctx, tableName, inputProps.SQL, outputProps)
		if err != nil {
			return nil, fmt.Errorf("failed to incrementally insert into table: %w", err)
		}
	}

	duration := time.Since(start)

	// Build result properties
	resultProps := &ModelResultProperties{
		Table:         tableName,
		View:          asView,
		UsedModelName: usedModelName,
	}
	resultPropsMap := map[string]interface{}{}
	if err := mapstructure.WeakDecode(resultProps, &resultPropsMap); err != nil {
		return nil, fmt.Errorf("failed to encode result properties: %w", err)
	}

	return &drivers.ModelResult{
		Connector:    opts.OutputConnector,
		Properties:   resultPropsMap,
		Table:        tableName,
		ExecDuration: duration,
	}, nil
}

// ModelInputProperties defines input properties for StarRocks models.
type ModelInputProperties struct {
	SQL string `mapstructure:"sql"`
}

// ModelOutputProperties defines output properties for StarRocks models.
type ModelOutputProperties struct {
	// Table is the output table name.
	Table string `mapstructure:"table"`
	// Materialize can be "TABLE" or "VIEW".
	Materialize string `mapstructure:"materialize"`
	// Engine specifies the StarRocks table engine (e.g., "DUPLICATE", "PRIMARY", "UNIQUE", "AGGREGATE").
	Engine string `mapstructure:"engine"`
	// Keys specifies the key columns for PRIMARY/UNIQUE/AGGREGATE engines.
	Keys string `mapstructure:"keys"`
	// DistributedBy specifies the distribution key.
	DistributedBy string `mapstructure:"distributed_by"`
	// Buckets specifies the number of buckets for distribution.
	Buckets int `mapstructure:"buckets"`
	// Properties specifies additional table properties.
	Properties string `mapstructure:"properties"`
	// IncrementalStrategy for incremental models.
	IncrementalStrategy drivers.IncrementalStrategy `mapstructure:"incremental_strategy"`
}

// ModelResultProperties defines result properties for StarRocks models.
type ModelResultProperties struct {
	Table         string `mapstructure:"table"`
	View          bool   `mapstructure:"view"`
	UsedModelName bool   `mapstructure:"used_model_name"`
}

// stagingTableNameFor returns a staging table name.
func stagingTableNameFor(name string) string {
	return "__rill_staging_" + name
}

// createTableAsSelect creates a table or view from a SELECT statement.
func (c *connection) createTableAsSelect(ctx context.Context, name, sql string, asView bool, props *ModelOutputProperties) error {
	db, err := c.getDB(ctx)
	if err != nil {
		return err
	}

	// Build fully-qualified table name using connector's database
	tableName := safeSQLName(name)
	if c.configProp.Database != "" {
		tableName = safeSQLName(c.configProp.Database) + "." + tableName
	}

	if asView {
		// Create view
		query := fmt.Sprintf("CREATE VIEW %s AS %s", tableName, sql)
		_, err = db.ExecContext(ctx, query)
		return err
	}

	// Create table using CREATE TABLE AS SELECT (CTAS)
	// StarRocks supports CTAS: https://docs.starrocks.io/docs/sql-reference/sql-statements/table_bucket_part_index/CREATE_TABLE_AS_SELECT/
	var builder strings.Builder
	builder.WriteString("CREATE TABLE ")
	builder.WriteString(tableName)

	// Add engine and key configuration
	if props.Engine != "" {
		builder.WriteString(" ENGINE=")
		builder.WriteString(props.Engine)
		if props.Keys != "" {
			builder.WriteString(" KEY(")
			builder.WriteString(props.Keys)
			builder.WriteString(")")
		}
	}

	// Add distribution
	if props.DistributedBy != "" {
		builder.WriteString(" DISTRIBUTED BY HASH(")
		builder.WriteString(props.DistributedBy)
		builder.WriteString(")")
		if props.Buckets > 0 {
			builder.WriteString(fmt.Sprintf(" BUCKETS %d", props.Buckets))
		}
	}

	// Add properties
	if props.Properties != "" {
		builder.WriteString(" PROPERTIES(")
		builder.WriteString(props.Properties)
		builder.WriteString(")")
	}

	builder.WriteString(" AS ")
	builder.WriteString(sql)

	_, err = db.ExecContext(ctx, builder.String())
	return err
}

// dropTable drops a table or view.
func (c *connection) dropTable(ctx context.Context, name string, isView bool) error {
	db, err := c.getDB(ctx)
	if err != nil {
		return err
	}

	// Build fully-qualified table name using connector's database
	tableName := safeSQLName(name)
	if c.configProp.Database != "" {
		tableName = safeSQLName(c.configProp.Database) + "." + tableName
	}

	var query string
	if isView {
		query = fmt.Sprintf("DROP VIEW IF EXISTS %s", tableName)
	} else {
		query = fmt.Sprintf("DROP TABLE IF EXISTS %s", tableName)
	}
	_, err = db.ExecContext(ctx, query)
	return err
}

// renameTable renames a table or recreates a view.
func (c *connection) renameTable(ctx context.Context, oldName, newName string, isView bool) error {
	db, err := c.getDB(ctx)
	if err != nil {
		return err
	}

	// Build fully-qualified table names using connector's database
	oldTableName := safeSQLName(oldName)
	newTableName := safeSQLName(newName)
	if c.configProp.Database != "" {
		oldTableName = safeSQLName(c.configProp.Database) + "." + oldTableName
		newTableName = safeSQLName(c.configProp.Database) + "." + newTableName
	}

	if isView {
		// StarRocks doesn't support RENAME VIEW, so we need to recreate
		// First get the view definition
		var createStmt string
		row := db.QueryRowContext(ctx, fmt.Sprintf("SHOW CREATE VIEW %s", oldTableName))
		var viewName string
		if err := row.Scan(&viewName, &createStmt); err != nil {
			return fmt.Errorf("failed to get view definition: %w", err)
		}

		// Extract the SELECT part from CREATE VIEW statement
		selectIdx := strings.Index(strings.ToUpper(createStmt), " AS ")
		if selectIdx == -1 {
			return fmt.Errorf("failed to parse view definition")
		}
		selectStmt := createStmt[selectIdx+4:]

		// Drop old view
		if _, err := db.ExecContext(ctx, fmt.Sprintf("DROP VIEW IF EXISTS %s", oldTableName)); err != nil {
			return err
		}

		// Drop target view if exists
		if _, err := db.ExecContext(ctx, fmt.Sprintf("DROP VIEW IF EXISTS %s", newTableName)); err != nil {
			return err
		}

		// Create new view
		_, err = db.ExecContext(ctx, fmt.Sprintf("CREATE VIEW %s AS %s", newTableName, selectStmt))
		return err
	}

	// For tables, use ALTER TABLE RENAME
	// First drop target if exists
	if _, err := db.ExecContext(ctx, fmt.Sprintf("DROP TABLE IF EXISTS %s", newTableName)); err != nil {
		return err
	}

	// Rename table
	_, err = db.ExecContext(ctx, fmt.Sprintf("ALTER TABLE %s RENAME %s", oldTableName, safeSQLName(newName)))
	return err
}

// insertIntoTable inserts data into an existing table.
func (c *connection) insertIntoTable(ctx context.Context, name, sql string, props *ModelOutputProperties) error {
	db, err := c.getDB(ctx)
	if err != nil {
		return err
	}

	// Build fully-qualified table name using connector's database
	tableName := safeSQLName(name)
	if c.configProp.Database != "" {
		tableName = safeSQLName(c.configProp.Database) + "." + tableName
	}

	strategy := props.IncrementalStrategy
	if strategy == "" || strategy == drivers.IncrementalStrategyAppend {
		// Append strategy: simple INSERT
		query := fmt.Sprintf("INSERT INTO %s %s", tableName, sql)
		_, err = db.ExecContext(ctx, query)
		return err
	}

	return fmt.Errorf("incremental strategy %q not supported for StarRocks", strategy)
}

// safeSQLName escapes a SQL identifier.
func safeSQLName(name string) string {
	// Use backticks for StarRocks/MySQL compatibility
	return "`" + strings.ReplaceAll(name, "`", "``") + "`"
}
