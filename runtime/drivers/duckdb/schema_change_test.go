package duckdb

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/storage"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestInsertTableAsSelectSchemaChangeModes(t *testing.T) {
	tests := []struct {
		name       string
		mode       drivers.OnSchemaChange
		sourceSQL  string
		wantErr    string
		wantCols   []string
		assertRows func(*testing.T, *connection, string)
	}{
		{
			name:      "ignore added column",
			mode:      drivers.OnSchemaChangeIgnore,
			sourceSQL: `SELECT 1 AS id, 1 AS partition_id, 'new-1' AS old_value, 42 AS new_value`,
			wantCols:  []string{"id", "partition_id", "old_value"},
			assertRows: func(t *testing.T, c *connection, table string) {
				requireQueryStrings(t, c, fmt.Sprintf(`SELECT old_value FROM %s ORDER BY id`, safeSQLName(table)), [][]string{{"new-1"}, {"old-2"}})
			},
		},
		{
			name:      "ignore removed column",
			mode:      drivers.OnSchemaChangeIgnore,
			sourceSQL: `SELECT 1 AS id, 1 AS partition_id`,
			wantCols:  []string{"id", "partition_id", "old_value"},
			assertRows: func(t *testing.T, c *connection, table string) {
				requireQueryStrings(t, c, fmt.Sprintf(`SELECT coalesce(old_value, 'NULL') FROM %s ORDER BY id`, safeSQLName(table)), [][]string{{"NULL"}, {"old-2"}})
			},
		},
		{
			name:      "ignore retains target type",
			mode:      drivers.OnSchemaChangeIgnore,
			sourceSQL: `SELECT 1 AS id, 1 AS partition_id, 123::INTEGER AS old_value`,
			wantCols:  []string{"id", "partition_id", "old_value"},
			assertRows: func(t *testing.T, c *connection, table string) {
				requireQueryStrings(t, c, fmt.Sprintf(`SELECT typeof(old_value), old_value FROM %s WHERE id = 1`, safeSQLName(table)), [][]string{{"VARCHAR", "123"}})
			},
		},
		{
			name:       "fail added column",
			mode:       drivers.OnSchemaChangeFail,
			sourceSQL:  `SELECT 1 AS id, 1 AS partition_id, 'new-1' AS old_value, 42 AS new_value`,
			wantErr:    "does not match the schema of table",
			wantCols:   []string{"id", "partition_id", "old_value"},
			assertRows: assertOriginalSchemaChangeRows,
		},
		{
			name:       "omitted mode is rejected",
			sourceSQL:  `SELECT 1 AS id, 1 AS partition_id, 'new-1' AS old_value, 42 AS new_value`,
			wantErr:    `invalid on_schema_change mode ""`,
			wantCols:   []string{"id", "partition_id", "old_value"},
			assertRows: assertOriginalSchemaChangeRows,
		},
		{
			name:       "fail removed column",
			mode:       drivers.OnSchemaChangeFail,
			sourceSQL:  `SELECT 1 AS id, 1 AS partition_id`,
			wantErr:    "does not match the schema of table",
			wantCols:   []string{"id", "partition_id", "old_value"},
			assertRows: assertOriginalSchemaChangeRows,
		},
		{
			name:      "fail retains target type",
			mode:      drivers.OnSchemaChangeFail,
			sourceSQL: `SELECT 1 AS id, 1 AS partition_id, 123::INTEGER AS old_value`,
			wantCols:  []string{"id", "partition_id", "old_value"},
			assertRows: func(t *testing.T, c *connection, table string) {
				requireQueryStrings(t, c, fmt.Sprintf(`SELECT typeof(old_value), old_value FROM %s ORDER BY id`, safeSQLName(table)), [][]string{{"VARCHAR", "123"}, {"VARCHAR", "old-2"}})
			},
		},
		{
			name:      "append added and retain removed columns",
			mode:      drivers.OnSchemaChangeAppendNewColumns,
			sourceSQL: `SELECT 1 AS id, 1 AS partition_id, 42 AS new_value`,
			wantCols:  []string{"id", "partition_id", "old_value", "new_value"},
			assertRows: func(t *testing.T, c *connection, table string) {
				requireQueryStrings(t, c, fmt.Sprintf(`SELECT coalesce(old_value, 'NULL'), coalesce(new_value::VARCHAR, 'NULL') FROM %s ORDER BY id`, safeSQLName(table)), [][]string{{"NULL", "42"}, {"old-2", "NULL"}})
			},
		},
		{
			name:      "append retains target type",
			mode:      drivers.OnSchemaChangeAppendNewColumns,
			sourceSQL: `SELECT 1 AS id, 1 AS partition_id, 123::INTEGER AS old_value, 42 AS new_value`,
			wantCols:  []string{"id", "partition_id", "old_value", "new_value"},
			assertRows: func(t *testing.T, c *connection, table string) {
				requireQueryStrings(t, c, fmt.Sprintf(`SELECT typeof(old_value), old_value, coalesce(new_value::VARCHAR, 'NULL') FROM %s ORDER BY id`, safeSQLName(table)), [][]string{{"VARCHAR", "123", "42"}, {"VARCHAR", "old-2", "NULL"}})
			},
		},
	}

	for _, strategy := range []drivers.IncrementalStrategy{drivers.IncrementalStrategyMerge, drivers.IncrementalStrategyPartitionOverwrite} {
		for _, tt := range tests {
			t.Run(fmt.Sprintf("%s/%s", strategy, tt.name), func(t *testing.T) {
				c := newSchemaChangeTestConnection(t)
				table := strings.ReplaceAll(fmt.Sprintf("schema_%s_%s", strategy, tt.name), " ", "_")
				_, err := c.createTableAsSelect(context.Background(), table, `
					SELECT 1 AS id, 1 AS partition_id, 'old-1' AS old_value
					UNION ALL
					SELECT 2 AS id, 2 AS partition_id, 'old-2' AS old_value
				`, &createTableOptions{})
				require.NoError(t, err)

				_, err = c.insertTableAsSelect(context.Background(), table, tt.sourceSQL, schemaChangeInsertOptions(strategy, tt.mode))
				if tt.wantErr != "" {
					require.ErrorContains(t, err, tt.wantErr)
				} else {
					require.NoError(t, err)
				}
				require.Equal(t, tt.wantCols, schemaChangeColumnNames(t, c, table))
				tt.assertRows(t, c, table)
			})
		}
	}
}

func TestInsertTableAsSelectHandlesReorderedQuotedColumns(t *testing.T) {
	for _, strategy := range []drivers.IncrementalStrategy{drivers.IncrementalStrategyMerge, drivers.IncrementalStrategyPartitionOverwrite} {
		t.Run(string(strategy), func(t *testing.T) {
			c := newSchemaChangeTestConnection(t)
			table := "quoted_" + string(strategy)
			_, err := c.createTableAsSelect(context.Background(), table, `SELECT 1 AS "Key", 10 AS "Partition Key", 'old' AS "select"`, &createTableOptions{})
			require.NoError(t, err)

			opts := schemaChangeInsertOptions(strategy, drivers.OnSchemaChangeFail)
			opts.UniqueKey = []string{"KEY"}
			opts.PartitionBy = `"Partition Key"`
			_, err = c.insertTableAsSelect(context.Background(), table, `SELECT 'new' AS "SELECT", 1 AS "key", 10 AS "partition key"`, opts)
			require.NoError(t, err)
			requireQueryStrings(t, c, fmt.Sprintf(`SELECT "Key"::VARCHAR, "Partition Key"::VARCHAR, "select" FROM %s`, safeSQLName(table)), [][]string{{"1", "10", "new"}})
		})
	}
}

// TestInsertTableAsSelectAppendNewColumnsAcrossInserts covers the steady state after a column has been appended:
// a later insert must write into the appended column instead of adding it again, and an insert that stops
// producing it must leave it NULL.
func TestInsertTableAsSelectAppendNewColumnsAcrossInserts(t *testing.T) {
	for _, strategy := range []drivers.IncrementalStrategy{drivers.IncrementalStrategyMerge, drivers.IncrementalStrategyPartitionOverwrite} {
		t.Run(string(strategy), func(t *testing.T) {
			c := newSchemaChangeTestConnection(t)
			table := "successive_" + string(strategy)
			_, err := c.createTableAsSelect(context.Background(), table, `SELECT 1 AS id, 1 AS partition_id`, &createTableOptions{})
			require.NoError(t, err)

			opts := schemaChangeInsertOptions(strategy, drivers.OnSchemaChangeAppendNewColumns)
			_, err = c.insertTableAsSelect(context.Background(), table, `SELECT 2 AS id, 2 AS partition_id, 'two' AS new_value`, opts)
			require.NoError(t, err)
			require.Equal(t, []string{"id", "partition_id", "new_value"}, schemaChangeColumnNames(t, c, table))

			// The column now exists, so this insert populates it rather than appending a duplicate.
			_, err = c.insertTableAsSelect(context.Background(), table, `SELECT 3 AS id, 3 AS partition_id, 'three' AS new_value`, opts)
			require.NoError(t, err)
			require.Equal(t, []string{"id", "partition_id", "new_value"}, schemaChangeColumnNames(t, c, table))

			// The column is no longer produced, so it is retained and left NULL for the new row.
			_, err = c.insertTableAsSelect(context.Background(), table, `SELECT 4 AS id, 4 AS partition_id`, opts)
			require.NoError(t, err)
			require.Equal(t, []string{"id", "partition_id", "new_value"}, schemaChangeColumnNames(t, c, table))

			requireQueryStrings(t, c, fmt.Sprintf(`SELECT id::VARCHAR, coalesce(new_value, 'NULL') FROM %s ORDER BY id`, safeSQLName(table)),
				[][]string{{"1", "NULL"}, {"2", "two"}, {"3", "three"}, {"4", "NULL"}})
		})
	}
}

// TestInsertTableAsSelectUncastableValue documents that the target's column types are retained,
// so DuckDB fails the insert when it cannot convert an incoming value.
func TestInsertTableAsSelectUncastableValue(t *testing.T) {
	for _, mode := range []drivers.OnSchemaChange{drivers.OnSchemaChangeFail, drivers.OnSchemaChangeIgnore, drivers.OnSchemaChangeAppendNewColumns} {
		t.Run(string(mode), func(t *testing.T) {
			c := newSchemaChangeTestConnection(t)
			table := "uncastable_" + string(mode)
			_, err := c.createTableAsSelect(context.Background(), table, `SELECT 1 AS id, 1 AS partition_id, 10 AS value`, &createTableOptions{})
			require.NoError(t, err)

			_, err = c.insertTableAsSelect(context.Background(), table, `SELECT 2 AS id, 2 AS partition_id, 'abc' AS value`, schemaChangeInsertOptions(drivers.IncrementalStrategyMerge, mode))
			require.ErrorContains(t, err, "Conversion Error")
			require.Equal(t, []string{"id", "partition_id", "value"}, schemaChangeColumnNames(t, c, table))
			requireQueryStrings(t, c, fmt.Sprintf(`SELECT id::VARCHAR, value::VARCHAR FROM %s`, safeSQLName(table)), [][]string{{"1", "10"}})
		})
	}
}

// TestInsertTableAsSelectKeyColumnDropped covers new data that no longer contains the unique_key or partition_by
// column. The partition_overwrite DELETE references the partition expression unqualified in a subquery, so without
// an explicit check it binds to the target table as a correlated reference and deletes every row.
func TestInsertTableAsSelectKeyColumnDropped(t *testing.T) {
	tests := []struct {
		strategy drivers.IncrementalStrategy
		wantErr  string
	}{
		{drivers.IncrementalStrategyMerge, `the new data does not contain the "id" column from "unique_key"`},
		{drivers.IncrementalStrategyPartitionOverwrite, `failed to resolve the "partition_by" expression "partition_id" against the new data`},
	}
	for _, tt := range tests {
		for _, mode := range []drivers.OnSchemaChange{drivers.OnSchemaChangeIgnore, drivers.OnSchemaChangeAppendNewColumns} {
			t.Run(fmt.Sprintf("%s/%s", tt.strategy, mode), func(t *testing.T) {
				c := newSchemaChangeTestConnection(t)
				table := fmt.Sprintf("keydrop_%s_%s", tt.strategy, mode)
				_, err := c.createTableAsSelect(context.Background(), table, `
					SELECT 1 AS id, 1 AS partition_id, 'a' AS value
					UNION ALL SELECT 2 AS id, 2 AS partition_id, 'b' AS value
				`, &createTableOptions{})
				require.NoError(t, err)

				_, err = c.insertTableAsSelect(context.Background(), table, `SELECT 'new' AS value`, schemaChangeInsertOptions(tt.strategy, mode))
				require.ErrorContains(t, err, tt.wantErr)

				// The target table must be left untouched.
				require.Equal(t, []string{"id", "partition_id", "value"}, schemaChangeColumnNames(t, c, table))
				requireQueryStrings(t, c, fmt.Sprintf(`SELECT id::VARCHAR, partition_id::VARCHAR, value FROM %s ORDER BY id`, safeSQLName(table)),
					[][]string{{"1", "1", "a"}, {"2", "2", "b"}})
			})
		}
	}
}

// TestInsertTableAsSelectKeyColumnMissingFromTarget covers new data that contains a unique_key column the target lacks.
// Adding it would give every existing row a NULL key, and merge treats NULL keys as equal,
// so a NULL incoming key would otherwise delete every existing row.
func TestInsertTableAsSelectKeyColumnMissingFromTarget(t *testing.T) {
	for _, mode := range []drivers.OnSchemaChange{drivers.OnSchemaChangeFail, drivers.OnSchemaChangeIgnore, drivers.OnSchemaChangeAppendNewColumns} {
		t.Run(string(mode), func(t *testing.T) {
			c := newSchemaChangeTestConnection(t)
			table := "keymissing_" + string(mode)
			_, err := c.createTableAsSelect(context.Background(), table, `
				SELECT 1 AS partition_id, 'a' AS value
				UNION ALL SELECT 2 AS partition_id, 'b' AS value
			`, &createTableOptions{})
			require.NoError(t, err)

			_, err = c.insertTableAsSelect(context.Background(), table, `SELECT NULL::INTEGER AS id, 1 AS partition_id, 'new' AS value`, schemaChangeInsertOptions(drivers.IncrementalStrategyMerge, mode))
			require.ErrorContains(t, err, `does not contain the "id" column from "unique_key"`)

			require.Equal(t, []string{"partition_id", "value"}, schemaChangeColumnNames(t, c, table))
			requireQueryStrings(t, c, fmt.Sprintf(`SELECT partition_id::VARCHAR, value FROM %s ORDER BY partition_id`, safeSQLName(table)),
				[][]string{{"1", "a"}, {"2", "b"}})
		})
	}
}

// TestInsertTableAsSelectPartitionByExpression checks that a partition_by SQL expression, rather than a plain
// column reference, still resolves against the new data.
func TestInsertTableAsSelectPartitionByExpression(t *testing.T) {
	c := newSchemaChangeTestConnection(t)
	table := "partition_expr"
	_, err := c.createTableAsSelect(context.Background(), table, `
		SELECT 1 AS id, '2024-01-01 05:00:00'::TIMESTAMP AS ts, 'a' AS value
		UNION ALL SELECT 2 AS id, '2024-01-02 05:00:00'::TIMESTAMP AS ts, 'b' AS value
	`, &createTableOptions{})
	require.NoError(t, err)

	opts := schemaChangeInsertOptions(drivers.IncrementalStrategyPartitionOverwrite, drivers.OnSchemaChangeAppendNewColumns)
	opts.PartitionBy = "date_trunc('day', ts)"
	_, err = c.insertTableAsSelect(context.Background(), table, `
		SELECT 3 AS id, '2024-01-01 09:00:00'::TIMESTAMP AS ts, 'replaced' AS value, 42 AS extra
	`, opts)
	require.NoError(t, err)

	// Only the 2024-01-01 partition was overwritten.
	requireQueryStrings(t, c, fmt.Sprintf(`SELECT id::VARCHAR, value, coalesce(extra::VARCHAR, 'NULL') FROM %s ORDER BY id`, safeSQLName(table)),
		[][]string{{"2", "b", "NULL"}, {"3", "replaced", "42"}})
}

func TestInsertTableAsSelectEmptyPartitionSkipsSchemaHandling(t *testing.T) {
	for _, strategy := range []drivers.IncrementalStrategy{drivers.IncrementalStrategyMerge, drivers.IncrementalStrategyPartitionOverwrite} {
		t.Run(string(strategy), func(t *testing.T) {
			c := newSchemaChangeTestConnection(t)
			table := "empty_" + string(strategy)
			_, err := c.createTableAsSelect(context.Background(), table, `SELECT 1 AS id, 1 AS partition_id, 'old' AS value`, &createTableOptions{})
			require.NoError(t, err)

			_, err = c.insertTableAsSelect(context.Background(), table, `SELECT 1 AS id, 42 AS added WHERE false`, schemaChangeInsertOptions(strategy, drivers.OnSchemaChangeFail))
			require.NoError(t, err)
			require.Equal(t, []string{"id", "partition_id", "value"}, schemaChangeColumnNames(t, c, table))
			requireQueryStrings(t, c, fmt.Sprintf(`SELECT id::VARCHAR, partition_id::VARCHAR, value FROM %s`, safeSQLName(table)), [][]string{{"1", "1", "old"}})
		})
	}
}

func newSchemaChangeTestConnection(t *testing.T) *connection {
	t.Helper()
	handle, err := Driver{}.Open("", "default", map[string]any{}, storage.MustNew(t.TempDir(), nil), activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)
	c := handle.(*connection)
	require.NoError(t, c.Migrate(context.Background()))
	c.AsOLAP("default")
	return c
}

func schemaChangeInsertOptions(strategy drivers.IncrementalStrategy, mode drivers.OnSchemaChange) *InsertTableOptions {
	return &InsertTableOptions{
		Strategy:       strategy,
		UniqueKey:      []string{"id"},
		PartitionBy:    "partition_id",
		OnSchemaChange: mode,
	}
}

func schemaChangeColumnNames(t *testing.T, c *connection, table string) []string {
	t.Helper()
	tbl, err := c.InformationSchema().Lookup(context.Background(), "", "", table)
	require.NoError(t, err)
	columns := make([]string, len(tbl.Schema.Fields))
	for i, field := range tbl.Schema.Fields {
		columns[i] = field.Name
	}
	return columns
}

func requireQueryStrings(t *testing.T, c *connection, query string, want [][]string) {
	t.Helper()
	rows, err := c.Query(context.Background(), &drivers.Statement{Query: query})
	require.NoError(t, err)
	defer rows.Close()

	got := make([][]string, 0, len(want))
	for rows.Next() {
		row := make([]string, len(want[0]))
		values := make([]any, len(row))
		for i := range row {
			values[i] = &row[i]
		}
		require.NoError(t, rows.Scan(values...))
		got = append(got, row)
	}
	require.NoError(t, rows.Err())
	require.Equal(t, want, got)
}

func assertOriginalSchemaChangeRows(t *testing.T, c *connection, table string) {
	t.Helper()
	requireQueryStrings(t, c, fmt.Sprintf(`SELECT old_value FROM %s ORDER BY id`, safeSQLName(table)), [][]string{{"old-1"}, {"old-2"}})
}

func TestOnSchemaChangeValidation(t *testing.T) {
	tests := []struct {
		name    string
		props   ModelOutputProperties
		opts    drivers.ModelExecuteOptions
		wantErr string
		want    drivers.OnSchemaChange
	}{
		{
			name:  "unset defaults to fail for merge",
			props: ModelOutputProperties{UniqueKey: []string{"id"}},
			opts:  drivers.ModelExecuteOptions{Incremental: true},
			want:  drivers.OnSchemaChangeFail,
		},
		{
			name:  "unset defaults to fail for partitioned run",
			props: ModelOutputProperties{},
			opts:  drivers.ModelExecuteOptions{Incremental: true, PartitionRun: true, PartitionKey: "p1"},
			want:  drivers.OnSchemaChangeFail,
		},
		{
			name:  "unset stays unset for append",
			props: ModelOutputProperties{},
			opts:  drivers.ModelExecuteOptions{Incremental: true},
		},
		{
			name:  "merge without partitions",
			props: ModelOutputProperties{UniqueKey: []string{"id"}, OnSchemaChange: drivers.OnSchemaChangeAppendNewColumns},
			opts:  drivers.ModelExecuteOptions{Incremental: true},
			want:  drivers.OnSchemaChangeAppendNewColumns,
		},
		{
			name:  "partition_overwrite without partitions",
			props: ModelOutputProperties{PartitionBy: "day", IncrementalStrategy: drivers.IncrementalStrategyPartitionOverwrite, OnSchemaChange: drivers.OnSchemaChangeIgnore},
			opts:  drivers.ModelExecuteOptions{Incremental: true},
			want:  drivers.OnSchemaChangeIgnore,
		},
		{
			name:  "partitioned run",
			props: ModelOutputProperties{OnSchemaChange: drivers.OnSchemaChangeFail},
			opts:  drivers.ModelExecuteOptions{Incremental: true, PartitionRun: true, PartitionKey: "p1"},
			want:  drivers.OnSchemaChangeFail,
		},
		{
			name:    "invalid mode",
			props:   ModelOutputProperties{UniqueKey: []string{"id"}, OnSchemaChange: "replace"},
			opts:    drivers.ModelExecuteOptions{Incremental: true},
			wantErr: `invalid on_schema_change mode "replace"`,
		},
		{
			name:    "append strategy",
			props:   ModelOutputProperties{IncrementalStrategy: drivers.IncrementalStrategyAppend, OnSchemaChange: drivers.OnSchemaChangeIgnore},
			opts:    drivers.ModelExecuteOptions{Incremental: true},
			wantErr: `only supported for the "merge" and "partition_overwrite" incremental strategies`,
		},
		{
			name:    "non incremental model",
			props:   ModelOutputProperties{OnSchemaChange: drivers.OnSchemaChangeIgnore},
			opts:    drivers.ModelExecuteOptions{},
			wantErr: `only supported for the "merge" and "partition_overwrite" incremental strategies`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			props := tt.props
			opts := tt.opts
			opts.ModelExecutorOptions = &drivers.ModelExecutorOptions{InputConnector: "duckdb", OutputConnector: "duckdb"}
			err := props.validateAndApplyDefaults(&opts, &ModelInputProperties{SQL: "SELECT 1"})
			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, props.OnSchemaChange)
		})
	}
}
