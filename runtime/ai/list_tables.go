package ai

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rilldata/rill/runtime"
)

const ListTablesName = "list_tables"

type ListTables struct {
	Runtime *runtime.Runtime
}

var _ Tool[*ListTablesArgs, *ListTablesResult] = (*ListTables)(nil)

type ListTablesArgs struct {
	Connector     string `json:"connector,omitempty" jsonschema:"Optional OLAP connector name. Defaults to the instance's default OLAP connector."`
	SearchPattern string `json:"search_pattern,omitempty" jsonschema:"Optional pattern to filter table names (uses SQL LIKE syntax)."`
}

type ListTablesResult struct {
	Tables []TableInfo `json:"tables"`
}

type TableInfo struct {
	Database       string `json:"database,omitempty"`
	DatabaseSchema string `json:"database_schema,omitempty"`
	Name           string `json:"name"`
	IsView         bool   `json:"is_view"`
}

func (t *ListTables) Spec() *mcp.Tool {
	return &mcp.Tool{
		Name:        ListTablesName,
		Title:       "List Tables",
		Description: "List tables and views in an OLAP connector.",
		Meta: map[string]any{
			"openai/toolInvocation/invoking": "Listing tables...",
			"openai/toolInvocation/invoked":  "Listed tables",
		},
	}
}

func (t *ListTables) CheckAccess(ctx context.Context) (bool, error) {
	return checkDeveloperAgentAccess(ctx, t.Runtime)
}

func (t *ListTables) Handler(ctx context.Context, args *ListTablesArgs) (*ListTablesResult, error) {
	s := GetSession(ctx)

	// Get OLAP handle
	olap, release, err := t.Runtime.OLAP(ctx, s.InstanceID(), args.Connector)
	if err != nil {
		return nil, err
	}
	defer release()

	// List tables via information schema
	const pageSize = 1000
	tables, _, err := olap.InformationSchema().All(ctx, args.SearchPattern, pageSize, "")
	if err != nil {
		return nil, err
	}

	// Convert to result format
	result := &ListTablesResult{
		Tables: make([]TableInfo, 0, len(tables)),
	}
	for _, table := range tables {
		result.Tables = append(result.Tables, TableInfo{
			Database:       table.Database,
			DatabaseSchema: table.DatabaseSchema,
			Name:           table.Name,
			IsView:         table.View,
		})
	}

	return result, nil
}
