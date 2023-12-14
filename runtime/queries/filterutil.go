package queries

import (
	"github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"google.golang.org/protobuf/types/known/structpb"
)

func FilterColumn(col string) *runtimev1.Expression {
	return &runtimev1.Expression{
		Expression: &runtimev1.Expression_Ident{
			Ident: col,
		},
	}
}

func FilterValue(val *structpb.Value) *runtimev1.Expression {
	return &runtimev1.Expression{
		Expression: &runtimev1.Expression_Val{
			Val: val,
		},
	}
}

func FilterInClause(col *runtimev1.Expression, values []*runtimev1.Expression) *runtimev1.Expression {
	return &runtimev1.Expression{
		Expression: &runtimev1.Expression_Cond{
			Cond: &runtimev1.Condition{
				Op:    runtimev1.Operation_OPERATION_IN,
				Exprs: append([]*runtimev1.Expression{col}, values...),
			},
		},
	}
}

func FilterNotInClause(col *runtimev1.Expression, values []*runtimev1.Expression) *runtimev1.Expression {
	return &runtimev1.Expression{
		Expression: &runtimev1.Expression_Cond{
			Cond: &runtimev1.Condition{
				Op:    runtimev1.Operation_OPERATION_NIN,
				Exprs: append([]*runtimev1.Expression{col}, values...),
			},
		},
	}
}

func FilterLikeClause(col, val *runtimev1.Expression) *runtimev1.Expression {
	return &runtimev1.Expression{
		Expression: &runtimev1.Expression_Cond{
			Cond: &runtimev1.Condition{
				Op:    runtimev1.Operation_OPERATION_LIKE,
				Exprs: []*runtimev1.Expression{col, val},
			},
		},
	}
}

func FilterNotLikeClause(col, val *runtimev1.Expression) *runtimev1.Expression {
	return &runtimev1.Expression{
		Expression: &runtimev1.Expression_Cond{
			Cond: &runtimev1.Condition{
				Op:    runtimev1.Operation_OPERATION_NLIKE,
				Exprs: []*runtimev1.Expression{col, val},
			},
		},
	}
}

func FilterAndClause(values []*runtimev1.Expression) *runtimev1.Expression {
	return &runtimev1.Expression{
		Expression: &runtimev1.Expression_Cond{
			Cond: &runtimev1.Condition{
				Op:    runtimev1.Operation_OPERATION_AND,
				Exprs: values,
			},
		},
	}
}

func FilterOrClause(values []*runtimev1.Expression) *runtimev1.Expression {
	return &runtimev1.Expression{
		Expression: &runtimev1.Expression_Cond{
			Cond: &runtimev1.Condition{
				Op:    runtimev1.Operation_OPERATION_OR,
				Exprs: values,
			},
		},
	}
}
