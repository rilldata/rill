package metricsview

import (
	"testing"
)

func TestExpressionToSQL(t *testing.T) {
	tests := []struct {
		name    string
		e       *Expression
		want    string
		wantErr bool
	}{
		{
			name: "empty expression",
			e:    nil,
			want: "",
		},
		{
			name: "name expression",
			e:    &Expression{Name: "foo"},
			want: "foo",
		},
		{
			name: "value expression",
			e:    &Expression{Value: 42},
			want: "42",
		},
		{
			name: "subquery expression",
			e: &Expression{Subquery: &Subquery{
				Dimension: Dimension{Name: "dim"},
				Measures:  []Measure{{Name: "count"}},
				Having: &Expression{Condition: &Condition{
					Operator: OperatorEq,
					Expressions: []*Expression{
						{Name: "count"},
						{Value: 10},
					},
				}},
			}},
			want: "(SELECT dim FROM metrics_view HAVING count = 10)",
		},
		{
			name: "condition expression",
			e: &Expression{
				Condition: &Condition{
					Operator: OperatorEq,
					Expressions: []*Expression{
						{Name: "foo"},
						{Value: 42},
					},
				},
			},
			want: "foo = 42",
		},
		{
			name: "and expression",
			e: &Expression{
				Condition: &Condition{
					Operator: OperatorAnd,
					Expressions: []*Expression{
						{Name: "foo"},
						{Name: "bar"},
					},
				},
			},
			want: "foo AND bar",
		},
		{
			name: "and or expression",
			e: &Expression{
				Condition: &Condition{
					Operator: OperatorAnd,
					Expressions: []*Expression{
						{Name: "foo"},
						{
							Condition: &Condition{
								Operator: OperatorOr,
								Expressions: []*Expression{
									{Name: "bar"},
									{Name: "baz"},
								},
							},
						},
					},
				},
			},
			want: "foo AND (bar OR baz)",
		},
		{
			name: "in expression",
			e: &Expression{
				Condition: &Condition{
					Operator: OperatorIn,
					Expressions: []*Expression{
						{Name: "foo"},
						{Value: []int{1, 2, 3}},
					},
				},
			},
			want: "foo IN [1,2,3]",
		},
		{
			name: "is null",
			e: &Expression{
				Condition: &Condition{
					Operator: OperatorEq,
					Expressions: []*Expression{
						{Name: "foo"},
						{},
					},
				},
			},
			want: "foo IS NULL",
		},
		{
			name: "or is null expression",
			e: &Expression{
				Condition: &Condition{
					Operator: OperatorOr,
					Expressions: []*Expression{
						{Name: "foo"},
						{
							Condition: &Condition{
								Operator: OperatorEq,
								Expressions: []*Expression{
									{Name: "bar"},
									{Value: nil},
								},
							},
						},
						{
							Condition: &Condition{
								Operator: OperatorEq,
								Expressions: []*Expression{
									{Name: "baz"},
									{Value: 42},
								},
							},
						},
					},
				},
			},
			want: "foo OR (bar IS NULL) OR (baz = 42)",
		},
		{
			name: "or is not null expression",
			e: &Expression{
				Condition: &Condition{
					Operator: OperatorOr,
					Expressions: []*Expression{
						{Name: "foo"},
						{
							Condition: &Condition{
								Operator: OperatorNeq,
								Expressions: []*Expression{
									{Name: "bar"},
									{Value: nil},
								},
							},
						},
						{
							Condition: &Condition{
								Operator: OperatorEq,
								Expressions: []*Expression{
									{Name: "baz"},
									{Value: 42},
								},
							},
						},
					},
				},
			},
			want: "foo OR (bar IS NOT NULL) OR (baz = 42)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExpressionToSQL(tt.e)
			if err != nil {
				t.Errorf("ExpressionToSQL: got error: %v, want nil", err)
				return
			}
			if got != tt.want {
				t.Errorf("ExpressionToSQL: got %q, want %q", got, tt.want)
			}
		})
	}
}
