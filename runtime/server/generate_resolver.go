package server

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	aiv1 "github.com/rilldata/rill/proto/gen/rill/ai/v1"
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
	)
	s.addInstanceRequestAttributes(ctx, req.InstanceId)

	// Must have edit permissions on the repo
	if !auth.GetClaims(ctx, req.InstanceId).Can(runtime.EditRepo) {
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

	var resolver string
	var resolverProps map[string]interface{}

	if req.Table != "" && req.Connector != "" {
		// Get table info
		tbl, err := olap.InformationSchema().Lookup(ctx, "", "", req.Table)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "table not found")
		}

		start := time.Now()

		resolver, resolverProps, err = s.generateResolverForTable(ctx, req.InstanceId, req.Prompt, tbl.Name, dialect, tbl.Schema)
		attrs := []attribute.KeyValue{attribute.Int("table_column_count", len(tbl.Schema.Fields))}
		attrs = append(attrs, attribute.Int64("elapsed_ms", time.Since(start).Milliseconds()))
		if err != nil {
			attrs = append(attrs, attribute.Bool("succeeded", false))
			s.activity.Record(ctx, activity.EventTypeLog, "ai_generated_resolver", attrs...)
			return nil, err
		}

		attrs = append(attrs, attribute.Bool("succeeded", true))
		s.activity.Record(ctx, activity.EventTypeLog, "ai_generated_resolver", attrs...)
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

		start := time.Now()

		resolver, resolverProps, err = s.generateResolverForMetricsView(ctx, req.InstanceId, req.Prompt, req.MetricsView, dialect, q.Result.Schema)
		attrs := []attribute.KeyValue{attribute.Int("table_column_count", len(q.Result.Schema.Fields))}
		attrs = append(attrs, attribute.Int64("elapsed_ms", time.Since(start).Milliseconds()))
		if err != nil {
			attrs = append(attrs, attribute.Bool("succeeded", false))
			s.activity.Record(ctx, activity.EventTypeLog, "ai_generated_resolver", attrs...)
			return nil, err
		}

		attrs = append(attrs, attribute.Bool("succeeded", true))
		s.activity.Record(ctx, activity.EventTypeLog, "ai_generated_resolver", attrs...)
	} else {
		return nil, errors.New("one of table or metrics_view should be provided")
	}

	resolverPropsPB, err := structpb.NewStruct(resolverProps)
	if err != nil {
		return nil, err
	}

	return &runtimev1.GenerateResolverResponse{
		Resolver:           resolver,
		ResolverProperties: resolverPropsPB,
	}, nil
}

var semiColonRegex = regexp.MustCompile(`(?m);\s*$`)

func (s *Server) generateResolverForTable(ctx context.Context, instanceID, userPrompt, tblName, dialect string, schema *runtimev1.StructType) (string, map[string]interface{}, error) {
	// Build messages
	systemPrompt := resolverForTableSystemPrompt()
	fullUserPrompt := resolverUserPrompt(userPrompt, tblName, dialect, schema)

	msgs := []*aiv1.CompletionMessage{
		{
			Role: "system",
			Content: []*aiv1.ContentBlock{
				{
					BlockType: &aiv1.ContentBlock_Text{
						Text: systemPrompt,
					},
				},
			},
		},
		{
			Role: "user",
			Content: []*aiv1.ContentBlock{
				{
					BlockType: &aiv1.ContentBlock_Text{
						Text: fullUserPrompt,
					},
				},
			},
		},
	}

	// Connect to the AI service configured for the instance
	ai, release, err := s.runtime.AI(ctx, instanceID)
	if err != nil {
		return "", nil, err
	}
	defer release()

	// Apply timeout
	ctx, cancel := context.WithTimeout(ctx, aiGenerateTimeout)
	defer cancel()

	// Call AI service to infer a metrics view YAML
	res, err := ai.Complete(ctx, &drivers.CompleteOptions{
		Messages: msgs,
	})
	if err != nil {
		return "", nil, err
	}

	// Extract text from content blocks
	var responseText string
	for _, block := range res.Message.Content {
		switch blockType := block.GetBlockType().(type) {
		case *aiv1.ContentBlock_Text:
			if text := blockType.Text; text != "" {
				responseText += text
			}
		default:
			// For resolver generation, we only expect text responses
			return "", nil, fmt.Errorf("unexpected content block type in AI response: %T", blockType)
		}
	}

	// The AI may produce Markdown output. Remove the code tags around the SQL.
	responseText = strings.TrimPrefix(responseText, "```sql")
	responseText = strings.TrimPrefix(responseText, "```")
	responseText = strings.TrimSuffix(responseText, "```")
	// Remove the trailing semicolon
	responseText = semiColonRegex.ReplaceAllString(responseText, "")

	return "sql", map[string]interface{}{
		"sql": responseText,
	}, nil
}

func resolverForTableSystemPrompt() string {
	return `
You are an agent whose only task is to suggest an SQL query to get data based on a table schema.
Your output should consist of valid SQL in the format below:

<SQL query to get the data in the requested SQL dialect>
`
}

var aggregateCorrections = regexp.MustCompile(`(?i)AGGREGATE(.*?)\s*AS\s*`)

// generateResolverForMetricsView uses AI to generate a MetricsSQL resolver
func (s *Server) generateResolverForMetricsView(ctx context.Context, instanceID, userPrompt, metricsView, dialect string, schema *runtimev1.StructType) (string, map[string]interface{}, error) {
	// Build messages
	systemPrompt := resolverForMetricsViewSystemPrompt()
	fullUserPrompt := resolverUserPrompt(userPrompt, metricsView, dialect, schema)

	msgs := []*aiv1.CompletionMessage{
		{
			Role: "system",
			Content: []*aiv1.ContentBlock{
				{
					BlockType: &aiv1.ContentBlock_Text{
						Text: systemPrompt,
					},
				},
			},
		},
		{
			Role: "user",
			Content: []*aiv1.ContentBlock{
				{
					BlockType: &aiv1.ContentBlock_Text{
						Text: fullUserPrompt,
					},
				},
			},
		},
	}

	// Connect to the AI service configured for the instance
	ai, release, err := s.runtime.AI(ctx, instanceID)
	if err != nil {
		return "", nil, err
	}
	defer release()

	// Apply timeout
	ctx, cancel := context.WithTimeout(ctx, aiGenerateTimeout)
	defer cancel()

	// Call AI service to infer a metrics view YAML
	res, err := ai.Complete(ctx, &drivers.CompleteOptions{
		Messages: msgs,
	})
	if err != nil {
		return "", nil, err
	}

	// Extract text from content blocks
	var responseText string
	for _, block := range res.Message.Content {
		switch blockType := block.GetBlockType().(type) {
		case *aiv1.ContentBlock_Text:
			if text := blockType.Text; text != "" {
				responseText += text
			}
		default:
			// For resolver generation, we only expect text responses
			return "", nil, fmt.Errorf("unexpected content block type in AI response: %T", blockType)
		}
	}

	// The AI may produce Markdown output. Remove the code tags around the SQL.
	responseText = strings.TrimPrefix(responseText, "```sql")
	responseText = strings.TrimPrefix(responseText, "```")
	responseText = strings.TrimSuffix(responseText, "```")
	// Remove the trailing semicolon
	responseText = semiColonRegex.ReplaceAllString(responseText, "")

	// Asking chatgpt to not add aggregations is not always honoured.
	// It is more consistent to ask it to wrap with AGGREGATE and strip it.
	responseText = aggregateCorrections.ReplaceAllString(responseText, "")

	return "metrics_sql", map[string]interface{}{
		"sql": responseText,
	}, nil
}

func resolverForMetricsViewSystemPrompt() string {
	return `
You are an agent whose only task is to suggest an SQL query to get data based on a table schema.
Wrap aggregations with "AGGREGATE", EG: AGGREGATE(impressions).
Do not use any complex aggregations and do not use WHERE or FILTER.
Do not add GROUP BY.

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
