package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/server/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) GenerateChartSpec(ctx context.Context, req *runtimev1.GenerateChartSpecRequest) (*runtimev1.GenerateChartSpecResponse, error) {
	// TODO: telemetry

	// Must have edit permissions on the repo
	if !auth.GetClaims(ctx).CanInstance(req.InstanceId, auth.EditRepo) {
		return nil, ErrForbidden
	}

	var schema *runtimev1.StructType
	var table string
	var err error

	if req.Chart != "" {
		// Get chart
		chartSpec, err := s.getChart(ctx, req.InstanceId, req.Chart)
		if err != nil {
			return nil, err
		}

		// Resolve schema of the chart
		schema, err = s.runtime.ResolveSchema(ctx, &runtime.ResolveOptions{
			InstanceID:         req.InstanceId,
			Resolver:           chartSpec.Resolver,
			ResolverProperties: chartSpec.ResolverProperties.AsMap(),
			Args:               nil,
			UserAttributes:     auth.GetClaims(ctx).Attributes(),
		})
		if err != nil {
			return nil, err
		}

		table = req.Chart // not needed but better not to leave it empty
	} else if req.Table != "" {
		// Get instance
		inst, err := s.runtime.Instance(ctx, req.InstanceId)
		if err != nil {
			return nil, err
		}

		// Connect to connector and check it's an OLAP db
		handle, release, err := s.runtime.AcquireHandle(ctx, req.InstanceId, inst.ResolveOLAPConnector())
		if err != nil {
			return nil, err
		}
		defer release()
		olap, ok := handle.AsOLAP(req.InstanceId)
		if !ok {
			return nil, status.Error(codes.InvalidArgument, "connector is not an OLAP connector")
		}

		// Get table info
		tbl, err := olap.InformationSchema().Lookup(ctx, req.Table)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "table not found")
		}

		schema = tbl.Schema
		table = req.Table
	} else {
		// TODO: support metrics view
		return nil, errors.New("at lest one of chart or table must be specified")
	}

	resp, err := s.generateChartWithAI(ctx, req.InstanceId, table, req.Prompt, schema)
	if err != nil {
		return nil, err
	}

	// Convert the vega lite spec json to string
	spec, err := json.Marshal(resp.VegaLiteSpec)
	if err != nil {
		return nil, err
	}

	return &runtimev1.GenerateChartSpecResponse{
		VegaLiteSpec: string(spec),
		Sql:          resp.SQL,
	}, nil
}

type chartAIResponse struct {
	VegaLiteSpec map[string]interface{} `json:"vega_lite_spec"`
	SQL          string                 `json:"sql"`
}

// generateChartWithAI attempts to generate a vega lite chart spec and a SQL used to display the chart.
func (s *Server) generateChartWithAI(ctx context.Context, instanceID, tblName, userPrompt string, schema *runtimev1.StructType) (*chartAIResponse, error) {
	// Build messages
	msgs := []*drivers.CompletionMessage{
		{Role: "system", Data: chartYAMLSystemPrompt()},
		{Role: "user", Data: chartYAMLUserPrompt(tblName, userPrompt, schema)},
	}

	// Connect to the AI service configured for the instance
	ai, release, err := s.runtime.AI(ctx, instanceID)
	if err != nil {
		return nil, err
	}
	defer release()

	// Apply timeout
	ctx, cancel := context.WithTimeout(ctx, aiGenerateTimeout)
	defer cancel()

	// Call AI service to infer a metrics view YAML
	res, err := ai.Complete(ctx, msgs)
	if err != nil {
		return nil, err
	}

	// The AI may produce Markdown output. Remove the code tags around the YAML.
	res.Data = strings.TrimPrefix(res.Data, "```yaml")
	res.Data = strings.TrimPrefix(res.Data, "```")
	res.Data = strings.TrimSuffix(res.Data, "```")

	res.Data = strings.TrimPrefix(res.Data, "```json")
	res.Data = strings.TrimPrefix(res.Data, "```")
	res.Data = strings.TrimSuffix(res.Data, "```")

	res.Data = strings.TrimPrefix(res.Data, "```sql")
	res.Data = strings.TrimPrefix(res.Data, "```")
	res.Data = strings.TrimSuffix(res.Data, "```")

	resp := &chartAIResponse{}
	err = json.Unmarshal([]byte(res.Data), resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// chartYAMLSystemPrompt returns the static system prompt for the chart generation AI.
func chartYAMLSystemPrompt() string {
	return `
You are an agent whose only task is to suggest relevant chart based on a table schema.
You should suggest the vega config for the chart based on the data.
Replace the data field in vega lite json with,
{ "name": "table" }

Your output should consist of valid JSON in the format below:
{
  "vega_lite_spec": <vega lite json in the format: https://vega.github.io/schema/vega-lite/v5.json >

  "sql": "<the sql query used to fetch the data for the above chart. replace new line with \n>"
}
`
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
