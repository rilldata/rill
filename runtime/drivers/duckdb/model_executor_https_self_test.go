package duckdb

import (
	"context"
	"encoding/csv"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rilldata/rill/runtime/drivers"
	activity "github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/storage"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	_ "github.com/rilldata/rill/runtime/drivers/https"
)

// CSV Data used in the test
var testCSVData = [][]string{
	{"Name", "Age", "City"},
	{"Alice", "30", "New York"},
	{"Bob", "25", "Los Angeles"},
	{"Charlie", "35", "Chicago"},
}

// CSV Handler function for the test server
func csvHandler(w http.ResponseWriter, r *http.Request) {
	apiKey := r.Header.Get("X-API-Key")
	if apiKey != "test-key" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=data.csv")

	writer := csv.NewWriter(w)
	defer writer.Flush()
	for _, row := range testCSVData {
		writer.Write(row)
	}
}

func TestHTTPToDuckDBTransfer(t *testing.T) {
	// Create a test server using the handler
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/data.csv" {
			http.NotFound(w, r)
			return
		}
		csvHandler(w, r)
	}))
	defer server.Close() // Ensure server shuts down after test

	to, err := drivers.Open("duckdb", "default", map[string]any{}, storage.MustNew(t.TempDir(), nil), activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)

	inputHandle, err := drivers.Open("https", "default", map[string]any{}, storage.MustNew(t.TempDir(), nil), activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)

	opts := &drivers.ModelExecutorOptions{
		InputHandle:     inputHandle,
		InputConnector:  "https",
		OutputHandle:    to,
		OutputConnector: "duckdb",
		Env: &drivers.ModelEnv{
			AllowHostAccess: false,
			StageChanges:    true,
			AcquireConnector: func(ctx context.Context, name string) (drivers.Handle, func(), error) {
				if name == "https" {
					return inputHandle, func() {}, nil
				}
				return nil, nil, fmt.Errorf("unsupported name: %s", name)
			},
		},
		PreliminaryInputProperties: map[string]any{
			"path":    server.URL + "/data.csv",
			"headers": map[string]any{"X-API-Key": "test-key"},
		},
		PreliminaryOutputProperties: map[string]any{
			"table": "sink",
		},
	}

	me, err := to.AsModelExecutor("default", opts)
	require.NoError(t, err)

	execOpts := &drivers.ModelExecuteOptions{
		ModelExecutorOptions: opts,
		InputProperties:      opts.PreliminaryInputProperties,
		OutputProperties:     opts.PreliminaryOutputProperties,
	}

	_, err = me.Execute(context.Background(), execOpts)
	require.NoError(t, err)

	olap, ok := to.AsOLAP("default")
	require.True(t, ok)

	res, err := olap.Query(context.Background(), &drivers.Statement{Query: "select count(*) from sink"})
	require.NoError(t, err)
	for res.Next() {
		var count int
		err = res.Rows.Scan(&count)
		require.NoError(t, err)
		require.Equal(t, 3, count)
	}
	require.NoError(t, res.Err())
	require.NoError(t, res.Close())
}
