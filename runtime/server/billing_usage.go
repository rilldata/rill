package server

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
)

// recordAPICallUsage emits the billable `api_calls` metric for a programmatic request to the given instance.
// The instance attributes (org_id, project_id, etc.) are attached so the usage can be attributed to the right
// organization and project; requests without a resolvable instance are ignored downstream.
//
// This covers the programmatic surfaces that are cleanly distinguishable from interactive dashboard usage:
// the custom REST API, the MCP server, and server-side alert/report query execution. Dashboard UI queries and
// SDK/CLI queries share the same gRPC endpoints with user tokens and are intentionally not counted here.
func (s *Server) recordAPICallUsage(ctx context.Context, instanceID, source string) {
	if instanceID == "" {
		return
	}
	attrs := s.runtime.GetInstanceAttributes(ctx, instanceID)
	attrs = append(attrs, attribute.String("api_source", source))
	s.activity.RecordMetric(ctx, "api_calls", 1, attrs...)
}
