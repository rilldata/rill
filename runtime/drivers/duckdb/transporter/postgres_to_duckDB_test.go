package transporter

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

var sqlStmt = `CREATE TABLE all_datatypes (
	id serial PRIMARY KEY,
	name text,
	age integer,
	salary numeric(10,2),
	is_married boolean,
	date_of_birth date,
	time_of_day time,
	created_at timestamp,
	ip_address inet,
	mac_address macaddr,
	json_data jsonb,
	bit bit,
	bit_varying bit varying,
	character character,
	character_varying character varying,
	daterange daterange,
	geometric point,
	interval interval,
	money money,
	numeric numeric,
	real real,
	smallint smallint,
	text text,
	timetz timetz,
	timestamptz timestamptz,
	uuid uuid
  );
  INSERT INTO all_datatypes (name, age, salary, is_married, date_of_birth, time_of_day, created_at, ip_address, mac_address, json_data, bit, bit_varying, character, character_varying, daterange, geometric, interval, money, numeric, real, smallint, text, timetz, timestamptz, uuid)
  VALUES
	('John Doe', 30, 100000, true, '1983-03-08', '12:00:00', '2023-09-12 12:46:55', '192.168.1.1', 'AA:BB:CC:DD:EE:FF', '{"name": "John Doe", "age": 30, "salary": 100000}', b'1', b'10101010', 'a', 'ab', daterange('2018-01-01', '2018-12-31', '[]'), '(1,2)', interval '1 DAY', '100.00', 123.45, 6.78, 123, 'This is a text string.', '2022-09-12 12:46:55+00:00', '2023-09-12 12:46:55+05:30', '292a485f-a56a-4938-8f1a-bbbbbbbbbbb1'::UUID);  
  `

func TestTransfer(t *testing.T) {
	pg := pgtestcontainer.New(t)
	defer pg.Terminate(t)

	db, err := sql.Open("pgx", pg.DatabaseURL)
	require.NoError(t, err)
	defer db.Close()

	t.Run("AllDataTypes", func(t *testing.T) { allDataTypesTest(t, db, pg.DatabaseURL) })
	t.Run("CompositeTypes", func(t *testing.T) { compositeTypesTest(t, db, pg.DatabaseURL) })
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
}

var compositeSQl = `CREATE TYPE inventory_item AS (
    name            text,
    supplier_id     integer,
    price           numeric
);
CREATE TABLE on_hand (
    item      inventory_item,
    count     integer
);
INSERT INTO on_hand VALUES (ROW('fuzzy dice', 42, 1.99), 1000);`

func compositeTypesTest(t *testing.T, db *sql.DB, dbURL string) {
	ctx := context.Background()
	_, err := db.ExecContext(ctx, compositeSQl)
	require.NoError(t, err)

	handle, err := drivers.Open("postgres", map[string]any{"database_url": dbURL}, false, activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)
	require.NotNil(t, handle)

	sqlStore, _ := handle.AsSQLStore()
	to, err := drivers.Open("duckdb", map[string]any{"dsn": ""}, false, activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)
	olap, _ := to.AsOLAP("")

	tr := NewSQLStoreToDuckDB(sqlStore, olap, zap.NewNop())
	err = tr.Transfer(ctx, map[string]any{"sql": "select * from on_hand;"}, map[string]any{"table": "sink"}, &drivers.TransferOptions{Progress: drivers.NoOpProgress{}})
	require.NoError(t, err)
}
