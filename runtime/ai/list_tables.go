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
	PageSize      int    `json:"page_size,omitempty" jsonschema:"Maximum number of tables to return. Defaults to 100."`
	PageToken     string `json:"page_token,omitempty" jsonschema:"Token for pagination. Use next_page_token from previous response to get next page."`
}

type ListTablesResult struct {
	Tables        []TableInfo `json:"tables"`
	NextPageToken string      `json:"next_page_token,omitempty"`
}

type TableInfo struct {
	Database                string `json:"database,omitempty"`
	DatabaseSchema          string `json:"database_schema,omitempty"`
	IsDefaultDatabaseSchema bool   `json:"is_default_database_schema,omitempty"`
	Name                    string `json:"name"`
	IsView                  bool   `json:"is_view"`
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
	return checkDeveloperAccess(ctx, t.Runtime, false)
}

func (t *ListTables) Handler(ctx context.Context, args *ListTablesArgs) (*ListTablesResult, error) {
	s := GetSession(ctx)

	// Get OLAP handle
	olap, release, err := t.Runtime.OLAP(ctx, s.InstanceID(), args.Connector)
	if err != nil {
		return nil, err
	}
	defer release()

	// Apply defaults
	pageSize := args.PageSize
	if pageSize <= 0 {
		pageSize = 100
	}

	// List tables via information schema
	tables, nextToken, err := olap.InformationSchema().All(ctx, args.SearchPattern, uint32(pageSize), args.PageToken)
	if err != nil {
		return nil, err
	}

	// Convert to result format
	result := &ListTablesResult{
		Tables:        make([]TableInfo, 0, len(tables)),
		NextPageToken: nextToken,
	}
	for _, table := range tables {
		result.Tables = append(result.Tables, TableInfo{
			Database:                table.Database,
			DatabaseSchema:          table.DatabaseSchema,
			IsDefaultDatabaseSchema: table.IsDefaultDatabaseSchema,
			Name:                    table.Name,
			IsView:                  table.View,
		})
	}

	return result, nil
}
