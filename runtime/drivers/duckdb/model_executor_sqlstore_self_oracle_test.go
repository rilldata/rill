package duckdb_test

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/storage"
	_ "github.com/sijms/go-ora/v2"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/zap"

	_ "github.com/rilldata/rill/runtime/drivers/duckdb"
	_ "github.com/rilldata/rill/runtime/drivers/oracle"
)

var oracleInitStmt = `
CREATE TABLE all_data_types_table (
    id NUMBER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    sample_number NUMBER,
    sample_float BINARY_FLOAT,
    sample_double BINARY_DOUBLE,
    sample_integer INTEGER,
    sample_varchar2 VARCHAR2(255),
    sample_char CHAR(10),
    sample_clob CLOB,
    sample_date DATE,
    sample_timestamp TIMESTAMP,
    sample_timestamp_tz TIMESTAMP WITH TIME ZONE
);
INSERT INTO all_data_types_table (sample_number, sample_float, sample_double, sample_integer, sample_varchar2, sample_char, sample_clob, sample_date, sample_timestamp, sample_timestamp_tz)
VALUES (42.5, 3.14, 2.718281828, 100, 'Hello', 'World', 'CLOB text', TO_DATE('2024-02-14', 'YYYY-MM-DD'), TO_TIMESTAMP('2025-02-14 12:34:56.789', 'YYYY-MM-DD HH24:MI:SS.FF3'), TO_TIMESTAMP_TZ('2025-02-14 12:34:56.789 +00:00', 'YYYY-MM-DD HH24:MI:SS.FF3 TZH:TZM'));
INSERT INTO all_data_types_table (sample_number, sample_float, sample_double, sample_integer, sample_varchar2, sample_char, sample_clob, sample_date, sample_timestamp, sample_timestamp_tz)
VALUES (NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL)
`

func TestOracleToDuckDBTransfer(t *testing.T) {
	ctx := context.Background()
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		Started: true,
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "gvenzl/oracle-free:slim-faststart",
			ExposedPorts: []string{"1521/tcp"},
			Env: map[string]string{
				"ORACLE_PASSWORD": "oracle",
			},
			WaitingFor: wait.ForLog("DATABASE IS READY TO USE!").WithStartupTimeout(5 * time.Minute),
		},
	})
	require.NoError(t, err)
	defer container.Terminate(ctx)

	host, err := container.Host(ctx)
	require.NoError(t, err)
	port, err := container.MappedPort(ctx, "1521/tcp")
	require.NoError(t, err)

	dsn := fmt.Sprintf("oracle://system:oracle@%s:%s/FREEPDB1", host, port.Port())

	db, err := sql.Open("oracle", dsn)
	require.NoError(t, err)
	defer db.Close()

	// Execute init statements one at a time
	stmts := strings.Split(oracleInitStmt, ";")
	for _, stmt := range stmts {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		_, err = db.ExecContext(ctx, stmt)
		require.NoError(t, err)
	}

	t.Run("model_executor_oracle_to_duckDB", func(t *testing.T) {
		oracleToDuckDB(t, dsn)
	})
}

func oracleToDuckDB(t *testing.T, dsn string) {
	duckDB, err := drivers.Open("duckdb", "default", map[string]any{"data_dir": t.TempDir()}, storage.MustNew(t.TempDir(), nil), activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)

	inputHandle, err := drivers.Open("oracle", "default", map[string]any{"dsn": dsn}, storage.MustNew(t.TempDir(), nil), activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)

	opts := &drivers.ModelExecutorOptions{
		InputHandle:     inputHandle,
		InputConnector:  "oracle",
		OutputHandle:    duckDB,
		OutputConnector: "duckdb",
		Env: &drivers.ModelEnv{
			AllowHostAccess: false,
			StageChanges:    true,
		},
		PreliminaryInputProperties: map[string]any{
			"sql": "SELECT * FROM all_data_types_table",
			"dsn": dsn,
		},
		PreliminaryOutputProperties: map[string]any{
			"table": "sink",
		},
	}

	me, err := duckDB.AsModelExecutor("default", opts)
	require.NoError(t, err)

	execOpts := &drivers.ModelExecuteOptions{
		ModelExecutorOptions: opts,
		InputProperties:      opts.PreliminaryInputProperties,
		OutputProperties:     opts.PreliminaryOutputProperties,
	}
	_, err = me.Execute(context.Background(), execOpts)
	require.NoError(t, err)

	olap, ok := duckDB.AsOLAP("default")
	require.True(t, ok)

	res, err := olap.Query(context.Background(), &drivers.Statement{Query: "select count(*) from sink"})
	require.NoError(t, err)
	for res.Next() {
		var count int
		err = res.Rows.Scan(&count)
		require.NoError(t, err)
		require.Equal(t, 2, count)
	}
	require.NoError(t, res.Err())
	require.NoError(t, res.Close())

	tbl, err := olap.InformationSchema().Lookup(context.Background(), "", "", "sink")
	require.NoError(t, err)
	require.False(t, tbl.View)
	require.NoError(t, duckDB.Close())
}
