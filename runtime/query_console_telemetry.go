package runtime

import (
	"context"
	"time"

	"github.com/rilldata/rill/runtime/pkg/activity"
)

// Query console telemetry event type constants.
const (
	QueryConsoleEventQueryExecuted              = "query_executed"
	QueryConsoleEventQueryFailed                = "query_failed"
	QueryConsoleEventQueryWarnedSoftLimit       = "query_warned_soft_limit"
	QueryConsoleEventQueryBlockedHardLimit      = "query_blocked_hard_limit"
	QueryConsoleEventModelPublished             = "model_published"
	QueryConsoleEventModelPublishFailedDupeName = "model_publish_failed_duplicate_name"
)

// QueryEventDimensions captures the structured fields emitted with each query console telemetry event.
type QueryEventDimensions struct {
	// BackendType identifies the execution backend (e.g. "embedded_engine" or "external_warehouse").
	BackendType string
	// ProjectID is the Rill project identifier.
	ProjectID string
	// UserID is the authenticated user who triggered the event.
	UserID string
	// InstanceID is the runtime instance identifier.
	InstanceID string
	// BytesScanned is the number of bytes scanned by the query (0 if unknown).
	BytesScanned int64
	// ExecutionTimeMS is the wall-clock execution time in milliseconds.
	ExecutionTimeMS int64
	// SQL is the (possibly truncated) query text. Useful for debugging but should
	// never contain credentials. Truncated to a safe length before emission.
	SQL string
	// ModelName is populated only for model-publish events.
	ModelName string
	// ModelType is populated only for model-publish events (e.g. "source_model", "derived_model").
	ModelType string
	// ErrorMessage is populated for failure events. Truncated to a safe length.
	ErrorMessage string
}

// maxSQLTelemetryLen is the maximum length of a SQL string included in telemetry.
const maxSQLTelemetryLen = 1024

// maxErrorTelemetryLen is the maximum length of an error message included in telemetry.
const maxErrorTelemetryLen = 512

// EmitQueryEvent emits a structured telemetry event for the query console.
// It is safe to call with a nil activity client — the call becomes a no-op.
func EmitQueryEvent(ctx context.Context, client *activity.Client, eventType string, dims QueryEventDimensions) {
	if client == nil {
		return
	}

	// Truncate potentially large fields to keep telemetry payloads bounded.
	sql := truncate(dims.SQL, maxSQLTelemetryLen)
	errMsg := truncate(dims.ErrorMessage, maxErrorTelemetryLen)

	ev := &activity.Event{
		EventType: eventType,
		EventTime: time.Now(),
	}

	// Build the dimension map. We always include the core fields even if empty
	// so downstream consumers have a stable schema.
	ev.Dimensions = map[string]string{
		"backend_type": dims.BackendType,
		"project_id":   dims.ProjectID,
		"user_id":      dims.UserID,
		"instance_id":  dims.InstanceID,
	}

	// Numeric measures carried as dimensions (string-encoded) for simplicity,
	// matching patterns used elsewhere in the codebase.
	ev.Measures = map[string]float64{
		"bytes_scanned":     float64(dims.BytesScanned),
		"execution_time_ms": float64(dims.ExecutionTimeMS),
	}

	// Optional fields — only set when non-empty to keep payloads clean.
	if sql != "" {
		ev.Dimensions["sql"] = sql
	}
	if dims.ModelName != "" {
		ev.Dimensions["model_name"] = dims.ModelName
	}
	if dims.ModelType != "" {
		ev.Dimensions["model_type"] = dims.ModelType
	}
	if errMsg != "" {
		ev.Dimensions["error_message"] = errMsg
	}

	client.Record(ctx, ev)
}

// EmitQueryExecuted is a convenience wrapper for a successful query execution event.
func EmitQueryExecuted(ctx context.Context, client *activity.Client, dims QueryEventDimensions) {
	EmitQueryEvent(ctx, client, QueryConsoleEventQueryExecuted, dims)
}

// EmitQueryFailed is a convenience wrapper for a failed query execution event.
func EmitQueryFailed(ctx context.Context, client *activity.Client, dims QueryEventDimensions) {
	EmitQueryEvent(ctx, client, QueryConsoleEventQueryFailed, dims)
}

// EmitQueryWarnedSoftLimit is a convenience wrapper for a soft-limit warning event.
func EmitQueryWarnedSoftLimit(ctx context.Context, client *activity.Client, dims QueryEventDimensions) {
	EmitQueryEvent(ctx, client, QueryConsoleEventQueryWarnedSoftLimit, dims)
}

// EmitQueryBlockedHardLimit is a convenience wrapper for a hard-limit blocking event.
func EmitQueryBlockedHardLimit(ctx context.Context, client *activity.Client, dims QueryEventDimensions) {
	EmitQueryEvent(ctx, client, QueryConsoleEventQueryBlockedHardLimit, dims)
}

// EmitModelPublished is a convenience wrapper for a successful model publish event.
func EmitModelPublished(ctx context.Context, client *activity.Client, dims QueryEventDimensions) {
	EmitQueryEvent(ctx, client, QueryConsoleEventModelPublished, dims)
}

// EmitModelPublishFailedDupeName is a convenience wrapper for a model publish
// failure due to a duplicate name.
func EmitModelPublishFailedDupeName(ctx context.Context, client *activity.Client, dims QueryEventDimensions) {
	EmitQueryEvent(ctx, client, QueryConsoleEventModelPublishFailedDupeName, dims)
}

// truncate returns s truncated to at most maxLen bytes. If truncated, an
// ellipsis marker is appended. It is safe for empty strings and zero maxLen.
func truncate(s string, maxLen int) string {
	if maxLen <= 0 || len(s) <= maxLen {
		return s
	}
	const ellipsis = "..."
	if maxLen <= len(ellipsis) {
		return s[:maxLen]
	}
	return s[:maxLen-len(ellipsis)] + ellipsis
}
