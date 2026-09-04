package metricsview

import (
	"fmt"
	"strings"
	"testing"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/stretchr/testify/require"
)

func TestParseMeasureExpressionAccepts(t *testing.T) {
	cases := []struct {
		expr string
		refs []string
	}{
		{`revenue - cost`, []string{"revenue", "cost"}},
		{`revenue-cost`, []string{"revenue", "cost"}},
		{`(a + b) * 2`, []string{"a", "b"}},
		{`-a`, []string{"a"}},
		{`a % b`, []string{"a", "b"}},
		{`a * -1`, []string{"a"}},
		{`power(a, 2)`, []string{"a"}},
		{`coalesce(a, 0)`, []string{"a"}},
		{`coalesce(a, NULL)`, []string{"a"}},
		{`nullif(cost, 0)`, []string{"cost"}},
		{`round(a / b, 2)`, []string{"a", "b"}},
		{`abs(a) + floor(b) + ceil(c) + sqrt(d) + ln(e) + exp(f)`, []string{"a", "b", "c", "d", "e", "f"}},
		{`greatest(a, b, c)`, []string{"a", "b", "c"}},
		{`least(a, b)`, []string{"a", "b"}},
		{`"count" - a`, []string{"count", "a"}},
		{`1e6 * a`, []string{"a"}},
		{`0.4 * revenue`, []string{"revenue"}},
		{`revenue - revenue`, []string{"revenue"}},
		{`((((a))))`, []string{"a"}},
		{`"a--b" + x`, []string{"a--b", "x"}},
		{`"a/*b" + x`, []string{"a/*b", "x"}},
	}
	for _, c := range cases {
		t.Run(c.expr, func(t *testing.T) {
			e, err := ParseMeasureExpression(c.expr)
			require.NoError(t, err)
			require.Equal(t, c.refs, e.Refs())
		})
	}
}

func TestParseMeasureExpressionRejects(t *testing.T) {
	cases := []struct {
		expr        string
		errContains string
	}{
		{``, "empty"},
		{`   `, "empty"},
		{`revenue); DROP TABLE x;--`, "comments are not allowed"},
		{`revenue); DROP TABLE x`, "invalid expression"},
		{`a; SELECT 1`, "single expression"},
		{`(select secret from t)`, "subqueries"},
		{`exists(select 1)`, "unsupported"},
		{`sum(revenue)`, "aggregate function"},
		{`count(*)`, "aggregate function"},
		{`t.revenue`, "qualified references"},
		{`db.t.revenue`, "qualified references"},
		{`case when a > 0 then 1 end`, "CASE"},
		{`a > b`, "unsupported operator"},
		{`a and b`, "unsupported operator"},
		{`not a`, "unsupported unary operator"},
		{`'str'`, "string literals"},
		{`nullif(a, 'x')`, "string literals"},
		{`nullif(a, '--')`, "string literals"},
		{`a || b`, "unsupported operator"},
		{`@@version`, "unsupported expression"},
		{`@v`, "unsupported expression"},
		{`foo(a)`, "unsupported function"},
		{`round(a, 1, 2)`, "does not accept 3 argument"},
		{`coalesce(a)`, "does not accept 1 argument"},
		{`100`, "must reference at least one measure"},
		{`1 + 2`, "must reference at least one measure"},
		{`NULL`, "must reference at least one measure"},
		{`cast(a as double)`, "unsupported"},
		{`a as profit`, "alias is not allowed"},
		{`revenue -- cost`, "comments are not allowed"},
		{`revenue # x`, "comments are not allowed"},
		{`revenue /*- cost*/`, "comments are not allowed"},
		{`a /* inline comment */ + b`, "comments are not allowed"},
		{`*`, "wildcard"},
		{`a from t`, "SQL clauses"},
		{`distinct a`, "SQL clauses"},
		{strings.Repeat("a+", 1024) + "a", "maximum length"},
	}
	for _, c := range cases {
		t.Run(c.expr, func(t *testing.T) {
			_, err := ParseMeasureExpression(c.expr)
			require.Error(t, err)
			require.Contains(t, err.Error(), c.errContains)
		})
	}
}

func TestParseMeasureExpressionDepthLimit(t *testing.T) {
	expr := strings.Repeat("(", 100) + "a" + strings.Repeat(")", 100)
	_, err := ParseMeasureExpression(expr)
	require.Error(t, err)
	require.Contains(t, err.Error(), "deeply nested")
}

func TestMeasureExpressionRender(t *testing.T) {
	opts := MeasureExpressionRenderOptions{
		EscapeRef:  func(name string) string { return fmt.Sprintf("<%s>", name) },
		SafeDivide: func(num, den string) string { return fmt.Sprintf("safe_div(%s, %s)", num, den) },
	}

	cases := []struct {
		expr string
		want string
	}{
		{`revenue - cost`, `(<revenue> - <cost>)`},
		{`a + b * c`, `(<a> + (<b> * <c>))`},
		{`(a + b) * c`, `((<a> + <b>) * <c>)`},
		{`a / b`, `safe_div(<a>, <b>)`},
		{`round(a / b, 2)`, `round(safe_div(<a>, <b>), 2)`},
		{`-a`, `(-<a>)`},
		{`a * -2`, `(<a> * (-2))`},
		{`coalesce(a, NULL, 0.5)`, `coalesce(<a>, NULL, 0.5)`},
		{`a % b`, `(<a> % <b>)`},
		{`"weird""name" + 1`, `(<weird"name> + 1)`},
	}
	for _, c := range cases {
		t.Run(c.expr, func(t *testing.T) {
			e, err := ParseMeasureExpression(c.expr)
			require.NoError(t, err)
			require.Equal(t, c.want, e.Render(opts))
		})
	}
}

func TestMeasureExpressionRenderDialectFuncOverrides(t *testing.T) {
	e, err := ParseMeasureExpression(`round(a, 2) + round(b)`)
	require.NoError(t, err)

	// Pinot's "round" buckets to a multiple; decimal rounding is "roundDecimal".
	require.Equal(t, `(roundDecimal(a, 2) + roundDecimal(b))`, e.Render(MeasureExpressionRenderOptions{Dialect: drivers.DialectNamePinot}))
	require.Equal(t, `(round(a, 2) + round(b))`, e.Render(MeasureExpressionRenderOptions{Dialect: drivers.DialectNameDuckDB}))
}

func TestMeasureExpressionRenderDefaults(t *testing.T) {
	e, err := ParseMeasureExpression(`a / b`)
	require.NoError(t, err)
	require.Equal(t, `(a)/(b)`, e.Render(MeasureExpressionRenderOptions{}))
}
