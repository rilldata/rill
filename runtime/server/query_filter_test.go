package server_test

import (
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"google.golang.org/protobuf/types/known/structpb"
)

func filterColumn(col string) *runtimev1.Expression {
	return &runtimev1.Expression{
		Expression: &runtimev1.Expression_Ident{
			Ident: col,
		},
	}
}

func filterValue(val *structpb.Value) *runtimev1.Expression {
	return &runtimev1.Expression{
		Expression: &runtimev1.Expression_Val{
			Val: val,
		},
	}
}

func filterInClause(col *runtimev1.Expression, values []*runtimev1.Expression) *runtimev1.Expression {
	return &runtimev1.Expression{
		Expression: &runtimev1.Expression_Cond{
			Cond: &runtimev1.Condition{
				Op:    runtimev1.Operation_OPERATION_IN,
				Exprs: append([]*runtimev1.Expression{col}, values...),
			},
		},
	}
}

func filterNotInClause(col *runtimev1.Expression, values []*runtimev1.Expression) *runtimev1.Expression {
	return &runtimev1.Expression{
		Expression: &runtimev1.Expression_Cond{
			Cond: &runtimev1.Condition{
				Op:    runtimev1.Operation_OPERATION_NIN,
				Exprs: append([]*runtimev1.Expression{col}, values...),
			},
		},
	}
}

func filterLikeClause(col *runtimev1.Expression, val *runtimev1.Expression) *runtimev1.Expression {
	return &runtimev1.Expression{
		Expression: &runtimev1.Expression_Cond{
			Cond: &runtimev1.Condition{
				Op:    runtimev1.Operation_OPERATION_LIKE,
				Exprs: []*runtimev1.Expression{col, val},
			},
		},
	}
}

func filterNotLikeClause(col *runtimev1.Expression, val *runtimev1.Expression) *runtimev1.Expression {
	return &runtimev1.Expression{
		Expression: &runtimev1.Expression_Cond{
			Cond: &runtimev1.Condition{
				Op:    runtimev1.Operation_OPERATION_NLIKE,
				Exprs: []*runtimev1.Expression{col, val},
			},
		},
	}
}

func filterAndClause(values []*runtimev1.Expression) *runtimev1.Expression {
	return &runtimev1.Expression{
		Expression: &runtimev1.Expression_Cond{
			Cond: &runtimev1.Condition{
				Op:    runtimev1.Operation_OPERATION_AND,
				Exprs: values,
			},
		},
	}
}

func filterOrClause(values []*runtimev1.Expression) *runtimev1.Expression {
	return &runtimev1.Expression{
		Expression: &runtimev1.Expression_Cond{
			Cond: &runtimev1.Condition{
				Op:    runtimev1.Operation_OPERATION_OR,
				Exprs: values,
			},
		},
	}
}
