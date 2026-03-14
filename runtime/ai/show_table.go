package ai

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
)

const ShowTableName = "show_table"

type ShowTable struct {
	Runtime *runtime.Runtime
}

var _ Tool[*ShowTableArgs, *ShowTableResult] = (*ShowTable)(nil)

type ShowTableArgs struct {
	Connector      string `json:"connector,omitempty" jsonschema:"Optional OLAP connector name. Defaults to the instance's default OLAP connector."`
	Table          string `json:"table" jsonschema:"Name of the table to describe. Must be a simple table name; database/schema names should be provided using the separate fields."`
	Database       string `json:"database,omitempty" jsonschema:"Database that contains the table (defaults to the connector's default database if applicable)."`
	DatabaseSchema string `json:"database_schema,omitempty" jsonschema:"Database schema that contains the table (defaults to the connector's default schema if applicable)."`
}

type ShowTableResult struct {
	Name              string       `json:"name"`
	IsView            bool         `json:"is_view"`
	Columns           []ColumnInfo `json:"columns"`
	PhysicalSizeBytes int64        `json:"physical_size_bytes,omitempty" jsonschema:"The physical size of the table in bytes. If 0 or omitted, size information is not available."`
	DDL               string       `json:"ddl,omitempty" jsonschema:"The SQL DDL statement (CREATE TABLE/VIEW) for this table, if available."`
}

type ColumnInfo struct {
	Name string `json:"name" jsonschema:"The name of the column."`
	Type string `json:"type" jsonschema:"The data type of the column. This is a generic type code and does not exactly match the underlying SQL type."`
}

func (t *ShowTable) Spec() *mcp.Tool {
	return &mcp.Tool{
		Name:        ShowTableName,
		Title:       "Show Table",
		Description: "Show schema and column information for a table in an OLAP connector. Note: Table, schema and database names passed to this tool are case sensitive; if you get an error and you're working with a database that folds unquoted identifiers (e.g Snowflake folds to uppercase), you may need to retry with the casing adjusted accordingly.",
		Meta: map[string]any{
			"openai/toolInvocation/invoking": "Getting table schema...",
			"openai/toolInvocation/invoked":  "Got table schema",
		},
	}
}

func (t *ShowTable) CheckAccess(ctx context.Context) (bool, error) {
	return checkDeveloperAccess(ctx, t.Runtime, false)
}

func (t *ShowTable) Handler(ctx context.Context, args *ShowTableArgs) (*ShowTableResult, error) {
	if args.Table == "" {
		return nil, fmt.Errorf("table name is required")
	}

	s := GetSession(ctx)

	// Get OLAP handle
	olap, release, err := t.Runtime.OLAP(ctx, s.InstanceID(), args.Connector)
	if err != nil {
		return nil, err
	}
	defer release()

	// Lookup the table
	table, err := olap.InformationSchema().Lookup(ctx, args.Database, args.DatabaseSchema, args.Table)
	if err != nil {
		return nil, err
	}

	// Load physical size
	_ = olap.InformationSchema().LoadPhysicalSize(ctx, []*drivers.OlapTable{table})

	// Load DDL
	_ = olap.InformationSchema().LoadDDL(ctx, table)

	// Build result
	result := &ShowTableResult{
		Name:              table.Name,
		IsView:            table.View,
		PhysicalSizeBytes: table.PhysicalSizeBytes,
		DDL:               table.DDL,
		Columns:           make([]ColumnInfo, 0),
	}

	if table.Schema != nil {
		for _, field := range table.Schema.Fields {
			result.Columns = append(result.Columns, ColumnInfo{
				Name: field.Name,
				Type: field.Type.Code.String(),
			})
		}
	}

	return result, nil
}
