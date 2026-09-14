package snowflake_test

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/metricsview"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/storage"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/rilldata/rill/runtime/testruntime/testmode"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestOLAP(t *testing.T) {
	t.Skip("skipping due to inactive Snowflake account")
	testmode.Expensive(t)

	_, olap := acquireTestSnowflake(t)
	tests := []struct {
		query  string
		args   []any
		result map[string]any
	}{
		{
			"SELECT TRUE AS bool",
			nil,
			map[string]any{"BOOL": true},
		},
		{
			"SELECT FALSE AS bool",
			nil,
			map[string]any{"BOOL": false},
		},
		{
			"SELECT '2021-01-01'::DATE AS date",
			nil,
			map[string]any{"DATE": time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)},
		},
		{
			"SELECT '2025-01-31 23:59:59.999999'::TIMESTAMP_NTZ AS datetime",
			nil,
			map[string]any{"DATETIME": time.Date(2025, 1, 31, 23, 59, 59, 999999000, time.UTC)},
		},
		{
			"SELECT 99999999999999999999999999999999999999 AS integer",
			nil,
			map[string]any{"INTEGER": "99999999999999999999999999999999999999"},
		},
		{
			"SELECT 99999999999999999999999999999.999999999::NUMBER(38,9) AS number",
			nil,
			map[string]any{"NUMBER": "99999999999999999999999999999.999999999"},
		},
		{
			"SELECT 0.1::NUMBER(10,1) AS number",
			nil,
			map[string]any{"NUMBER": "0.1"},
		},
		{
			"SELECT 3.14::FLOAT AS number",
			nil,
			map[string]any{"NUMBER": 3.14},
		},
		{
			"SELECT ARRAY_CONSTRUCT(1, 2, 3) AS arr",
			nil,
			map[string]any{"ARR": "[\n  1,\n  2,\n  3\n]"},
		},
		{
			"SELECT OBJECT_CONSTRUCT('a', 1, 'b', 'abc') AS obj",
			nil,
			map[string]any{"OBJ": "{\n  \"a\": 1,\n  \"b\": \"abc\"\n}"},
		},
		{
			"SELECT '23:59:59.999999'::TIME AS t",
			nil,
			map[string]any{"T": time.Date(1, 1, 1, 23, 59, 59, 999999000, time.UTC)},
		},
		{
			"SELECT float_col FROM integration_test.public.all_datatypes WHERE int32_col = ?",
			[]any{2147483647},
			map[string]any{"FLOAT_COL": 3.14},
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

func TestEmptyRows(t *testing.T) {
	t.Skip("skipping due to inactive Snowflake account")
	testmode.Expensive(t)

	_, olap := acquireTestSnowflake(t)
	rows, err := olap.Query(t.Context(), &drivers.Statement{Query: "SELECT int32_col, float_col FROM integration_test.public.all_datatypes LIMIT 0"})
	require.NoError(t, err)
	defer rows.Close()

	sc := rows.Schema
	require.Len(t, sc.Fields, 2)
	require.Equal(t, "INT32_COL", sc.Fields[0].Name)
	require.Equal(t, "FLOAT_COL", sc.Fields[1].Name)
	require.False(t, rows.Next())
	require.Nil(t, rows.Err())
}

func TestComplexTypes(t *testing.T) {
	t.Skip("skipping due to inactive Snowflake account")
	testmode.Expensive(t)

	_, olap := acquireTestSnowflake(t)

	// Test complex data types (variant, array, object)
	rows, err := olap.Query(t.Context(), &drivers.Statement{
		Query: "SELECT variant_col, array_col, object_col FROM integration_test.public.all_datatypes WHERE id = 1",
	})
	require.NoError(t, err)
	defer rows.Close()

	require.True(t, rows.Next())
	res := make(map[string]any)
	err = rows.MapScan(res)
	require.NoError(t, err)

	// Verify values
	var varCol map[string]string
	err = json.Unmarshal([]byte(res["VARIANT_COL"].(string)), &varCol)
	require.NoError(t, err)
	require.Equal(t, map[string]string{"key": "value"}, varCol)

	var arrCol []int
	err = json.Unmarshal([]byte(res["ARRAY_COL"].(string)), &arrCol)
	require.NoError(t, err)
	require.Equal(t, []int{1, 2, 3}, arrCol)

	var objCol map[string]any
	err = json.Unmarshal([]byte(res["OBJECT_COL"].(string)), &objCol)
	require.NoError(t, err)
	require.Equal(t, map[string]any{"city": "New York"}, objCol)

	require.False(t, rows.Next())
	require.NoError(t, rows.Err())
}

func TestLoadDDL(t *testing.T) {
	t.Skip("skipping due to inactive Snowflake account")
	testmode.Expensive(t)
	_, olap := acquireTestSnowflake(t)

	table, err := olap.InformationSchema().Lookup(t.Context(), "INTEGRATION_TEST", "public", "all_datatypes")
	require.NoError(t, err)
	err = olap.InformationSchema().LoadDDL(t.Context(), table)
	require.NoError(t, err)
	require.Contains(t, table.DDL, "create")
	require.Contains(t, strings.ToUpper(table.DDL), "ALL_DATATYPES")
}

func TestDryRun(t *testing.T) {
	t.Skip("skipping due to inactive Snowflake account")
	testmode.Expensive(t)

	_, olap := acquireTestSnowflake(t)
	// Dry run query
	_, err := olap.Query(t.Context(), &drivers.Statement{
		Query:  "SELECT * FROM integration_test.public.all_datatypes WHERE int32_col = ?",
		Args:   []any{2147483647},
		DryRun: true,
	})
	require.NoError(t, err)
}

// TestUnnestDimension creates a table in the DSN's current database and schema, so the DSN must point at a writable schema.
func TestUnnestDimension(t *testing.T) {
	testmode.Expensive(t)
	_, olap := acquireTestSnowflake(t)

	// Rows with overlapping and empty arrays, so joins that duplicate source rows are detectable.
	// The driver returns NUMBER columns as strings.
	name := "test_unnest_" + uuid.New().String()[:8]
	t.Cleanup(func() {
		err := olap.Exec(context.Background(), &drivers.Statement{Query: "DROP TABLE IF EXISTS " + name})
		require.NoError(t, err)
	})
	err := olap.Exec(t.Context(), &drivers.Statement{Query: "CREATE TABLE " + name + " AS SELECT 1 AS id, ARRAY_CONSTRUCT('a', 'b') AS tags UNION ALL SELECT 2, ARRAY_CONSTRUCT('b') UNION ALL SELECT 3, ARRAY_CONSTRUCT('c') UNION ALL SELECT 4, ARRAY_CONSTRUCT()"})
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
				{"tags": "a", "count": "1"},
				{"tags": "b", "count": "2"},
				{"tags": "c", "count": "1"},
			},
		},
		{
			// Row 1 matches both values but must be counted once.
			name: "in filter counts each source row once",
			qry:  count(tagsFilter(metricsview.OperatorIn, []any{"a", "b"})),
			want: []map[string]any{{"count": "2"}},
		},
		{
			// Excludes rows containing 'a' even if they also contain other values; keeps the empty array.
			name: "nin filter excludes rows containing any listed value",
			qry:  count(tagsFilter(metricsview.OperatorNin, []any{"a"})),
			want: []map[string]any{{"count": "3"}},
		},
		{
			name: "eq filter",
			qry:  count(tagsFilter(metricsview.OperatorEq, "b")),
			want: []map[string]any{{"count": "2"}},
		},
		{
			name: "neq filter excludes rows containing the value",
			qry:  count(tagsFilter(metricsview.OperatorNeq, "b")),
			want: []map[string]any{{"count": "2"}},
		},
		{
			name: "ilike filter",
			qry:  count(tagsFilter(metricsview.OperatorIlike, "%B%")),
			want: []map[string]any{{"count": "2"}},
		},
		{
			name: "eq filter with no match",
			qry:  count(tagsFilter(metricsview.OperatorEq, "missing")),
			want: []map[string]any{{"count": "0"}},
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
				{"id": "1", "count": "1"},
				{"id": "2", "count": "1"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.qry.MetricsView = name
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

func acquireTestSnowflake(t *testing.T) (drivers.Handle, drivers.OLAPStore) {
	cfg := testruntime.AcquireConnector(t, "snowflake")
	conn, err := drivers.Open("snowflake", "", "default", cfg, storage.MustNew(t.TempDir(), nil), activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)
	t.Cleanup(func() { conn.Close() })

	olap, ok := conn.AsOLAP("default")
	require.True(t, ok)

	return conn, olap
}
