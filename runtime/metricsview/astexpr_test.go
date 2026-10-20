package metricsview

import (
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/drivers/clickhouse"
	"github.com/rilldata/rill/runtime/drivers/databricks"
	"github.com/rilldata/rill/runtime/drivers/duckdb"
	"github.com/rilldata/rill/runtime/drivers/snowflake"
	"github.com/stretchr/testify/require"
)

func TestArrayContainsCondition(t *testing.T) {
	mv := &runtimev1.MetricsViewSpec{
		Table:     "test_table",
		Database:  "",
		Connector: "",
		Dimensions: []*runtimev1.MetricsViewSpec_Dimension{
			{Name: "id", Column: "id"},
			{Name: "tags", Column: "tags", Unnest: true},
			{Name: "city", Column: "city"},
		},
		Measures: []*runtimev1.MetricsViewSpec_Measure{
			{Name: "count", Expression: "count(*)", Type: runtimev1.MetricsViewSpec_MEASURE_TYPE_SIMPLE},
		},
	}
	sec := skipMetricsViewSecurity{}

	tests := []struct {
		name     string
		dialect  drivers.Dialect
		dims     []Dimension
		where    *Expression
		wantSQL  string
		wantArgs []any
	}{
		{
			name:    "duckdb: in on unnest dim uses list_has_any",
			dialect: duckdb.DialectDuckDB,
			where: &Expression{Condition: &Condition{
				Operator: OperatorIn,
				Expressions: []*Expression{
					{Name: "tags"},
					{Value: []any{"a", "b", "c"}},
				},
			}},
			wantSQL:  `(list_has_any(("tags"), [?,?,?]))`,
			wantArgs: []any{"a", "b", "c"},
		},
		{
			name:    "duckdb: nin on unnest dim uses NOT list_has_any",
			dialect: duckdb.DialectDuckDB,
			where: &Expression{Condition: &Condition{
				Operator: OperatorNin,
				Expressions: []*Expression{
					{Name: "tags"},
					{Value: []any{"a", "b"}},
				},
			}},
			wantSQL:  `(NOT list_has_any(("tags"), [?,?]))`,
			wantArgs: []any{"a", "b"},
		},
		{
			name:    "duckdb: in on unnest dim with empty list",
			dialect: duckdb.DialectDuckDB,
			where: &Expression{Condition: &Condition{
				Operator: OperatorIn,
				Expressions: []*Expression{
					{Name: "tags"},
					{Value: []any{}},
				},
			}},
			wantSQL:  `FALSE`,
			wantArgs: nil,
		},
		{
			name:    "duckdb: nin on unnest dim with empty list",
			dialect: duckdb.DialectDuckDB,
			where: &Expression{Condition: &Condition{
				Operator: OperatorNin,
				Expressions: []*Expression{
					{Name: "tags"},
					{Value: []any{}},
				},
			}},
			wantSQL:  `TRUE`,
			wantArgs: nil,
		},
		{
			name:    "duckdb: in on unnest dim with null value in list",
			dialect: duckdb.DialectDuckDB,
			where: &Expression{Condition: &Condition{
				Operator: OperatorIn,
				Expressions: []*Expression{
					{Name: "tags"},
					{Value: []any{"a", nil, "b"}},
				},
			}},
			wantSQL:  `(list_has_any(("tags"), [?,?,?]))`,
			wantArgs: []any{"a", nil, "b"},
		},
		{
			name:    "clickhouse: in on unnest dim uses hasAny",
			dialect: clickhouse.DialectClickhouse,
			where: &Expression{Condition: &Condition{
				Operator: OperatorIn,
				Expressions: []*Expression{
					{Name: "tags"},
					{Value: []any{"a", "b"}},
				},
			}},
			wantSQL:  `(hasAny(("tags"), [?,?]))`,
			wantArgs: []any{"a", "b"},
		},
		{
			name:    "clickhouse: nin on unnest dim uses NOT hasAny",
			dialect: clickhouse.DialectClickhouse,
			where: &Expression{Condition: &Condition{
				Operator: OperatorNin,
				Expressions: []*Expression{
					{Name: "tags"},
					{Value: []any{"a", "b"}},
				},
			}},
			wantSQL:  `(NOT hasAny(("tags"), [?,?]))`,
			wantArgs: []any{"a", "b"},
		},
		{
			name:    "clickhouse: in on unnest dim with null values",
			dialect: clickhouse.DialectClickhouse,
			where: &Expression{Condition: &Condition{
				Operator: OperatorIn,
				Expressions: []*Expression{
					{Name: "tags"},
					{Value: []any{nil, "a"}},
				},
			}},
			wantSQL:  `(hasAny(("tags"), [?,?]))`,
			wantArgs: []any{nil, "a"},
		},
		{
			name:    "databricks: in on unnest dim uses arrays_overlap",
			dialect: databricks.DialectDatabricks,
			where: &Expression{Condition: &Condition{
				Operator: OperatorIn,
				Expressions: []*Expression{
					{Name: "tags"},
					{Value: []any{"a", "b"}},
				},
			}},
			wantSQL:  "(COALESCE(arrays_overlap((`tags`), array(?,?)), FALSE))",
			wantArgs: []any{"a", "b"},
		},
		{
			name:    "databricks: nin on unnest dim uses NOT arrays_overlap",
			dialect: databricks.DialectDatabricks,
			where: &Expression{Condition: &Condition{
				Operator: OperatorNin,
				Expressions: []*Expression{
					{Name: "tags"},
					{Value: []any{"a", "b"}},
				},
			}},
			wantSQL:  "(NOT COALESCE(arrays_overlap((`tags`), array(?,?)), FALSE))",
			wantArgs: []any{"a", "b"},
		},
		{
			name:    "databricks: in on unnest dim already in select falls back to normal IN",
			dialect: databricks.DialectDatabricks,
			dims:    []Dimension{{Name: "tags"}},
			where: &Expression{Condition: &Condition{
				Operator: OperatorIn,
				Expressions: []*Expression{
					{Name: "tags"},
					{Value: []any{"a", "b"}},
				},
			}},
			wantSQL:  "((`t0`.`tags`) IN (?,?))",
			wantArgs: []any{"a", "b"},
		},
		{
			name:    "snowflake: in on unnest dim uses FILTER",
			dialect: snowflake.DialectSnowflake,
			where: &Expression{Condition: &Condition{
				Operator: OperatorIn,
				Expressions: []*Expression{
					{Name: "tags"},
					{Value: []any{"a", "b"}},
				},
			}},
			wantSQL:  "COALESCE(ARRAY_SIZE(FILTER(tags, t2 -> ((t2::VARCHAR) IN (?,?)))) > 0, FALSE)",
			wantArgs: []any{"a", "b"},
		},
		{
			name:    "duckdb: in on non-unnest dim uses normal IN",
			dialect: duckdb.DialectDuckDB,
			where: &Expression{Condition: &Condition{
				Operator: OperatorIn,
				Expressions: []*Expression{
					{Name: "city"},
					{Value: []any{"NYC", "LA"}},
				},
			}},
			wantSQL:  `(("city") IN (?,?))`,
			wantArgs: []any{"NYC", "LA"},
		},
		{
			name:    "duckdb: nin on non-unnest dim uses normal NOT IN",
			dialect: duckdb.DialectDuckDB,
			where: &Expression{Condition: &Condition{
				Operator: OperatorNin,
				Expressions: []*Expression{
					{Name: "city"},
					{Value: []any{"NYC"}},
				},
			}},
			wantSQL:  `(("city") NOT IN (?) OR ("city") IS NULL)`,
			wantArgs: []any{"NYC"},
		},
		{
			name:    "duckdb: in on unnest dim nested in AND",
			dialect: duckdb.DialectDuckDB,
			where: &Expression{Condition: &Condition{
				Operator: OperatorAnd,
				Expressions: []*Expression{
					{Condition: &Condition{
						Operator: OperatorIn,
						Expressions: []*Expression{
							{Name: "tags"},
							{Value: []any{"a", "b"}},
						},
					}},
					{Condition: &Condition{
						Operator: OperatorEq,
						Expressions: []*Expression{
							{Name: "city"},
							{Value: "NYC"},
						},
					}},
				},
			}},
			wantSQL:  `((list_has_any(("tags"), [?,?])) AND (("city") = ?))`,
			wantArgs: []any{"a", "b", "NYC"},
		},
		{
			name:    "duckdb: in on unnest dim already in select falls back to normal IN",
			dialect: duckdb.DialectDuckDB,
			dims:    []Dimension{{Name: "tags"}},
			where: &Expression{Condition: &Condition{
				Operator: OperatorIn,
				Expressions: []*Expression{
					{Name: "tags"},
					{Value: []any{"a", "b"}},
				},
			}},
			// When tags is in the query dimensions, it's already unnested via lateral join, so falls back to normal IN.
			// The AST qualifies the column with a table alias from the lateral unnest join.
			wantSQL:  `(("t0"."tags") IN (?,?))`,
			wantArgs: []any{"a", "b"},
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

			sql, args, err := ast.SQLForExpression(tt.where, nil, false, false)
			require.NoError(t, err)
			require.Equal(t, tt.wantSQL, sql)
			require.Equal(t, tt.wantArgs, args)
		})
	}
}

func TestLookupDimensionFilter(t *testing.T) {
	mv := &runtimev1.MetricsViewSpec{
		Table: "facts",
		Dimensions: []*runtimev1.MetricsViewSpec_Dimension{
			{Name: "app_id", Column: "app_id"},
			{Name: "category", Column: "app_id", LookupTable: "app_dict", LookupKeyColumn: "app_id", LookupValueColumn: "category"},
			{Name: "category_or_unknown", Column: "app_id", LookupTable: "app_dict", LookupKeyColumn: "app_id", LookupValueColumn: "category", LookupDefaultExpression: "'Unknown'"},
		},
		Measures: []*runtimev1.MetricsViewSpec_Measure{
			{Name: "count", Expression: "count(*)", Type: runtimev1.MetricsViewSpec_MEASURE_TYPE_SIMPLE},
		},
	}
	sec := skipMetricsViewSecurity{}

	const lookupSQL = `dictGet('app_dict', 'category', "app_id")`
	const lookupWithDefaultSQL = `dictGetOrDefault('app_dict', 'category', "app_id", 'Unknown')`

	tests := []struct {
		name     string
		dims     []Dimension
		where    *Expression
		wantSQL  string
		wantArgs []any
	}{
		{
			name:     "in filters through the lookup table keys",
			where:    &Expression{Condition: &Condition{Operator: OperatorIn, Expressions: []*Expression{{Name: "category"}, {Value: []any{"Games"}}}}},
			wantSQL:  `"app_id" IN (SELECT "app_id" FROM "app_dict" WHERE ((category) IN (?)))`,
			wantArgs: []any{"Games"},
		},
		{
			name:     "ilike filters through the lookup table keys",
			where:    &Expression{Condition: &Condition{Operator: OperatorIlike, Expressions: []*Expression{{Name: "category"}, {Value: "%game%"}}}},
			wantSQL:  `"app_id" IN (SELECT "app_id" FROM "app_dict" WHERE ((category)::Nullable(TEXT) ILIKE ?::Nullable(TEXT)))`,
			wantArgs: []any{"%game%"},
		},
		{
			name:     "neq filters through the lookup table keys",
			where:    &Expression{Condition: &Condition{Operator: OperatorNeq, Expressions: []*Expression{{Name: "category"}, {Value: "Games"}}}},
			wantSQL:  `"app_id" IN (SELECT "app_id" FROM "app_dict" WHERE ((category) != ?))`,
			wantArgs: []any{"Games"},
		},
		{
			name:     "nin with null filters through the lookup table keys",
			where:    &Expression{Condition: &Condition{Operator: OperatorNin, Expressions: []*Expression{{Name: "category"}, {Value: []any{"Games", nil}}}}},
			wantSQL:  `"app_id" IN (SELECT "app_id" FROM "app_dict" WHERE ((category) NOT IN (?) AND (category) IS NOT NULL))`,
			wantArgs: []any{"Games"},
		},
		{
			name:     "in with null filters on the dimension expression",
			where:    &Expression{Condition: &Condition{Operator: OperatorIn, Expressions: []*Expression{{Name: "category"}, {Value: []any{nil}}}}},
			wantSQL:  `((` + lookupSQL + `) IS NULL)`,
			wantArgs: nil,
		},
		{
			name:     "in with null and values filters on the dimension expression",
			where:    &Expression{Condition: &Condition{Operator: OperatorIn, Expressions: []*Expression{{Name: "category"}, {Value: []any{"Games", nil}}}}},
			wantSQL:  `((` + lookupSQL + `) IN (?) OR (` + lookupSQL + `) IS NULL)`,
			wantArgs: []any{"Games"},
		},
		{
			name:     "eq null filters on the dimension expression",
			where:    &Expression{Condition: &Condition{Operator: OperatorEq, Expressions: []*Expression{{Name: "category"}, {Value: nil}}}},
			wantSQL:  `((` + lookupSQL + `) IS NULL)`,
			wantArgs: nil,
		},
		{
			name:     "nin filters on the dimension expression",
			where:    &Expression{Condition: &Condition{Operator: OperatorNin, Expressions: []*Expression{{Name: "category"}, {Value: []any{"Games"}}}}},
			wantSQL:  `((` + lookupSQL + `) NOT IN (?) OR (` + lookupSQL + `) IS NULL)`,
			wantArgs: []any{"Games"},
		},
		{
			name:     "nilike filters on the dimension expression",
			where:    &Expression{Condition: &Condition{Operator: OperatorNilike, Expressions: []*Expression{{Name: "category"}, {Value: "%game%"}}}},
			wantSQL:  `((` + lookupSQL + `)::Nullable(TEXT) NOT ILIKE ?::Nullable(TEXT) OR (` + lookupSQL + `) IS NULL)`,
			wantArgs: []any{"%game%"},
		},
		{
			name:     "nin on a selected dimension filters on the dimension expression",
			dims:     []Dimension{{Name: "category"}},
			where:    &Expression{Condition: &Condition{Operator: OperatorNin, Expressions: []*Expression{{Name: "category"}, {Value: []any{"Games"}}}}},
			wantSQL:  `((` + lookupSQL + `) NOT IN (?) OR (` + lookupSQL + `) IS NULL)`,
			wantArgs: []any{"Games"},
		},
		{
			name:     "in with default expression also checks the default",
			where:    &Expression{Condition: &Condition{Operator: OperatorIn, Expressions: []*Expression{{Name: "category_or_unknown"}, {Value: []any{"Games"}}}}},
			wantSQL:  `("app_id" IN (SELECT "app_id" FROM "app_dict" WHERE ((category) IN (?))) OR ((('Unknown') IN (?)) AND "app_id" NOT IN (SELECT "app_id" FROM "app_dict")))`,
			wantArgs: []any{"Games", "Games"},
		},
		{
			name:     "nin with default expression filters on the dimension expression",
			where:    &Expression{Condition: &Condition{Operator: OperatorNin, Expressions: []*Expression{{Name: "category_or_unknown"}, {Value: []any{"Games"}}}}},
			wantSQL:  `((` + lookupWithDefaultSQL + `) NOT IN (?) OR (` + lookupWithDefaultSQL + `) IS NULL)`,
			wantArgs: []any{"Games"},
		},
		{
			name:     "eq null with default expression filters on the dimension expression",
			where:    &Expression{Condition: &Condition{Operator: OperatorEq, Expressions: []*Expression{{Name: "category_or_unknown"}, {Value: nil}}}},
			wantSQL:  `((` + lookupWithDefaultSQL + `) IS NULL)`,
			wantArgs: nil,
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

			ast, err := NewAST(mv, sec, qry, clickhouse.DialectClickhouse)
			require.NoError(t, err)

			sql, args, err := ast.SQLForExpression(tt.where, nil, false, false)
			require.NoError(t, err)
			require.Equal(t, tt.wantSQL, sql)
			require.Equal(t, tt.wantArgs, args)
		})
	}
}
