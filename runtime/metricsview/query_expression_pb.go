package metricsview

import runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"

func NewExpressionFromProto(expr *runtimev1.Expression) *Expression {
	if expr == nil {
		return nil
	}

	res := &Expression{}

	switch e := expr.Expression.(type) {
	case *runtimev1.Expression_Ident:
		res.Name = e.Ident
	case *runtimev1.Expression_Val:
		res.Value = e.Val.AsInterface()
	case *runtimev1.Expression_Cond:
		var op Operator
		switch e.Cond.Op {
		case runtimev1.Operation_OPERATION_UNSPECIFIED:
			op = OperatorUnspecified
		case runtimev1.Operation_OPERATION_EQ:
			op = OperatorEq
		case runtimev1.Operation_OPERATION_NEQ:
			op = OperatorNeq
		case runtimev1.Operation_OPERATION_LT:
			op = OperatorLt
		case runtimev1.Operation_OPERATION_LTE:
			op = OperatorLte
		case runtimev1.Operation_OPERATION_GT:
			op = OperatorGt
		case runtimev1.Operation_OPERATION_GTE:
			op = OperatorGte
		case runtimev1.Operation_OPERATION_OR:
			op = OperatorOr
		case runtimev1.Operation_OPERATION_AND:
			op = OperatorAnd
		case runtimev1.Operation_OPERATION_IN:
			op = OperatorIn
		case runtimev1.Operation_OPERATION_NIN:
			op = OperatorNin
		case runtimev1.Operation_OPERATION_LIKE:
			op = OperatorIlike
		case runtimev1.Operation_OPERATION_NLIKE:
			op = OperatorNilike
		}

		exprs := make([]*Expression, 0, len(e.Cond.Exprs))
		for _, e := range e.Cond.Exprs {
			exprs = append(exprs, NewExpressionFromProto(e))
		}

		res.Condition = &Condition{
			Operator:    op,
			Expressions: exprs,
		}
	case *runtimev1.Expression_Subquery:
		measures := make([]Measure, 0, len(e.Subquery.Measures))
		for _, m := range e.Subquery.Measures {
			measures = append(measures, Measure{Name: m})
		}

		res.Subquery = &Subquery{
			Dimension: Dimension{Name: e.Subquery.Dimension},
			Measures:  measures,
			Where:     NewExpressionFromProto(e.Subquery.Where),
			Having:    NewExpressionFromProto(e.Subquery.Having),
		}
	}

	return res
}
