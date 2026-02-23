package ai

import (
	"context"
	"fmt"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rilldata/rill/runtime"
)

const QuerySQLName = "query_sql"

type QuerySQL struct {
	Runtime *runtime.Runtime
}

var _ Tool[*QuerySQLArgs, *QuerySQLResult] = (*QuerySQL)(nil)

type QuerySQLArgs struct {
	Connector      string `json:"connector,omitempty" jsonschema:"Optional OLAP connector name. Defaults to the instance's default OLAP connector."`
	SQL            string `json:"sql" jsonschema:"The SQL query to execute. You are strongly encouraged to use LIMIT in your query and to keep it as low as possible for your task (ideally below 100 rows). The server will truncate large results regardless of the limit (and return a warning if it does)."`
	TimeoutSeconds int    `json:"timeout_seconds,omitempty" jsonschema:"Query timeout in seconds. Defaults to 30."`
}

type QuerySQLResult struct {
	Schema            []SchemaField `json:"schema"`
	Data              [][]any       `json:"data"`
	TruncationWarning string        `json:"truncation_warning,omitempty"`
}

func (t *QuerySQL) Spec() *mcp.Tool {
	return &mcp.Tool{
		Name:        QuerySQLName,
		Title:       "Query SQL",
		Description: "Execute a raw SQL query against an OLAP connector to introspect data.",
		Meta: map[string]any{
			"openai/toolInvocation/invoking": "Executing SQL query...",
			"openai/toolInvocation/invoked":  "Executed SQL query",
		},
	}
}

func (t *QuerySQL) CheckAccess(ctx context.Context) (bool, error) {
	return checkDeveloperAccess(ctx, t.Runtime, false)
}

func (t *QuerySQL) Handler(ctx context.Context, args *QuerySQLArgs) (*QuerySQLResult, error) {
	if args.SQL == "" {
		return nil, fmt.Errorf("sql query is required")
	}

	s := GetSession(ctx)

	// Apply timeout default
	timeoutSeconds := args.TimeoutSeconds
	if timeoutSeconds <= 0 {
		timeoutSeconds = 30
	}

	// Apply a hard limit to prevent large results that bloat the context
	instance, err := t.Runtime.Instance(ctx, s.InstanceID())
	if err != nil {
		return nil, fmt.Errorf("failed to get instance: %w", err)
	}
	cfg, err := instance.Config()
	if err != nil {
		return nil, fmt.Errorf("failed to get instance config: %w", err)
	}
	hardLimit := cfg.AIMaxQueryLimit

	// Apply timeout
	ctx, cancel := context.WithTimeout(ctx, time.Duration(timeoutSeconds)*time.Second)
	defer cancel()

	// Build resolver properties with system limit
	props := map[string]any{
		"sql":   args.SQL,
		"limit": hardLimit,
	}
	if args.Connector != "" {
		props["connector"] = args.Connector
	}

	// Execute via the SQL resolver
	res, err := t.Runtime.Resolve(ctx, &runtime.ResolveOptions{
		InstanceID:         s.InstanceID(),
		Resolver:           "sql",
		ResolverProperties: props,
		Claims:             s.Claims(),
	})
	if err != nil {
		return nil, err
	}
	defer res.Close()

	// Collect results in tabular format
	schema, data, err := resolverResultToTabular(res)
	if err != nil {
		return nil, err
	}

	result := &QuerySQLResult{
		Schema: schema,
		Data:   data,
	}
	if int64(len(data)) >= hardLimit { // Add a warning if we hit the system limit
		result.TruncationWarning = fmt.Sprintf("The system truncated the result to %d rows", hardLimit)
	}
	return result, nil
}
