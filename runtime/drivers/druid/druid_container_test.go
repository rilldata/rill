package druid

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/storage"
	"github.com/rilldata/rill/runtime/testruntime/testmode"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/zap"
)

const testTable = "test_data"

var testCSV = strings.TrimSpace(`
id,timestamp,publisher,domain,bid_price
5000,2022-03-18T12:25:58.074Z,Facebook,facebook.com,4.19
9000,2022-03-15T11:17:23.530Z,Microsoft,msn.com,3.48
10000,2022-03-02T04:00:56.643Z,Microsoft,msn.com,3.57
11000,2022-01-16T00:26:44.770Z,,instagram.com,5.38
12000,2022-01-17T08:55:09.270Z,,msn.com,1.34
13000,2022-03-20T03:16:57.618Z,Yahoo,news.yahoo.com,1.05
14000,2022-01-29T19:05:33.545Z,Google,news.google.com,4.54
15000,2022-03-22T00:56:22.035Z,Yahoo,news.yahoo.com,1.13
16000,2022-01-24T13:41:43.527Z,,instagram.com,1.78
`)

var testIngestSpec = fmt.Sprintf(`{
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
		"dataSchema": {
			"dataSource": "%s",
			"timestampSpec": {
				"column": "timestamp",
				"format": "iso"
			},
			"transformSpec": {},
			"dimensionsSpec": {
				"dimensions": [
					{"type": "long", "name": "id"},
					"publisher",
					"domain",
					{"type": "double", "name": "bid_price"}
				]
			},
			"granularitySpec": {
				"queryGranularity": "none",
				"rollup": false,
				"segmentGranularity": "day"
			}
		}
	}
}`, strings.ReplaceAll(testCSV, "\n", "\\n"), testTable)

// TestContainer starts a Druid cluster using testcontainers, ingests data into it, then runs all other tests
// in this file as sub-tests (to prevent spawning many clusters).
//
// Unfortunately starting a Druid cluster with test containers is extremely slow.
// If you have access to our Druid test cluster, consider using the test_druid.go file instead.
func TestContainer(t *testing.T) {
	testmode.Expensive(t)

	ctx := context.Background()
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		Started: true,
		ContainerRequest: testcontainers.ContainerRequest{
			WaitingFor:   wait.ForHTTP("/status/health").WithPort("8081").WithStartupTimeout(time.Minute * 2),
			Image:        "gcr.io/rilldata/druid-micro:25.0.0",
			ExposedPorts: []string{"8081/tcp", "8082/tcp"},
			Cmd:          []string{"./bin/start-micro-quickstart"},
		},
	})
	require.NoError(t, err)
	defer container.Terminate(ctx)

	coordinatorURL, err := container.PortEndpoint(ctx, "8081/tcp", "http")
	require.NoError(t, err)

	t.Run("ingest", func(t *testing.T) { testIngest(t, coordinatorURL) })

	brokerURL, err := container.PortEndpoint(ctx, "8082/tcp", "http")
	require.NoError(t, err)

	druidAPIURL, err := url.JoinPath(brokerURL, "/druid/v2/sql")
	require.NoError(t, err)

	dd := &driver{}
	conn, err := dd.Open("default", map[string]any{"dsn": druidAPIURL}, storage.MustNew(t.TempDir(), nil), activity.NewNoopClient(), zap.Must(zap.NewDevelopment()))
	require.NoError(t, err)

	olap, ok := conn.AsOLAP("")
	require.True(t, ok)

	t.Run("count", func(t *testing.T) { testCount(t, olap) })
	t.Run("max", func(t *testing.T) { testMax(t, olap) })
	t.Run("time floor", func(t *testing.T) { testTimeFloor(t, olap) })

	require.NoError(t, conn.Close())
}

func testIngest(t *testing.T, coordinatorURL string) {
	timeout := 5 * time.Minute
	err := Ingest(coordinatorURL, testIngestSpec, testTable, timeout)
	require.NoError(t, err)
}

func testCount(t *testing.T, olap drivers.OLAPStore) {
	qry := fmt.Sprintf("SELECT count(*) FROM %s", testTable)
	rows, err := olap.Query(context.Background(), &drivers.Statement{Query: qry})
	require.NoError(t, err)

	var count int
	rows.Next()

	require.NoError(t, rows.Scan(&count))
	require.Equal(t, 9, count)
	require.NoError(t, rows.Close())
}

func testMax(t *testing.T, olap drivers.OLAPStore) {
	qry := fmt.Sprintf("SELECT max(id) FROM %s", testTable)
	expectedValue := 16000
	rows, err := olap.Query(context.Background(), &drivers.Statement{Query: qry})
	require.NoError(t, err)

	var count int
	rows.Next()
	require.NoError(t, rows.Scan(&count))
	require.Equal(t, expectedValue, count)
	require.NoError(t, rows.Close())
}

func testTimeFloor(t *testing.T, olap drivers.OLAPStore) {
	qry := fmt.Sprintf("SELECT time_floor(__time, 'P1D', null, CAST(? AS VARCHAR)) FROM %s", testTable)
	rows, err := olap.Query(context.Background(), &drivers.Statement{
		Query: qry,
		Args:  []any{"Asia/Kathmandu"},
	})
	require.NoError(t, err)
	defer rows.Close()

	var tmString string
	count := 0
	for rows.Next() {
		require.NoError(t, rows.Scan(&tmString))
		tm, err := time.Parse(time.RFC3339, tmString)
		require.NoError(t, err)
		require.Equal(t, 15, tm.Minute())
		count += 1
	}
	require.NoError(t, rows.Err())
	require.Equal(t, 9, count)
}
