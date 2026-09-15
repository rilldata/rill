package databricks_test

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/metricsview"
	"strings"
	"testing"
	"time"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/storage"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/rilldata/rill/runtime/testruntime/testmode"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestOLAP(t *testing.T) {
	t.Skip("skipping due to inactive Databricks account")
	testmode.Expensive(t)

	_, olap := acquireTestDatabricks(t)
	tests := []struct {
		query  string
		args   []any
		result map[string]any
	}{
		{
			"SELECT TRUE AS bool_val",
			nil,
			map[string]any{"bool_val": true},
		},
		{
			"SELECT FALSE AS bool_val",
			nil,
			map[string]any{"bool_val": false},
		},
		{
			"SELECT CAST('2021-01-01' AS DATE) AS date_val",
			nil,
			map[string]any{"date_val": time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)},
		},
		{
			"SELECT CAST(3.14 AS DOUBLE) AS float_val",
			nil,
			map[string]any{"float_val": 3.14},
		},
		{
			"SELECT 'hello' AS string_val",
			nil,
			map[string]any{"string_val": "hello"},
		},
		{
			"SELECT double_col FROM integration_test.all_datatypes WHERE int32_col = 2147483647",
			nil,
			map[string]any{"double_col": 2.718},
		},
	}
	for _, test := range tests {
		t.Run(test.query, func(t *testing.T) {
			rows, err := olap.Query(t.Context(), &drivers.Statement{Query: test.query, Args: test.args})
			require.NoError(t, err)
			defer rows.Close()
			for rows.Next() {
				res := make(map[string]any)
				err = rows.MapScan(res)
				require.NoError(t, err)
				require.Equal(t, test.result, res)
			}
			require.NoError(t, rows.Err())
		})
	}
}

func TestComplexTypes(t *testing.T) {
	t.Skip("skipping due to inactive Databricks account")
	testmode.Expensive(t)

	_, olap := acquireTestDatabricks(t)

	// Test non-null complex types (id=1)
	rows, err := olap.Query(t.Context(), &drivers.Statement{
		Query: "SELECT array_col, map_col, struct_col FROM integration_test.all_datatypes WHERE id = 1",
	})
	require.NoError(t, err)
	defer rows.Close()

	require.True(t, rows.Next())
	res := make(map[string]any)
	err = rows.MapScan(res)
	require.NoError(t, err)

	require.Equal(t, "[1,2,3]", res["array_col"])
	require.Equal(t, `{"city":"New York"}`, res["map_col"])
	require.Equal(t, `{"city":"New York","zip":10001}`, res["struct_col"])

	require.False(t, rows.Next())
	require.NoError(t, rows.Err())

	// Test null complex types (id=3)
	rows2, err := olap.Query(t.Context(), &drivers.Statement{
		Query: "SELECT array_col, map_col, struct_col FROM integration_test.all_datatypes WHERE id = 3",
	})
	require.NoError(t, err)
	defer rows2.Close()

	require.True(t, rows2.Next())
	res2 := make(map[string]any)
	err = rows2.MapScan(res2)
	require.NoError(t, err)

	require.Nil(t, res2["array_col"])
	require.Nil(t, res2["map_col"])
	// Databricks expands NULL structs into their fields with null values
	require.Equal(t, `{"city":null,"zip":null}`, res2["struct_col"])

	require.False(t, rows2.Next())
	require.NoError(t, rows2.Err())
}

func TestEmptyRows(t *testing.T) {
	t.Skip("skipping due to inactive Databricks account")
	testmode.Expensive(t)

	_, olap := acquireTestDatabricks(t)
	rows, err := olap.Query(t.Context(), &drivers.Statement{Query: "SELECT int32_col, double_col FROM integration_test.all_datatypes LIMIT 0"})
	require.NoError(t, err)
	defer rows.Close()

	sc := rows.Schema
	require.Len(t, sc.Fields, 2)
	require.Equal(t, "int32_col", sc.Fields[0].Name)
	require.Equal(t, "double_col", sc.Fields[1].Name)
	require.False(t, rows.Next())
	require.Nil(t, rows.Err())
}

func TestLoadDDL(t *testing.T) {
	t.Skip("skipping due to inactive Databricks account")
	testmode.Expensive(t)

	_, olap := acquireTestDatabricks(t)
	table, err := olap.InformationSchema().Lookup(t.Context(), "", "integration_test", "all_datatypes")
	require.NoError(t, err)
	err = olap.InformationSchema().LoadDDL(t.Context(), table)
	require.NoError(t, err)
	require.Contains(t, strings.ToUpper(table.DDL), "ALL_DATATYPES")
}

func TestDryRun(t *testing.T) {
	t.Skip("skipping due to inactive Databricks account")
	testmode.Expensive(t)

	_, olap := acquireTestDatabricks(t)
	_, err := olap.Query(t.Context(), &drivers.Statement{
		Query:  "SELECT * FROM integration_test.all_datatypes WHERE int32_col = 2147483647",
		DryRun: true,
	})
	require.NoError(t, err)
}

func TestQuerySchema(t *testing.T) {
	t.Skip("skipping due to inactive Databricks account")
	testmode.Expensive(t)

	_, olap := acquireTestDatabricks(t)
	schema, err := olap.QuerySchema(t.Context(), "SELECT int32_col, string_col FROM integration_test.all_datatypes", nil)
	require.NoError(t, err)
	require.Len(t, schema.Fields, 2)
	require.Equal(t, "int32_col", schema.Fields[0].Name)
	require.Equal(t, "string_col", schema.Fields[1].Name)
}

func TestUnnestDimension(t *testing.T) {
	t.Skip("skipping due to inactive Databricks account")
	testmode.Expensive(t)
	_, olap := acquireTestDatabricks(t)

	// Rows with overlapping and empty arrays, so joins that duplicate source rows are detectable.
	// The table is created in the DSN's current schema.
	name := "test_unnest_" + uuid.New().String()[:8]
	t.Cleanup(func() {
		err := olap.Exec(context.Background(), &drivers.Statement{Query: "DROP TABLE IF EXISTS " + name})
		require.NoError(t, err)
	})
	err := olap.Exec(t.Context(), &drivers.Statement{Query: "CREATE TABLE " + name + " AS SELECT CAST(id AS BIGINT) AS id, tags FROM VALUES (1, array('a', 'b')), (2, array('b')), (3, array('c')), (4, CAST(array() AS ARRAY<STRING>)) AS t(id, tags)"})
	require.NoError(t, err)

	mv := &runtimev1.MetricsViewSpec{
		Table: name,
		Dimensions: []*runtimev1.MetricsViewSpec_Dimension{
			{Name: "tags", Column: "tags", Unnest: true},
			{Name: "id", Column: "id"},
		},
		Measures: []*runtimev1.MetricsViewSpec_Measure{
			{Name: "count", Expression: "count(*)", Type: runtimev1.MetricsViewSpec_MEASURE_TYPE_SIMPLE},
		},
	}

	// Same query shape as the executor's dimension validation.
	dialect := olap.Dialect()
	escapeTable := dialect.EscapeTable(mv.Database, mv.DatabaseSchema, mv.Table)
	sel, unnestClause, err := dialect.DimensionSelect(escapeTable, mv.Dimensions[0])
	require.NoError(t, err)
	err = olap.Exec(t.Context(), &drivers.Statement{Query: fmt.Sprintf("SELECT %s FROM %s %s GROUP BY 1", sel, escapeTable, unnestClause), DryRun: true})
	require.NoError(t, err)

	tagsFilter := func(op metricsview.Operator, val any) *metricsview.Expression {
		return &metricsview.Expression{Condition: &metricsview.Condition{
			Operator:    op,
			Expressions: []*metricsview.Expression{{Name: "tags"}, {Value: val}},
		}}
	}
	count := func(where *metricsview.Expression) *metricsview.Query {
		return &metricsview.Query{Measures: []metricsview.Measure{{Name: "count"}}, Where: where}
	}

	tests := []struct {
		name string
		qry  *metricsview.Query
		want []map[string]any
	}{
		{
			name: "group by unnest dimension",
			qry: &metricsview.Query{
				Dimensions: []metricsview.Dimension{{Name: "tags"}},
				Measures:   []metricsview.Measure{{Name: "count"}},
				Sort:       []metricsview.Sort{{Name: "tags"}},
			},
			want: []map[string]any{
				{"tags": "a", "count": int64(1)},
				{"tags": "b", "count": int64(2)},
				{"tags": "c", "count": int64(1)},
			},
		},
		{
			// Row 1 matches both values but must be counted once.
			name: "in filter counts each source row once",
			qry:  count(tagsFilter(metricsview.OperatorIn, []any{"a", "b"})),
			want: []map[string]any{{"count": int64(2)}},
		},
		{
			// Excludes rows containing 'a' even if they also contain other values; keeps the empty array.
			name: "nin filter excludes rows containing any listed value",
			qry:  count(tagsFilter(metricsview.OperatorNin, []any{"a"})),
			want: []map[string]any{{"count": int64(3)}},
		},
		{
			name: "eq filter",
			qry:  count(tagsFilter(metricsview.OperatorEq, "b")),
			want: []map[string]any{{"count": int64(2)}},
		},
		{
			name: "neq filter excludes rows containing the value",
			qry:  count(tagsFilter(metricsview.OperatorNeq, "b")),
			want: []map[string]any{{"count": int64(2)}},
		},
		{
			name: "ilike filter",
			qry:  count(tagsFilter(metricsview.OperatorIlike, "%B%")),
			want: []map[string]any{{"count": int64(2)}},
		},
		{
			name: "eq filter with no match",
			qry:  count(tagsFilter(metricsview.OperatorEq, "missing")),
			want: []map[string]any{{"count": int64(0)}},
		},
		{
			name: "filter combined with group by on another dimension",
			qry: &metricsview.Query{
				Dimensions: []metricsview.Dimension{{Name: "id"}},
				Measures:   []metricsview.Measure{{Name: "count"}},
				Where:      tagsFilter(metricsview.OperatorIn, []any{"a", "b"}),
				Sort:       []metricsview.Sort{{Name: "id"}},
			},
			want: []map[string]any{
				{"id": int64(1), "count": int64(1)},
				{"id": int64(2), "count": int64(1)},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.qry.MetricsView = mv.Table
			ast, err := metricsview.NewAST(mv, allowAllSecurity{}, tt.qry, dialect)
			require.NoError(t, err)
			sql, args, err := ast.SQL()
			require.NoError(t, err)
			require.Equal(t, tt.want, queryRows(t, olap, sql, args))
		})
	}
}

func queryRows(t *testing.T, olap drivers.OLAPStore, query string, args []any) []map[string]any {
	rows, err := olap.Query(t.Context(), &drivers.Statement{Query: query, Args: args})
	require.NoError(t, err)
	defer rows.Close()
	var res []map[string]any
	for rows.Next() {
		row := make(map[string]any)
		require.NoError(t, rows.MapScan(row))
		res = append(res, row)
	}
	require.NoError(t, rows.Err())
	return res
}

type allowAllSecurity struct{}

func (allowAllSecurity) CanAccessField(string) bool         { return true }
func (allowAllSecurity) RowFilter() string                  { return "" }
func (allowAllSecurity) QueryFilter() *runtimev1.Expression { return nil }

func acquireTestDatabricks(t *testing.T) (drivers.Handle, drivers.OLAPStore) {
	cfg := testruntime.AcquireConnector(t, "databricks")
	conn, err := drivers.Open("databricks", "", "default", cfg, storage.MustNew(t.TempDir(), nil), activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)
	t.Cleanup(func() { conn.Close() })

	olap, ok := conn.AsOLAP("default")
	require.True(t, ok)

	return conn, olap
}
