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

	// Validate and apply defaults
	if err := validateAndApplyDefaults(opts, outputProps); err != nil {
		return nil, fmt.Errorf("invalid model properties: %w", err)
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

		// Drop staging table/view if exists (try both types in case type changed)
		_ = e.c.dropTableOrView(ctx, stagingTableName)

		// Create table/view
		err := e.c.createTableAsSelect(ctx, stagingTableName, inputProps.SQL, asView, outputProps)
		if err != nil {
			_ = e.c.dropTableOrView(ctx, stagingTableName)
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

	// Build result properties with catalog/database info for downstream propagation
	catalog := e.c.configProp.Catalog
	if catalog == "" {
		catalog = defaultCatalog
	}
	resultProps := &ModelResultProperties{
		Catalog:       catalog,
		Database:      e.c.configProp.Database,
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
// Reference: https://docs.starrocks.io/docs/sql-reference/sql-statements/table_bucket_part_index/CREATE_TABLE/
type ModelOutputProperties struct {
	// Table is the output table name.
	Table string `mapstructure:"table"`
	// Materialize can be "TABLE" or "VIEW".
	Materialize string `mapstructure:"materialize"`

	// === Table Model (Key Type) ===
	// Engine specifies the StarRocks table model:
	// - "DUPLICATE" (default): Duplicate Key model, stores all rows
	// - "AGGREGATE": Aggregate Key model, pre-aggregates data
	// - "UNIQUE": Unique Key model, deduplicates by key (Replace semantics)
	// - "PRIMARY": Primary Key model, supports real-time updates (Merge-on-Read)
	Engine string `mapstructure:"engine"`
	// Keys specifies the key columns for the table model.
	// Required for AGGREGATE/UNIQUE/PRIMARY, optional for DUPLICATE.
	Keys string `mapstructure:"keys"`
	// OrderBy specifies sort key columns (only for PRIMARY KEY model when different from keys).
	OrderBy string `mapstructure:"order_by"`

	// === Distribution ===
	// DistributedBy specifies the hash distribution key columns.
	// If empty, uses random distribution.
	DistributedBy string `mapstructure:"distributed_by"`
	// Buckets specifies the number of tablets (buckets) for distribution.
	// Default: automatically determined by StarRocks based on data size.
	Buckets int `mapstructure:"buckets"`

	// === Partitioning ===
	// PartitionBy specifies range/list partition expression.
	// Example: "RANGE(event_date)" or "LIST(city)"
	PartitionBy string `mapstructure:"partition_by"`
	// Partitions defines the partition values.
	// Example: "(PARTITION p1 VALUES LESS THAN ('2024-01-01'), PARTITION p2 VALUES LESS THAN ('2024-02-01'))"
	Partitions string `mapstructure:"partitions"`

	// === Table Properties ===
	// ReplicationNum sets the number of replicas (default: 3).
	ReplicationNum int `mapstructure:"replication_num"`
	// Properties specifies additional table properties in SQL format.
	// Example: "\"enable_persistent_index\" = \"true\", \"bloom_filter_columns\" = \"col1,col2\""
	Properties string `mapstructure:"properties"`
	// Comment adds a table comment.
	Comment string `mapstructure:"comment"`

	// === Incremental Processing ===
	// IncrementalStrategy for incremental models: "append" (default), "merge".
	IncrementalStrategy drivers.IncrementalStrategy `mapstructure:"incremental_strategy"`
}

// ModelResultProperties defines result properties for StarRocks models.
// StarRocks mapping: Catalog = Rill database, Database = Rill databaseSchema
type ModelResultProperties struct {
	Catalog       string `mapstructure:"catalog"`        // StarRocks catalog (maps to Rill database)
	Database      string `mapstructure:"database"`       // StarRocks database (maps to Rill databaseSchema)
	Table         string `mapstructure:"table"`
	View          bool   `mapstructure:"view"`
	UsedModelName bool   `mapstructure:"used_model_name"`
}

// starrocksToSelfExecutor executes models where input is from a different StarRocks connector
// (e.g., external catalog) and output is to this StarRocks connector (e.g., default catalog).
type starrocksToSelfExecutor struct {
	inputConn  *connection // Input connector (e.g., external catalog for reading)
	outputConn *connection // Output connector (e.g., default catalog for writing)
}

var _ drivers.ModelExecutor = &starrocksToSelfExecutor{}

// Concurrency returns the recommended concurrency for model execution.
func (e *starrocksToSelfExecutor) Concurrency(desired int) (int, bool) {
	if desired > 1 {
		return desired, true
	}
	return 1, true
}

// Execute runs the model SQL from input connector and materializes results in output connector.
// Uses input connector's catalog/database context for SELECT, output connector for CREATE.
func (e *starrocksToSelfExecutor) Execute(ctx context.Context, opts *drivers.ModelExecuteOptions) (*drivers.ModelResult, error) {
	// Parse input and output properties
	inputProps := &ModelInputProperties{}
	if err := mapstructure.WeakDecode(opts.InputProperties, inputProps); err != nil {
		return nil, fmt.Errorf("failed to parse input properties: %w", err)
	}
	outputProps := &ModelOutputProperties{}
	if err := mapstructure.WeakDecode(opts.OutputProperties, outputProps); err != nil {
		return nil, fmt.Errorf("failed to parse output properties: %w", err)
	}

	// Validate and apply defaults
	if err := validateAndApplyDefaults(opts, outputProps); err != nil {
		return nil, fmt.Errorf("invalid model properties: %w", err)
	}

	// Use model name as table name if not specified
	usedModelName := false
	if outputProps.Table == "" {
		outputProps.Table = opts.ModelName
		usedModelName = true
	}

	tableName := outputProps.Table
	asView := strings.EqualFold(outputProps.Materialize, "VIEW")

	// Views are not supported for cross-connector execution
	if asView {
		return nil, fmt.Errorf("VIEW materialization is not supported for cross-connector execution; use TABLE instead")
	}

	start := time.Now()

	// Determine output catalog/database
	outputCatalog := e.outputConn.configProp.Catalog
	if outputCatalog == "" {
		outputCatalog = defaultCatalog
	}
	outputDB := e.outputConn.configProp.Database

	// Execute the cross-connector model
	if !opts.IncrementalRun {
		// Full refresh: drop and recreate
		stagingTableName := tableName
		if opts.Env.StageChanges {
			stagingTableName = stagingTableNameFor(tableName)
		}

		// Drop staging table if exists (using output connector)
		_ = e.outputConn.dropTableOrView(ctx, stagingTableName)

		// Create table using cross-connector execution
		err := e.createTableAsSelectCrossConnector(ctx, stagingTableName, inputProps.SQL, outputProps, outputCatalog, outputDB)
		if err != nil {
			_ = e.outputConn.dropTableOrView(ctx, stagingTableName)
			return nil, fmt.Errorf("failed to create model: %w", err)
		}

		// Rename staging table to final table
		if stagingTableName != tableName {
			err = e.outputConn.renameTable(ctx, stagingTableName, tableName, false)
			if err != nil {
				return nil, fmt.Errorf("failed to rename staged model: %w", err)
			}
		}
	} else {
		// Incremental: insert into existing table
		err := e.insertIntoTableCrossConnector(ctx, tableName, inputProps.SQL, outputProps, outputCatalog, outputDB)
		if err != nil {
			return nil, fmt.Errorf("failed to incrementally insert into table: %w", err)
		}
	}

	duration := time.Since(start)

	// Build result properties with output catalog/database info
	resultProps := &ModelResultProperties{
		Catalog:       outputCatalog,
		Database:      outputDB,
		Table:         tableName,
		View:          false,
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

// createTableAsSelectCrossConnector creates a table from a SELECT using cross-connector execution.
// Sets input connector's catalog/database context for SELECT, uses fully qualified name for CREATE.
func (e *starrocksToSelfExecutor) createTableAsSelectCrossConnector(ctx context.Context, name, sql string, props *ModelOutputProperties, outputCatalog, outputDB string) error {
	// Use output connector's database connection
	db, err := e.outputConn.getDB(ctx)
	if err != nil {
		return err
	}

	conn, err := db.Connx(ctx)
	if err != nil {
		return fmt.Errorf("create connection: %w", err)
	}
	defer conn.Close()

	// Set INPUT connector's catalog/database context for SELECT
	// This allows unqualified table references in SQL to resolve to input catalog
	if err := switchCatalogContext(ctx, conn, e.inputConn.configProp.Catalog, e.inputConn.configProp.Database); err != nil {
		return fmt.Errorf("switch to input catalog context: %w", err)
	}

	// Build fully qualified OUTPUT table name: catalog.database.table
	fullTableName := safeSQLName(outputCatalog)
	if outputDB != "" {
		fullTableName += "." + safeSQLName(outputDB)
	}
	fullTableName += "." + safeSQLName(name)

	// Build CTAS query
	var builder strings.Builder
	builder.WriteString("CREATE TABLE ")
	builder.WriteString(fullTableName)

	// Build table configuration
	tableConfig := props.tblConfig()
	if tableConfig != "" {
		builder.WriteString(" ")
		builder.WriteString(tableConfig)
	}

	builder.WriteString(" AS ")
	builder.WriteString(sql)

	_, err = conn.ExecContext(ctx, builder.String())
	return err
}

// insertIntoTableCrossConnector inserts data into an existing table using cross-connector execution.
func (e *starrocksToSelfExecutor) insertIntoTableCrossConnector(ctx context.Context, name, sql string, props *ModelOutputProperties, outputCatalog, outputDB string) error {
	// Use output connector's database connection
	db, err := e.outputConn.getDB(ctx)
	if err != nil {
		return err
	}

	conn, err := db.Connx(ctx)
	if err != nil {
		return fmt.Errorf("create connection: %w", err)
	}
	defer conn.Close()

	// Set INPUT connector's catalog/database context for SELECT
	if err := switchCatalogContext(ctx, conn, e.inputConn.configProp.Catalog, e.inputConn.configProp.Database); err != nil {
		return fmt.Errorf("switch to input catalog context: %w", err)
	}

	// Build fully qualified OUTPUT table name
	fullTableName := safeSQLName(outputCatalog)
	if outputDB != "" {
		fullTableName += "." + safeSQLName(outputDB)
	}
	fullTableName += "." + safeSQLName(name)

	strategy := props.IncrementalStrategy
	if strategy == "" || strategy == drivers.IncrementalStrategyAppend {
		query := fmt.Sprintf("INSERT INTO %s %s", fullTableName, sql)
		_, err = conn.ExecContext(ctx, query)
		return err
	}

	return fmt.Errorf("incremental strategy %q not supported for cross-connector StarRocks execution", strategy)
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

	// Use a dedicated connection to ensure catalog/database context is maintained
	conn, err := db.Connx(ctx)
	if err != nil {
		return fmt.Errorf("create connection: %w", err)
	}
	defer conn.Close()

	// Set catalog and database context
	if err := c.setCatalogContext(ctx, conn); err != nil {
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
		_, err = conn.ExecContext(ctx, query)
		return err
	}

	// Create table using CREATE TABLE AS SELECT (CTAS)
	// StarRocks CTAS: https://docs.starrocks.io/docs/sql-reference/sql-statements/table_bucket_part_index/CREATE_TABLE_AS_SELECT/
	var builder strings.Builder
	builder.WriteString("CREATE TABLE ")
	builder.WriteString(tableName)

	// Build table configuration
	tableConfig := props.tblConfig()
	if tableConfig != "" {
		builder.WriteString(" ")
		builder.WriteString(tableConfig)
	}

	builder.WriteString(" AS ")
	builder.WriteString(sql)

	_, err = conn.ExecContext(ctx, builder.String())
	return err
}

// validateAndApplyDefaults validates and applies defaults to model properties.
func validateAndApplyDefaults(opts *drivers.ModelExecuteOptions, props *ModelOutputProperties) error {
	// Normalize materialize
	props.Materialize = strings.ToUpper(props.Materialize)

	// Incremental models must be tables
	if opts.Incremental {
		if props.Materialize == "VIEW" {
			return fmt.Errorf("incremental models must be materialized as TABLE, not VIEW")
		}
		props.Materialize = "TABLE"
	}

	// Default materialize to TABLE if not specified
	if props.Materialize == "" {
		if opts.Env.DefaultMaterialize {
			props.Materialize = "TABLE"
		} else {
			props.Materialize = "VIEW"
		}
	}

	// Validate materialize value
	if props.Materialize != "TABLE" && props.Materialize != "VIEW" {
		return fmt.Errorf("materialize must be TABLE or VIEW, got %q", props.Materialize)
	}

	// Normalize engine
	props.Engine = strings.ToUpper(props.Engine)

	// Validate engine if specified
	validEngines := map[string]bool{
		"":          true,
		"DUPLICATE": true,
		"DUP":       true,
		"AGGREGATE": true,
		"AGG":       true,
		"UNIQUE":    true,
		"PRIMARY":   true,
	}
	if !validEngines[props.Engine] {
		return fmt.Errorf("invalid engine %q, must be one of: DUPLICATE, AGGREGATE, UNIQUE, PRIMARY", props.Engine)
	}

	// AGGREGATE, UNIQUE, PRIMARY require keys
	if (props.Engine == "AGGREGATE" || props.Engine == "AGG" ||
		props.Engine == "UNIQUE" || props.Engine == "PRIMARY") && props.Keys == "" {
		return fmt.Errorf("%s KEY model requires keys to be specified", props.Engine)
	}

	// ORDER BY only makes sense for PRIMARY KEY model
	if props.OrderBy != "" && props.Engine != "PRIMARY" {
		return fmt.Errorf("order_by is only applicable for PRIMARY KEY model")
	}

	// Validate incremental strategy
	if props.IncrementalStrategy != "" &&
		props.IncrementalStrategy != drivers.IncrementalStrategyAppend &&
		props.IncrementalStrategy != drivers.IncrementalStrategyMerge {
		return fmt.Errorf("unsupported incremental strategy %q, use 'append' or 'merge'", props.IncrementalStrategy)
	}

	// Merge strategy requires PRIMARY KEY or UNIQUE KEY model
	if props.IncrementalStrategy == drivers.IncrementalStrategyMerge {
		if props.Engine != "PRIMARY" && props.Engine != "UNIQUE" {
			return fmt.Errorf("merge incremental strategy requires PRIMARY or UNIQUE KEY model")
		}
	}

	return nil
}

// tblConfig builds the table configuration SQL from ModelOutputProperties.
// Reference: https://docs.starrocks.io/docs/sql-reference/sql-statements/table_bucket_part_index/CREATE_TABLE/
func (props *ModelOutputProperties) tblConfig() string {
	var sb strings.Builder

	// 1. Table Model (Key Type)
	// Format: [ENGINE = key_type] [KEY (column1, column2, ...)]
	engine := strings.ToUpper(props.Engine)
	if engine != "" {
		switch engine {
		case "DUPLICATE", "DUP":
			if props.Keys != "" {
				fmt.Fprintf(&sb, "DUPLICATE KEY(%s)", props.Keys)
			}
		case "AGGREGATE", "AGG":
			if props.Keys != "" {
				fmt.Fprintf(&sb, "AGGREGATE KEY(%s)", props.Keys)
			}
		case "UNIQUE":
			if props.Keys != "" {
				fmt.Fprintf(&sb, "UNIQUE KEY(%s)", props.Keys)
			}
		case "PRIMARY":
			if props.Keys != "" {
				fmt.Fprintf(&sb, "PRIMARY KEY(%s)", props.Keys)
			}
		}
	}

	// 2. Comment (must come before PARTITION BY)
	if props.Comment != "" {
		if sb.Len() > 0 {
			sb.WriteString(" ")
		}
		// Escape backslashes first, then single quotes
		escaped := strings.ReplaceAll(props.Comment, "\\", "\\\\")
		escaped = strings.ReplaceAll(escaped, "'", "''")
		fmt.Fprintf(&sb, "COMMENT '%s'", escaped)
	}

	// 3. Partitioning
	// Format: PARTITION BY RANGE|LIST (column) (partition_definitions)
	if props.PartitionBy != "" {
		if sb.Len() > 0 {
			sb.WriteString(" ")
		}
		fmt.Fprintf(&sb, "PARTITION BY %s", props.PartitionBy)
		if props.Partitions != "" {
			fmt.Fprintf(&sb, " %s", props.Partitions)
		}
	}

	// 4. Distribution
	// Format: DISTRIBUTED BY HASH(columns) [BUCKETS n] | DISTRIBUTED BY RANDOM [BUCKETS n]
	if props.DistributedBy != "" {
		if sb.Len() > 0 {
			sb.WriteString(" ")
		}
		fmt.Fprintf(&sb, "DISTRIBUTED BY HASH(%s)", props.DistributedBy)
		if props.Buckets > 0 {
			fmt.Fprintf(&sb, " BUCKETS %d", props.Buckets)
		}
	} else if props.Buckets > 0 {
		// Random distribution with specified buckets
		if sb.Len() > 0 {
			sb.WriteString(" ")
		}
		fmt.Fprintf(&sb, "DISTRIBUTED BY RANDOM BUCKETS %d", props.Buckets)
	}

	// 5. Order By (for PRIMARY KEY model only)
	if props.OrderBy != "" && engine == "PRIMARY" {
		if sb.Len() > 0 {
			sb.WriteString(" ")
		}
		fmt.Fprintf(&sb, "ORDER BY (%s)", props.OrderBy)
	}

	// 6. Properties
	// Format: PROPERTIES ("key" = "value", ...)
	var propParts []string
	if props.ReplicationNum > 0 {
		propParts = append(propParts, fmt.Sprintf("\"replication_num\" = \"%d\"", props.ReplicationNum))
	}
	if props.Properties != "" {
		propParts = append(propParts, props.Properties)
	}
	if len(propParts) > 0 {
		if sb.Len() > 0 {
			sb.WriteString(" ")
		}
		fmt.Fprintf(&sb, "PROPERTIES (%s)", strings.Join(propParts, ", "))
	}

	return sb.String()
}

// dropTable drops a table or view.
func (c *connection) dropTable(ctx context.Context, name string, isView bool) error {
	db, err := c.getDB(ctx)
	if err != nil {
		return err
	}

	// Use a dedicated connection to ensure catalog/database context is maintained
	conn, err := db.Connx(ctx)
	if err != nil {
		return fmt.Errorf("create connection: %w", err)
	}
	defer conn.Close()

	// Set catalog and database context
	if err := c.setCatalogContext(ctx, conn); err != nil {
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
	_, err = conn.ExecContext(ctx, query)
	return err
}

// dropTableOrView drops a table or view regardless of its type.
// Tries both DROP TABLE and DROP VIEW to handle cases where the type changed.
func (c *connection) dropTableOrView(ctx context.Context, name string) error {
	db, err := c.getDB(ctx)
	if err != nil {
		return err
	}

	// Use a dedicated connection to ensure catalog/database context is maintained
	conn, err := db.Connx(ctx)
	if err != nil {
		return fmt.Errorf("create connection: %w", err)
	}
	defer conn.Close()

	// Set catalog and database context
	if err := c.setCatalogContext(ctx, conn); err != nil {
		return err
	}

	// Build fully-qualified table name using connector's database
	tableName := safeSQLName(name)
	if c.configProp.Database != "" {
		tableName = safeSQLName(c.configProp.Database) + "." + tableName
	}

	// Try dropping as table first, then as view
	// Both use IF EXISTS so they won't error if the object doesn't exist
	_, _ = conn.ExecContext(ctx, fmt.Sprintf("DROP TABLE IF EXISTS %s", tableName))
	_, _ = conn.ExecContext(ctx, fmt.Sprintf("DROP VIEW IF EXISTS %s", tableName))
	return nil
}

// renameTable renames a table or recreates a view.
func (c *connection) renameTable(ctx context.Context, oldName, newName string, isView bool) error {
	db, err := c.getDB(ctx)
	if err != nil {
		return err
	}

	// Use a dedicated connection to ensure catalog/database context is maintained
	conn, err := db.Connx(ctx)
	if err != nil {
		return fmt.Errorf("create connection: %w", err)
	}
	defer conn.Close()

	// Set catalog and database context
	if err := c.setCatalogContext(ctx, conn); err != nil {
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
		row := conn.QueryRowContext(ctx, fmt.Sprintf("SHOW CREATE VIEW %s", oldTableName))
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
		if _, err := conn.ExecContext(ctx, fmt.Sprintf("DROP VIEW IF EXISTS %s", oldTableName)); err != nil {
			return err
		}

		// Drop target view if exists
		if _, err := conn.ExecContext(ctx, fmt.Sprintf("DROP VIEW IF EXISTS %s", newTableName)); err != nil {
			return err
		}

		// Create new view
		_, err = conn.ExecContext(ctx, fmt.Sprintf("CREATE VIEW %s AS %s", newTableName, selectStmt))
		return err
	}

	// For tables, use ALTER TABLE RENAME
	// First drop target if exists
	if _, err := conn.ExecContext(ctx, fmt.Sprintf("DROP TABLE IF EXISTS %s", newTableName)); err != nil {
		return err
	}

	// Rename table
	_, err = conn.ExecContext(ctx, fmt.Sprintf("ALTER TABLE %s RENAME %s", oldTableName, safeSQLName(newName)))
	return err
}

// insertIntoTable inserts data into an existing table.
func (c *connection) insertIntoTable(ctx context.Context, name, sql string, props *ModelOutputProperties) error {
	db, err := c.getDB(ctx)
	if err != nil {
		return err
	}

	// Use a dedicated connection to ensure catalog/database context is maintained
	conn, err := db.Connx(ctx)
	if err != nil {
		return fmt.Errorf("create connection: %w", err)
	}
	defer conn.Close()

	// Set catalog and database context
	if err := c.setCatalogContext(ctx, conn); err != nil {
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
		_, err = conn.ExecContext(ctx, query)
		return err
	}

	return fmt.Errorf("incremental strategy %q not supported for StarRocks", strategy)
}

// safeSQLName escapes a SQL identifier.
func safeSQLName(name string) string {
	// Use backticks for StarRocks/MySQL compatibility
	return "`" + strings.ReplaceAll(name, "`", "``") + "`"
}
