package server

import (
	"context"
	"fmt"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/server/auth"
)

func (s *Server) GenerateChartFile(ctx context.Context, req *runtimev1.GenerateChartFileRequest) (*runtimev1.GenerateChartFileResponse, error) {
	// TODO: telemetry

	// Must have edit permissions on the repo
	if !auth.GetClaims(ctx).CanInstance(req.InstanceId, auth.EditRepo) {
		return nil, ErrForbidden
	}

	// Get chart
	chartSpec, err := s.getChart(ctx, req.InstanceId, req.Chart)
	if err != nil {
		return nil, err
	}

	schema, err := s.runtime.ResolveSchema(ctx, &runtime.ResolveOptions{
		InstanceID:         req.InstanceId,
		Resolver:           chartSpec.Resolver,
		ResolverProperties: chartSpec.ResolverProperties.AsMap(),
		Args:               nil,
		UserAttributes:     auth.GetClaims(ctx).Attributes(),
	})
	if err != nil {
		return nil, err
	}

	err = s.generateChartYAMLWithAI(ctx, req.InstanceId, req.Table, req.Prompt, schema)
	if err != nil {
		return nil, err
	}

	return &runtimev1.GenerateChartFileResponse{}, nil
}

func (s *Server) generateChartYAMLWithAI(ctx context.Context, instanceID, tblName, userPrompt string, schema *runtimev1.StructType) error {
	// Build messages
	msgs := []*drivers.CompletionMessage{
		{Role: "system", Data: chartYAMLSystemPrompt()},
		{Role: "user", Data: chartYAMLUserPrompt(tblName, userPrompt, schema)},
	}

	// Connect to the AI service configured for the instance
	ai, release, err := s.runtime.AI(ctx, instanceID)
	if err != nil {
		return err
	}
	defer release()

	// Apply timeout
	ctx, cancel := context.WithTimeout(ctx, aiGenerateTimeout)
	defer cancel()

	// Call AI service to infer a metrics view YAML
	res, err := ai.Complete(ctx, msgs)
	if err != nil {
		return err
	}

	fmt.Println(res.Data)

	return nil
}

// chartYAMLSystemPrompt returns the static system prompt for the chart generation AI.
func chartYAMLSystemPrompt() string {
	prompt := fmt.Sprintf(`
You are an agent whose only task is to suggest relevant chart based on a table schema.
Your output should only consist of valid JSON in the format below:

https://vega.github.io/schema/vega-lite/v5.json
`)

	return prompt
}

func chartYAMLUserPrompt(tblName, userPrompt string, schema *runtimev1.StructType) string {
	prompt := fmt.Sprintf(`
%s on the table named %q, which has the following schema:
`, userPrompt, tblName)
	for _, field := range schema.Fields {
		prompt += fmt.Sprintf("- column=%s, type=%s\n", field.Name, field.Type.Code.String())
	}
	return prompt
}
