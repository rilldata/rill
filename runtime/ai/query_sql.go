package ai

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/pkg/jsonval"
)

const QuerySQLName = "query_sql"

type QuerySQL struct {
	Runtime *runtime.Runtime
}

var _ Tool[*QuerySQLArgs, *QuerySQLResult] = (*QuerySQL)(nil)

type QuerySQLArgs struct {
	Connector      string `json:"connector,omitempty" jsonschema:"Optional OLAP connector name. Defaults to the instance's default OLAP connector."`
	SQL            string `json:"sql" jsonschema:"The SQL query to execute."`
	TimeoutSeconds int    `json:"timeout_seconds,omitempty" jsonschema:"Query timeout in seconds. Defaults to 30."`
}

type QuerySQLResult struct {
	Data []map[string]any `json:"data"`
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
	return checkDeveloperAgentAccess(ctx, t.Runtime)
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

	// Apply timeout
	ctx, cancel := context.WithTimeout(ctx, time.Duration(timeoutSeconds)*time.Second)
	defer cancel()

	// Build resolver properties with hard-coded limit
	props := map[string]any{
		"sql":   args.SQL,
		"limit": int64(1000),
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

	// Collect results
	var data []map[string]any
	schema := &runtimev1.Type{Code: runtimev1.Type_CODE_STRUCT, StructType: res.Schema()}
	for {
		row, err := res.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}

		// Convert types for JSON serialization
		v, err := jsonval.ToValue(row, schema)
		if err != nil {
			return nil, fmt.Errorf("failed to convert row: %w", err)
		}
		row, ok := v.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("expected row to be map[string]any, got %T", v)
		}

		data = append(data, row)
	}

	return &QuerySQLResult{Data: data}, nil
}
