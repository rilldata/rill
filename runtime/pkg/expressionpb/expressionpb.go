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

func String(str string) *runtimev1.Expression {
	return &runtimev1.Expression{
		Expression: &runtimev1.Expression_Val{
			Val: structpb.NewStringValue(str),
		},
	}
}

func Number(n float64) *runtimev1.Expression {
	return &runtimev1.Expression{
		Expression: &runtimev1.Expression_Val{
			Val: structpb.NewNumberValue(n),
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

func Eq(col, value string) *runtimev1.Expression {
	return &runtimev1.Expression{
		Expression: &runtimev1.Expression_Cond{
			Cond: &runtimev1.Condition{
				Op:    runtimev1.Operation_OPERATION_EQ,
				Exprs: []*runtimev1.Expression{Identifier(col), String(value)},
			},
		},
	}
}

func Gt(col string, n int) *runtimev1.Expression {
	return &runtimev1.Expression{
		Expression: &runtimev1.Expression_Cond{
			Cond: &runtimev1.Condition{
				Op:    runtimev1.Operation_OPERATION_GT,
				Exprs: []*runtimev1.Expression{Identifier(col), structpb.NewNumberValue(n)},
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

func AndAll(values ...*runtimev1.Expression) *runtimev1.Expression {
	return &runtimev1.Expression{
		Expression: &runtimev1.Expression_Cond{
			Cond: &runtimev1.Condition{
				Op:    runtimev1.Operation_OPERATION_AND,
				Exprs: values,
			},
		},
	}
}

func OrAll(values ...*runtimev1.Expression) *runtimev1.Expression {
	return &runtimev1.Expression{
		Expression: &runtimev1.Expression_Cond{
			Cond: &runtimev1.Condition{
				Op:    runtimev1.Operation_OPERATION_OR,
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
