package server

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	aiv1 "github.com/rilldata/rill/proto/gen/rill/ai/v1"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/rilldata/rill/runtime/server/auth"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/protobuf/types/known/structpb"
)

func (s *Server) GenerateRenderer(ctx context.Context, req *runtimev1.GenerateRendererRequest) (*runtimev1.GenerateRendererResponse, error) {
	rp, err := json.Marshal(req.ResolverProperties.AsMap())
	if err != nil {
		return nil, err
	}
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
		attribute.String("args.resolver", req.Resolver),
		attribute.String("args.resolver_property", string(rp)),
	)
	s.addInstanceRequestAttributes(ctx, req.InstanceId)

	// Must have edit permissions on the repo
	claims := auth.GetClaims(ctx, req.InstanceId)
	if !claims.Can(runtime.EditRepo) {
		return nil, ErrForbidden
	}

	res, err := s.runtime.Resolve(ctx, &runtime.ResolveOptions{
		InstanceID:         req.InstanceId,
		Resolver:           req.Resolver,
		ResolverProperties: req.ResolverProperties.AsMap(),
		Args:               nil,
		Claims:             claims,
	})
	if err != nil {
		return nil, err
	}
	defer res.Close()

	start := time.Now()
	renderer, props, err := s.generateRendererWithAI(ctx, req.InstanceId, req.Prompt, res.Schema())

	var propsPB *structpb.Struct
	if err == nil && props != nil {
		propsPB, err = structpb.NewStruct(props)
	}

	s.activity.Record(ctx, activity.EventTypeLog, "ai_generated_renderer",
		attribute.Int("table_column_count", len(res.Schema().Fields)),
		attribute.Int64("elapsed_ms", time.Since(start).Milliseconds()),
		attribute.Bool("succeeded", err == nil),
	)

	if err != nil {
		return nil, err
	}

	return &runtimev1.GenerateRendererResponse{
		Renderer:           renderer,
		RendererProperties: propsPB,
	}, nil
}

// generateRendererWithAI attempts to generate a component renderer based on a user-provided prompt and a data schema.
// It currently only supports generating a Vega lite render.
func (s *Server) generateRendererWithAI(ctx context.Context, instanceID, userPrompt string, schema *runtimev1.StructType) (string, map[string]any, error) {
	// Build messages
	systemPrompt := vegaSpecSystemPrompt()
	userPrompt = vegaSpecUserPrompt(userPrompt, schema)

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
						Text: userPrompt,
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
			// For chart generation, we only expect text responses
			return "", nil, fmt.Errorf("unexpected content block type in AI response: %T", blockType)
		}
	}

	// The AI may produce Markdown output. Remove the code tags around the JSON.
	responseText = strings.TrimPrefix(responseText, "```json")
	responseText = strings.TrimPrefix(responseText, "```")
	responseText = strings.TrimSuffix(responseText, "```")

	return "vega_lite", map[string]any{"spec": responseText}, nil
}

// vegaSpecSystemPrompt returns the static system prompt for the Vega spec generation AI.
func vegaSpecSystemPrompt() string {
	// `{ "name": "table" }` is our format to add data in the UI. To retain the formatting of the json it is better to ask AI to keep this as the `data` config.
	return `
You are an agent whose only task is to suggest relevant chart based on a table schema.
Replace the data field in vega lite json with,
{ "name": "table" }

Your output should consist of valid JSON in the format below:

<vega lite json in the format: https://vega.github.io/schema/vega-lite/v5.json >
`
}

func vegaSpecUserPrompt(userPrompt string, schema *runtimev1.StructType) string {
	prompt := fmt.Sprintf(`
Prompt provided by the user: %s:

Based on a table with schema:
`, userPrompt)
	for _, field := range schema.Fields {
		prompt += fmt.Sprintf("- column=%s, type=%s\n", field.Name, field.Type.Code.String())
	}
	return prompt
}
