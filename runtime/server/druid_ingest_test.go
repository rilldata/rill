package server_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers/druid"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/pkg/ratelimit"
	"github.com/rilldata/rill/runtime/queries"
	"github.com/rilldata/rill/runtime/server"
	"github.com/rilldata/rill/runtime/server/auth"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
)

const DruidIngestUrl = "http://localhost:8888"

func Test_IngestToDruid(t *testing.T) {
	// uncomment and run to ingest testdata folder into druid
	// Note: druid should be started outside of this
	t.Skip()

	ctx := auth.WithOpen(context.Background())

	ingestProjectIntoDruid(ctx, t, "ad_bids")
	ingestProjectIntoDruid(ctx, t, "timeseries")
}

func ingestProjectIntoDruid(ctx context.Context, t *testing.T, projectName string) {
	rt, instanceID := testruntime.NewInstanceForProject(t, projectName)
	srv, err := server.NewServer(context.Background(), &server.Options{}, rt, nil, ratelimit.NewNoop(), activity.NewNoopClient())
	require.NoError(t, err)

	modelsResp, err := srv.ListResources(ctx, &runtimev1.ListResourcesRequest{
		InstanceId: instanceID,
		Kind:       runtime.ResourceKindModel,
	})
	require.NoError(t, err)

	for _, model := range modelsResp.Resources {
		if model.Meta.ReconcileError != "" {
			continue
		}

		tableName := model.GetModel().State.Table
		if tableName == "" {
			continue
		}

		dataPath := exportModelForDruid(ctx, t, rt, instanceID, tableName)

		colsResp, err := srv.TableColumns(ctx, &runtimev1.TableColumnsRequest{
			InstanceId: instanceID,
			TableName:  tableName,
		})
		require.NoError(t, err)
		ingestModelIntoDruid(t, dataPath, tableName, colsResp.ProfileColumns)
	}
}

func exportModelForDruid(ctx context.Context, t *testing.T, rt *runtime.Runtime, instanceID, name string) string {
	f, err := os.Create(filepath.Join(os.TempDir(), fmt.Sprintf("%s.csv", name)))
	require.NoError(t, err)
	defer f.Close()
	q := queries.TableHead{
		TableName: name,
		Limit:     10000,
	}
	require.NoError(t, q.Export(ctx, rt, instanceID, f, &runtime.ExportOptions{
		Format: runtimev1.ExportFormat_EXPORT_FORMAT_CSV,
	}))

	return f.Name()
}

func ingestModelIntoDruid(t *testing.T, dataPath, name string, cols []*runtimev1.ProfileColumn) {
	schema := getDruidModelSpec(name, cols)
	data, err := os.ReadFile(dataPath)
	require.NoError(t, err)
	defer os.Remove(dataPath)

	fmt.Printf("Ingesting model %s\n", name)
	require.NoError(t, druid.Ingest(
		DruidIngestUrl,
		getDruidIngestionSpec(string(data), schema),
		name,
		5*time.Minute,
	))
}

func getDruidIngestionSpec(data, schema string) string {
	return fmt.Sprintf(
		`{
			"type": "index_parallel",
			"spec": {
				"ioConfig": {
					"type": "index_parallel",
					"inputSource": {
						"type": "inline",
						"data": "%s"
					},
					"inputFormat": {
						"type": "csv",
						"findColumnsFromHeader": true
					}
				},
				"tuningConfig": {
					"type": "index_parallel",
					"partitionsSpec": {
						"type": "dynamic"
					}
				},
				"dataSchema": %s
			}
		}`,
		strings.ReplaceAll(data, "\n", "\\n"),
		schema,
	)
}

func getDruidModelSpec(name string, cols []*runtimev1.ProfileColumn) string {
	var timestampCol string
	var dimensions []string
	for _, col := range cols {
		if col.Type == "TIMESTAMP" || col.Name == "timestamp" {
			timestampCol = col.Name
		} else if col.Type == "VARCHAR" {
			dimensions = append(dimensions, fmt.Sprintf("%q", col.Name))
		} else {
			dimensions = append(dimensions, fmt.Sprintf(`{"type": %q, "name": %q}`, mapColTypeForDruid(col.Type), col.Name))
		}
	}

	return fmt.Sprintf(
		`{
			"dataSource": "%s",
			"timestampSpec": {
				"column": "%s",
				"format": "auto"
			},
			"transformSpec": {},
			"dimensionsSpec": {
				"dimensions": [
					%s
				]
			},
			"granularitySpec": {
				"queryGranularity": "none",
				"rollup": false,
				"segmentGranularity": "day"
			}
		}`,
		name,
		timestampCol,
		strings.Join(dimensions, ",\n"),
	)
}

func mapColTypeForDruid(colType string) string {
	if colType == "BIGINT" {
		return "long"
	}
	return strings.ToLower(colType)
}
