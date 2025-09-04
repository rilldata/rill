package clickhouse

import (
	"context"
	"fmt"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
)

const _defaultConcurrentInserts = 1

type ModelInputProperties struct {
	SQL string `mapstructure:"sql"`
}

type ModelOutputProperties struct {
	// Table is the name of the table to create. If not specified, the model name is used.
	Table string `mapstructure:"table"`
	// Materialize is a flag to indicate if the model should be materialized as a physical table or dictionary.
	// If false, the model will be created as a view.
	// If unspcecified, a default is selected based on context.
	Materialize *bool `mapstructure:"materialize"`
	// IncrementalStrategy is the strategy to use for incremental inserts.
	IncrementalStrategy drivers.IncrementalStrategy `mapstructure:"incremental_strategy"`
	// UniqueKey is the unique key for the model. This is used for the incremental strategy "merge".
	UniqueKey []string `mapstructure:"unique_key"`
	// Typ to materialize the model into. Possible values include `TABLE`, `VIEW` or `DICTIONARY`. Optional.
	Typ string `mapstructure:"type"`
	// Columns sets the column names and data types. If unspecified these are detected from the select query by clickhouse.
	// It is also possible to set indexes with this property.
	// Example : (id UInt32, username varchar, email varchar, created_at datetime, INDEX idx1 username TYPE set(100) GRANULARITY 3)
	Columns string `mapstructure:"columns"`
	// EngineFull can be used to set the table parameters like engine, partition key in SQL format without setting individual properties.
	// It also allows creating dictionaries using a source.
	// Example:
	//  ENGINE = MergeTree
	//	PARTITION BY toYYYYMM(__time)
	//	ORDER BY __time
	//	TTL d + INTERVAL 1 MONTH DELETE
	EngineFull string `mapstructure:"engine_full"`
	// Engine sets the table engine. Default: MergeTree
	Engine string `mapstructure:"engine"`
	// OrderBy sets the order by clause. Default: tuple() for MergeTree and not set for other engines
	OrderBy string `mapstructure:"order_by"`
	// PartitionBy sets the partition by clause.
	PartitionBy string `mapstructure:"partition_by"`
	// PrimaryKey sets the primary key clause.
	PrimaryKey string `mapstructure:"primary_key"`
	// SampleBy sets the sample by clause.
	SampleBy string `mapstructure:"sample_by"`
	// TTL sets ttl for column and table.
	TTL string `mapstructure:"ttl"`
	// TableSettings set the table specific settings.
	TableSettings string `mapstructure:"table_settings"`
	// QuerySettings sets the settings clause used in insert/create table as select queries.
	QuerySettings string `mapstructure:"query_settings"`
	// DistributedSettings is table settings for distributed table.
	DistributedSettings string `mapstructure:"distributed_settings"`
	// DistributedShardingKey is the sharding key for distributed table.
	DistributedShardingKey string `mapstructure:"distributed_sharding_key"`
	// DictionarySourceUser is the user that case access the source dictionary table. Only used when typ is DICTIONARY.
	DictionarySourceUser string `mapstructure:"dictionary_source_user"`
	// DictionarySourcePassword is the password for the user that can access the source dictionary table. Only used when typ is DICTIONARY.
	DictionarySourcePassword string `mapstructure:"dictionary_source_password"`
}

// validateAndApplyDefaults validates the model input and output properties and applies defaults.
// The inputProps may optionally be nil to allow for connectors that don't use the base SQL-centric input properties.
// The outputProps may not be nil.
func (c *Connection) validateAndApplyDefaults(opts *drivers.ModelExecuteOptions, ip *ModelInputProperties, op *ModelOutputProperties) error {
	// EngineFull is not compatible with most other individual properties.
	if op.EngineFull != "" {
		if op.Engine != "" || op.OrderBy != "" || op.PartitionBy != "" || op.PrimaryKey != "" || op.SampleBy != "" || op.TTL != "" || op.TableSettings != "" || op.DictionarySourceUser != "" || op.DictionarySourcePassword != "" {
			return fmt.Errorf("`engine_full` property cannot be used with individual properties")
		}
	}

	// Handle materialize and type properties. We want to gracefully handle the cases where either or both are set in a non-contradictory way.
	// This gets extra tricky since materialize=true is compatible with both TABLE and DICTIONARY types (just not VIEW).
	op.Typ = strings.ToUpper(op.Typ)
	if op.Materialize != nil {
		if *op.Materialize {
			if op.Typ == "VIEW" {
				return fmt.Errorf("the `type` and `materialize` properties contradict each other")
			} else if op.Typ == "" {
				op.Typ = "TABLE"
			}
		} else {
			if op.Typ == "" {
				op.Typ = "VIEW"
			} else if op.Typ != "VIEW" {
				return fmt.Errorf("the `type` and `materialize` properties contradict each other")
			}
		}
	}
	if opts.Incremental || opts.PartitionRun { // Incremental or partitioned models default to TABLE.
		if op.Typ != "" && op.Typ != "TABLE" {
			return fmt.Errorf("incremental or partitioned models must be materialized as a table")
		}
		op.Typ = "TABLE"
	}
	if op.Typ == "" { // Apply default for plain unannotated models.
		if opts.Env.DefaultMaterialize {
			op.Typ = "TABLE"
		} else {
			op.Typ = "VIEW"
		}
	}
	if op.Typ != "TABLE" && op.Typ != "VIEW" && op.Typ != "DICTIONARY" {
		return fmt.Errorf("invalid type %q, must be one of TABLE, VIEW or DICTIONARY", op.Typ)
	}

	// For tables, apply a default table engine.
	if op.Engine == "" && (op.Typ == "TABLE" || op.Typ == "DICTIONARY") {
		if c.config.Cluster != "" {
			op.Engine = "ReplicatedMergeTree"
		} else {
			op.Engine = "MergeTree"
		}
	}

	// Validate it's a known incremental strategy.
	switch op.IncrementalStrategy {
	case drivers.IncrementalStrategyUnspecified, drivers.IncrementalStrategyAppend, drivers.IncrementalStrategyPartitionOverwrite, drivers.IncrementalStrategyMerge:
	default:
		return fmt.Errorf("invalid incremental strategy %q", op.IncrementalStrategy)
	}

	// partition_by is required for the partition_overwrite incremental strategy.
	if op.IncrementalStrategy == drivers.IncrementalStrategyPartitionOverwrite && op.PartitionBy == "" {
		return fmt.Errorf(`must specify a "partition_by" when "incremental_strategy" is %q`, op.IncrementalStrategy)
	}

	// ClickHouse enforces the requirement of either a primary key or an ORDER BY clause for the ReplacingMergeTree engine.
	// When using the incremental strategy as 'merge', the engine must be ReplacingMergeTree.
	// This ensures that duplicate rows are eventually replaced, maintaining data consistency.
	if op.IncrementalStrategy == drivers.IncrementalStrategyMerge && !(strings.Contains(op.Engine, "ReplacingMergeTree") || strings.Contains(op.EngineFull, "ReplacingMergeTree")) {
		return fmt.Errorf(`must use "ReplacingMergeTree" engine when "incremental_strategy" is %q`, op.IncrementalStrategy)
	}

	// We want to use partition_overwrite as the default incremental strategy for models with partitions.
	// This requires us to inject the partition key into the SQL query, so this only works for SQL models.
	if op.IncrementalStrategy == drivers.IncrementalStrategyUnspecified && opts.PartitionRun && ip != nil && ip.SQL != "" {
		if op.EngineFull != "" {
			return fmt.Errorf("you must provide an explicit `incremental_strategy` when using `engine_full` with a partitioned model")
		}
		// `use_structure_from_insertion_table_in_table_functions = 0` is a workaround for https://github.com/ClickHouse/ClickHouse/issues/83257
		ip.SQL = fmt.Sprintf("SELECT %s AS __rill_partition, * FROM (%s\n) SETTINGS use_structure_from_insertion_table_in_table_functions = 0", safeSQLString(opts.PartitionKey), ip.SQL)
		op.IncrementalStrategy = drivers.IncrementalStrategyPartitionOverwrite
		op.PartitionBy = "__rill_partition"
	}

	// If we failed to apply a better incremental strategy, fall back to append.
	if op.IncrementalStrategy == drivers.IncrementalStrategyUnspecified {
		op.IncrementalStrategy = drivers.IncrementalStrategyAppend
	}

	// The input props are optional, but if set, check there's a valid SQL query.
	if ip != nil {
		if ip.SQL == "" && op.Typ != "DICTIONARY" {
			return fmt.Errorf("input SQL is required")
		}
	}

	// Add query settings to the SQL query.
	// We do this last since the SQL query may be modified in some of the above steps.
	if op.QuerySettings != "" {
		if ip == nil || ip.SQL == "" {
			return fmt.Errorf("cannot set query_settings without a SQL query")
		}
		ip.SQL = ip.SQL + " SETTINGS " + op.QuerySettings
	}

	return nil
}

func (p *ModelOutputProperties) tblConfig() string {
	if p.EngineFull != "" {
		return p.EngineFull
	}

	var sb strings.Builder

	// engine
	if p.Engine != "" {
		fmt.Fprintf(&sb, "ENGINE = %s", p.Engine)
	}

	// order_by
	if p.OrderBy != "" {
		fmt.Fprintf(&sb, " ORDER BY %s", p.OrderBy)
	} else if p.PrimaryKey != "" {
		fmt.Fprintf(&sb, " ORDER BY %s", p.PrimaryKey)
	} else if p.Engine == "MergeTree" || p.Engine == "ReplicatedMergeTree" {
		// need ORDER BY for MergeTree
		// it is optional for many other engines
		fmt.Fprintf(&sb, " ORDER BY tuple()")
	}

	// partition_by
	if p.PartitionBy != "" {
		fmt.Fprintf(&sb, " PARTITION BY %s", p.PartitionBy)
	}

	// primary_key
	if p.PrimaryKey != "" {
		fmt.Fprintf(&sb, " PRIMARY KEY %s", p.PrimaryKey)
	}

	// sample_by
	if p.SampleBy != "" {
		fmt.Fprintf(&sb, " SAMPLE BY %s", p.SampleBy)
	}

	// ttl
	if p.TTL != "" {
		fmt.Fprintf(&sb, " TTL %s", p.TTL)
	}

	// settings
	if p.TableSettings != "" {
		// Backwards compatibility: previously we did not automatically add the `SETTINGS` keyword.
		// So we only add it if it's not already there.
		if strings.HasPrefix(strings.TrimSpace(p.TableSettings), "SETTINGS") || strings.HasPrefix(strings.TrimSpace(p.TableSettings), "settings") {
			fmt.Fprintf(&sb, " %s", p.TableSettings)
		} else {
			fmt.Fprintf(&sb, " SETTINGS %s", p.TableSettings)
		}
	}

	return sb.String()
}

type ModelResultProperties struct {
	Table         string `mapstructure:"table"`
	View          bool   `mapstructure:"view"`
	Typ           string `mapstructure:"type"`
	UsedModelName bool   `mapstructure:"used_model_name"`
}

func (c *Connection) Rename(ctx context.Context, res *drivers.ModelResult, newName string, env *drivers.ModelEnv) (*drivers.ModelResult, error) {
	resProps := &ModelResultProperties{}
	if err := mapstructure.WeakDecode(res.Properties, resProps); err != nil {
		return nil, fmt.Errorf("failed to parse previous result properties: %w", err)
	}

	if !resProps.UsedModelName {
		return res, nil
	}

	err := c.forceRenameTable(ctx, resProps.Table, resProps.View, newName)
	if err != nil {
		return nil, fmt.Errorf("failed to rename model: %w", err)
	}

	resProps.Table = newName
	resPropsMap := map[string]interface{}{}
	err = mapstructure.WeakDecode(resProps, &resPropsMap)
	if err != nil {
		return nil, fmt.Errorf("failed to encode result properties: %w", err)
	}

	return &drivers.ModelResult{
		Connector:  res.Connector,
		Properties: resPropsMap,
		Table:      newName,
	}, nil
}

func (c *Connection) Exists(ctx context.Context, res *drivers.ModelResult) (bool, error) {
	olap, ok := c.AsOLAP(c.instanceID)
	if !ok {
		return false, fmt.Errorf("connector is not an OLAP")
	}

	_, err := olap.InformationSchema().Lookup(ctx, c.config.Database, "", res.Table)
	return err == nil, nil
}

func (c *Connection) Delete(ctx context.Context, res *drivers.ModelResult) error {
	olap, ok := c.AsOLAP(c.instanceID)
	if !ok {
		return fmt.Errorf("connector is not an OLAP")
	}

	_ = c.dropTable(ctx, stagingTableNameFor(res.Table))

	table, err := olap.InformationSchema().Lookup(ctx, c.config.Database, "", res.Table)
	if err != nil {
		return err
	}

	return c.dropTable(ctx, table.Name)
}

func (c *Connection) MergePartitionResults(a, b *drivers.ModelResult) (*drivers.ModelResult, error) {
	if a.Table != b.Table {
		return nil, fmt.Errorf("cannot merge partitioned results that output to different table names (%q != %q)", a.Table, b.Table)
	}
	return a, nil
}

// forceRenameTable renames a table or view from fromName to toName.
// If a view or table already exists with toName, it is overwritten.
func (c *Connection) forceRenameTable(ctx context.Context, fromName string, fromIsView bool, toName string) error {
	if fromName == "" || toName == "" {
		return fmt.Errorf("cannot rename empty table name: fromName=%q, toName=%q", fromName, toName)
	}

	if fromName == toName {
		return nil
	}

	// Infer SQL keyword for the table type
	var typ string
	if fromIsView {
		typ = "VIEW"
	} else {
		typ = "TABLE"
	}

	// Renaming a table to the same name with different casing is not supported. Workaround by renaming to a temporary name first.
	if strings.EqualFold(fromName, toName) {
		tmpName := fmt.Sprintf("__rill_tmp_rename_%s_%s", typ, toName)
		err := c.renameEntity(ctx, fromName, tmpName)
		if err != nil {
			return err
		}
		fromName = tmpName
	}

	// Do the rename
	return c.renameEntity(ctx, fromName, toName)
}

func boolPtr(b bool) *bool {
	return &b
}

// stagingTableName returns a stable temporary table name for a destination table.
// By using a stable temporary table name, we can ensure proper garbage collection without managing additional state.
func stagingTableNameFor(table string) string {
	return "__rill_tmp_model_" + table
}
