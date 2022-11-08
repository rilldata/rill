package server

import (
	"context"
	"testing"

	// drivers "github.com/rilldata/rill/runtime/drivers"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"

	"github.com/rilldata/rill/runtime/api"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/graceful"
	"go.uber.org/zap"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestServer_Database(t *testing.T) {
	server, _, err := getTestServer()
	if err != nil {
		t.Fatal(err)
	}
	rows := createTestTable(server, t)
	for rows.Next() {
		var count int
		if err := rows.Scan(&count); err != nil {
			t.Fatal(err)
		}
		if count != 1 {
			t.Fatalf("expected 1, got %d", count)
		}
	}
	rows.Close()
}

func createTestTable(server *Server, t *testing.T) *sqlx.Rows {
	rows, err := server.os.Execute(context.Background(), &drivers.Statement{
		Query: "create table test (a int)",
	})
	if err != nil {
		t.Fatal(err)
	}
	rows.Close()
	rows, _ = server.os.Execute(context.Background(), &drivers.Statement{
		Query: "insert into test values (1)",
	})
	rows.Close()
	rows, err = server.os.Execute(context.Background(), &drivers.Statement{
		Query: "select count(*) from test",
	})
	if err != nil {
		t.Fatal(err)
	}
	return rows
}

func TestServer_Cardinality(t *testing.T) {
	logger, _ := zap.NewDevelopment(zap.IncreaseLevel(zap.InfoLevel))
	server, _, err := getTestServer()
	rows := createTestTable(server, t)
	rows.Close()
	ctx := graceful.WithCancelOnTerminate(context.Background())
	server.ServeNoWait(ctx)
	if err != nil {
		logger.Error("server crashed", zap.Error(err))
	}
	// var opts []grpc.DialOption
	// ServerAddr
	conn, err := grpc.Dial("localhost:9090", grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		logger.Error("did not connect", zap.Error(err))
	}
	defer conn.Close()
	client := api.NewRuntimeServiceClient(conn)
	cr, _ := client.Cardinality(context.Background(), &api.CardinalityRequest{
		InstanceId: "test",
		TableName:  "test",
	})

	require.Equal(t, int64(1), cr.Cardinality)

}
