package druid

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"testing"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
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

// TestDruid starts a Druid cluster using testcontainers, ingests data into it, then runs all other tests
// in this file as sub-tests (to prevent spawning many clusters).
func TestDruid(t *testing.T) {
	if testing.Short() {
		t.Skip("druid: skipping test in short mode")
	}

	ctx := context.Background()
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		Started: true,
		ContainerRequest: testcontainers.ContainerRequest{
			WaitingFor:   wait.ForHTTP("/status/health").WithPort("8081"),
			Image:        "druid-micro:latest",
			ExposedPorts: []string{"8081/tcp", "8082/tcp"},
		},
	})
	require.NoError(t, err)
	defer container.Terminate(ctx)

	coordinatorURL, err := container.PortEndpoint(ctx, "8081/tcp", "http")
	require.NoError(t, err)

	t.Run("ingest", func(t *testing.T) { testIngest(t, coordinatorURL) })

	brokerURL, err := container.PortEndpoint(ctx, "8082/tcp", "http")
	require.NoError(t, err)

	avaticaURL, err := url.JoinPath(brokerURL, "/druid/v2/sql/avatica-protobuf/")
	require.NoError(t, err)

	conn, err := driver{}.Open(avaticaURL, zap.NewNop())
	require.NoError(t, err)

	olap, ok := conn.OLAPStore()
	require.True(t, ok)

	t.Run("count", func(t *testing.T) { testCount(t, olap) })
	t.Run("max", func(t *testing.T) { testMax(t, olap) })
	t.Run("schema all", func(t *testing.T) { testSchemaAll(t, olap) })
	t.Run("schema lookup", func(t *testing.T) { testSchemaLookup(t, olap) })
	// Add new tests here

	require.NoError(t, conn.Close())
	require.Error(t, conn.(*connection).db.Ping())
}

func testIngest(t *testing.T, coordinatorURL string) {
	timeout := 5 * time.Minute
	err := Ingest(coordinatorURL, testIngestSpec, testTable, timeout)
	require.NoError(t, err)
}

func testCount(t *testing.T, olap drivers.OLAPStore) {
	qry := fmt.Sprintf("SELECT count(*) FROM %s", testTable)
	rows, err := olap.Execute(context.Background(), &drivers.Statement{Query: qry})
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
	rows, err := olap.Execute(context.Background(), &drivers.Statement{Query: qry})
	require.NoError(t, err)

	var count int
	rows.Next()
	require.NoError(t, rows.Scan(&count))
	require.Equal(t, expectedValue, count)
	require.NoError(t, rows.Close())
}

func testSchemaAll(t *testing.T, olap drivers.OLAPStore) {
	tables, err := olap.InformationSchema().All(context.Background())
	require.NoError(t, err)

	require.Equal(t, 1, len(tables))
	require.Equal(t, testTable, tables[0].Name)

	require.Equal(t, 5, len(tables[0].Schema.Fields))

	mp := make(map[string]*runtimev1.StructType_Field)
	for _, f := range tables[0].Schema.Fields {
		mp[f.Name] = f
	}

	f := mp["__time"]
	require.Equal(t, "__time", f.Name)
	require.Equal(t, runtimev1.Type_CODE_TIMESTAMP, f.Type.Code)
	require.Equal(t, false, f.Type.Nullable)
	f = mp["bid_price"]
	require.Equal(t, runtimev1.Type_CODE_FLOAT64, f.Type.Code)
	require.Equal(t, false, f.Type.Nullable)
	f = mp["domain"]
	require.Equal(t, runtimev1.Type_CODE_STRING, f.Type.Code)
	require.Equal(t, true, f.Type.Nullable)
	f = mp["id"]
	require.Equal(t, runtimev1.Type_CODE_INT64, f.Type.Code)
	require.Equal(t, false, f.Type.Nullable)
	f = mp["publisher"]
	require.Equal(t, runtimev1.Type_CODE_STRING, f.Type.Code)
	require.Equal(t, true, f.Type.Nullable)
}

func testSchemaLookup(t *testing.T, olap drivers.OLAPStore) {
	ctx := context.Background()
	table, err := olap.InformationSchema().Lookup(ctx, testTable)
	require.NoError(t, err)
	require.Equal(t, testTable, table.Name)

	_, err = olap.InformationSchema().Lookup(ctx, "foo")
	require.Equal(t, drivers.ErrNotFound, err)
}
