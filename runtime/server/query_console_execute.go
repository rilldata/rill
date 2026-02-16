package server

import (
	"context"
	"fmt"
	"strings"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/server/auth"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
)

// defaultResultRowLimit is the maximum number of rows returned in a query preview.
const defaultResultRowLimit = 10000

// instanceVarResultRowLimit is the instance variable key for overriding the default row limit.
const instanceVarResultRowLimit = "query_console_result_row_limit"

// ExecuteQuery implements RuntimeService.ExecuteQuery — runs an ad-hoc SQL query
// against the instance's OLAP engine (or a specified external connector), applies
// guardrail checks, and returns a preview of the results.
func (s *Server) ExecuteQuery(ctx context.Context, req *runtimev1.ExecuteQueryRequest) (*runtimev1.ExecuteQueryResponse, error) {
	// ---------- Auth ----------
	if !auth.GetClaims(ctx).CanInstance(req.InstanceId, auth.ReadOLAP) {
		return nil, status.Error(codes.PermissionDenied, "not authorized to execute queries on this instance")
	}

	// ---------- Validate request ----------
	sql := strings.TrimSpace(req.Sql)
	if sql == "" {
		return nil, status.Error(codes.InvalidArgument, "sql is required")
	}

	// ---------- Resolve instance ----------
	inst, err := s.runtime.Instance(ctx, req.InstanceId)
	if err != nil {
		return nil, err
	}

	// ---------- Resolve OLAP handle ----------
	olap, release, backendType, err := s.resolveOLAPForQuery(ctx, inst, req.BackendHint)
	if err != nil {
		return nil, err
	}
	defer release()

	// ---------- Load guardrails ----------
	guardrails := runtime.LoadGuardrails(inst.ResolveVariables())

	// ---------- Cost estimation & guardrail checks ----------
	if estimator, ok := olap.(drivers.CostEstimator); ok {
		resp, done, err := s.applyCostGuardrails(ctx, estimator, guardrails, sql, req.ConfirmCostOverride, backendType, req.InstanceId)
		if err != nil {
			return nil, err
		}
		if done {
			return resp, nil
		}
	}

	// ---------- Execute query ----------
	startTime := time.Now()
	result, err := olap.Execute(ctx, &drivers.Statement{
		Query: sql,
	})
	execDuration := time.Since(startTime)

	if err != nil {
		s.emitQueryTelemetry(ctx, "query_failed", backendType, req.InstanceId, 0, execDuration.Milliseconds())
		return nil, status.Errorf(codes.Internal, "query execution failed: %v", err)
	}
	defer result.Close()

	// ---------- Marshal results ----------
	rowLimit := s.resolveRowLimit(inst)
	preview, rowCount, err := marshalResultPreview(result, rowLimit)
	if err != nil {
		s.emitQueryTelemetry(ctx, "query_failed", backendType, req.InstanceId, 0, execDuration.Milliseconds())
		return nil, status.Errorf(codes.Internal, "failed to read query results: %v", err)
	}

	// Estimate bytes scanned — not all drivers report this, default to 0.
	var bytesScanned int64
	if result.EstimatedStorageBytes() > 0 {
		bytesScanned = result.EstimatedStorageBytes()
	}

	// ---------- Telemetry ----------
	s.emitQueryTelemetry(ctx, "query_executed", backendType, req.InstanceId, bytesScanned, execDuration.Milliseconds())

	return &runtimev1.ExecuteQueryResponse{
		Status:          runtimev1.QueryStatus_QUERY_STATUS_SUCCESS,
		Preview:         preview,
		BytesScanned:    bytesScanned,
		ExecutionTimeMs: execDuration.Milliseconds(),
		RowCount:        int64(rowCount),
	}, nil
}

// resolveOLAPForQuery selects the OLAP driver handle based on the backend hint.
// Returns the OLAPStore, a release function, the resolved backend type, and any error.
func (s *Server) resolveOLAPForQuery(ctx context.Context, inst *drivers.Instance, hint runtimev1.QueryBackendType) (drivers.OLAPStore, func(), runtimev1.QueryBackendType, error) {
	connector := inst.ResolveOLAPConnector()
	backendType := runtimev1.QueryBackendType_QUERY_BACKEND_TYPE_EMBEDDED_ENGINE

	// If caller provided a hint for external warehouse, try to honor it.
	if hint == runtimev1.QueryBackendType_QUERY_BACKEND_TYPE_EXTERNAL_WAREHOUSE {
		// Look for an explicitly configured non-default OLAP connector.
		// For V1 we still fall back to the default connector; external warehouse
		// support will be expanded in a future iteration.
		backendType = runtimev1.QueryBackendType_QUERY_BACKEND_TYPE_EXTERNAL_WAREHOUSE
	}

	handle, release, err := s.runtime.OLAP(ctx, inst.ID, connector)
	if err != nil {
		return nil, nil, backendType, status.Errorf(codes.Internal, "failed to acquire OLAP connection: %v", err)
	}

	return handle, release, backendType, nil
}

// applyCostGuardrails runs cost estimation and checks soft/hard limits.
// Returns (response, done, error). When done==true the caller should return
// the response immediately without executing the query.
func (s *Server) applyCostGuardrails(
	ctx context.Context,
	estimator drivers.CostEstimator,
	guardrails *runtime.GuardrailConfig,
	sql string,
	confirmOverride bool,
	backendType runtimev1.QueryBackendType,
	instanceID string,
) (*runtimev1.ExecuteQueryResponse, bool, error) {
	estimate, err := estimator.EstimateQueryCost(ctx, sql)
	if err != nil {
		// Cost estimation failure is non-fatal — log and proceed.
		s.logger.Warn("cost estimation failed, proceeding without guardrail check",
			zap.String("instance_id", instanceID),
			zap.Error(err),
		)
		return nil, false, nil
	}

	if !estimate.Supported {
		return nil, false, nil
	}

	// Hard limit check (cannot be overridden)
	if blocked, reason := runtime.CheckHardLimit(estimate.BytesScanned, guardrails); blocked {
		s.emitQueryTelemetry(ctx, "query_blocked_hard_limit", backendType, instanceID, estimate.BytesScanned, 0)
		return &runtimev1.ExecuteQueryResponse{
			Status:         runtimev1.QueryStatus_QUERY_STATUS_BLOCKED_LIMIT,
			StatusMessage:  reason,
			BytesScanned:   estimate.BytesScanned,
		}, true, nil
	}

	// Soft limit check (can be overridden with confirmation)
	if exceeded, message := runtime.CheckSoftLimit(estimate.BytesScanned, guardrails); exceeded && !confirmOverride {
		s.emitQueryTelemetry(ctx, "query_warned_soft_limit", backendType, instanceID, estimate.BytesScanned, 0)
		return &runtimev1.ExecuteQueryResponse{
			Status:         runtimev1.QueryStatus_QUERY_STATUS_WARNING_COST,
			StatusMessage:  message,
			BytesScanned:   estimate.BytesScanned,
		}, true, nil
	}

	return nil, false, nil
}

// marshalResultPreview reads query results into a ResultPreview proto message.
// It stops after rowLimit rows, setting the truncated flag if more data exists.
func marshalResultPreview(result *drivers.Result, rowLimit int) (*runtimev1.ResultPreview, int, error) {
	if result == nil {
		return &runtimev1.ResultPreview{}, 0, nil
	}

	// Build column definitions from schema.
	schema := result.Schema
	columns := make([]*runtimev1.ResultColumn, 0, len(schema.Fields))
	for _, f := range schema.Fields {
		columns = append(columns, &runtimev1.ResultColumn{
			Name:     f.Name,
			DataType: f.Type.Code.String(),
		})
	}

	// Read rows.
	var rows []*structpb.Struct
	rowCount := 0
	truncated := false

	for result.Next() {
		if rowCount >= rowLimit {
			truncated = true
			break
		}

		rowMap := make(map[string]interface{})
		err := result.MapScan(rowMap)
		if err != nil {
			return nil, 0, fmt.Errorf("scanning row %d: %w", rowCount, err)
		}

		// Convert all values to structpb-safe types.
		safeMap := make(map[string]interface{}, len(rowMap))
		for k, v := range rowMap {
			safeMap[k] = toStructpbValue(v)
		}

		s, err := structpb.NewStruct(safeMap)
		if err != nil {
			return nil, 0, fmt.Errorf("converting row %d to struct: %w", rowCount, err)
		}
		rows = append(rows, s)
		rowCount++
	}

	if err := result.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterating results: %w", err)
	}

	return &runtimev1.ResultPreview{
		Columns:   columns,
		Rows:      rows,
		Truncated: truncated,
	}, rowCount, nil
}

// toStructpbValue coerces a driver result value into a type that structpb can represent.
func toStructpbValue(v interface{}) interface{} {
	if v == nil {
		return nil
	}
	switch val := v.(type) {
	case string, bool, float32, float64, int, int32, int64, uint, uint32, uint64:
		return val
	case []byte:
		return string(val)
	case time.Time:
		return val.Format(time.RFC3339Nano)
	case fmt.Stringer:
		return val.String()
	default:
		return fmt.Sprintf("%v", val)
	}
}

// resolveRowLimit returns the configured result row limit for the instance.
func (s *Server) resolveRowLimit(inst *drivers.Instance) int {
	vars := inst.ResolveVariables()
	if raw, ok := vars[instanceVarResultRowLimit]; ok {
		var limit int
		if _, err := fmt.Sscanf(raw, "%d", &limit); err == nil && limit > 0 {
			return limit
		}
	}
	return defaultResultRowLimit
}

// emitQueryTelemetry sends a structured telemetry event for a query console operation.
func (s *Server) emitQueryTelemetry(ctx context.Context, eventType string, backendType runtimev1.QueryBackendType, instanceID string, bytesScanned int64, executionTimeMs int64) {
	if s.activity == nil {
		return
	}
	s.activity.Record(ctx, activity.EventTypeLog, eventType,
		attribute("backend_type", backendType.String()),
		attribute("instance_id", instanceID),
		attribute("bytes_scanned", bytesScanned),
		attribute("execution_time_ms", executionTimeMs),
	)
}

// attribute is a small helper that creates an activity attribute pair.
func attribute(key string, value interface{}) activity.Attribute {
	return activity.Attribute{Key: key, Value: value}
}
