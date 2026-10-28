package claude

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	aiv1 "github.com/rilldata/rill/proto/gen/rill/ai/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/metricsview"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestComplete(t *testing.T) {
	cases := []struct {
		name            string
		config          map[string]any
		wantTemperature any
	}{
		{
			// Current models reject temperature (or non-default values), so it's only sent if configured.
			name:            "default",
			config:          map[string]any{},
			wantTemperature: nil,
		},
		{
			name:            "temperature configured",
			config:          map[string]any{"temperature": 0.5},
			wantTemperature: 0.5,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Current models return thinking blocks by default, with empty text unless display is "summarized".
			body, url := newFakeClaude(t, `{
				"id": "msg_1", "type": "message", "role": "assistant", "model": "test-model",
				"content": [
					{"type": "thinking", "thinking": "", "signature": "sig"},
					{"type": "text", "text": "Revenue is 100."}
				],
				"stop_reason": "end_turn",
				"usage": {"input_tokens": 10, "output_tokens": 3}
			}`)

			cfg := map[string]any{"api_key": "test-key", "base_url": url}
			for k, v := range tc.config {
				cfg[k] = v
			}
			h, err := driver{}.Open(t.Context(), "", "default", cfg, nil, nil, zap.NewNop())
			require.NoError(t, err)
			ai, ok := h.AsAI("default")
			require.True(t, ok)

			res, err := ai.Complete(t.Context(), &drivers.CompleteOptions{
				Messages: []*aiv1.CompletionMessage{
					{Role: "user", Content: []*aiv1.ContentBlock{textBlock("What is revenue?")}},
					// An empty text block, e.g. from a thinking block in a conversation saved before the fix.
					{Role: "assistant", Content: []*aiv1.ContentBlock{textBlock("")}},
				},
				Tools: []*aiv1.Tool{{Name: "query_metrics_view", InputSchema: metricsview.QueryJSONSchema}},
			})
			require.NoError(t, err)

			// Temperature is only sent if configured.
			require.Equal(t, tc.wantTemperature, (*body)["temperature"])

			// The empty text block isn't sent.
			msgs := (*body)["messages"].([]any)
			require.Len(t, msgs, 1)

			// The tool schema keeps $defs, which the $ref parameters point to.
			schema := (*body)["tools"].([]any)[0].(map[string]any)["input_schema"].(map[string]any)
			require.Contains(t, schema["$defs"], "TimeRange")
			timeRange := schema["properties"].(map[string]any)["time_range"].(map[string]any)
			require.Equal(t, "#/$defs/TimeRange", timeRange["$ref"])

			// The thinking block is returned as reasoning (with its signature), not as a text block.
			require.Len(t, res.Message.Content, 2)
			require.Equal(t, "sig", res.Message.Content[0].GetReasoning().GetSignature())
			require.Equal(t, "Revenue is 100.", res.Message.Content[1].GetText())
		})
	}
}

func TestPromptCaching(t *testing.T) {
	for _, enabled := range []bool{false, true} {
		t.Run(fmt.Sprintf("enabled=%v", enabled), func(t *testing.T) {
			body, url := newFakeClaude(t, `{
				"id": "msg_1", "type": "message", "role": "assistant", "model": "test-model",
				"content": [{"type": "text", "text": "hi"}],
				"stop_reason": "end_turn",
				"usage": {"input_tokens": 10, "output_tokens": 3, "cache_read_input_tokens": 100, "cache_creation_input_tokens": 50}
			}`)

			h, err := driver{}.Open(t.Context(), "", "default", map[string]any{"api_key": "test-key", "base_url": url, "prompt_caching": enabled}, nil, nil, zap.NewNop())
			require.NoError(t, err)
			ai, _ := h.AsAI("default")
			res, err := ai.Complete(t.Context(), &drivers.CompleteOptions{
				Messages: []*aiv1.CompletionMessage{
					{Role: "system", Content: []*aiv1.ContentBlock{textBlock("You are an analyst.")}},
					{Role: "user", Content: []*aiv1.ContentBlock{textBlock("What is revenue?")}},
				},
				Tools: []*aiv1.Tool{{Name: "first"}, {Name: "last"}},
			})
			require.NoError(t, err)

			// Cache reads and writes are reported regardless of whether caching is enabled.
			require.Equal(t, 100, res.CachedInputTokens)
			require.Equal(t, 50, res.CacheWriteInputTokens)

			// If enabled, the last tool, system block, and message block are marked as cache breakpoints.
			tools := (*body)["tools"].([]any)
			system := (*body)["system"].([]any)
			msgs := (*body)["messages"].([]any)
			lastMsgContent := msgs[len(msgs)-1].(map[string]any)["content"].([]any)
			marked := []any{
				tools[len(tools)-1],
				system[len(system)-1],
				lastMsgContent[len(lastMsgContent)-1],
			}
			for _, block := range marked {
				if enabled {
					require.Equal(t, map[string]any{"type": "ephemeral"}, block.(map[string]any)["cache_control"])
				} else {
					require.NotContains(t, block, "cache_control")
				}
			}
			require.NotContains(t, tools[0], "cache_control")
		})
	}
}

func TestReasoning(t *testing.T) {
	// Claude requires thinking blocks to be passed back within a tool call loop,
	// but a thinking block is bound to the tools offered when it was produced, so they're removed when no tools are offered.
	for _, toolsOffered := range []bool{true, false} {
		t.Run(fmt.Sprintf("tools=%v", toolsOffered), func(t *testing.T) {
			body, url := newFakeClaude(t, `{
				"id": "msg_1", "type": "message", "role": "assistant", "model": "test-model",
				"content": [
					{"type": "thinking", "thinking": "Thinking about revenue.", "signature": "sig-2"},
					{"type": "redacted_thinking", "data": "encrypted"},
					{"type": "text", "text": "Revenue is 100."}
				],
				"stop_reason": "end_turn",
				"usage": {"input_tokens": 10, "output_tokens": 3}
			}`)
			h, err := driver{}.Open(t.Context(), "", "default", map[string]any{"api_key": "test-key", "base_url": url}, nil, nil, zap.NewNop())
			require.NoError(t, err)
			ai, _ := h.AsAI("default")

			input, err := structpb.NewStruct(map[string]any{})
			require.NoError(t, err)
			opts := &drivers.CompleteOptions{
				Messages: []*aiv1.CompletionMessage{
					{Role: "user", Content: []*aiv1.ContentBlock{textBlock("What is revenue?")}},
					{Role: "assistant", Content: []*aiv1.ContentBlock{{BlockType: &aiv1.ContentBlock_Reasoning{Reasoning: &aiv1.Reasoning{Signature: "sig-1"}}}}},
					{Role: "assistant", Content: []*aiv1.ContentBlock{{BlockType: &aiv1.ContentBlock_ToolCall{ToolCall: &aiv1.ToolCall{Id: "call-1", Name: "query", Input: input}}}}},
					{Role: "user", Content: []*aiv1.ContentBlock{{BlockType: &aiv1.ContentBlock_ToolResult{ToolResult: &aiv1.ToolResult{Id: "call-1", Content: "100"}}}}},
				},
			}
			if toolsOffered {
				opts.Tools = []*aiv1.Tool{{Name: "query"}}
			}
			res, err := ai.Complete(t.Context(), opts)
			require.NoError(t, err)

			// Thinking and redacted thinking blocks are returned as reasoning.
			require.Len(t, res.Message.Content, 3)
			require.Equal(t, "Thinking about revenue.", res.Message.Content[0].GetReasoning().GetText())
			require.Equal(t, "sig-2", res.Message.Content[0].GetReasoning().GetSignature())
			require.Equal(t, "encrypted", res.Message.Content[1].GetReasoning().GetRedactedData())

			// The reasoning in the history is passed back as a thinking block before its tool call, unless no tools are offered.
			var types []string
			for _, msg := range (*body)["messages"].([]any) {
				for _, block := range msg.(map[string]any)["content"].([]any) {
					block := block.(map[string]any)
					types = append(types, block["type"].(string))
					if block["type"] == "thinking" {
						require.Equal(t, "sig-1", block["signature"])
					}
				}
			}
			if toolsOffered {
				require.Equal(t, []string{"text", "thinking", "tool_use", "tool_result"}, types)
			} else {
				require.Equal(t, []string{"text", "tool_use", "tool_result"}, types)
			}
		})
	}
}

// newFakeClaude starts a fake Claude API server that returns the given response.
// It returns a pointer to the decoded body of the last request, and the server's URL to use as the base URL.
func newFakeClaude(t *testing.T, response string) (*map[string]any, string) {
	body := &map[string]any{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		require.NoError(t, json.Unmarshal(data, body))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(response))
	}))
	t.Cleanup(srv.Close)
	return body, srv.URL
}

func textBlock(s string) *aiv1.ContentBlock {
	return &aiv1.ContentBlock{BlockType: &aiv1.ContentBlock_Text{Text: s}}
}
