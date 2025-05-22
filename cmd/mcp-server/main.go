package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	mcpgo "github.com/mark3labs/mcp-go/mcp"
	"github.com/rilldata/rill/runtime/server/mcp"
)

func main() {
	// Create a new MCP server
	server, err := mcp.NewMCPServer()
	if err != nil {
		log.Fatalf("Failed to create MCP server: %v", err)
	}

	// Create a new MCP instance
	mcpServer := mcpgo.NewServer()

	// Register handlers
	mcpServer.RegisterTool("list_metrics_views", server.handleListMetricsViews)
	mcpServer.RegisterTool("get_metrics_view_spec", server.handleGetMetricsViewSpec)
	mcpServer.RegisterTool("get_metrics_view_time_range_summary", server.handleGetMetricsViewTimeRangeSummary)
	mcpServer.RegisterTool("get_metrics_view_aggregation", server.handleGetMetricsViewAggregation)
	mcpServer.RegisterTool("generate_chart", server.handleGenerateChart)

	// Create a context that will be canceled on SIGINT or SIGTERM
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Start the server
	if err := mcpServer.Start(ctx); err != nil {
		log.Fatalf("Failed to start MCP server: %v", err)
	}
}
