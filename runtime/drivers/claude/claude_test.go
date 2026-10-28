package claude

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	aiv1 "github.com/rilldata/rill/proto/gen/rill/ai/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/metricsview"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
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
			var body map[string]any
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				data, err := io.ReadAll(r.Body)
				require.NoError(t, err)
				require.NoError(t, json.Unmarshal(data, &body))
				w.Header().Set("Content-Type", "application/json")
				// Current models return thinking blocks by default, with empty text unless display is "summarized".
				_, _ = w.Write([]byte(`{
					"id": "msg_1", "type": "message", "role": "assistant", "model": "test-model",
					"content": [
						{"type": "thinking", "thinking": "", "signature": "sig"},
						{"type": "text", "text": "Revenue is 100."}
					],
					"stop_reason": "end_turn",
					"usage": {"input_tokens": 10, "output_tokens": 3}
				}`))
			}))
			defer srv.Close()

			cfg := map[string]any{"api_key": "test-key", "base_url": srv.URL}
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
			require.Equal(t, tc.wantTemperature, body["temperature"])

			// The empty text block isn't sent.
			msgs := body["messages"].([]any)
			require.Len(t, msgs, 1)

			// The tool schema keeps $defs, which the $ref parameters point to.
			schema := body["tools"].([]any)[0].(map[string]any)["input_schema"].(map[string]any)
			require.Contains(t, schema["$defs"], "TimeRange")
			timeRange := schema["properties"].(map[string]any)["time_range"].(map[string]any)
			require.Equal(t, "#/$defs/TimeRange", timeRange["$ref"])

			// The thinking block isn't turned into a text block.
			require.Len(t, res.Message.Content, 1)
			require.Equal(t, "Revenue is 100.", res.Message.Content[0].GetText())
		})
	}
}

func textBlock(s string) *aiv1.ContentBlock {
	return &aiv1.ContentBlock{BlockType: &aiv1.ContentBlock_Text{Text: s}}
}
