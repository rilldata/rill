package openai

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	aiv1 "github.com/rilldata/rill/proto/gen/rill/ai/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestTemperature(t *testing.T) {
	cases := []struct {
		name   string
		config map[string]any
		want   any
	}{
		{
			name:   "default",
			config: map[string]any{},
			want:   0.1,
		},
		{
			// Reasoning models reject temperature unless reasoning effort is "none", so it's omitted when reasoning effort is set.
			name:   "reasoning effort",
			config: map[string]any{"reasoning_effort": "medium"},
			want:   nil,
		},
		{
			name:   "configured",
			config: map[string]any{"temperature": 0.5},
			want:   0.5,
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
				_, _ = w.Write([]byte(`{
					"id": "chatcmpl-1", "object": "chat.completion", "created": 1, "model": "test-model",
					"choices": [{"index": 0, "message": {"role": "assistant", "content": "hi"}, "finish_reason": "stop"}],
					"usage": {"prompt_tokens": 1, "completion_tokens": 1, "total_tokens": 2}
				}`))
			}))
			defer srv.Close()

			cfg := map[string]any{"api_key": "test-key", "base_url": srv.URL}
			for k, v := range tc.config {
				cfg[k] = v
			}
			h, err := driver{}.Open(t.Context(), "", "default", cfg, nil, nil, zap.NewNop())
			require.NoError(t, err)
			ai, _ := h.AsAI("default")
			_, err = ai.Complete(t.Context(), &drivers.CompleteOptions{
				Messages: []*aiv1.CompletionMessage{{Role: "user", Content: []*aiv1.ContentBlock{{BlockType: &aiv1.ContentBlock_Text{Text: "hi"}}}}},
			})
			require.NoError(t, err)
			require.Equal(t, tc.want, body["temperature"])
		})
	}
}
