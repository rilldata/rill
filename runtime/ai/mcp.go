package ai

import "github.com/modelcontextprotocol/go-sdk/mcp"

const mcpInstructions = `
# Rill MCP Server
This server exposes APIs for querying **metrics views**, which represent Rill's metrics layer.
## Workflow Overview
1. **List metrics views:** Use "list_metrics_views" to discover available metrics views in the project.
2. **Get metrics view spec:** Use "get_metrics_view" to fetch a metrics view's specification. This is important to understand all the dimensions and measures in a metrics view.
3. **Query the time range:** Use "query_metrics_view_time_range" to obtain the available time range for a metrics view. This is important to understand what time range the data spans.
4. **Query the metrics:** Use "query_metrics_view" to run queries to get aggregated results.
In the workflow, do not proceed with the next step until the previous step has been completed. If the information from the previous step is already known (let's say for subsequent queries), you can skip it.
If a response contains an "ai_instructions" field, you should interpret it as additional instructions for how to behave in subsequent responses that relate to that tool call.
`

func (s *Session) MCPServer() *mcp.Server {
	impl := &mcp.Implementation{
		Name:    "rill",
		Title:   "Rill MCP Server",
		Version: s.runner.Runtime.Version().String(),
	}

	srv := mcp.NewServer(impl, &mcp.ServerOptions{
		Instructions: mcpInstructions,
	})

	for _, t := range s.runner.Tools {
		if t.checkAccess != nil && !t.checkAccess(s.claims) {
			continue
		}
		t.registerWithMCPServer(srv)
	}

	return srv
}
