package duckdb

import (
	"context"
	"database/sql"
	"testing"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"fmt"
	_ "github.com/rilldata/rill/runtime/drivers/mysql"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"time"
)

var mysqlInitStmt = `
CREATE TABLE all_data_types_table (
    id INT AUTO_INCREMENT PRIMARY KEY,
    sample_char CHAR(1),
    sample_varchar VARCHAR(100),
    sample_tinytext TINYTEXT,
    sample_text TEXT,
    sample_mediumtext MEDIUMTEXT,
    sample_longtext LONGTEXT,
    sample_binary BINARY(1),
    sample_varbinary VARBINARY(100),
    sample_tinyblob TINYBLOB,
    sample_blob BLOB,
    sample_mediumblob MEDIUMBLOB,
    sample_longblob LONGBLOB,
    sample_enum ENUM('value1', 'value2'),
    sample_set SET('value1', 'value2'),
    sample_bit BIT(8),
    sample_tinyint TINYINT,
    sample_tinyint_unsigned TINYINT UNSIGNED NOT NULL,
    sample_smallint SMALLINT,
    sample_smallint_unsigned SMALLINT UNSIGNED NOT NULL,
    sample_mediumint MEDIUMINT,
    sample_mediumint_unsigned MEDIUMINT UNSIGNED NOT NULL,
    sample_int INT,
    sample_int_unsigned INT UNSIGNED NOT NULL,
    sample_bigint BIGINT,
    sample_bigint_unsigned BIGINT UNSIGNED NOT NULL,
    sample_float FLOAT,
    sample_double DOUBLE,
    sample_decimal DECIMAL(10,2),
    sample_date DATE,
    sample_datetime DATETIME,
    sample_timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    sample_time TIME,
    sample_year YEAR,
    sample_json JSON
);

INSERT INTO all_data_types_table (sample_char, sample_varchar, sample_tinytext, sample_text, sample_mediumtext, sample_longtext, sample_binary, sample_varbinary, sample_tinyblob, sample_blob, sample_mediumblob, sample_longblob, sample_enum, sample_set, sample_bit, sample_tinyint, sample_tinyint_unsigned, sample_smallint, sample_smallint_unsigned, sample_mediumint, sample_mediumint_unsigned, sample_int, sample_int_unsigned, sample_bigint, sample_bigint_unsigned, sample_float, sample_double, sample_decimal, sample_date, sample_datetime, sample_timestamp, sample_time, sample_year, sample_json) 
VALUES ('A', 'Sample Text', 'Tiny Text', 'Some Longer Text.', 'Medium Length Text', 'This is an example of really long text for the LONGTEXT column.', BINARY '1', 'Sample Binary', 'Tiny Blob Data', 'Sample Blob Data', 'Medium Blob Data', 'Long Blob Data', 'value1', 'value1,value2', b'10101010', -128, 255, -32768, 65535, -8388608, 16777215, -2147483648, 4294967295, -9223372036854775808, 18446744073709551615, 123.45, 1234567890.123, 12345.67, '2023-01-01', '2023-01-01 12:00:00', CURRENT_TIMESTAMP, '12:00:00', 2023, JSON_OBJECT('key', 'value'));

INSERT INTO all_data_types_table (sample_char, sample_varchar, sample_tinytext, sample_text, sample_mediumtext, sample_longtext, sample_binary, sample_varbinary, sample_tinyblob, sample_blob, sample_mediumblob, sample_longblob, sample_enum, sample_set, sample_bit, sample_tinyint, sample_tinyint_unsigned, sample_smallint, sample_smallint_unsigned, sample_mediumint, sample_mediumint_unsigned, sample_int, sample_int_unsigned, sample_bigint, sample_bigint_unsigned, sample_float, sample_double, sample_decimal, sample_date, sample_datetime, sample_timestamp, sample_time, sample_year, sample_json) 
VALUES (NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, 0, NULL, 0, NULL, 0, NULL, 0, NULL, 0, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL);
`

func TestMySQLToDuckDBTransfer(t *testing.T) {
	ctx := context.Background()
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		Started: true,
		ContainerRequest: testcontainers.ContainerRequest{
			WaitingFor:   wait.ForLog("mysqld: ready for connections").WithOccurrence(2).WithStartupTimeout(15 * time.Second),
			Image:        "mysql:8.3.0",
			ExposedPorts: []string{"3306/tcp"},
			Env: map[string]string{
				"MYSQL_ROOT_PASSWORD": "mypassword",
				"MYSQL_DATABASE":      "mydb",
				"MYSQL_USER":          "myuser",
				"MYSQL_PASSWORD":      "mypassword",
			},
		},
	})
	require.NoError(t, err)
	defer container.Terminate(ctx)

	host, err := container.Host(ctx)
	require.NoError(t, err)
	port, err := container.MappedPort(ctx, "3306/tcp")
	require.NoError(t, err)

	dsn := fmt.Sprintf("myuser:mypassword@tcp(%s:%d)/mydb?multiStatements=true", host, port.Int())

	db, err := sql.Open("mysql", dsn)
	require.NoError(t, err)
	defer db.Close()

	t.Run("AllDataTypes", func(t *testing.T) { allMySQLDataTypesTest(t, db, dsn) })
}

func allMySQLDataTypesTest(t *testing.T, db *sql.DB, dsn string) {
	ctx := context.Background()
	_, err := db.ExecContext(ctx, mysqlInitStmt)
	require.NoError(t, err)

	handle, err := drivers.Open("mysql", map[string]any{"dsn": dsn}, false, activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)
	require.NotNil(t, handle)

	sqlStore, _ := handle.AsSQLStore()
	to, err := drivers.Open("duckdb", map[string]any{"dsn": ":memory:"}, false, activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)
	olap, _ := to.AsOLAP("")

	tr := NewSQLStoreToDuckDB(sqlStore, olap, zap.NewNop())
	err = tr.Transfer(ctx, map[string]any{"sql": "select * from all_data_types_table;"}, map[string]any{"table": "sink"}, &drivers.TransferOptions{Progress: drivers.NoOpProgress{}})
	require.NoError(t, err)
	res, err := olap.Execute(context.Background(), &drivers.Statement{Query: "select count(*) from sink"})
	require.NoError(t, err)
	for res.Next() {
		var count int
		err = res.Rows.Scan(&count)
		require.NoError(t, err)
		require.Equal(t, count, 2)
	}
	require.NoError(t, res.Close())
	require.NoError(t, to.Close())
}
