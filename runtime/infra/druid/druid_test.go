package druid

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/rilldata/rill/runtime/infra"
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

// TestDruid starts a Druid cluster using testcontainers, ingests data into it, then runs all other tests
// in this file as sub-tests (to prevent spawning many clusters).
func TestDruid(t *testing.T) {
	ctx := context.Background()
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		Started: true,
		ContainerRequest: testcontainers.ContainerRequest{
			ExposedPorts: []string{"8081/tcp", "8082/tcp"},
			WaitingFor:   wait.ForHTTP("/status/health").WithPort("8081"),
			FromDockerfile: testcontainers.FromDockerfile{
				Context:       ".",
				Dockerfile:    "Dockerfile",
				PrintBuildLog: true,
			},
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

	conn, err := driver{}.Open(avaticaURL)
	require.NoError(t, err)

	time.Sleep(30 * time.Second)

	t.Run("count", func(t *testing.T) { testCount(t, conn) })
	t.Run("max", func(t *testing.T) { testMax(t, conn) })
	// Add new tests here

	require.NoError(t, conn.Close())
	require.Error(t, conn.(*connection).db.Ping())
}

func testIngest(t *testing.T, coordinatorURL string) {
	escapedCSV := strings.ReplaceAll(testCSV, "\n", "\\n")
	ingestSpec := fmt.Sprintf(`{
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
				{
				  "type": "long",
				  "name": "id"
				},
				"publisher",
				"domain",
				{
				  "type": "double",
				  "name": "bid_price"
				}
			  ]
			},
			"granularitySpec": {
			  "queryGranularity": "none",
			  "rollup": false,
			  "segmentGranularity": "day"
			}
		  }
		}
	  }`, escapedCSV, testTable)

	timeout := 5 * time.Minute
	err := Ingest(coordinatorURL, ingestSpec, testTable, timeout)
	require.NoError(t, err)
}

func testCount(t *testing.T, conn infra.Connection) {
	qry := fmt.Sprintf("SELECT count(*) FROM %s", testTable)
	rows, err := conn.Execute(context.Background(), &infra.Statement{Query: qry})
	require.NoError(t, err)

	var count int
	rows.Next()

	require.NoError(t, rows.Scan(&count))
	require.Equal(t, 9, count)
	require.NoError(t, rows.Close())
}

func testMax(t *testing.T, conn infra.Connection) {
	qry := fmt.Sprintf("SELECT max(id) FROM %s", testTable)
	expectedValue := 16000
	rows, err := conn.Execute(context.Background(), &infra.Statement{Query: qry})
	require.NoError(t, err)

	var count int
	rows.Next()
	require.NoError(t, rows.Scan(&count))
	require.Equal(t, expectedValue, count)
	require.NoError(t, rows.Close())
}
