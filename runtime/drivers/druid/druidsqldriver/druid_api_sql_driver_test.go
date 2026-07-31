package druidsqldriver

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

// testServer is a Druid SQL API stub that captures each request body,
// and responds with a minimal valid arrayLines result (header, types header, one row).
type testServer struct {
	*httptest.Server
	mu       sync.Mutex
	requests []DruidRequest
}

func newTestServer(t *testing.T) *testServer {
	t.Helper()
	ts := &testServer{}
	ts.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var dr DruidRequest
		if err := json.Unmarshal(body, &dr); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		ts.mu.Lock()
		ts.requests = append(ts.requests, dr)
		ts.mu.Unlock()

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("[\"n\"]\n[\"BIGINT\"]\n[1]\n"))
	}))
	t.Cleanup(ts.Close)
	return ts
}

// query runs a query against the stub server with the given config and returns the captured request's context.
func (ts *testServer) query(t *testing.T, queryCfg *QueryConfig) map[string]any {
	t.Helper()
	db, err := sql.Open("druid", ts.URL)
	require.NoError(t, err)
	defer db.Close()

	ctx := context.Background()
	if queryCfg != nil {
		ctx = WithQueryConfig(ctx, queryCfg)
	}

	rows, err := db.QueryContext(ctx, "SELECT 1")
	require.NoError(t, err)
	require.NoError(t, rows.Close())

	ts.mu.Lock()
	defer ts.mu.Unlock()
	require.NotEmpty(t, ts.requests)
	return ts.requests[len(ts.requests)-1].Context
}

func TestQueryContextDefaults(t *testing.T) {
	qctx := newTestServer(t).query(t, nil)

	require.NotEmpty(t, qctx["sqlQueryId"])
	require.Equal(t, true, qctx["enableTimeBoundaryPlanning"])
	// Optional keys must be omitted when unset.
	require.Len(t, qctx, 2)
}

func TestQueryContextConfig(t *testing.T) {
	useCache := false
	populateCache := true
	qctx := newTestServer(t).query(t, &QueryConfig{
		UseCache:      &useCache,
		PopulateCache: &populateCache,
		Priority:      3,
	})

	// A pointer to false must serialize as false, not be omitted.
	require.Equal(t, false, qctx["useCache"])
	require.Equal(t, true, qctx["populateCache"])
	require.Equal(t, float64(3), qctx["priority"])
}

func TestQueryContextAttributes(t *testing.T) {
	qctx := newTestServer(t).query(t, &QueryConfig{
		Priority: 3,
		Attributes: map[string]string{
			"userEmail": "user@example.com",
			// Keys set by the driver itself cannot be overridden.
			"sqlQueryId": "hijack",
			"priority":   "999",
		},
	})

	require.Equal(t, "user@example.com", qctx["userEmail"])
	require.NotEqual(t, "hijack", qctx["sqlQueryId"])
	require.Equal(t, float64(3), qctx["priority"])
}
