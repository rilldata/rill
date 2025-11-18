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
	"github.com/rilldata/rill/runtime/parser"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/rilldata/rill/runtime/resolvers"
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

	templateData := parser.TemplateData{
		User:      claims.UserAttributes,
		Variables: inst.ResolveVariables(false),
		State:     make(map[string]any),
	}

	funcMap := sprig.TxtFuncMap()
	delete(funcMap, "env")
	delete(funcMap, "expandenv")

	// Register the metrics_sql custom function
	funcMap["metrics_sql"] = func(sql string) (string, error) {
		value, metricsViewName, fieldName, err := s.executeMetricsSQL(ctx, req.InstanceId, claims, sql, req.AdditionalWhereByMetricsView, req.AdditionalTimeRange)
		if err != nil {
			return "", err
		}

		// Return format token or raw value based on request
		if req.UseFormatTokens {
			return fmt.Sprintf(`__RILL__FORMAT__(%q, %q, %v)`, metricsViewName, fieldName, value), nil
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

// executeMetricsSQL executes a metrics SQL query and returns a single scalar value
func (s *Server) executeMetricsSQL(ctx context.Context, instanceID string, claims *runtime.SecurityClaims, sql string, additionalWhereByMetricsView map[string]*runtimev1.Expression, additionalTimeRange *runtimev1.TimeRange) (value any, metricsViewName, fieldName string, err error) {
	compiler, err := resolvers.CreateMetricsSQLCompiler(ctx, s.runtime, instanceID, claims, 0)
	if err != nil {
		return nil, "", "", err
	}

	// Parse the metrics SQL query
	query, err := compiler.Parse(ctx, sql)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to parse metrics SQL: %w", err)
	}

	// Apply additional filters if provided for this metrics view
	if additionalWhere, ok := additionalWhereByMetricsView[query.MetricsView]; ok {
		additionalWhereMV := convertProtoExpression(additionalWhere)
		query.Where = resolvers.ApplyAdditionalWhere(query.Where, additionalWhereMV)
	}

	// Apply additional time range if provided
	if additionalTimeRange != nil {
		additionalTimeRangeMV := convertProtoTimeRange(additionalTimeRange)
		query.TimeRange = resolvers.ApplyAdditionalTimeRange(query.TimeRange, additionalTimeRangeMV)
	}

	// Get the metrics view resource and resolve security
	ctrl, err := s.runtime.Controller(ctx, instanceID)
	if err != nil {
		return nil, "", "", err
	}

	mv, err := ctrl.Get(ctx, &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: query.MetricsView}, false)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to get metrics view %q: %w", query.MetricsView, err)
	}

	sec, err := s.runtime.ResolveSecurity(ctx, ctrl.InstanceID, claims, mv)
	if err != nil {
		return nil, "", "", err
	}

	// Create executor and execute the query
	exec, err := executor.New(ctx, s.runtime, instanceID, mv.GetMetricsView().State.ValidSpec, false, sec, 0)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to create executor: %w", err)
	}
	defer exec.Close()

	// Check for cancellation before executing query
	if ctx.Err() != nil {
		return nil, "", "", status.Error(codes.Canceled, "query was cancelled")
	}

	res, err := exec.Query(ctx, query, nil)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return nil, "", "", status.Error(codes.Canceled, "query was cancelled")
		}
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, "", "", status.Error(codes.DeadlineExceeded, "query timed out")
		}
		return nil, "", "", fmt.Errorf("failed to execute query: %w", err)
	}
	defer res.Close()

	// Read and validate result - must be exactly one row with one column
	if !res.Next() {
		if errors.Is(res.Err(), context.Canceled) {
			return nil, "", "", status.Error(codes.Canceled, "query was cancelled")
		}
		return nil, "", "", errors.New("metrics_sql query must return exactly one row with one value")
	}

	row := make(map[string]any)
	if err := res.MapScan(row); err != nil {
		return nil, "", "", fmt.Errorf("failed to scan query result: %w", err)
	}

	if len(row) != 1 {
		return nil, "", "", fmt.Errorf("metrics_sql query must return exactly one value, got %d values", len(row))
	}

	if res.Next() {
		return nil, "", "", errors.New("metrics_sql query must return exactly one row, got multiple rows")
	}

	if err := res.Err(); err != nil {
		if errors.Is(err, context.Canceled) {
			return nil, "", "", status.Error(codes.Canceled, "query was cancelled")
		}
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, "", "", status.Error(codes.DeadlineExceeded, "query timed out")
		}
		return nil, "", "", fmt.Errorf("query execution error: %w", err)
	}

	// Extract the single value
	var val any
	var field string
	for k, v := range row {
		field = k
		val = v
		break
	}

	return val, query.MetricsView, field, nil
}

// convertProtoExpression converts runtimev1.Expression to metricsview.Expression.
// These converter functions are specific to the ResolveTemplatedString RPC and handle
// the translation from proto definitions to internal metricsview types.
func convertProtoExpression(expr *runtimev1.Expression) *metricsview.Expression {
	if expr == nil {
		return nil
	}

	switch e := expr.Expression.(type) {
	case *runtimev1.Expression_Ident:
		return &metricsview.Expression{Name: e.Ident}
	case *runtimev1.Expression_Val:
		return &metricsview.Expression{Value: e.Val.AsInterface()}
	case *runtimev1.Expression_Cond:
		cond := &metricsview.Condition{Operator: convertProtoOperator(e.Cond.Op)}
		for _, exp := range e.Cond.Exprs {
			cond.Expressions = append(cond.Expressions, convertProtoExpression(exp))
		}
		return &metricsview.Expression{Condition: cond}
	case *runtimev1.Expression_Subquery:
		return nil // Subqueries not yet supported
	default:
		return nil
	}
}

// convertProtoOperator converts runtimev1.Operation to metricsview.Operator
func convertProtoOperator(op runtimev1.Operation) metricsview.Operator {
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

// convertProtoTimeRange converts a proto TimeRange to a metricsview TimeRange
func convertProtoTimeRange(tr *runtimev1.TimeRange) *metricsview.TimeRange {
	if tr == nil {
		return nil
	}

	res := &metricsview.TimeRange{
		Expression:    tr.Expression,
		IsoDuration:   tr.IsoDuration,
		IsoOffset:     tr.IsoOffset,
		RoundToGrain:  metricsview.TimeGrainFromProto(tr.RoundToGrain),
		TimeDimension: tr.TimeDimension,
	}
	if tr.Start != nil {
		res.Start = tr.Start.AsTime()
	}
	if tr.End != nil {
		res.End = tr.End.AsTime()
	}
	return res
}
