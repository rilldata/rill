package metricsview

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateExpressionNarrows(t *testing.T) {
	// cond builds a comparison of a dimension to one or more value expressions, such as cond(OperatorIn, "country", "DK", "US").
	cond := func(op Operator, dimension string, values ...any) *Expression {
		exprs := []*Expression{{Name: dimension}}
		for _, v := range values {
			exprs = append(exprs, &Expression{Value: v})
		}
		return &Expression{Condition: &Condition{Operator: op, Expressions: exprs}}
	}
	and := func(exprs ...*Expression) *Expression {
		return &Expression{Condition: &Condition{Operator: OperatorAnd, Expressions: exprs}}
	}
	or := func(exprs ...*Expression) *Expression {
		return &Expression{Condition: &Condition{Operator: OperatorOr, Expressions: exprs}}
	}
	subquery := &Expression{Subquery: &Subquery{
		Dimension: Dimension{Name: "country"},
		Measures:  []Measure{{Name: "revenue"}},
		Having:    cond(OperatorGt, "revenue", 100),
	}}

	tests := []struct {
		name    string
		expr    *Expression
		base    *Expression
		wantErr string
	}{
		{
			name: "no base",
			expr: cond(OperatorEq, "country", "DK"),
		},
		{
			name:    "no expression",
			base:    cond(OperatorEq, "country", "DK"),
			wantErr: `no filter on "country"`,
		},
		{
			name: "same filter",
			expr: cond(OperatorIn, "country", []any{"DK"}),
			base: cond(OperatorIn, "country", []any{"DK"}),
		},
		{
			name: "in with a scalar restated as eq",
			expr: cond(OperatorEq, "country", "DK"),
			base: cond(OperatorIn, "country", "DK"),
		},
		{
			name: "in with a list restated as flattened values",
			expr: cond(OperatorIn, "country", "DK", "US"),
			base: cond(OperatorIn, "country", []any{"DK", "US"}),
		},
		{
			name: "in with flattened values restated as a list",
			expr: cond(OperatorIn, "country", []any{"DK", "US"}),
			base: cond(OperatorIn, "country", "DK", "US"),
		},
		{
			name: "in with a typed list",
			expr: cond(OperatorIn, "country", []string{"DK"}),
			base: cond(OperatorIn, "country", []string{"DK", "US"}),
		},
		{
			name: "narrowed to fewer values",
			expr: cond(OperatorEq, "country", "DK"),
			base: cond(OperatorIn, "country", []any{"DK", "US"}),
		},
		{
			name:    "widened to more values",
			expr:    cond(OperatorIn, "country", []any{"DK", "US"}),
			base:    cond(OperatorIn, "country", []any{"DK"}),
			wantErr: `country IN ('DK', 'US') is not within country IN ('DK')`,
		},
		{
			name:    "changed to another value",
			expr:    cond(OperatorEq, "country", "US"),
			base:    cond(OperatorEq, "country", "DK"),
			wantErr: `country = 'US' is not within country = 'DK'`,
		},
		{
			name: "extra filter alongside",
			expr: and(cond(OperatorEq, "country", "DK"), cond(OperatorEq, "device", "mobile")),
			base: cond(OperatorEq, "country", "DK"),
		},
		{
			name:    "unrelated filter instead",
			expr:    cond(OperatorEq, "device", "mobile"),
			base:    cond(OperatorEq, "country", "DK"),
			wantErr: `no filter on "country"`,
		},
		{
			name:    "measure filter instead",
			expr:    cond(OperatorGt, "revenue", 50),
			base:    cond(OperatorEq, "country", "DK"),
			wantErr: `no filter on "country"`,
		},
		{
			name: "nested ands on both sides",
			expr: and(and(cond(OperatorEq, "device", "mobile"), cond(OperatorEq, "country", "DK")), cond(OperatorGt, "revenue", 50)),
			base: and(cond(OperatorEq, "country", "DK"), and(cond(OperatorEq, "device", "mobile"))),
		},
		{
			name:    "one of several base filters dropped",
			expr:    cond(OperatorEq, "country", "DK"),
			base:    and(cond(OperatorEq, "country", "DK"), cond(OperatorEq, "device", "mobile")),
			wantErr: `no filter on "device"`,
		},
		{
			name:    "filter inside an or does not count",
			expr:    or(cond(OperatorEq, "country", "DK"), cond(OperatorEq, "device", "mobile")),
			base:    cond(OperatorEq, "country", "DK"),
			wantErr: `no filter on "country"`,
		},
		{
			name: "or in base is skipped",
			expr: cond(OperatorEq, "device", "mobile"),
			base: and(or(cond(OperatorEq, "country", "DK"), cond(OperatorEq, "country", "US")), cond(OperatorEq, "device", "mobile")),
		},
		{
			name: "subquery in base is skipped",
			expr: cond(OperatorEq, "device", "mobile"),
			base: and(subquery, cond(OperatorEq, "device", "mobile")),
		},
		{
			name: "exclusion restated",
			expr: cond(OperatorNin, "country", "DK"),
			base: cond(OperatorNin, "country", []any{"DK"}),
		},
		{
			name: "exclusion restated as a selection",
			expr: cond(OperatorEq, "country", "US"),
			base: cond(OperatorNin, "country", []any{"DK"}),
		},
		{
			name:    "exclusion dropped",
			expr:    cond(OperatorEq, "device", "mobile"),
			base:    cond(OperatorNin, "country", []any{"DK"}),
			wantErr: `no filter on "country"`,
		},
		{
			name: "pattern restated as a selection",
			expr: cond(OperatorEq, "country", "DK"),
			base: cond(OperatorIlike, "country", "D%"),
		},
		{
			name: "value-first comparison",
			expr: &Expression{Condition: &Condition{Operator: OperatorEq, Expressions: []*Expression{{Value: "DK"}, {Name: "country"}}}},
			base: cond(OperatorEq, "country", "DK"),
		},
		{
			name: "null value",
			expr: cond(OperatorEq, "country", nil),
			base: cond(OperatorIn, "country", []any{nil, "DK"}),
		},
		{
			name: "numbers of different types",
			expr: cond(OperatorEq, "year", 2024),
			base: cond(OperatorIn, "year", []any{float64(2024), float64(2025)}),
		},
		{
			name: "comparison of two dimensions is skipped",
			expr: cond(OperatorEq, "device", "mobile"),
			base: &Expression{Condition: &Condition{Operator: OperatorEq, Expressions: []*Expression{{Name: "country"}, {Name: "country_of_origin"}}}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateExpressionNarrows(tt.expr, tt.base)
			if tt.wantErr == "" {
				require.NoError(t, err)
				return
			}
			require.ErrorContains(t, err, tt.wantErr)
		})
	}
}
