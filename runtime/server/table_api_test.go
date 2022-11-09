package server

import (
	"context"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"

	"github.com/rilldata/rill/runtime/api"
	"github.com/rilldata/rill/runtime/drivers"
)

func getSingleValue(rows *sqlx.Rows) int {
	var val int
	for rows.Next() {
		if err := rows.Scan(&val); err != nil {
			panic(err)
		}
	}
	rows.Close()
	return val
}

func TestServer_Database(t *testing.T) {
	server, instanceId, err := getTestServer()
	if err != nil {
		t.Fatal(err)
	}
	rows := createTestTable(server, instanceId, t)
	require.Equal(t, 1, getSingleValue(rows))
	rows, err = server.query(context.Background(), instanceId, &drivers.Statement{
		Query: "select count(*) from test",
	})
	if err != nil {
		t.Fatal(err)
	}
	require.Equal(t, 1, getSingleValue(rows))
}

func createTestTable(server *Server, instanceId string, t *testing.T) *sqlx.Rows {
	rows, err := server.query(context.Background(), instanceId, &drivers.Statement{
		Query: "create table test (a int)",
	})
	if err != nil {
		t.Fatal(err)
	}
	rows.Close()
	rows, _ = server.query(context.Background(), instanceId, &drivers.Statement{
		Query: "insert into test values (1)",
	})
	rows.Close()
	rows, err = server.query(context.Background(), instanceId, &drivers.Statement{
		Query: "select count(*) from test",
	})
	if err != nil {
		t.Fatal(err)
	}
	return rows
}

func TestServer_Cardinality(t *testing.T) {
	server, instanceId, err := getTestServer()
	if err != nil {
		t.Fatal(err)
	}
	rows := createTestTable(server, instanceId, t)
	rows.Close()
	cr, err := server.TableCardinality(context.Background(), &api.CardinalityRequest{
		InstanceId: instanceId,
		TableName:  "test",
	})
	if err != nil {
		t.Fatal(err)
	}
	require.Equal(t, int64(1), cr.Cardinality)
}

func TestServer_ProfileColumns(t *testing.T) {
	server, instanceId, err := getTestServer()
	if err != nil {
		t.Fatal(err)
	}
	rows := createTestTable(server, instanceId, t)
	rows.Close()
	cr, err := server.ProfileColumns(context.Background(), &api.ProfileColumnsRequest{
		InstanceId: instanceId,
		TableName:  "test",
	})
	if err != nil {
		t.Fatal(err)
	}
	require.Equal(t, 0, len(cr.ProfileColumn))
}
