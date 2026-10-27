package bedrock

import (
	"encoding/json"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/document"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
	aiv1 "github.com/rilldata/rill/proto/gen/rill/ai/v1"
	"github.com/rilldata/rill/runtime/metricsview"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestConvertMessages(t *testing.T) {
	input, err := structpb.NewStruct(map[string]any{"query": "revenue", "limit": 10})
	require.NoError(t, err)

	msgs := []*aiv1.CompletionMessage{
		{Role: "system", Content: []*aiv1.ContentBlock{text("You are helpful.")}},
		{Role: "user", Content: []*aiv1.ContentBlock{text("Hello"), text("")}},
		{Role: "user", Content: []*aiv1.ContentBlock{text("What is revenue?")}},
		{Role: "assistant", Content: []*aiv1.ContentBlock{
			text("Let me check."),
			{BlockType: &aiv1.ContentBlock_ToolCall{ToolCall: &aiv1.ToolCall{Id: "call_1", Name: "query", Input: input}}},
		}},
		{Role: "user", Content: []*aiv1.ContentBlock{
			{BlockType: &aiv1.ContentBlock_ToolResult{ToolResult: &aiv1.ToolResult{Id: "call_1", Content: "boom", IsError: true}}},
		}},
		{Role: "assistant", Content: []*aiv1.ContentBlock{text("Revenue is 100.")}},
	}

	system, res, err := convertMessages(msgs, true)
	require.NoError(t, err)

	require.Len(t, system, 1)
	require.Equal(t, "You are helpful.", system[0].(*types.SystemContentBlockMemberText).Value)

	// Consecutive user messages are merged, blank text is dropped, and roles alternate.
	require.Len(t, res, 4)
	require.Equal(t, types.ConversationRoleUser, res[0].Role)
	require.Len(t, res[0].Content, 2)
	require.Equal(t, "What is revenue?", res[0].Content[1].(*types.ContentBlockMemberText).Value)

	require.Equal(t, types.ConversationRoleAssistant, res[1].Role)
	require.Len(t, res[1].Content, 2)
	toolUse := res[1].Content[1].(*types.ContentBlockMemberToolUse).Value
	require.Equal(t, "call_1", aws.ToString(toolUse.ToolUseId))
	require.Equal(t, "query", aws.ToString(toolUse.Name))
	data, err := toolUse.Input.MarshalSmithyDocument()
	require.NoError(t, err)
	require.JSONEq(t, `{"query":"revenue","limit":10}`, string(data))

	require.Equal(t, types.ConversationRoleUser, res[2].Role)
	toolResult := res[2].Content[0].(*types.ContentBlockMemberToolResult).Value
	require.Equal(t, "call_1", aws.ToString(toolResult.ToolUseId))
	require.Equal(t, types.ToolResultStatusError, toolResult.Status)
	require.Equal(t, "boom", toolResult.Content[0].(*types.ToolResultContentBlockMemberText).Value)

	require.Equal(t, types.ConversationRoleAssistant, res[3].Role)
}

func TestConvertTools(t *testing.T) {
	tools, err := convertTools([]*aiv1.Tool{
		{Name: "no_args", Description: "Takes no arguments"},
		{Name: "query", InputSchema: `{"type":"object","properties":{"sql":{"type":"string"}},"required":["sql"]}`},
	})
	require.NoError(t, err)
	require.Len(t, tools, 2)

	spec := tools[0].(*types.ToolMemberToolSpec).Value
	require.Equal(t, "no_args", aws.ToString(spec.Name))
	require.Equal(t, "Takes no arguments", aws.ToString(spec.Description))
	data, err := spec.InputSchema.(*types.ToolInputSchemaMemberJson).Value.MarshalSmithyDocument()
	require.NoError(t, err)
	require.JSONEq(t, `{"type":"object","properties":{}}`, string(data))

	spec = tools[1].(*types.ToolMemberToolSpec).Value
	require.Nil(t, spec.Description)
	data, err = spec.InputSchema.(*types.ToolInputSchemaMemberJson).Value.MarshalSmithyDocument()
	require.NoError(t, err)
	require.JSONEq(t, `{"type":"object","properties":{"sql":{"type":"string"}},"required":["sql"]}`, string(data))

	_, err = convertTools([]*aiv1.Tool{{Name: "bad", InputSchema: "{"}})
	require.Error(t, err)
}

func TestConvertToolsRefTypes(t *testing.T) {
	// Some Bedrock models treat $ref parameters without a type as strings, so the type is copied from the definition.
	tools, err := convertTools([]*aiv1.Tool{
		{Name: "query_metrics_view", InputSchema: metricsview.QueryJSONSchema},
		{Name: "other", InputSchema: `{"type":"object","properties":{"external":{"$ref":"https://example.com/schema.json"}}}`},
	})
	require.NoError(t, err)

	schemaOf := func(tool types.Tool) map[string]any {
		data, err := tool.(*types.ToolMemberToolSpec).Value.InputSchema.(*types.ToolInputSchemaMemberJson).Value.MarshalSmithyDocument()
		require.NoError(t, err)
		var schema map[string]any
		require.NoError(t, json.Unmarshal(data, &schema))
		return schema
	}

	schema := schemaOf(tools[0])
	props := schema["properties"].(map[string]any)
	for _, name := range []string{"time_range", "comparison_time_range", "where"} {
		prop := props[name].(map[string]any)
		require.Equal(t, "object", prop["type"], name)
		require.NotEmpty(t, prop["$ref"], name)
	}
	require.Equal(t, "array", props["dimensions"].(map[string]any)["type"])
	require.Contains(t, schema["$defs"], "TimeRange")

	// Refs that can't be resolved locally are left as-is.
	external := schemaOf(tools[1])["properties"].(map[string]any)["external"].(map[string]any)
	require.NotContains(t, external, "type")
}

func TestConvertResponseMessage(t *testing.T) {
	input := document.NewLazyDocument(map[string]any{"sql": "SELECT 1", "limit": 5, "nested": map[string]any{"x": 1.5}})

	msg, err := convertResponseMessage(&types.Message{
		Role: types.ConversationRoleAssistant,
		Content: []types.ContentBlock{
			&types.ContentBlockMemberReasoningContent{},
			&types.ContentBlockMemberText{Value: "Running a query."},
			&types.ContentBlockMemberToolUse{Value: types.ToolUseBlock{ToolUseId: aws.String("call_2"), Name: aws.String("query"), Input: input}},
		},
	})
	require.NoError(t, err)
	require.Equal(t, "assistant", msg.Role)
	require.Len(t, msg.Content, 2)
	require.Equal(t, "Running a query.", msg.Content[0].GetText())

	call := msg.Content[1].GetToolCall()
	require.Equal(t, "call_2", call.Id)
	require.Equal(t, "query", call.Name)
	require.Equal(t, map[string]any{"sql": "SELECT 1", "limit": float64(5), "nested": map[string]any{"x": 1.5}}, call.Input.AsMap())
}

func TestConvertMessagesWithoutTools(t *testing.T) {
	// On the final iteration of a tool call loop, no tools are offered, but the history contains tool calls.
	// Converse rejects tool blocks without tools, so they must be rendered as text.
	input, err := structpb.NewStruct(map[string]any{"sql": "SELECT 1"})
	require.NoError(t, err)
	msgs := []*aiv1.CompletionMessage{
		{Role: "user", Content: []*aiv1.ContentBlock{text("What is revenue?")}},
		{Role: "assistant", Content: []*aiv1.ContentBlock{
			{BlockType: &aiv1.ContentBlock_ToolCall{ToolCall: &aiv1.ToolCall{Id: "call_1", Name: "query", Input: input}}},
		}},
		{Role: "user", Content: []*aiv1.ContentBlock{
			{BlockType: &aiv1.ContentBlock_ToolResult{ToolResult: &aiv1.ToolResult{Id: "call_1", Content: "bad query", IsError: true}}},
		}},
		{Role: "user", Content: []*aiv1.ContentBlock{text("Tool call limit reached. Provide a final response without additional tool calls.")}},
	}

	_, res, err := convertMessages(msgs, false)
	require.NoError(t, err)
	require.Len(t, res, 3)
	for _, msg := range res {
		for _, block := range msg.Content {
			require.IsType(t, &types.ContentBlockMemberText{}, block)
		}
	}
	require.Equal(t, types.ConversationRoleAssistant, res[1].Role)
	require.Equal(t, `Called tool "query" with input: {"sql":"SELECT 1"}`, res[1].Content[0].(*types.ContentBlockMemberText).Value)
	require.Equal(t, types.ConversationRoleUser, res[2].Role)
	require.Len(t, res[2].Content, 2)
	require.Equal(t, "Error from tool \"query\":\nbad query", res[2].Content[0].(*types.ContentBlockMemberText).Value)
}

func text(s string) *aiv1.ContentBlock {
	return &aiv1.ContentBlock{BlockType: &aiv1.ContentBlock_Text{Text: s}}
}
