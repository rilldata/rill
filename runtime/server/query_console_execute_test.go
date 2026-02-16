package server

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ---------- mock OLAP driver ----------

// mockOLAPRows implements drivers.Result for testing.
type mockOLAPRows struct {
	schema *runtimev1.StructType
	data   []*runtimev1.Struct
	idx    int
	closed bool
}

func (m *mockOLAPRows) Close() error {
	m.closed = true
	return nil
}

func (m *mockOLAPRows) Schema(ctx context.Context) (*runtimev1.StructType, error) {
	return m.schema, nil
}

func (m *mockOLAPRows) Next() (*runtimev1.Struct, error) {
	if m.idx >= len(m.data) {
		return nil, nil
	}
	row := m.data[m.idx]
	m.idx++
	return row, nil
}

// mockOLAPStore implements the minimal OLAPStore interface for query execution tests.
type mockOLAPStore struct {
	drivers.OLAPStore
	executeResult *mockOLAPRows
	executeErr    error
	executeSQL    string // captures last SQL for assertion
}

func (m *mockOLAPStore) Execute(ctx context.Context, stmt *drivers.Statement) (*drivers.Result, error) {
	m.executeSQL = stmt.Query
	if m.executeErr != nil {
		return nil, m.executeErr
	}
	result := &drivers.Result{
		Schema:         m.executeResult.schema,
		Rows:           m.executeResult,
	}
	return result, nil
}

// mockCostEstimatorOLAP is an OLAP store that also implements CostEstimator.
type mockCostEstimatorOLAP struct {
	mockOLAPStore
	costEstimate *drivers.CostEstimate
	costErr      error
}

func (m *mockCostEstimatorOLAP) EstimateQueryCost(ctx context.Context, sql string) (*drivers.CostEstimate, error) {
	if m.costErr != nil {
		return nil, m.costErr
	}
	return m.costEstimate, nil
}

// ---------- mock activity client ----------

type recordedEvent struct {
	eventType string
	dimensions map[string]string
}

type mockActivityClient struct {
	activity.Client
	events []recordedEvent
}

func (m *mockActivityClient) Record(ctx context.Context, eventType string, dimensions map[string]string) {
	m.events = append(m.events, recordedEvent{eventType: eventType, dimensions: dimensions})
}

func (m *mockActivityClient) RecordRaw(eventType string, dims map[string]string) {
	m.events = append(m.events, recordedEvent{eventType: eventType, dimensions: dims})
}

// ---------- helpers ----------

func makeStructRow(vals map[string]string) *runtimev1.Struct {
	fields := make(map[string]*runtimev1.Value, len(vals))
	for k, v := range vals {
		fields[k] = &runtimev1.Value{
			Kind: &runtimev1.Value_StringValue{StringValue: v},
		}
	}
	return &runtimev1.Struct{Fields: fields}
}

func makeSchema(cols ...string) *runtimev1.StructType {
	fields := make([]*runtimev1.StructType_Field, len(cols))
	for i, c := range cols {
		fields[i] = &runtimev1.StructType_Field{
			Name: c,
			Type: runtimev1.Type_CODE_STRING,
		}
	}
	return &runtimev1.StructType{Fields: fields}
}

func generateRows(n int) []*runtimev1.Struct {
	rows := make([]*runtimev1.Struct, n)
	for i := 0; i < n; i++ {
		rows[i] = makeStructRow(map[string]string{"id": fmt.Sprintf("%d", i)})
	}
	return rows
}

// buildServer creates a minimal Server struct suitable for testing ExecuteQuery.
// The caller provides the OLAP store and optionally an activity client.
func buildTestServer(olap drivers.OLAPStore, ac activity.Client, instanceVars map[string]string) *Server {
	logger, _ := zap.NewDevelopment()
	if ac == nil {
		ac = activity.NewNoopClient()
	}

	// We create a minimal runtime.Runtime wrapper via interface; the handler
	// only needs OLAPForInstance and InstanceVariables, so we use a shim.
	rt := &testRuntimeShim{
		olap:         olap,
		instanceVars: instanceVars,
	}

	s := &Server{
		runtime:  rt,
		logger:   logger,
		activity: ac,
	}
	return s
}

// testRuntimeShim satisfies the subset of the runtime.Runtime interface that
// the execute handler needs.
type testRuntimeShim struct {
	runtime.RuntimeInterface
	olap         drivers.OLAPStore
	instanceVars map[string]string
}

func (t *testRuntimeShim) OLAPForInstance(ctx context.Context, instanceID string) (drivers.OLAPStore, func(), error) {
	return t.olap, func() {}, nil
}

func (t *testRuntimeShim) InstanceVariables(ctx context.Context, instanceID string) (map[string]string, error) {
	return t.instanceVars, nil
}

func (t *testRuntimeShim) Resolve(ctx context.Context, instanceID string, connector string) (drivers.OLAPStore, func(), error) {
	return t.olap, func() {}, nil
}

// ---------- tests ----------

func TestExecuteQuery_Success(t *testing.T) {
	schema := makeSchema("id", "name")
	rows := []*runtimev1.Struct{
		makeStructRow(map[string]string{"id": "1", "name": "alice"}),
		makeStructRow(map[string]string{"id": "2", "name": "bob"}),
	}

	olap := &mockOLAPStore{
		executeResult: &mockOLAPRows{schema: schema, data: rows},
	}

	mockAC := &mockActivityClient{}
	s := buildTestServer(olap, mockAC, nil)

	resp, err := s.ExecuteQuery(context.Background(), &runtimev1.ExecuteQueryRequest{
		InstanceId: "test-instance",
		Sql:        "SELECT id, name FROM users",
	})

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, runtimev1.QueryStatus_QUERY_STATUS_SUCCESS, resp.Status)
	require.NotNil(t, resp.Result)
	require.Len(t, resp.Result.Columns, 2)
	require.Len(t, resp.Result.Rows, 2)
	require.False(t, resp.Result.Truncated)
	require.Greater(t, resp.ExecutionTimeMs, int64(0))
	require.Equal(t, "SELECT id, name FROM users", olap.executeSQL)
}

func TestExecuteQuery_EmptyResult(t *testing.T) {
	schema := makeSchema("col1")
	olap := &mockOLAPStore{
		executeResult: &mockOLAPRows{schema: schema, data: nil},
	}

	s := buildTestServer(olap, nil, nil)

	resp, err := s.ExecuteQuery(context.Background(), &runtimev1.ExecuteQueryRequest{
		InstanceId: "test-instance",
		Sql:        "SELECT col1 FROM empty_table",
	})

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, runtimev1.QueryStatus_QUERY_STATUS_SUCCESS, resp.Status)
	require.Len(t, resp.Result.Rows, 0)
	require.False(t, resp.Result.Truncated)
}

func TestExecuteQuery_DriverError(t *testing.T) {
	olap := &mockOLAPStore{
		executeErr: fmt.Errorf("syntax error at position 42"),
	}

	s := buildTestServer(olap, nil, nil)

	resp, err := s.ExecuteQuery(context.Background(), &runtimev1.ExecuteQueryRequest{
		InstanceId: "test-instance",
		Sql:        "SELEKT bad_sql",
	})

	// Implementation may return gRPC error or a response with FAILED status.
	// Accept either pattern.
	if err != nil {
		st, ok := status.FromError(err)
		require.True(t, ok)
		require.Contains(t, st.Message(), "syntax error")
	} else {
		require.Equal(t, runtimev1.QueryStatus_QUERY_STATUS_FAILED, resp.Status)
		require.Contains(t, resp.ErrorMessage, "syntax error")
	}
}

func TestExecuteQuery_EmptySQL(t *testing.T) {
	olap := &mockOLAPStore{}
	s := buildTestServer(olap, nil, nil)

	_, err := s.ExecuteQuery(context.Background(), &runtimev1.ExecuteQueryRequest{
		InstanceId: "test-instance",
		Sql:        "",
	})

	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.InvalidArgument, st.Code())
}

func TestExecuteQuery_MissingInstanceID(t *testing.T) {
	olap := &mockOLAPStore{}
	s := buildTestServer(olap, nil, nil)

	_, err := s.ExecuteQuery(context.Background(), &runtimev1.ExecuteQueryRequest{
		InstanceId: "",
		Sql:        "SELECT 1",
	})

	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.InvalidArgument, st.Code())
}

// ---------- result truncation tests ----------

func TestExecuteQuery_ResultTruncation(t *testing.T) {
	tests := []struct {
		name         string
		numRows      int
		rowLimit     int64
		wantRows     int
		wantTruncated bool
	}{
		{
			name:          "within_limit",
			numRows:       50,
			rowLimit:      100,
			wantRows:      50,
			wantTruncated: false,
		},
		{
			name:          "exactly_at_limit",
			numRows:       100,
			rowLimit:      100,
			wantRows:      100,
			wantTruncated: false,
		},
		{
			name:          "exceeds_limit",
			numRows:       200,
			rowLimit:      100,
			wantRows:      100,
			wantTruncated: true,
		},
		{
			name:          "default_limit_10000",
			numRows:       150,
			rowLimit:      0, // should use default
			wantRows:      150,
			wantTruncated: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := makeSchema("id")
			rows := generateRows(tt.numRows)

			olap := &mockOLAPStore{
				executeResult: &mockOLAPRows{schema: schema, data: rows},
			}

			s := buildTestServer(olap, nil, nil)

			req := &runtimev1.ExecuteQueryRequest{
				InstanceId: "test-instance",
				Sql:        "SELECT id FROM big_table",
			}
			if tt.rowLimit > 0 {
				req.RowLimit = tt.rowLimit
			}

			resp, err := s.ExecuteQuery(context.Background(), req)
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.Equal(t, runtimev1.QueryStatus_QUERY_STATUS_SUCCESS, resp.Status)
			require.Len(t, resp.Result.Rows, tt.wantRows)
			require.Equal(t, tt.wantTruncated, resp.Result.Truncated)
		})
	}
}

// ---------- guardrail: soft limit warning tests ----------

func TestExecuteQuery_SoftLimitWarning(t *testing.T) {
	schema := makeSchema("x")
	rows := []*runtimev1.Struct{makeStructRow(map[string]string{"x": "1"})}

	olap := &mockCostEstimatorOLAP{
		mockOLAPStore: mockOLAPStore{
			executeResult: &mockOLAPRows{schema: schema, data: rows},
		},
		costEstimate: &drivers.CostEstimate{
			BytesScanned: 600 * 1024 * 1024, // 600 MB — above a soft limit of 500 MB
			Supported:    true,
		},
	}

	// Set a soft limit of 500 MB in instance vars
	instVars := map[string]string{
		"query_console.soft_limit_bytes_scanned": fmt.Sprintf("%d", 500*1024*1024),
		"query_console.hard_limit_bytes_scanned": fmt.Sprintf("%d", 2*1024*1024*1024),
	}

	mockAC := &mockActivityClient{}
	s := buildTestServer(olap, mockAC, instVars)

	// First call: no confirmation → should get WARNING_COST
	resp, err := s.ExecuteQuery(context.Background(), &runtimev1.ExecuteQueryRequest{
		InstanceId:          "test-instance",
		Sql:                 "SELECT * FROM huge_table",
		ConfirmCostOverride: false,
	})

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, runtimev1.QueryStatus_QUERY_STATUS_WARNING_COST, resp.Status)
	// The query should NOT have been executed
	require.Nil(t, resp.Result)
	require.NotEmpty(t, resp.WarningMessage)
}

func TestExecuteQuery_SoftLimitWarning_ConfirmOverride(t *testing.T) {
	schema := makeSchema("x")
	rows := []*runtimev1.Struct{makeStructRow(map[string]string{"x": "1"})}

	olap := &mockCostEstimatorOLAP{
		mockOLAPStore: mockOLAPStore{
			executeResult: &mockOLAPRows{schema: schema, data: rows},
		},
		costEstimate: &drivers.CostEstimate{
			BytesScanned: 600 * 1024 * 1024,
			Supported:    true,
		},
	}

	instVars := map[string]string{
		"query_console.soft_limit_bytes_scanned": fmt.Sprintf("%d", 500*1024*1024),
		"query_console.hard_limit_bytes_scanned": fmt.Sprintf("%d", 2*1024*1024*1024),
	}

	s := buildTestServer(olap, nil, instVars)

	// With confirm_cost_override = true, should execute despite soft limit
	resp, err := s.ExecuteQuery(context.Background(), &runtimev1.ExecuteQueryRequest{
		InstanceId:          "test-instance",
		Sql:                 "SELECT * FROM huge_table",
		ConfirmCostOverride: true,
	})

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, runtimev1.QueryStatus_QUERY_STATUS_SUCCESS, resp.Status)
	require.NotNil(t, resp.Result)
	require.Len(t, resp.Result.Rows, 1)
}

// ---------- guardrail: hard limit blocking tests ----------

func TestExecuteQuery_HardLimitBlocked(t *testing.T) {
	schema := makeSchema("x")
	rows := []*runtimev1.Struct{makeStructRow(map[string]string{"x": "1"})}

	olap := &mockCostEstimatorOLAP{
		mockOLAPStore: mockOLAPStore{
			executeResult: &mockOLAPRows{schema: schema, data: rows},
		},
		costEstimate: &drivers.CostEstimate{
			BytesScanned: 5 * 1024 * 1024 * 1024, // 5 GB — above hard limit of 2 GB
			Supported:    true,
		},
	}

	instVars := map[string]string{
		"query_console.soft_limit_bytes_scanned": fmt.Sprintf("%d", 500*1024*1024),
		"query_console.hard_limit_bytes_scanned": fmt.Sprintf("%d", 2*1024*1024*1024),
	}

	mockAC := &mockActivityClient{}
	s := buildTestServer(olap, mockAC, instVars)

	// Even with confirm override true, hard limit cannot be bypassed
	resp, err := s.ExecuteQuery(context.Background(), &runtimev1.ExecuteQueryRequest{
		InstanceId:          "test-instance",
		Sql:                 "SELECT * FROM enormous_table",
		ConfirmCostOverride: true,
	})

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, runtimev1.QueryStatus_QUERY_STATUS_BLOCKED_LIMIT, resp.Status)
	require.Nil(t, resp.Result)
	require.NotEmpty(t, resp.ErrorMessage)
}

func TestExecuteQuery_HardLimitBlocked_OverrideIgnored(t *testing.T) {
	// Confirms that confirm_cost_override is irrelevant for hard limits
	olap := &mockCostEstimatorOLAP{
		mockOLAPStore: mockOLAPStore{
			executeResult: &mockOLAPRows{schema: makeSchema("a"), data: nil},
		},
		costEstimate: &drivers.CostEstimate{
			BytesScanned: 10 * 1024 * 1024 * 1024, // 10 GB
			Supported:    true,
		},
	}

	instVars := map[string]string{
		"query_console.hard_limit_bytes_scanned": fmt.Sprintf("%d", 1*1024*1024*1024), // 1 GB
	}

	for _, override := range []bool{false, true} {
		t.Run(fmt.Sprintf("confirm_%v", override), func(t *testing.T) {
			s := buildTestServer(olap, nil, instVars)
			resp, err := s.ExecuteQuery(context.Background(), &runtimev1.ExecuteQueryRequest{
				InstanceId:          "test-instance",
				Sql:                 "SELECT * FROM huge",
				ConfirmCostOverride: override,
			})

			require.NoError(t, err)
			require.Equal(t, runtimev1.QueryStatus_QUERY_STATUS_BLOCKED_LIMIT, resp.Status)
		})
	}
}

// ---------- cost estimation not supported ----------

func TestExecuteQuery_NoCostEstimator_SkipsGuardrails(t *testing.T) {
	// Standard mockOLAPStore does not implement CostEstimator.
	// Guardrails should be skipped and query executed normally.
	schema := makeSchema("val")
	rows := []*runtimev1.Struct{makeStructRow(map[string]string{"val": "hello"})}

	olap := &mockOLAPStore{
		executeResult: &mockOLAPRows{schema: schema, data: rows},
	}

	instVars := map[string]string{
		"query_console.soft_limit_bytes_scanned": "1",  // Extremely low limit
		"query_console.hard_limit_bytes_scanned": "10", // Extremely low limit
	}

	s := buildTestServer(olap, nil, instVars)

	resp, err := s.ExecuteQuery(context.Background(), &runtimev1.ExecuteQueryRequest{
		InstanceId: "test-instance",
		Sql:        "SELECT val FROM t",
	})

	// Should succeed because driver doesn't support cost estimation
	require.NoError(t, err)
	require.Equal(t, runtimev1.QueryStatus_QUERY_STATUS_SUCCESS, resp.Status)
	require.Len(t, resp.Result.Rows, 1)
}

func TestExecuteQuery_CostEstimationError_ProceedsWithExecution(t *testing.T) {
	// If cost estimation fails, we should still execute the query (fail-open)
	schema := makeSchema("v")
	rows := []*runtimev1.Struct{makeStructRow(map[string]string{"v": "ok"})}

	olap := &mockCostEstimatorOLAP{
		mockOLAPStore: mockOLAPStore{
			executeResult: &mockOLAPRows{schema: schema, data: rows},
		},
		costErr: fmt.Errorf("estimation service unavailable"),
	}

	instVars := map[string]string{
		"query_console.soft_limit_bytes_scanned": "1",
	}

	s := buildTestServer(olap, nil, instVars)

	resp, err := s.ExecuteQuery(context.Background(), &runtimev1.ExecuteQueryRequest{
		InstanceId: "test-instance",
		Sql:        "SELECT v FROM t",
	})

	require.NoError(t, err)
	require.Equal(t, runtimev1.QueryStatus_QUERY_STATUS_SUCCESS, resp.Status)
	require.Len(t, resp.Result.Rows, 1)
}

func TestExecuteQuery_CostEstimateNotSupported_SkipsGuardrails(t *testing.T) {
	// Driver implements CostEstimator but returns Supported=false
	schema := makeSchema("z")
	rows := []*runtimev1.Struct{makeStructRow(map[string]string{"z": "42"})}

	olap := &mockCostEstimatorOLAP{
		mockOLAPStore: mockOLAPStore{
			executeResult: &mockOLAPRows{schema: schema, data: rows},
		},
		costEstimate: &drivers.CostEstimate{
			Supported: false,
		},
	}

	instVars := map[string]string{
		"query_console.soft_limit_bytes_scanned": "1",
		"query_console.hard_limit_bytes_scanned": "1",
	}

	s := buildTestServer(olap, nil, instVars)

	resp, err := s.ExecuteQuery(context.Background(), &runtimev1.ExecuteQueryRequest{
		InstanceId: "test-instance",
		Sql:        "SELECT z FROM t",
	})

	require.NoError(t, err)
	require.Equal(t, runtimev1.QueryStatus_QUERY_STATUS_SUCCESS, resp.Status)
}

// ---------- guardrails: edge case with thresholds ----------

func TestExecuteQuery_BelowSoftLimit_Executes(t *testing.T) {
	schema := makeSchema("a")
	rows := []*runtimev1.Struct{makeStructRow(map[string]string{"a": "1"})}

	olap := &mockCostEstimatorOLAP{
		mockOLAPStore: mockOLAPStore{
			executeResult: &mockOLAPRows{schema: schema, data: rows},
		},
		costEstimate: &drivers.CostEstimate{
			BytesScanned: 100 * 1024 * 1024, // 100 MB
			Supported:    true,
		},
	}

	instVars := map[string]string{
		"query_console.soft_limit_bytes_scanned": fmt.Sprintf("%d", 500*1024*1024),
		"query_console.hard_limit_bytes_scanned": fmt.Sprintf("%d", 2*1024*1024*1024),
	}

	s := buildTestServer(olap, nil, instVars)

	resp, err := s.ExecuteQuery(context.Background(), &runtimev1.ExecuteQueryRequest{
		InstanceId: "test-instance",
		Sql:        "SELECT a FROM small_table",
	})

	require.NoError(t, err)
	require.Equal(t, runtimev1.QueryStatus_QUERY_STATUS_SUCCESS, resp.Status)
	require.NotNil(t, resp.Result)
}

func TestExecuteQuery_ExactlyAtSoftLimit(t *testing.T) {
	schema := makeSchema("a")
	rows := []*runtimev1.Struct{makeStructRow(map[string]string{"a": "1"})}

	softLimit := int64(500 * 1024 * 1024)

	olap := &mockCostEstimatorOLAP{
		mockOLAPStore: mockOLAPStore{
			executeResult: &mockOLAPRows{schema: schema, data: rows},
		},
		costEstimate: &drivers.CostEstimate{
			BytesScanned: softLimit, // exactly at soft limit
			Supported:    true,
		},
	}

	instVars := map[string]string{
		"query_console.soft_limit_bytes_scanned": fmt.Sprintf("%d", softLimit),
		"query_console.hard_limit_bytes_scanned": fmt.Sprintf("%d", 2*softLimit),
	}

	s := buildTestServer(olap, nil, instVars)

	resp, err := s.ExecuteQuery(context.Background(), &runtimev1.ExecuteQueryRequest{
		InstanceId: "test-instance",
		Sql:        "SELECT a FROM t",
	})

	// At exactly the threshold: implementation may warn or allow.
	// We accept either SUCCESS or WARNING_COST.
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.True(t,
		resp.Status == runtimev1.QueryStatus_QUERY_STATUS_SUCCESS ||
			resp.Status == runtimev1.QueryStatus_QUERY_STATUS_WARNING_COST,
		"expected SUCCESS or WARNING_COST, got %v", resp.Status,
	)
}

// ---------- telemetry event emission tests ----------

func TestExecuteQuery_Telemetry_SuccessEvent(t *testing.T) {
	schema := makeSchema("c")
	rows := []*runtimev1.Struct{makeStructRow(map[string]string{"c": "1"})}

	olap := &mockOLAPStore{
		executeResult: &mockOLAPRows{schema: schema, data: rows},
	}

	mockAC := &mockActivityClient{}
	s := buildTestServer(olap, mockAC, nil)

	_, err := s.ExecuteQuery(context.Background(), &runtimev1.ExecuteQueryRequest{
		InstanceId: "test-instance",
		Sql:        "SELECT c FROM t",
	})
	require.NoError(t, err)

	// Verify at least one telemetry event was recorded
	require.NotEmpty(t, mockAC.events, "expected telemetry event to be emitted")

	// Find the query execution event
	var found bool
	for _, ev := range mockAC.events {
		if strings.Contains(ev.eventType, "query_executed") || strings.Contains(ev.eventType, "query") {
			found = true
			// Verify key dimensions are present
			require.Contains(t, ev.dimensions, "instance_id")
			break
		}
	}
	// If no specific query event found, just verify events were emitted
	if !found {
		require.NotEmpty(t, mockAC.events, "expected at least one telemetry event")
	}
}

func TestExecuteQuery_Telemetry_WarningEvent(t *testing.T) {
	olap := &mockCostEstimatorOLAP{
		mockOLAPStore: mockOLAPStore{
			executeResult: &mockOLAPRows{schema: makeSchema("x"), data: nil},
		},
		costEstimate: &drivers.CostEstimate{
			BytesScanned: 600 * 1024 * 1024,
			Supported:    true,
		},
	}

	instVars := map[string]string{
		"query_console.soft_limit_bytes_scanned": fmt.Sprintf("%d", 500*1024*1024),
		"query_console.hard_limit_bytes_scanned": fmt.Sprintf("%d", 2*1024*1024*1024),
	}

	mockAC := &mockActivityClient{}
	s := buildTestServer(olap, mockAC, instVars)

	resp, err := s.ExecuteQuery(context.Background(), &runtimev1.ExecuteQueryRequest{
		InstanceId: "test-instance",
		Sql:        "SELECT * FROM big_table",
	})
	require.NoError(t, err)
	require.Equal(t, runtimev1.QueryStatus_QUERY_STATUS_WARNING_COST, resp.Status)

	// Should have emitted a warning telemetry event
	require.NotEmpty(t, mockAC.events, "expected telemetry events for soft limit warning")
}

func TestExecuteQuery_Telemetry_BlockedEvent(t *testing.T) {
	olap := &mockCostEstimatorOLAP{
		mockOLAPStore: mockOLAPStore{
			executeResult: &mockOLAPRows{schema: makeSchema("x"), data: nil},
		},
		costEstimate: &drivers.CostEstimate{
			BytesScanned: 10 * 1024 * 1024 * 1024, // 10 GB
			Supported:    true,
		},
	}

	instVars := map[string]string{
		"query_console.hard_limit_bytes_scanned": fmt.Sprintf("%d", 1*1024*1024*1024),
	}

	mockAC := &mockActivityClient{}
	s := buildTestServer(olap, mockAC, instVars)

	resp, err := s.ExecuteQuery(context.Background(), &runtimev1.ExecuteQueryRequest{
		InstanceId: "test-instance",
		Sql:        "SELECT * FROM enormous_table",
	})
	require.NoError(t, err)
	require.Equal(t, runtimev1.QueryStatus_QUERY_STATUS_BLOCKED_LIMIT, resp.Status)

	// Should have emitted a blocked telemetry event
	require.NotEmpty(t, mockAC.events, "expected telemetry events for hard limit block")
}

func TestExecuteQuery_Telemetry_FailedEvent(t *testing.T) {
	olap := &mockOLAPStore{
		executeErr: fmt.Errorf("relation does not exist"),
	}

	mockAC := &mockActivityClient{}
	s := buildTestServer(olap, mockAC, nil)

	resp, err := s.ExecuteQuery(context.Background(), &runtimev1.ExecuteQueryRequest{
		InstanceId: "test-instance",
		Sql:        "SELECT * FROM nonexistent",
	})

	// Accept either gRPC error or response with FAILED status
	if err == nil {
		require.Equal(t, runtimev1.QueryStatus_QUERY_STATUS_FAILED, resp.Status)
	}

	// Should have emitted failure telemetry
	require.NotEmpty(t, mockAC.events, "expected telemetry events for failed query")
}

// ---------- execution timing test ----------

func TestExecuteQuery_ExecutionTimeTracked(t *testing.T) {
	schema := makeSchema("x")
	rows := []*runtimev1.Struct{makeStructRow(map[string]string{"x": "1"})}

	olap := &mockOLAPStore{
		executeResult: &mockOLAPRows{schema: schema, data: rows},
	}

	s := buildTestServer(olap, nil, nil)

	start := time.Now()
	resp, err := s.ExecuteQuery(context.Background(), &runtimev1.ExecuteQueryRequest{
		InstanceId: "test-instance",
		Sql:        "SELECT x FROM t",
	})
	elapsed := time.Since(start)

	require.NoError(t, err)
	require.NotNil(t, resp)
	// Execution time should be non-negative and reasonable
	require.GreaterOrEqual(t, resp.ExecutionTimeMs, int64(0))
	require.LessOrEqual(t, resp.ExecutionTimeMs, elapsed.Milliseconds()+100) // allow some slack
}

// ---------- context cancellation test ----------

func TestExecuteQuery_ContextCancelled(t *testing.T) {
	// Simulates a cancelled context
	olap := &mockOLAPStore{
		executeErr: context.Canceled,
	}

	s := buildTestServer(olap, nil, nil)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	_, err := s.ExecuteQuery(ctx, &runtimev1.ExecuteQueryRequest{
		InstanceId: "test-instance",
		Sql:        "SELECT 1",
	})

	// Should get an error (either context cancelled or gRPC cancelled)
	require.Error(t, err)
}

// ---------- no guardrails configured ----------

func TestExecuteQuery_NoGuardrailsConfigured(t *testing.T) {
	schema := makeSchema("x")
	rows := []*runtimev1.Struct{makeStructRow(map[string]string{"x": "1"})}

	olap := &mockCostEstimatorOLAP{
		mockOLAPStore: mockOLAPStore{
			executeResult: &mockOLAPRows{schema: schema, data: rows},
		},
		costEstimate: &drivers.CostEstimate{
			BytesScanned: 999 * 1024 * 1024 * 1024, // Huge amount
			Supported:    true,
		},
	}

	// No instance variables → no guardrails configured → use defaults
	s := buildTestServer(olap, nil, nil)

	resp, err := s.ExecuteQuery(context.Background(), &runtimev1.ExecuteQueryRequest{
		InstanceId: "test-instance",
		Sql:        "SELECT x FROM t",
	})

	// Behavior depends on default thresholds. If defaults are permissive or zero,
	// the query should execute. If defaults are set, it may warn/block.
	// We just verify no panic and a valid response.
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.True(t,
		resp.Status == runtimev1.QueryStatus_QUERY_STATUS_SUCCESS ||
			resp.Status == runtimev1.QueryStatus_QUERY_STATUS_WARNING_COST ||
			resp.Status == runtimev1.QueryStatus_QUERY_STATUS_BLOCKED_LIMIT,
	)
}
