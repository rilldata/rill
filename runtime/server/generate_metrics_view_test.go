package server_test

import (
	"context"
	"testing"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/pkg/ratelimit"
	"github.com/rilldata/rill/runtime/server"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestGenerateMetricsView(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"rill.yaml": ``,
			// Normal model
			"ad_bids.sql": `SELECT now() AS time, 'DA' AS country, 3.141 as price`,
			// Create a non-default duckdb connector
			"custom_duckdb.yaml": `
type: connector
driver: duckdb
`,
		},
	})

	// Create some externally managed tables
	olapExecAdhoc(t, rt, instanceID, "duckdb", "CREATE TABLE IF NOT EXISTS foo AS SELECT now() AS time, 'DA' AS country, 3.141 as price")
	olapExecAdhoc(t, rt, instanceID, "custom_duckdb", "CREATE TABLE IF NOT EXISTS foo AS SELECT now() AS time, 'DA' AS country, 3.141 as price")

	ctx, cancel := context.WithTimeout(testCtx(), 25*time.Second)
	defer cancel()

	repo, release, err := rt.Repo(ctx, instanceID)
	require.NoError(t, err)
	defer release()

	server, err := server.NewServer(ctx, &server.Options{}, rt, zap.NewNop(), ratelimit.NewNoop(), activity.NewNoopClient())
	require.NoError(t, err)

	tt := []struct {
		name      string
		model     string
		connector string
		table     string
		contains  []string
	}{
		{
			name:  "model passed in request",
			model: "ad_bids",
			contains: []string{
				"model: ad_bids",
				"measures:",
				"format_preset: humanize",
			},
		},
		{
			name:     "model passed in request that matches a table",
			model:    "foo",
			contains: []string{"model: foo"},
		},
		{
			name:     "table passed in request that matches a model",
			table:    "ad_bids",
			contains: []string{"model: ad_bids"},
		},
		{
			name:     "table passed in request that does not match a model",
			table:    "foo",
			contains: []string{"model: foo"},
		},
		{
			name:      "table in non-default connector passed in request",
			table:     "foo",
			connector: "custom_duckdb",
			contains:  []string{"connector: custom_duckdb", "model: foo"},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			err = repo.Delete(ctx, "/metrics/generated_metrics_view.yaml", true)
			require.NoError(t, err)

			_, err = server.GenerateMetricsViewFile(ctx, &runtimev1.GenerateMetricsViewFileRequest{
				InstanceId: instanceID,
				Model:      tc.model,
				Connector:  tc.connector,
				Table:      tc.table,
				Path:       "/metrics/generated_metrics_view.yaml",
				UseAi:      false,
			})
			require.NoError(t, err)

			data, err := repo.Get(ctx, "/metrics/generated_metrics_view.yaml")
			require.NoError(t, err)

			for _, c := range tc.contains {
				require.Contains(t, data, c)
			}
		})
	}
}

func olapExecAdhoc(t *testing.T, rt *runtime.Runtime, instanceID, connector, query string) {
	h, release, err := rt.AcquireHandle(context.Background(), instanceID, connector)
	require.NoError(t, err)
	defer release()
	olap, _ := h.AsOLAP(instanceID)
	err = olap.Exec(context.Background(), &drivers.Statement{Query: query})
	require.NoError(t, err)
}
