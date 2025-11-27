package server

import (
	"context"
	"fmt"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/metricsview"
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

	templateData := parser.TemplateData{
		User:      claims.UserAttributes,
		Variables: inst.ResolveVariables(false),
		State:     make(map[string]any),
		Resolve: func(ref parser.ResourceName) (string, error) {
			return ref.Name, nil
		},
	}

	templateData.ExtraFuncs = map[string]any{
		"metrics_sql": func(sql string) (string, error) {
			// Resolve any templates in the SQL string
			resolvedSQL, err := parser.ResolveTemplate(sql, templateData, false)
			if err != nil {
				return "", fmt.Errorf("failed to resolve SQL template: %w", err)
			}

			value, metricsViewName, fieldName, err := s.executeMetricsSQL(ctx, req.InstanceId, claims, resolvedSQL, req.AdditionalWhereByMetricsView, req.AdditionalTimeRange)
			if err != nil {
				return "", err
			}

			// Return format token or raw value based on request
			if req.UseFormatTokens {
				return fmt.Sprintf(`__RILL__FORMAT__(%q, %q, %v)`, metricsViewName, fieldName, value), nil
			}

			return fmt.Sprintf("%v", value), nil
		},
	}

	// Resolve the template
	body, err := parser.ResolveTemplate(req.Body, templateData, false)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to resolve template: %s", err.Error())
	}

	return &runtimev1.ResolveTemplatedStringResponse{
		Body: body,
	}, nil
}

// executeMetricsSQL executes a metrics SQL query and returns a single scalar value
func (s *Server) executeMetricsSQL(ctx context.Context, instanceID string, claims *runtime.SecurityClaims, sql string, additionalWhereByMetricsView map[string]*runtimev1.Expression, additionalTimeRange *runtimev1.Expression) (value any, metricsViewName, fieldName string, err error) {
	ctrl, err := s.runtime.Controller(ctx, instanceID)
	if err != nil {
		return nil, "", "", err
	}

	compiler := metricssql.New(&metricssql.CompilerOptions{
		GetMetricsView: func(ctx context.Context, name string) (*runtimev1.Resource, error) {
			mv, err := ctrl.Get(ctx, &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: name}, false)
			if err != nil {
				return nil, err
			}
			sec, err := s.runtime.ResolveSecurity(ctx, instanceID, claims, mv)
			if err != nil {
				return nil, err
			}
			if !sec.CanAccess() {
				return nil, runtime.ErrForbidden
			}
			return mv, nil
		},
	})

	query, err := compiler.Parse(ctx, sql)
	if err != nil {
		return nil, "", "", err
	}

	metricsViewName = query.MetricsView

	// Resolve using the metrics_sql resolver
	opts := &runtime.ResolveOptions{
		InstanceID: instanceID,
		Resolver:   "metrics_sql",
		ResolverProperties: map[string]any{
			"sql": sql,
		},
		Claims: claims,
	}

	var combinedWhere *metricsview.Expression
	if additionalWhere, ok := additionalWhereByMetricsView[metricsViewName]; ok && additionalWhere != nil {
		combinedWhere = metricsview.NewExpressionFromProto(additionalWhere)
	}

	if additionalTimeRange != nil {
		timeExpr := metricsview.NewExpressionFromProto(additionalTimeRange)
		if combinedWhere != nil {
			combinedWhere = &metricsview.Expression{
				Condition: &metricsview.Condition{
					Operator:    metricsview.OperatorAnd,
					Expressions: []*metricsview.Expression{combinedWhere, timeExpr},
				},
			}
		} else {
			combinedWhere = timeExpr
		}
	}
	if combinedWhere != nil {
		opts.ResolverProperties["additional_where"] = combinedWhere
	}

	resolveRes, err := s.runtime.Resolve(ctx, opts)
	if err != nil {
		return nil, "", "", err
	}
	defer resolveRes.Close()

	row, err := resolveRes.Next()
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to get result: %w", err)
	}

	if len(row) != 1 {
		return nil, "", "", fmt.Errorf("metrics_sql in templating only allows one result field, got %d", len(row))
	}

	var val any
	for _, v := range row {
		val = v
		break
	}

	// Check no more rows
	_, err = resolveRes.Next()
	if err == nil {
		return nil, "", "", fmt.Errorf("metrics_sql in templating must return one row, but the query returned multiple")
	}

	// Get field name from schema
	schema := resolveRes.Schema()
	if len(schema.Fields) != 1 {
		return nil, "", "", fmt.Errorf("expected one field, got %d", len(schema.Fields))
	}

	fieldName = schema.Fields[0].Name

	return val, metricsViewName, fieldName, nil
}
