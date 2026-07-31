package druidsqldriver

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

// newTestServer returns a Druid SQL API stub that captures each request body into requests,
// and responds with a minimal valid arrayLines result (header, types header, one row).
func newTestServer(t *testing.T, requests *[]DruidRequest) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)

		var dr DruidRequest
		require.NoError(t, json.Unmarshal(body, &dr))
		*requests = append(*requests, dr)

		w.WriteHeader(http.StatusOK)
		_, err = w.Write([]byte("[\"n\"]\n[\"BIGINT\"]\n[1]\n"))
		require.NoError(t, err)
	}))
}

func TestQueryContext(t *testing.T) {
	var requests []DruidRequest
	srv := newTestServer(t, &requests)
	defer srv.Close()

	db, err := sql.Open("druid", srv.URL)
	require.NoError(t, err)
	defer db.Close()

	useCache := true
	ctx := WithQueryConfig(context.Background(), &QueryConfig{
		UseCache:     &useCache,
		Priority:     3,
		UserEmail:    "user@example.com",
		ServiceToken: "etl-bot",
	})

	rows, err := db.QueryContext(ctx, "SELECT 1")
	require.NoError(t, err)
	require.NoError(t, rows.Close())

	require.Len(t, requests, 1)
	qc := requests[0].Context
	require.NotEmpty(t, qc.SQLQueryID)
	require.True(t, qc.EnableTimeBoundaryPlanning)
	require.NotNil(t, qc.UseCache)
	require.True(t, *qc.UseCache)
	require.Nil(t, qc.PopulateCache)
	require.Equal(t, 3, qc.Priority)
	require.Equal(t, "user@example.com", qc.UserEmail)
	require.Equal(t, "etl-bot", qc.ServiceToken)
}

func TestQueryContextDefaults(t *testing.T) {
	var requests []DruidRequest
	srv := newTestServer(t, &requests)
	defer srv.Close()

	db, err := sql.Open("druid", srv.URL)
	require.NoError(t, err)
	defer db.Close()

	rows, err := db.QueryContext(context.Background(), "SELECT 1")
	require.NoError(t, err)
	require.NoError(t, rows.Close())

	require.Len(t, requests, 1)
	qc := requests[0].Context
	require.NotEmpty(t, qc.SQLQueryID)
	require.Nil(t, qc.UseCache)
	require.Nil(t, qc.PopulateCache)
	require.Zero(t, qc.Priority)
	require.Empty(t, qc.UserEmail)

	// Fields with omitempty must not be serialized when unset.
	b, err := json.Marshal(qc)
	require.NoError(t, err)
	var m map[string]any
	require.NoError(t, json.Unmarshal(b, &m))
	require.NotContains(t, m, "rillUserEmail")
	require.NotContains(t, m, "rillServiceToken")
	require.NotContains(t, m, "priority")
	require.NotContains(t, m, "useCache")
}
