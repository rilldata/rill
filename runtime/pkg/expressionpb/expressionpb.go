package expressionpb

import (
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"google.golang.org/protobuf/types/known/structpb"
)

/**
 * Helper utils to simplify using the expression structures.
 */

func Identifier(col string) *runtimev1.Expression {
	return &runtimev1.Expression{
		Expression: &runtimev1.Expression_Ident{
			Ident: col,
		},
	}
}

func Value(val *structpb.Value) *runtimev1.Expression {
	return &runtimev1.Expression{
		Expression: &runtimev1.Expression_Val{
			Val: val,
		},
	}
}

func In(col *runtimev1.Expression, values []*runtimev1.Expression) *runtimev1.Expression {
	return &runtimev1.Expression{
		Expression: &runtimev1.Expression_Cond{
			Cond: &runtimev1.Condition{
				Op:    runtimev1.Operation_OPERATION_IN,
				Exprs: append([]*runtimev1.Expression{col}, values...),
			},
		},
	}
}

func NotIn(col *runtimev1.Expression, values []*runtimev1.Expression) *runtimev1.Expression {
	return &runtimev1.Expression{
		Expression: &runtimev1.Expression_Cond{
			Cond: &runtimev1.Condition{
				Op:    runtimev1.Operation_OPERATION_NIN,
				Exprs: append([]*runtimev1.Expression{col}, values...),
			},
		},
	}
}

func Like(col, val *runtimev1.Expression) *runtimev1.Expression {
	return &runtimev1.Expression{
		Expression: &runtimev1.Expression_Cond{
			Cond: &runtimev1.Condition{
				Op:    runtimev1.Operation_OPERATION_LIKE,
				Exprs: []*runtimev1.Expression{col, val},
			},
		},
	}
}

func NotLike(col, val *runtimev1.Expression) *runtimev1.Expression {
	return &runtimev1.Expression{
		Expression: &runtimev1.Expression_Cond{
			Cond: &runtimev1.Condition{
				Op:    runtimev1.Operation_OPERATION_NLIKE,
				Exprs: []*runtimev1.Expression{col, val},
			},
		},
	}
}

func And(values []*runtimev1.Expression) *runtimev1.Expression {
	return &runtimev1.Expression{
		Expression: &runtimev1.Expression_Cond{
			Cond: &runtimev1.Condition{
				Op:    runtimev1.Operation_OPERATION_AND,
				Exprs: values,
			},
		},
	}
}

func Or(values []*runtimev1.Expression) *runtimev1.Expression {
	return &runtimev1.Expression{
		Expression: &runtimev1.Expression_Cond{
			Cond: &runtimev1.Condition{
				Op:    runtimev1.Operation_OPERATION_OR,
				Exprs: values,
			},
		},
	}
}
