package druid

import (
	"context"
	"fmt"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/rilldata/rill/runtime/infra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const dataSourceName = "test_data"

func TestIngestDataDruid(t *testing.T) {
	dataJson := fmt.Sprintf(`{
		"type": "index_parallel",
		"spec": {
		  "ioConfig": {
			"type": "index_parallel",
			"inputSource": {
			  "type": "inline",
			  "data": "id,timestamp,publisher,domain,bid_price\n5000,2022-03-18T12:25:58.074Z,Facebook,facebook.com,4.19\n9000,2022-03-15T11:17:23.530Z,Microsoft,msn.com,3.48\n10000,2022-03-02T04:00:56.643Z,Microsoft,msn.com,3.57\n11000,2022-01-16T00:26:44.770Z,,instagram.com,5.38\n12000,2022-01-17T08:55:09.270Z,,msn.com,1.34\n13000,2022-03-20T03:16:57.618Z,Yahoo,news.yahoo.com,1.05\n14000,2022-01-29T19:05:33.545Z,Google,news.google.com,4.54\n15000,2022-03-22T00:56:22.035Z,Yahoo,news.yahoo.com,1.13\n16000,2022-01-24T13:41:43.527Z,,instagram.com,1.78"
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
	  }`, dataSourceName)

	druidCoordinatorUrl := "http://localhost:8081"
	ingestionStats, err := Ingest(druidCoordinatorUrl, dataJson, dataSourceName)

	require.NoError(t, err)
	assert.Equal(t, 200, ingestionStats.StatusCode)

	conn := prepareConn(t)
	var qry string
	qry = fmt.Sprintf("SELECT count(*) FROM %s", dataSourceName)
	rows, err := conn.Execute(context.Background(), 0, qry)
	require.NoError(t, err)

	var count int
	rows.Next()

	require.NoError(t, rows.Scan(&count))
	require.Equal(t, 9, count)
	require.NoError(t, rows.Close())

}

func TestExecute(t *testing.T) {
	conn := prepareConn(t)

	var qry string
	qry = fmt.Sprintf("SELECT max(id) FROM %s", dataSourceName)
	rows, err := conn.Execute(context.Background(), 0, qry)
	require.NoError(t, err)

	var count int
	expectedValue := 16000
	rows.Next()
	require.NoError(t, rows.Scan(&count))
	require.Equal(t, expectedValue, count)
	require.NoError(t, rows.Close())

	err = conn.Close()
	require.NoError(t, err)
	err = conn.(*connection).db.Ping()
	require.Error(t, err)

}

func TestQueryAvaticaDriver(t *testing.T) {
	db, err := sqlx.Open("avatica", "http://localhost:8082/druid/v2/sql/avatica-protobuf/")
	require.NoError(t, err)

	rows, err := db.Queryx(`SELECT 'Foo' as domain`)
	require.NoError(t, err)

	defer func() {
		if err := rows.Close(); err != nil {
			require.NoError(t, err)
		}
	}()

	var domain string
	rows.Next()

	require.NoError(t, rows.Scan(&domain))
	require.Equal(t, "Foo", domain)
	require.NoError(t, rows.Close())
}

func prepareConn(t *testing.T) infra.Connection {
	conn, err := driver{}.Open("http://localhost:8082/druid/v2/sql/avatica-protobuf/")
	require.NoError(t, err)

	return conn
}
