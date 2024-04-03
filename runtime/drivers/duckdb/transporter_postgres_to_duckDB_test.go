package duckdb

import (
	"context"
	"database/sql"
	"testing"

	"github.com/rilldata/rill/admin/pkg/pgtestcontainer"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	// Load postgres driver
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/rilldata/rill/runtime/drivers/postgres"
)

var sqlStmt = `CREATE TYPE country AS ENUM ('IND', 'AUS', 'SA', 'NZ');
  CREATE TABLE all_datatypes (
	id serial PRIMARY KEY,
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
  INSERT INTO all_datatypes (name, age, is_married, date_of_birth, time_of_day, created_at, json, json_data, bit,bit_varying, character, character_varying, bpchar, smallint, text, timestamptz, float4, float8, int2, int4, int8, int8_array, timestamptz_array, emp_salary, country)
  VALUES
	('John Doe', 30, true, '1983-03-08', '12:35:00', '2023-09-12 12:46:55', '{"name": "John Doe", "age": 30, "salary": 100000}', '{"name": "John Doe", "age": 30, "salary": 100000}', b'1',b'10101010', 'a', 'ab', 'abcd', 123, 'This is a text string.', '2023-09-12 12:46:55+05:30', 23.2, 123.45, 1, 1234, 1234567, Array[1234567, 7654312], Array[timestamp'2023-09-12 12:46:55+05:30', timestamp'2023-10-12 12:46:55+05:30'], 38500000000000.71256565656563, 'IND');  
  `

func TestTransfer(t *testing.T) {
	pg := pgtestcontainer.New(t)
	defer pg.Terminate(t)

	db, err := sql.Open("pgx", pg.DatabaseURL)
	require.NoError(t, err)
	defer db.Close()

	t.Run("AllDataTypes", func(t *testing.T) { allDataTypesTest(t, db, pg.DatabaseURL) })
}

func allDataTypesTest(t *testing.T, db *sql.DB, dbURL string) {
	ctx := context.Background()
	_, err := db.ExecContext(ctx, sqlStmt)
	require.NoError(t, err)

	handle, err := drivers.Open("postgres", map[string]any{"database_url": dbURL}, false, activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)
	require.NotNil(t, handle)

	sqlStore, _ := handle.AsSQLStore()
	to, err := drivers.Open("duckdb", map[string]any{"dsn": ""}, false, activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)
	olap, _ := to.AsOLAP("")

	tr := NewSQLStoreToDuckDB(sqlStore, olap, zap.NewNop())
	err = tr.Transfer(ctx, map[string]any{"sql": "select * from all_datatypes;"}, map[string]any{"table": "sink"}, &drivers.TransferOptions{Progress: drivers.NoOpProgress{}})
	require.NoError(t, err)
	res, err := olap.Execute(context.Background(), &drivers.Statement{Query: "select count(*) from sink"})
	require.NoError(t, err)
	for res.Next() {
		var count int
		err = res.Rows.Scan(&count)
		require.NoError(t, err)
		require.Equal(t, 1, count)
	}
	require.NoError(t, res.Close())
	require.NoError(t, to.Close())
}
