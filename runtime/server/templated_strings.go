package server

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/metricsview"
	"github.com/rilldata/rill/runtime/metricsview/executor"
	"github.com/rilldata/rill/runtime/metricsview/metricssql"
	"github.com/rilldata/rill/runtime/parser"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/rilldata/rill/runtime/server/auth"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) ResolveTemplatedString(ctx context.Context, req *runtimev1.ResolveTemplatedStringRequest) (*runtimev1.ResolveTemplatedStringResponse, error) {
	s.addInstanceRequestAttributes(ctx, req.InstanceId)
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
		attribute.Bool("args.use_format_tokens", req.UseFormatTokens),
	)

	claims := auth.GetClaims(ctx, req.InstanceId)
	if !claims.Can(runtime.ReadAPI) {
		return nil, status.Errorf(codes.FailedPrecondition, "does not have access to query data")
	}

	inst, err := s.runtime.Instance(ctx, req.InstanceId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	ctrl, err := s.runtime.Controller(ctx, req.InstanceId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	templateData := parser.TemplateData{
		User:      claims.UserAttributes,
		Variables: inst.ResolveVariables(false),
		State:     make(map[string]any),
	}

	// Create base func map with Sprig functions (excluding env functions)
	funcMap := sprig.TxtFuncMap()
	delete(funcMap, "env")
	delete(funcMap, "expandenv")

	// Register the metrics_sql custom function
	funcMap["metrics_sql"] = func(sql string) (string, error) {
		// Create a metrics SQL compiler
		compiler := metricssql.New(&metricssql.CompilerOptions{
			GetMetricsView: func(ctx context.Context, name string) (*runtimev1.Resource, error) {
				mv, err := ctrl.Get(ctx, &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: name}, false)
				if err != nil {
					return nil, err
				}
				sec, err := s.runtime.ResolveSecurity(ctx, ctrl.InstanceID, claims, mv)
				if err != nil {
					return nil, err
				}
				if !sec.CanAccess() {
					return nil, runtime.ErrForbidden
				}
				return mv, nil
			},
			GetTimestamps: func(ctx context.Context, mv *runtimev1.Resource, timeDim string) (metricsview.TimestampsResult, error) {
				sec, err := s.runtime.ResolveSecurity(ctx, ctrl.InstanceID, claims, mv)
				if err != nil {
					return metricsview.TimestampsResult{}, err
				}
				e, err := executor.New(ctx, s.runtime, req.InstanceId, mv.GetMetricsView().State.ValidSpec, false, sec, 0)
				if err != nil {
					return metricsview.TimestampsResult{}, err
				}
				defer e.Close()
				return e.Timestamps(ctx, timeDim)
			},
		})

		// Parse the metrics SQL query
		query, err := compiler.Parse(ctx, sql)
		if err != nil {
			return "", fmt.Errorf("failed to parse metrics SQL: %w", err)
		}

		// Apply additional filters if provided
		if req.AdditionalWhere != nil {
			query.Where = applyAdditionalWhere(query.Where, req.AdditionalWhere)
		}

		// Apply additional time range if provided
		// Note: AdditionalTimeRange is an Expression that needs to be converted to a TimeRange
		if req.AdditionalTimeRange != nil {
			// For now, we don't support time range expressions in this context
			// This would require parsing the expression and extracting time range information
			// which is more complex than the basic where clause handling
			return "", errors.New("additional_time_range not yet supported")
		}

		// Get the metrics view resource
		mv, err := ctrl.Get(ctx, &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: query.MetricsView}, false)
		if err != nil {
			return "", fmt.Errorf("failed to get metrics view %q: %w", query.MetricsView, err)
		}

		// Check security
		sec, err := s.runtime.ResolveSecurity(ctx, ctrl.InstanceID, claims, mv)
		if err != nil {
			return "", err
		}
		if !sec.CanAccess() {
			return "", runtime.ErrForbidden
		}

		// Create executor
		exec, err := executor.New(ctx, s.runtime, req.InstanceId, mv.GetMetricsView().State.ValidSpec, false, sec, 0)
		if err != nil {
			return "", fmt.Errorf("failed to create executor: %w", err)
		}
		defer exec.Close()

		// Execute the query
		res, err := exec.Query(ctx, query, nil)
		if err != nil {
			return "", fmt.Errorf("failed to execute query: %w", err)
		}
		defer res.Close()

		// Read the result - must be exactly one row and one column
		if !res.Next() {
			return "", errors.New("metrics_sql query must return exactly one row with one value")
		}

		// Scan the row
		row := make(map[string]any)
		if err := res.MapScan(row); err != nil {
			return "", fmt.Errorf("failed to scan query result: %w", err)
		}

		// Check for exactly one value in the row
		if len(row) != 1 {
			return "", fmt.Errorf("metrics_sql query must return exactly one value, got %d values", len(row))
		}

		// Check for second row (should not exist)
		if res.Next() {
			return "", errors.New("metrics_sql query must return exactly one row, got multiple rows")
		}

		// Check for errors
		if err := res.Err(); err != nil {
			return "", fmt.Errorf("query execution error: %w", err)
		}

		// Extract the single value
		var value any
		var fieldName string
		for k, v := range row {
			fieldName = k
			value = v
			break
		}

		// Return format token or raw value based on request
		if req.UseFormatTokens {
			return fmt.Sprintf(`__RILL__FORMAT__(%q, %q, %v)`, query.MetricsView, fieldName, value), nil
		}

		return fmt.Sprintf("%v", value), nil
	}

	// Resolve the template
	tmpl, err := template.New("templated_string").Funcs(funcMap).Parse(req.Data)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to parse template: %s", err.Error())
	}

	var result strings.Builder
	err = tmpl.Execute(&result, templateData)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to execute template: %s", err.Error())
	}

	return &runtimev1.ResolveTemplatedStringResponse{
		ResolvedData: result.String(),
	}, nil
}

// applyAdditionalWhere combines the existing where clause with the additional where clause
func applyAdditionalWhere(current *metricsview.Expression, additional *runtimev1.Expression) *metricsview.Expression {
	if additional == nil {
		return current
	}

	// Convert runtimev1.Expression to metricsview.Expression
	additionalMV := convertExpression(additional)
	if additionalMV == nil {
		return current
	}

	if current == nil {
		return additionalMV
	}

	// Combine with AND
	return &metricsview.Expression{
		Condition: &metricsview.Condition{
			Operator: metricsview.OperatorAnd,
			Expressions: []*metricsview.Expression{
				current,
				additionalMV,
			},
		},
	}
}

// convertExpression converts runtimev1.Expression to metricsview.Expression
func convertExpression(expr *runtimev1.Expression) *metricsview.Expression {
	if expr == nil {
		return nil
	}

	switch e := expr.Expression.(type) {
	case *runtimev1.Expression_Ident:
		return &metricsview.Expression{
			Name: e.Ident,
		}
	case *runtimev1.Expression_Val:
		return &metricsview.Expression{
			Value: e.Val.AsInterface(),
		}
	case *runtimev1.Expression_Cond:
		cond := &metricsview.Condition{
			Operator: convertOperator(e.Cond.Op),
		}
		for _, exp := range e.Cond.Exprs {
			cond.Expressions = append(cond.Expressions, convertExpression(exp))
		}
		return &metricsview.Expression{
			Condition: cond,
		}
	case *runtimev1.Expression_Subquery:
		// Subqueries not yet supported
		return nil
	default:
		return nil
	}
}

// convertOperator converts runtimev1.Operation to metricsview.Operator
func convertOperator(op runtimev1.Operation) metricsview.Operator {
	switch op {
	case runtimev1.Operation_OPERATION_EQ:
		return metricsview.OperatorEq
	case runtimev1.Operation_OPERATION_NEQ:
		return metricsview.OperatorNeq
	case runtimev1.Operation_OPERATION_LT:
		return metricsview.OperatorLt
	case runtimev1.Operation_OPERATION_LTE:
		return metricsview.OperatorLte
	case runtimev1.Operation_OPERATION_GT:
		return metricsview.OperatorGt
	case runtimev1.Operation_OPERATION_GTE:
		return metricsview.OperatorGte
	case runtimev1.Operation_OPERATION_IN:
		return metricsview.OperatorIn
	case runtimev1.Operation_OPERATION_NIN:
		return metricsview.OperatorNin
	case runtimev1.Operation_OPERATION_LIKE:
		return metricsview.OperatorIlike
	case runtimev1.Operation_OPERATION_NLIKE:
		return metricsview.OperatorNilike
	case runtimev1.Operation_OPERATION_OR:
		return metricsview.OperatorOr
	case runtimev1.Operation_OPERATION_AND:
		return metricsview.OperatorAnd
	default:
		return metricsview.OperatorUnspecified
	}
}
