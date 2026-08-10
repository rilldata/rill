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
		mode       OnSchemaChange
		sourceSQL  string
		wantErr    string
		wantCols   []string
		assertRows func(*testing.T, *connection, string)
	}{
		{
			name:      "ignore added column",
			mode:      OnSchemaChangeIgnore,
			sourceSQL: `SELECT 1 AS id, 1 AS partition_id, 'new-1' AS old_value, 42 AS new_value`,
			wantCols:  []string{"id", "partition_id", "old_value"},
			assertRows: func(t *testing.T, c *connection, table string) {
				requireQueryStrings(t, c, fmt.Sprintf(`SELECT old_value FROM %s ORDER BY id`, safeSQLName(table)), [][]string{{"new-1"}, {"old-2"}})
			},
		},
		{
			name:       "ignore removed column",
			mode:       OnSchemaChangeIgnore,
			sourceSQL:  `SELECT 1 AS id, 1 AS partition_id`,
			wantErr:    "target columns to be absent from the source",
			wantCols:   []string{"id", "partition_id", "old_value"},
			assertRows: assertOriginalSchemaChangeRows,
		},
		{
			name:      "ignore retains target type",
			mode:      OnSchemaChangeIgnore,
			sourceSQL: `SELECT 1 AS id, 1 AS partition_id, 123::INTEGER AS old_value`,
			wantCols:  []string{"id", "partition_id", "old_value"},
			assertRows: func(t *testing.T, c *connection, table string) {
				requireQueryStrings(t, c, fmt.Sprintf(`SELECT typeof(old_value), old_value FROM %s WHERE id = 1`, safeSQLName(table)), [][]string{{"VARCHAR", "123"}})
			},
		},
		{
			name:       "fail added column",
			mode:       OnSchemaChangeFail,
			sourceSQL:  `SELECT 1 AS id, 1 AS partition_id, 'new-1' AS old_value, 42 AS new_value`,
			wantErr:    "detected schema differences",
			wantCols:   []string{"id", "partition_id", "old_value"},
			assertRows: assertOriginalSchemaChangeRows,
		},
		{
			name:       "omitted mode defaults to fail",
			sourceSQL:  `SELECT 1 AS id, 1 AS partition_id, 'new-1' AS old_value, 42 AS new_value`,
			wantErr:    "detected schema differences",
			wantCols:   []string{"id", "partition_id", "old_value"},
			assertRows: assertOriginalSchemaChangeRows,
		},
		{
			name:       "fail removed column",
			mode:       OnSchemaChangeFail,
			sourceSQL:  `SELECT 1 AS id, 1 AS partition_id`,
			wantErr:    "detected schema differences",
			wantCols:   []string{"id", "partition_id", "old_value"},
			assertRows: assertOriginalSchemaChangeRows,
		},
		{
			name:       "fail type change",
			mode:       OnSchemaChangeFail,
			sourceSQL:  `SELECT 1 AS id, 1 AS partition_id, 123::INTEGER AS old_value`,
			wantErr:    "type changes",
			wantCols:   []string{"id", "partition_id", "old_value"},
			assertRows: assertOriginalSchemaChangeRows,
		},
		{
			name:      "append added and retain removed columns",
			mode:      OnSchemaChangeAppendNewColumns,
			sourceSQL: `SELECT 1 AS id, 1 AS partition_id, 42 AS new_value`,
			wantCols:  []string{"id", "partition_id", "old_value", "new_value"},
			assertRows: func(t *testing.T, c *connection, table string) {
				requireQueryStrings(t, c, fmt.Sprintf(`SELECT coalesce(old_value, 'NULL'), coalesce(new_value::VARCHAR, 'NULL') FROM %s ORDER BY id`, safeSQLName(table)), [][]string{{"NULL", "42"}, {"old-2", "NULL"}})
			},
		},
		{
			name:      "append retains target type",
			mode:      OnSchemaChangeAppendNewColumns,
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

			opts := schemaChangeInsertOptions(strategy, OnSchemaChangeFail)
			opts.UniqueKey = []string{"KEY"}
			opts.PartitionBy = `"Partition Key"`
			_, err = c.insertTableAsSelect(context.Background(), table, `SELECT 'new' AS "SELECT", 1 AS "key", 10 AS "partition key"`, opts)
			require.NoError(t, err)
			requireQueryStrings(t, c, fmt.Sprintf(`SELECT "Key"::VARCHAR, "Partition Key"::VARCHAR, "select" FROM %s`, safeSQLName(table)), [][]string{{"1", "10", "new"}})
		})
	}
}

func TestInsertTableAsSelectEmptyPartitionSkipsSchemaHandling(t *testing.T) {
	for _, strategy := range []drivers.IncrementalStrategy{drivers.IncrementalStrategyMerge, drivers.IncrementalStrategyPartitionOverwrite} {
		t.Run(string(strategy), func(t *testing.T) {
			c := newSchemaChangeTestConnection(t)
			table := "empty_" + string(strategy)
			_, err := c.createTableAsSelect(context.Background(), table, `SELECT 1 AS id, 1 AS partition_id, 'old' AS value`, &createTableOptions{})
			require.NoError(t, err)

			_, err = c.insertTableAsSelect(context.Background(), table, `SELECT 1 AS id, 42 AS added WHERE false`, schemaChangeInsertOptions(strategy, OnSchemaChangeFail))
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

func schemaChangeInsertOptions(strategy drivers.IncrementalStrategy, mode OnSchemaChange) *InsertTableOptions {
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
