package metricsview

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExpressionToExport(t *testing.T) {
	tests := []struct {
		name string
		e    *Expression
		want string
	}{
		{
			name: "in expression",
			e: &Expression{Condition: &Condition{
				Operator:    OperatorIn,
				Expressions: []*Expression{{Name: "foo"}, {Value: []any{1, 2}}},
			}},
			want: "foo = 1, 2",
		},
		{
			name: "flattened in expression",
			e: &Expression{Condition: &Condition{
				Operator:    OperatorIn,
				Expressions: []*Expression{{Name: "foo"}, {Value: 1}, {Value: 2}},
			}},
			want: "foo = 1, 2",
		},
		{
			// A flattened list with no values is the left hand side alone; it is written like an empty list.
			name: "flattened in expression without values",
			e: &Expression{Condition: &Condition{
				Operator:    OperatorIn,
				Expressions: []*Expression{{Name: "foo"}},
			}},
			want: "foo = ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExpressionToExport(tt.e)
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}
