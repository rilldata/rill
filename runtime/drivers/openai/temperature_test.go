package openai

import (
	"context"
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

func TestCompleteOmitsTemperatureUnlessConfigured(t *testing.T) {
	// Current models don't support temperature unless reasoning effort is "none".
	body := completeAgainstFakeServer(t, nil)
	require.NotContains(t, body, "temperature")

	body = completeAgainstFakeServer(t, map[string]any{"temperature": 0.5})
	require.Equal(t, 0.5, body["temperature"])
}

// completeAgainstFakeServer runs a completion against a fake Chat Completions API and returns the request body it got.
func completeAgainstFakeServer(t *testing.T, config map[string]any) map[string]any {
	var body map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		require.NoError(t, json.Unmarshal(b, &body))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"1","object":"chat.completion","model":"test","choices":[{"index":0,"finish_reason":"stop","message":{"role":"assistant","content":"hello"}}],"usage":{"prompt_tokens":1,"completion_tokens":1}}`))
	}))
	t.Cleanup(srv.Close)

	conf := map[string]any{"api_key": "test", "base_url": srv.URL}
	for k, v := range config {
		conf[k] = v
	}
	h, err := driver{}.Open(context.Background(), "", "default", conf, nil, nil, zap.NewNop())
	require.NoError(t, err)
	ai, ok := h.AsAI("default")
	require.True(t, ok)
	_, err = ai.Complete(t.Context(), &drivers.CompleteOptions{
		Messages: []*aiv1.CompletionMessage{{Role: "user", Content: []*aiv1.ContentBlock{{BlockType: &aiv1.ContentBlock_Text{Text: "hi"}}}}},
	})
	require.NoError(t, err)
	return body
}
