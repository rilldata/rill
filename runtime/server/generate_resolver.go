package server

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/rilldata/rill/runtime/queries"
	"github.com/rilldata/rill/runtime/server/auth"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
)

func (s *Server) GenerateResolver(ctx context.Context, req *runtimev1.GenerateResolverRequest) (*runtimev1.GenerateResolverResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
		attribute.String("args.connector", req.Connector),
		attribute.String("args.table", req.Table),
		attribute.String("args.metrics_view", req.MetricsView),
		// attribute.String("args.prompt", req.Prompt), // Adding this might be a privacy issue
	)
	s.addInstanceRequestAttributes(ctx, req.InstanceId)

	// Must have edit permissions on the repo
	if !auth.GetClaims(ctx).CanInstance(req.InstanceId, auth.EditRepo) {
		return nil, ErrForbidden
	}

	connector := req.Connector
	if connector == "" {
		// Get instance
		inst, err := s.runtime.Instance(ctx, req.InstanceId)
		if err != nil {
			return nil, err
		}

		connector = inst.ResolveOLAPConnector()
	}

	// Connect to connector and check it's an OLAP db
	handle, release, err := s.runtime.AcquireHandle(ctx, req.InstanceId, connector)
	if err != nil {
		return nil, err
	}
	defer release()
	olap, ok := handle.AsOLAP(req.InstanceId)
	if !ok {
		return nil, status.Error(codes.InvalidArgument, "connector is not an OLAP connector")
	}

	dialect := olap.Dialect().String()

	ctrl, err := s.runtime.Controller(ctx, req.InstanceId)
	if err != nil {
		return nil, err
	}

	if req.Table != "" && req.Connector != "" {
		// Get table info
		tbl, err := olap.InformationSchema().Lookup(ctx, req.Table)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "table not found")
		}

		start := time.Now()

		sql, err := s.generateResolverForTable(ctx, req.InstanceId, req.Prompt, tbl.Name, dialect, tbl.Schema)
		attrs := []attribute.KeyValue{attribute.Int("table_column_count", len(tbl.Schema.Fields))}
		attrs = append(attrs, attribute.Int64("elapsed_ms", time.Since(start).Milliseconds()))
		if err != nil {
			attrs = append(attrs, attribute.Bool("succeeded", false))
			s.activity.Record(ctx, activity.EventTypeLog, "ai_generated_resolver", attrs...)
			return nil, err
		}

		attrs = append(attrs, attribute.Bool("succeeded", true))
		s.activity.Record(ctx, activity.EventTypeLog, "ai_generated_resolver", attrs...)

		resolverPropertiesPB, err := structpb.NewStruct(map[string]interface{}{
			"sql": sql,
		})
		if err != nil {
			return nil, err
		}

		return &runtimev1.GenerateResolverResponse{
			Resolver:           "SQL",
			ResolverProperties: resolverPropertiesPB,
		}, nil
	} else if req.MetricsView != "" {
		res, err := ctrl.Get(ctx, &runtimev1.ResourceName{Name: req.MetricsView, Kind: runtime.ResourceKindMetricsView}, true)
		if err != nil {
			return nil, err
		}

		mv := res.GetMetricsView()
		mvs := mv.GetState().GetValidSpec()
		if mvs == nil {
			return nil, fmt.Errorf("metrics view %q not found", req.MetricsView)
		}

		q := &queries.MetricsViewSchema{
			MetricsViewName: req.MetricsView,
		}
		err = s.runtime.Query(ctx, req.InstanceId, q, 0)
		if err != nil {
			return nil, err
		}

		sql, err := s.generateResolverForMetricsView(ctx, req.InstanceId, req.Prompt, req.MetricsView, dialect, q.Result.Schema)
		if err != nil {
			return nil, err
		}

		resolverPropertiesPB, err := structpb.NewStruct(map[string]interface{}{
			"sql": sql,
		})
		if err != nil {
			return nil, err
		}

		return &runtimev1.GenerateResolverResponse{
			Resolver:           "MetricsSQL",
			ResolverProperties: resolverPropertiesPB,
		}, nil
	}

	return nil, errors.New("one of table or metrics_view should be provided")
}

func (s *Server) generateResolverForTable(ctx context.Context, instanceID, userPrompt, tblName, dialect string, schema *runtimev1.StructType) (string, error) {
	// Build messages
	msgs := []*drivers.CompletionMessage{
		{Role: "system", Data: resolverForTableSystemPrompt()},
		{Role: "user", Data: resolverUserPrompt(userPrompt, tblName, dialect, schema)},
	}

	// Connect to the AI service configured for the instance
	ai, release, err := s.runtime.AI(ctx, instanceID)
	if err != nil {
		return "", err
	}
	defer release()

	// Apply timeout
	ctx, cancel := context.WithTimeout(ctx, aiGenerateTimeout)
	defer cancel()

	// Call AI service to infer a metrics view YAML
	res, err := ai.Complete(ctx, msgs)
	if err != nil {
		return "", err
	}

	// The AI may produce Markdown output. Remove the code tags around the SQL.
	res.Data = strings.TrimPrefix(res.Data, "```sql")
	res.Data = strings.TrimPrefix(res.Data, "```")
	res.Data = strings.TrimSuffix(res.Data, "```")

	return res.Data, nil
}

func resolverForTableSystemPrompt() string {
	return `
You are an agent whose only task is to suggest an SQL query to get data based on a table schema.
Your output should consist of valid SQL in the format below:

<SQL query to get the data in the requested SQL dialect>
`
}

// generateResolverForMetricsView uses AI to generate a MetricsSQL resolver
func (s *Server) generateResolverForMetricsView(ctx context.Context, instanceID, userPrompt, metricsView, dialect string, schema *runtimev1.StructType) (string, error) {
	// Build messages
	msgs := []*drivers.CompletionMessage{
		{Role: "system", Data: resolverForMetricsViewSystemPrompt()},
		{Role: "user", Data: resolverUserPrompt(userPrompt, metricsView, dialect, schema)},
	}

	// Connect to the AI service configured for the instance
	ai, release, err := s.runtime.AI(ctx, instanceID)
	if err != nil {
		return "", err
	}
	defer release()

	// Apply timeout
	ctx, cancel := context.WithTimeout(ctx, aiGenerateTimeout)
	defer cancel()

	// Call AI service to infer a metrics view YAML
	res, err := ai.Complete(ctx, msgs)
	if err != nil {
		return "", err
	}

	// The AI may produce Markdown output. Remove the code tags around the SQL.
	res.Data = strings.TrimPrefix(res.Data, "```sql")
	res.Data = strings.TrimPrefix(res.Data, "```")
	res.Data = strings.TrimSuffix(res.Data, "```")

	return res.Data, nil
}

func resolverForMetricsViewSystemPrompt() string {
	return `
You are an agent whose only task is to suggest an SQL query to get data based on a table schema.
Wrap aggregations with "AGGREGATE", EG: AGGREGATE(impressions).
Do not use any complex aggregations and do not use WHERE or FILTER.

Your output should consist of valid SQL in the format below:

<SQL query to get the data in the requested SQL dialect>
`
}

func resolverUserPrompt(userPrompt, tblName, dialect string, schema *runtimev1.StructType) string {
	prompt := fmt.Sprintf(`
Prompt provided by the user: %s:

Based on the table named %q using the %q SQL dialect with schema:
`, userPrompt, tblName, dialect)
	for _, field := range schema.Fields {
		prompt += fmt.Sprintf("- column=%s, type=%s\n", field.Name, field.Type.Code.String())
	}
	return prompt
}
