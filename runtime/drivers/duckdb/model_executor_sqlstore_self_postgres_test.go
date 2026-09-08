package duckdb_test

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"testing"

	"github.com/rilldata/rill/admin/pkg/pgtestcontainer"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/storage"
	"github.com/rilldata/rill/runtime/testruntime/testmode"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	_ "github.com/rilldata/rill/runtime/drivers/duckdb"
	_ "github.com/rilldata/rill/runtime/drivers/postgres"
)

var sqlStmt = `CREATE TYPE country AS ENUM ('IND', 'AUS', 'SA', 'NZ');
  CREATE TABLE all_datatypes (
	id serial PRIMARY KEY,
	uuid UUID,
	name text,
	age integer,
	is_married boolean,
	date_of_birth date,
	time_of_day time,
	created_at timestamp,
	json json,
	json_data jsonb,
	bit bit,
	bit_varying bit varying,
	character character,
	character_varying character varying,
	bpchar bpchar(10),
	smallint smallint,
	text text,
	timestamptz timestamptz,
	float4 float4,
	float8 float8,
	int2 int2,
	int4 int4,
	int8 int8,
	int8_array int8[],
	timestamptz_array timestamptz[],
	emp_salary NUMERIC,
	country country
  );
  INSERT INTO all_datatypes (uuid, name, age, is_married, date_of_birth, time_of_day, created_at, json, json_data, bit,bit_varying, character, character_varying, bpchar, smallint, text, timestamptz, float4, float8, int2, int4, int8, int8_array, timestamptz_array, emp_salary, country)
  VALUES
	(gen_random_uuid(), 'John Doe', 30, true, '1983-03-08', '12:35:00', '2023-09-12 12:46:55', '{"name": "John Doe", "age": 30, "salary": 100000}', '{"name": "John Doe", "age": 30, "salary": 100000}', b'1',b'10101010', 'a', 'ab', 'abcd', 123, 'This is a text string.', '2023-09-12 12:46:55+05:30', 23.2, 123.45, 1, 1234, 1234567, Array[1234567, 7654312], Array[timestamp'2023-09-12 12:46:55+05:30', timestamp'2023-10-12 12:46:55+05:30'], 38500000000000.71256565656563, 'IND');  
  `

func TestTransfer(t *testing.T) {
	testmode.Expensive(t)
	pg := pgtestcontainer.New(t)
	defer pg.Terminate(t)

	db, err := sql.Open("pgx", pg.DatabaseURL)
	require.NoError(t, err)
	defer db.Close()

	_, err = db.ExecContext(context.Background(), sqlStmt)
	require.NoError(t, err)

	t.Run("model_executor_postgres_to_duckDB", func(t *testing.T) { pgxToDuckDB(t, db, pg.DatabaseURL) })
}

func pgxToDuckDB(t *testing.T, pgdb *sql.DB, dbURL string) {
	duckDB, err := drivers.Open("duckdb", "", "default", map[string]any{"data_dir": t.TempDir()}, storage.MustNew(t.TempDir(), nil), activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)

	inputHandle, err := drivers.Open("postgres", "", "default", map[string]any{"database_url": dbURL}, storage.MustNew(t.TempDir(), nil), activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)

	opts := &drivers.ModelExecutorOptions{
		InputHandle:     inputHandle,
		InputConnector:  "postgres",
		OutputHandle:    duckDB,
		OutputConnector: "duckdb",
		Env: &drivers.ModelEnv{
			AllowHostAccess: false,
			StageChanges:    true,
		},
		PreliminaryInputProperties: map[string]any{
			"sql": "SELECT * FROM all_datatypes;",
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
		require.Equal(t, 1, count)
	}
	require.NoError(t, res.Err())
	require.NoError(t, res.Close())

	// ingest some more data in postges
	_, err = pgdb.Exec("INSERT INTO all_datatypes(uuid, created_at) VALUES (gen_random_uuid(), '2024-01-02 12:46:55');")
	require.NoError(t, err)

	// drop older data from postgres
	_, err = pgdb.Exec("DELETE FROM all_datatypes WHERE created_at < '2024-01-01 00:00:00';")
	require.NoError(t, err)

	// incremental run
	execOpts.IncrementalRun = true
	execOpts.InputProperties["sql"] = "SELECT * FROM all_datatypes WHERE created_at > '2024-01-01 00:00:00';"
	_, err = me.Execute(context.Background(), execOpts)
	require.NoError(t, err)

	res, err = olap.Query(context.Background(), &drivers.Statement{Query: "select count(*) from sink"})
	require.NoError(t, err)
	for res.Next() {
		var count int
		err = res.Rows.Scan(&count)
		require.NoError(t, err)
		require.Equal(t, 2, count)
	}
	require.NoError(t, res.Err())
	require.NoError(t, res.Close())

	require.NoError(t, duckDB.Close())
}

func TestPostgresToDuckLakeTransfer(t *testing.T) {
	// testmode.Expensive(t)
	pg := pgtestcontainer.New(t)
	defer pg.Terminate(t)

	db, err := sql.Open("pgx", pg.DatabaseURL)
	require.NoError(t, err)
	defer db.Close()

	_, err = db.ExecContext(t.Context(), sqlStmt)
	require.NoError(t, err)
	_, err = db.ExecContext(t.Context(), "CREATE DATABASE ducklake_catalog;")
	require.NoError(t, err)

	catalogURL, err := url.Parse(pg.DatabaseURL)
	require.NoError(t, err)
	catalogURL.Path = "/ducklake_catalog"
	dataPath := t.TempDir()
	outputConfig := map[string]any{
		"attach":        fmt.Sprintf("'ducklake:postgres:%s' AS ducklake (DATA_PATH '%s', DATA_INLINING_ROW_LIMIT 0)", catalogURL.String(), dataPath),
		"database_name": "ducklake",
		"mode":          "readwrite",
	}

	t.Run("retry_after_missing_column", func(t *testing.T) {
		duckLake, err := drivers.Open("duckdb", "", "retry", outputConfig, storage.MustNew(t.TempDir(), nil), activity.NewNoopClient(), zap.NewNop())
		require.NoError(t, err)
		inputHandle, err := drivers.Open("postgres", "", "retry", map[string]any{"database_url": pg.DatabaseURL}, storage.MustNew(t.TempDir(), nil), activity.NewNoopClient(), zap.NewNop())
		require.NoError(t, err)

		requireDuckLakeRetryAfterFailure(t, duckLake, inputHandle, "postgres", map[string]any{
			"sql": "SELECT missing_column FROM all_datatypes;",
		}, "SELECT * FROM all_datatypes;", 1)
		require.NoError(t, inputHandle.Close())
		require.NoError(t, duckLake.Close())
	})

	t.Run("ingestion", func(t *testing.T) {
		postgresToDuckLake(t, db, pg.DatabaseURL, outputConfig)
	})
}

func postgresToDuckLake(t *testing.T, pgdb *sql.DB, dbURL string, outputConfig map[string]any) {
	duckLake, err := drivers.Open("duckdb", "", "default", outputConfig, storage.MustNew(t.TempDir(), nil), activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)

	inputHandle, err := drivers.Open("postgres", "", "default", map[string]any{"database_url": dbURL}, storage.MustNew(t.TempDir(), nil), activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)

	opts := &drivers.ModelExecutorOptions{
		InputHandle:     inputHandle,
		InputConnector:  "postgres",
		OutputHandle:    duckLake,
		OutputConnector: "duckdb",
		Env: &drivers.ModelEnv{
			AllowHostAccess: false,
			StageChanges:    true,
		},
		PreliminaryInputProperties: map[string]any{
			"sql": "SELECT * FROM all_datatypes;",
		},
		PreliminaryOutputProperties: map[string]any{
			"table": "sink",
		},
	}

	me, err := duckLake.AsModelExecutor("default", opts)
	require.NoError(t, err)
	execOpts := &drivers.ModelExecuteOptions{
		ModelExecutorOptions: opts,
		InputProperties:      opts.PreliminaryInputProperties,
		OutputProperties:     opts.PreliminaryOutputProperties,
	}

	_, err = me.Execute(t.Context(), execOpts)
	require.NoError(t, err)
	requireDuckLakeRowCount(t, duckLake, 1)

	_, err = pgdb.ExecContext(t.Context(), "INSERT INTO all_datatypes(uuid, created_at) VALUES (gen_random_uuid(), '2024-01-02 12:46:55');")
	require.NoError(t, err)
	_, err = pgdb.ExecContext(t.Context(), "DELETE FROM all_datatypes WHERE created_at < '2024-01-01 00:00:00';")
	require.NoError(t, err)

	execOpts.IncrementalRun = true
	execOpts.InputProperties["sql"] = "SELECT * FROM all_datatypes WHERE created_at > '2024-01-01 00:00:00';"
	_, err = me.Execute(t.Context(), execOpts)
	require.NoError(t, err)
	requireDuckLakeRowCount(t, duckLake, 2)
	require.NoError(t, duckLake.Close())
}

func requireDuckLakeRowCount(t *testing.T, handle drivers.Handle, expected int) {
	olap, ok := handle.AsOLAP("default")
	require.True(t, ok)
	res, err := olap.Query(t.Context(), &drivers.Statement{Query: "SELECT count(*) FROM sink"})
	require.NoError(t, err)
	defer res.Close()
	require.True(t, res.Next())
	var count int
	require.NoError(t, res.Rows.Scan(&count))
	require.Equal(t, expected, count)
	require.NoError(t, res.Err())
}

func requireDuckLakeRetryAfterFailure(t *testing.T, duckLake, inputHandle drivers.Handle, inputConnector string, inputProperties map[string]any, validSQL string, expectedRows int) {
	opts := &drivers.ModelExecutorOptions{
		InputHandle:     inputHandle,
		InputConnector:  inputConnector,
		OutputHandle:    duckLake,
		OutputConnector: "duckdb",
		Env: &drivers.ModelEnv{
			AllowHostAccess: false,
			StageChanges:    true,
		},
		PreliminaryInputProperties: inputProperties,
		PreliminaryOutputProperties: map[string]any{
			"table": "sink",
		},
	}
	me, err := duckLake.AsModelExecutor("default", opts)
	require.NoError(t, err)
	execOpts := &drivers.ModelExecuteOptions{
		ModelExecutorOptions: opts,
		InputProperties:      opts.PreliminaryInputProperties,
		OutputProperties:     opts.PreliminaryOutputProperties,
	}

	_, err = me.Execute(t.Context(), execOpts)
	require.Error(t, err)
	require.ErrorContains(t, err, "missing_column")

	execOpts.InputProperties["sql"] = validSQL
	_, err = me.Execute(t.Context(), execOpts)
	require.NoError(t, err)
	requireDuckLakeRowCount(t, duckLake, expectedRows)
}
