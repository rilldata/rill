package gemini

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
			// Google warns that lowering temperature on Gemini 3 models can cause looping, so it's only sent if configured.
			name:   "default",
			config: map[string]any{},
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
					"candidates": [{"content": {"role": "model", "parts": [{"text": "hi"}]}}],
					"usageMetadata": {"promptTokenCount": 1, "candidatesTokenCount": 1}
				}`))
			}))
			defer srv.Close()
			t.Setenv("GOOGLE_GEMINI_BASE_URL", srv.URL)

			cfg := map[string]any{"api_key": "test-key"}
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

			genConfig, _ := body["generationConfig"].(map[string]any)
			if tc.want == nil {
				require.NotContains(t, genConfig, "temperature")
			} else {
				require.Equal(t, tc.want, genConfig["temperature"])
			}
		})
	}
}
