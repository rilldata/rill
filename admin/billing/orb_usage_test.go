package billing

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/orbcorp/orb-go"
	"github.com/orbcorp/orb-go/option"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestOrbReportUsageBatchBoundaries(t *testing.T) {
	// Orb accepts at most 500 events per ingestion request. These exact boundaries
	// guard against empty requests, dropped remainders, and accidental 501-event batches.
	tests := []struct {
		count       int
		wantBatches []int
	}{
		{count: 0, wantBatches: nil},
		{count: 1, wantBatches: []int{1}},
		{count: 500, wantBatches: []int{500}},
		{count: 501, wantBatches: []int{500, 1}},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%d events", tt.count), func(t *testing.T) {
			server := newOrbUsageServer(t, func(_ int, _ orbUsageRequest) (int, string) {
				return http.StatusOK, `{"validation_failed":[]}`
			})
			biller := newTestOrbUsageBiller(server.server.URL)

			err := biller.ReportUsage(t.Context(), makeOrbUsageFixtures(tt.count))
			require.NoError(t, err)
			require.Equal(t, tt.wantBatches, server.batchSizes())
		})
	}
}

func TestOrbReportUsageSerializesStableLogicalEvent(t *testing.T) {
	// Retries and repeated checkpoints must serialize the same logical bucket to
	// the same key, regardless of metadata map insertion order.
	server := newOrbUsageServer(t, func(_ int, _ orbUsageRequest) (int, string) {
		return http.StatusOK, `{"validation_failed":[]}`
	})
	biller := newTestOrbUsageBiller(server.server.URL)
	end := time.Date(2026, time.January, 2, 4, 0, 0, 0, time.UTC)
	first := &Usage{
		CustomerID: "org-1", MetricName: "api_calls", Value: 42,
		ReportingGrain: UsageReportingGranularityHour, EndTime: end,
		Metadata: map[string]interface{}{"region": "us-east", "tier": "pro"},
	}
	second := &Usage{
		CustomerID: "org-1", MetricName: "api_calls", Value: 42,
		ReportingGrain: UsageReportingGranularityHour, EndTime: end,
		Metadata: map[string]interface{}{"tier": "pro", "region": "us-east"},
	}

	require.NoError(t, biller.ReportUsage(t.Context(), []*Usage{first}))
	require.NoError(t, biller.ReportUsage(t.Context(), []*Usage{second}))
	requests := server.requestsSnapshot()
	require.Len(t, requests, 2)
	require.Len(t, requests[0].Events, 1)
	require.Equal(t, requests[0].Events[0].IdempotencyKey, requests[1].Events[0].IdempotencyKey)
	require.Equal(t, "api_calls_hour", requests[0].Events[0].EventName)
	require.Equal(t, "org-1", requests[0].Events[0].ExternalCustomerID)
	require.Equal(t, end.Add(-time.Second), requests[0].Events[0].Timestamp)
	require.EqualValues(t, 42, requests[0].Events[0].Properties["amount"])
	require.Equal(t, "us-east", requests[0].Events[0].Properties["region"])
}

func TestOrbReportUsageCorrectionKeepsLogicalIdempotencyKey(t *testing.T) {
	// A corrected value represents the same customer/metric/time bucket. Keeping
	// its identity stable prevents a correction retry from becoming an additive event.
	server := newOrbUsageServer(t, func(_ int, _ orbUsageRequest) (int, string) {
		return http.StatusOK, `{"validation_failed":[]}`
	})
	biller := newTestOrbUsageBiller(server.server.URL)
	fixtures := makeOrbUsageFixtures(1)
	require.NoError(t, biller.ReportUsage(t.Context(), fixtures))
	fixtures[0].Value = 99
	require.NoError(t, biller.ReportUsage(t.Context(), fixtures))

	requests := server.requestsSnapshot()
	require.Equal(t, requests[0].Events[0].IdempotencyKey, requests[1].Events[0].IdempotencyKey)
	require.EqualValues(t, 1, requests[0].Events[0].Properties["amount"])
	require.EqualValues(t, 99, requests[1].Events[0].Properties["amount"])
}

func TestOrbReportUsageRetriesOnlyTransientProviderFailures(t *testing.T) {
	// The generated Orb client is configured without its own retry loop here so
	// the application retry classifier and its exact request count are observable.
	tests := []struct {
		name         string
		firstStatus  int
		wantRequests int
		wantErr      bool
	}{
		{name: "500 retries", firstStatus: http.StatusInternalServerError, wantRequests: 2},
		{name: "429 retries", firstStatus: http.StatusTooManyRequests, wantRequests: 2},
		{name: "400 fails immediately", firstStatus: http.StatusBadRequest, wantRequests: 1, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := newOrbUsageServer(t, func(attempt int, _ orbUsageRequest) (int, string) {
				if attempt == 1 {
					return tt.firstStatus, orbUsageErrorBody(tt.firstStatus)
				}
				return http.StatusOK, `{"validation_failed":[]}`
			})
			biller := newTestOrbUsageBiller(server.server.URL)

			err := biller.ReportUsage(t.Context(), makeOrbUsageFixtures(1))
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, tt.wantRequests, len(server.requestsSnapshot()))
		})
	}
}

func orbUsageErrorBody(status int) string {
	return fmt.Sprintf(`{"status":%d,"title":"provider failure","type":"test_error","validation_errors":[],"detail":"provider failure"}`, status)
}

func TestOrbReportUsageReturnsRecordValidationDetails(t *testing.T) {
	// A 200 response may still reject individual records; the error must identify
	// their idempotency keys so operators can repair the correct usage checkpoint.
	server := newOrbUsageServer(t, func(_ int, request orbUsageRequest) (int, string) {
		return http.StatusOK, fmt.Sprintf(`{"validation_failed":[{"idempotency_key":%q,"validation_errors":["amount must be positive"]}]}`, request.Events[0].IdempotencyKey)
	})
	biller := newTestOrbUsageBiller(server.server.URL)

	err := biller.ReportUsage(t.Context(), makeOrbUsageFixtures(1))
	require.ErrorContains(t, err, "validation failure for 1 events")
	require.ErrorContains(t, err, "amount must be positive")
	require.ErrorContains(t, err, "org-0")
	require.Len(t, server.requestsSnapshot(), 1)
}

type orbUsageRequest struct {
	Events []struct {
		EventName          string                 `json:"event_name"`
		IdempotencyKey     string                 `json:"idempotency_key"`
		ExternalCustomerID string                 `json:"external_customer_id"`
		Timestamp          time.Time              `json:"timestamp"`
		Properties         map[string]interface{} `json:"properties"`
	} `json:"events"`
}

type orbUsageServer struct {
	server *httptest.Server

	mu       sync.Mutex
	requests []orbUsageRequest
}

func newOrbUsageServer(t *testing.T, response func(attempt int, request orbUsageRequest) (int, string)) *orbUsageServer {
	t.Helper()
	fixture := &orbUsageServer{}
	fixture.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/ingest", r.URL.Path)
		var request orbUsageRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&request))
		fixture.mu.Lock()
		fixture.requests = append(fixture.requests, request)
		attempt := len(fixture.requests)
		fixture.mu.Unlock()
		status, body := response(attempt, request)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(fixture.server.Close)
	return fixture
}

func (s *orbUsageServer) requestsSnapshot() []orbUsageRequest {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]orbUsageRequest(nil), s.requests...)
}

func (s *orbUsageServer) batchSizes() []int {
	requests := s.requestsSnapshot()
	if len(requests) == 0 {
		return nil
	}
	res := make([]int, len(requests))
	for i, request := range requests {
		res[i] = len(request.Events)
	}
	return res
}

func newTestOrbUsageBiller(baseURL string) *Orb {
	client := orb.NewClient(
		option.WithAPIKey("test-key"),
		option.WithBaseURL(baseURL),
		option.WithMaxRetries(0),
	)
	return &Orb{
		client:            client,
		logger:            zap.NewNop(),
		usageRetryBackoff: []time.Duration{0, 0, 0, 0, 0},
	}
}

func makeOrbUsageFixtures(count int) []*Usage {
	usage := make([]*Usage, count)
	for i := range usage {
		usage[i] = &Usage{
			CustomerID:     fmt.Sprintf("org-%d", i),
			MetricName:     "api_calls",
			Value:          float64(i + 1),
			ReportingGrain: UsageReportingGranularityHour,
			StartTime:      time.Date(2026, time.January, 2, 3, 0, 0, 0, time.UTC),
			EndTime:        time.Date(2026, time.January, 2, 4, 0, 0, 0, time.UTC),
			Metadata:       map[string]interface{}{"region": "us-east"},
		}
	}
	return usage
}
