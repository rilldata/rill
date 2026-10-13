package metricsview

import (
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/drivers/bigquery"
	"github.com/rilldata/rill/runtime/drivers/clickhouse"
	"github.com/rilldata/rill/runtime/drivers/databricks"
	"github.com/rilldata/rill/runtime/drivers/druid"
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

// Subquery filters remain supported for scalar dimensions and existing general unnest paths.
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
	base := drivers.NewBaseDialect(drivers.DialectNamePostgres, drivers.DoubleQuotesEscapeIdentifier, drivers.DoubleQuotesEscapeIdentifier)
	tests := []struct {
		dialect drivers.Dialect
		wantErr string
	}{
		{duckdb.DialectDuckDB, "the right value must be a list of values for an array IN condition"},
		{clickhouse.DialectClickhouse, "the right value must be a list of values for an array IN condition"},
		{databricks.DialectDatabricks, "the right value must be a list of values for an array IN condition"},
		{snowflake.DialectSnowflake, `dialect snowflake does not support subquery filters on unnest dimension "tags"`},
		{bigquery.DialectBigQuery, `dialect bigquery does not support subquery filters on unnest dimension "tags"`},
		{druid.DialectDruid, ""},
		{&base, ""},
	}
	for _, tt := range tests {
		for _, op := range []Operator{OperatorIn, OperatorNin} {
			for _, shape := range []struct {
				name string
				dim  string
				dims []Dimension
			}{
				{"unselected unnest dimension", "tags", []Dimension{{Name: "city"}}},
				{"selected unnest dimension", "tags", []Dimension{{Name: "tags"}}},
				{"scalar dimension", "city", []Dimension{{Name: "city"}}},
			} {
				t.Run(tt.dialect.String()+"/"+string(op)+"/"+shape.name, func(t *testing.T) {
					where := &Expression{Condition: &Condition{
						Operator: op,
						Expressions: []*Expression{
							{Name: shape.dim},
							{Subquery: &Subquery{
								Dimension: Dimension{Name: shape.dim},
								Measures:  []Measure{{Name: "count"}},
								Having:    &Expression{Condition: &Condition{Operator: OperatorGt, Expressions: []*Expression{{Name: "count"}, {Value: 10}}}},
							}},
						},
					}}
					qry := &Query{MetricsView: "test", Dimensions: shape.dims, Measures: []Measure{{Name: "count"}}, Where: where}
					ast, err := NewAST(mv, skipMetricsViewSecurity{}, qry, tt.dialect)
					if shape.name == "unselected unnest dimension" && tt.wantErr != "" {
						require.ErrorContains(t, err, tt.wantErr)
						return
					}
					require.NoError(t, err)
					sql, args, err := ast.SQL()
					require.NoError(t, err)
					require.Contains(t, sql, " IN (SELECT "+tt.dialect.EscapeAlias(shape.dim)+" FROM (")
					require.Equal(t, []any{10}, args)
				})
			}
		}
	}
}
