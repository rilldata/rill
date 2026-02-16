package server

import (
	"context"
	"time"

	"go.uber.org/zap"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/pkg/activity"
)

// MeterEventQueryExecution is the meter event name for query console executions.
const MeterEventQueryExecution = "query_console_execution"

// QueryMeterRecord captures the metering-relevant fields from a query execution.
type QueryMeterRecord struct {
	// BackendType indicates whether the query ran on the embedded engine or an external warehouse.
	BackendType runtimev1.QueryBackendType
	// BytesScanned is the number of bytes read during query execution.
	BytesScanned int64
	// ExecutionTimeMs is the wall-clock duration of the query in milliseconds.
	ExecutionTimeMs int64
	// UserID is the authenticated user who executed the query.
	UserID string
	// ProjectID is the Rill project (instance) the query was executed against.
	ProjectID string
	// Timestamp is the time the query completed.
	Timestamp time.Time
	// Success indicates whether the query finished without error.
	Success bool
}

// Billable returns true if the query execution should be counted toward billing.
// Per product requirements, only embedded-engine queries are billable.
// External warehouse queries are not billable because the cost is borne by the
// customer's own warehouse account.
func (r *QueryMeterRecord) Billable() bool {
	return r.BackendType == runtimev1.QueryBackendType_QUERY_BACKEND_TYPE_EMBEDDED_ENGINE
}

// RecordQueryMeter emits a metering event for a query console execution.
// It records usage dimensions needed for billing and usage analytics.
// If the activity client is nil the call is a no-op, which keeps the function
// safe for use in local/development mode where metering is not configured.
func RecordQueryMeter(ctx context.Context, client *activity.Client, logger *zap.Logger, record *QueryMeterRecord) {
	if client == nil {
		return
	}
	if record == nil {
		return
	}

	billable := record.Billable()

	backendStr := backendTypeString(record.BackendType)

	event := &activity.Event{
		EventType: MeterEventQueryExecution,
		EventTime: record.Timestamp,
		Dimensions: map[string]string{
			"backend_type": backendStr,
			"user_id":      record.UserID,
			"project_id":   record.ProjectID,
			"billable":     boolToMeterString(billable),
			"success":      boolToMeterString(record.Success),
		},
		Metrics: map[string]float64{
			"bytes_scanned":     float64(record.BytesScanned),
			"execution_time_ms": float64(record.ExecutionTimeMs),
		},
	}

	client.RecordMetric(ctx, event)

	if logger != nil {
		logger.Debug("query console meter recorded",
			zap.String("event_type", MeterEventQueryExecution),
			zap.String("backend_type", backendStr),
			zap.Int64("bytes_scanned", record.BytesScanned),
			zap.Int64("execution_time_ms", record.ExecutionTimeMs),
			zap.String("user_id", record.UserID),
			zap.String("project_id", record.ProjectID),
			zap.Bool("billable", billable),
			zap.Bool("success", record.Success),
		)
	}
}

// NewQueryMeterRecordFromExecution builds a QueryMeterRecord from the typical
// values available at the end of a query console execution. This is a
// convenience constructor used by the ExecuteQuery handler.
func NewQueryMeterRecordFromExecution(
	backendType runtimev1.QueryBackendType,
	bytesScanned int64,
	executionTimeMs int64,
	userID string,
	projectID string,
	success bool,
) *QueryMeterRecord {
	return &QueryMeterRecord{
		BackendType:     backendType,
		BytesScanned:    bytesScanned,
		ExecutionTimeMs: executionTimeMs,
		UserID:          userID,
		ProjectID:       projectID,
		Timestamp:       time.Now(),
		Success:         success,
	}
}

// backendTypeString converts the proto enum to a human-readable meter dimension value.
func backendTypeString(bt runtimev1.QueryBackendType) string {
	switch bt {
	case runtimev1.QueryBackendType_QUERY_BACKEND_TYPE_EMBEDDED_ENGINE:
		return "embedded_engine"
	case runtimev1.QueryBackendType_QUERY_BACKEND_TYPE_EXTERNAL_WAREHOUSE:
		return "external_warehouse"
	default:
		return "unknown"
	}
}

// boolToMeterString converts a boolean to a string suitable for meter dimensions.
func boolToMeterString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
