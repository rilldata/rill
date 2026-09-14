package metricsview

import (
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/drivers/bigquery"
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
			wantSQL:  "SELECT (`city`) AS `city`, (count(*)) AS `count` FROM `test_table`, UNNEST(`tags`) AS `tags` WHERE ((`tags`) = ?) GROUP BY 1",
			wantArgs: []any{"a"},
		},
		{
			name:    "snowflake: group by unnest dim",
			dialect: snowflake.DialectSnowflake,
			dims:    []Dimension{{Name: "tags"}},
			wantSQL: `SELECT (t0.tags) AS "tags", (count(*)) AS "count" FROM test_table, LATERAL (SELECT VALUE::VARCHAR AS tags FROM TABLE(FLATTEN(INPUT => tags))) t0 GROUP BY 1`,
		},
		{
			name:     "snowflake: filter on unnest dim not in select",
			dialect:  snowflake.DialectSnowflake,
			dims:     []Dimension{{Name: "city"}},
			where:    tagsEqA,
			wantSQL:  `SELECT (city) AS "city", (count(*)) AS "count" FROM test_table WHERE EXISTS (SELECT 1 FROM LATERAL (SELECT VALUE::VARCHAR AS tags FROM TABLE(FLATTEN(INPUT => tags))) t0 WHERE ((t0.tags) = ?)) GROUP BY 1`,
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
