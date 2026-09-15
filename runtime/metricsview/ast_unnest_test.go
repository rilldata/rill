package metricsview

import (
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/drivers/bigquery"
	"github.com/rilldata/rill/runtime/drivers/databricks"
	"github.com/rilldata/rill/runtime/drivers/duckdb"
	"github.com/rilldata/rill/runtime/drivers/snowflake"
	"github.com/stretchr/testify/require"
)

func TestUnnestSQL(t *testing.T) {
	mv := &runtimev1.MetricsViewSpec{
		Table: "test_table",
		Dimensions: []*runtimev1.MetricsViewSpec_Dimension{
			{Name: "tags", Column: "tags", Unnest: true},
			{Name: "city", Column: "city"},
		},
		Measures: []*runtimev1.MetricsViewSpec_Measure{
			{Name: "count", Expression: "count(*)", Type: runtimev1.MetricsViewSpec_MEASURE_TYPE_SIMPLE},
		},
	}
	sec := skipMetricsViewSecurity{}

	tagsEqA := &Expression{Condition: &Condition{
		Operator:    OperatorEq,
		Expressions: []*Expression{{Name: "tags"}, {Value: "a"}},
	}}

	tests := []struct {
		name     string
		dialect  drivers.Dialect
		dims     []Dimension
		where    *Expression
		wantSQL  string
		wantArgs []any
	}{
		{
			name:    "bigquery: group by unnest dim",
			dialect: bigquery.DialectBigQuery,
			dims:    []Dimension{{Name: "tags"}},
			wantSQL: "SELECT (`tags`) AS `tags`, (count(*)) AS `count` FROM `test_table`, UNNEST(`tags`) AS `tags` GROUP BY 1",
		},
		{
			name:     "bigquery: filter on unnest dim not in select",
			dialect:  bigquery.DialectBigQuery,
			dims:     []Dimension{{Name: "city"}},
			where:    tagsEqA,
			wantSQL:  "SELECT (`city`) AS `city`, (count(*)) AS `count` FROM `test_table` WHERE EXISTS (SELECT 1 FROM UNNEST(`tags`) AS `t0` WHERE ((`t0`) = ?)) GROUP BY 1",
			wantArgs: []any{"a"},
		},
		{
			name:    "snowflake: group by unnest dim",
			dialect: snowflake.DialectSnowflake,
			dims:    []Dimension{{Name: "tags"}},
			wantSQL: `SELECT (t0.tags::VARCHAR) AS "tags", (count(*)) AS "count" FROM test_table, LATERAL FLATTEN(INPUT => tags) t0 (seq, key, path, index, tags, this) GROUP BY 1`,
		},
		{
			name:     "snowflake: filter on unnest dim not in select",
			dialect:  snowflake.DialectSnowflake,
			dims:     []Dimension{{Name: "city"}},
			where:    tagsEqA,
			wantSQL:  `SELECT (city) AS "city", (count(*)) AS "count" FROM test_table WHERE COALESCE(ARRAY_SIZE(FILTER(tags, t0 -> ((t0::VARCHAR) = ?))) > 0, FALSE) GROUP BY 1`,
			wantArgs: []any{"a"},
		},
		{
			name:    "snowflake: nin filter on unnest dim not in select",
			dialect: snowflake.DialectSnowflake,
			dims:    []Dimension{{Name: "city"}},
			where: &Expression{Condition: &Condition{
				Operator:    OperatorNin,
				Expressions: []*Expression{{Name: "tags"}, {Value: []any{"a", "b"}}},
			}},
			wantSQL:  `SELECT (city) AS "city", (count(*)) AS "count" FROM test_table WHERE NOT COALESCE(ARRAY_SIZE(FILTER(tags, t0 -> ((t0::VARCHAR) IN (?,?)))) > 0, FALSE) GROUP BY 1`,
			wantArgs: []any{"a", "b"},
		},
		{
			name:    "databricks: group by unnest dim",
			dialect: databricks.DialectDatabricks,
			dims:    []Dimension{{Name: "tags"}},
			wantSQL: "SELECT (`t0`.`tags`) AS `tags`, (count(*)) AS `count` FROM `test_table` LATERAL VIEW EXPLODE(`tags`) t0 AS `tags` GROUP BY 1",
		},
		{
			name:     "databricks: filter on unnest dim not in select",
			dialect:  databricks.DialectDatabricks,
			dims:     []Dimension{{Name: "city"}},
			where:    tagsEqA,
			wantSQL:  "SELECT (`city`) AS `city`, (count(*)) AS `count` FROM `test_table` WHERE COALESCE(EXISTS(`tags`, t0 -> ((t0) = ?)), FALSE) GROUP BY 1",
			wantArgs: []any{"a"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qry := &Query{
				MetricsView: "test",
				Dimensions:  tt.dims,
				Measures:    []Measure{{Name: "count"}},
				Where:       tt.where,
			}

			ast, err := NewAST(mv, sec, qry, tt.dialect)
			require.NoError(t, err)

			sql, args, err := ast.SQL()
			require.NoError(t, err)
			require.Equal(t, tt.wantSQL, sql)
			require.Equal(t, tt.wantArgs, args)
		})
	}
}

// Measure filters produce "dim IN (subquery)". Dialects with an array-contains fast path must still handle them.
func TestUnnestSubqueryFilterSQL(t *testing.T) {
	mv := &runtimev1.MetricsViewSpec{
		Table: "test_table",
		Dimensions: []*runtimev1.MetricsViewSpec_Dimension{
			{Name: "tags", Column: "tags", Unnest: true},
			{Name: "city", Column: "city"},
		},
		Measures: []*runtimev1.MetricsViewSpec_Measure{
			{Name: "count", Expression: "count(*)", Type: runtimev1.MetricsViewSpec_MEASURE_TYPE_SIMPLE},
		},
	}
	where := &Expression{Condition: &Condition{
		Operator: OperatorNin,
		Expressions: []*Expression{
			{Name: "tags"},
			{Subquery: &Subquery{
				Dimension: Dimension{Name: "tags"},
				Measures:  []Measure{{Name: "count"}},
				Having:    &Expression{Condition: &Condition{Operator: OperatorGt, Expressions: []*Expression{{Name: "count"}, {Value: 10}}}},
			}},
		},
	}}
	// The subquery is the metrics view grouped by the unnest dimension, with the having clause applied in an outer select.
	sub := map[string]string{
		"duckdb":     `(SELECT "tags" FROM (SELECT ("t2"."tags") AS "tags", ("t2"."count") AS "count" FROM (SELECT ("t0"."tags") AS "tags", (count(*)) AS "count" FROM "test_table", LATERAL UNNEST("tags") t0("tags") GROUP BY 1) t2 WHERE (("t2"."count") > ?)))`,
		"databricks": "(SELECT `tags` FROM (SELECT (`t2`.`tags`) AS `tags`, (`t2`.`count`) AS `count` FROM (SELECT (`t0`.`tags`) AS `tags`, (count(*)) AS `count` FROM `test_table` LATERAL VIEW EXPLODE(`tags`) t0 AS `tags` GROUP BY 1) t2 WHERE ((`t2`.`count`) > ?)))",
		"snowflake":  `(SELECT "tags" FROM (SELECT (t2."tags") AS "tags", (t2."count") AS "count" FROM (SELECT (t0.tags::VARCHAR) AS "tags", (count(*)) AS "count" FROM test_table, LATERAL FLATTEN(INPUT => tags) t0 (seq, key, path, index, tags, this) GROUP BY 1) t2 WHERE ((t2."count") > ?)))`,
		"bigquery":   "(SELECT `tags` FROM (SELECT (`t2`.`tags`) AS `tags`, (`t2`.`count`) AS `count` FROM (SELECT (`tags`) AS `tags`, (count(*)) AS `count` FROM `test_table`, UNNEST(`tags`) AS `tags` GROUP BY 1) t2 WHERE ((`t2`.`count`) > ?)))",
	}
	tests := []struct {
		dialect drivers.Dialect
		want    string
	}{
		// No native form: correlated EXISTS over the unnest join.
		{duckdb.DialectDuckDB, `WHERE NOT EXISTS (SELECT 1 FROM LATERAL UNNEST("tags") t0("tags") WHERE (("t0"."tags") IN ` + sub["duckdb"] + `)) GROUP BY 1`},
		// Lambdas cannot contain subqueries: aggregate the subquery into an array.
		{databricks.DialectDatabricks, "WHERE (NOT COALESCE(arrays_overlap((`tags`), (SELECT collect_list(s.`tags`) FROM " + sub["databricks"] + " AS s)), FALSE)) GROUP BY 1"},
		// IN subquery inside FILTER hits an internal error: aggregate into an array and use ARRAY_CONTAINS.
		{snowflake.DialectSnowflake, `WHERE (NOT COALESCE(ARRAY_SIZE(FILTER((tags), x -> ARRAY_CONTAINS(x::VARCHAR::VARIANT, (SELECT ARRAY_AGG(s."tags") FROM ` + sub["snowflake"] + ` AS s)))) > 0, FALSE)) GROUP BY 1`},
		// A table-referencing subquery inside correlated EXISTS cannot be de-correlated: join the subquery to the unnested array instead.
		{bigquery.DialectBigQuery, "WHERE (NOT EXISTS (SELECT 1 FROM UNNEST((`tags`)) AS e JOIN " + sub["bigquery"] + " AS s ON e = s.`tags`)) GROUP BY 1"},
	}
	for _, tt := range tests {
		t.Run(tt.dialect.String(), func(t *testing.T) {
			qry := &Query{MetricsView: "test", Dimensions: []Dimension{{Name: "city"}}, Measures: []Measure{{Name: "count"}}, Where: where}
			ast, err := NewAST(mv, skipMetricsViewSecurity{}, qry, tt.dialect)
			require.NoError(t, err)
			sql, args, err := ast.SQL()
			require.NoError(t, err)
			require.Contains(t, sql, tt.want)
			require.Equal(t, []any{10}, args)
		})
	}
}
