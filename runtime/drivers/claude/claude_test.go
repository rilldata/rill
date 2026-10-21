package claude

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
	// Current models reject any sampling parameter with a 400.
	h, srv := newTestHandle(t, nil, textResponse)
	_, err := h.Complete(t.Context(), userPrompt("hi"))
	require.NoError(t, err)
	require.NotContains(t, srv.lastBody(t), "temperature")

	h, srv = newTestHandle(t, map[string]any{"temperature": 0.5}, textResponse)
	_, err = h.Complete(t.Context(), userPrompt("hi"))
	require.NoError(t, err)
	require.Equal(t, 0.5, srv.lastBody(t)["temperature"])
}

func TestCompleteKeepsSchemaDefs(t *testing.T) {
	// Without $defs, every $ref is dangling and the model sends nested objects as strings.
	h, srv := newTestHandle(t, nil, textResponse)
	opts := userPrompt("hi")
	opts.Tools = []*aiv1.Tool{{
		Name:        "query",
		InputSchema: `{"type":"object","properties":{"time_range":{"$ref":"#/$defs/TimeRange"}},"required":["time_range"],"$defs":{"TimeRange":{"type":"object","properties":{"start":{"type":"string"}}}}}`,
	}}
	_, err := h.Complete(t.Context(), opts)
	require.NoError(t, err)

	tools := srv.lastBody(t)["tools"].([]any)
	schema := tools[0].(map[string]any)["input_schema"].(map[string]any)
	require.Equal(t, "object", schema["type"])
	require.Equal(t, []any{"time_range"}, schema["required"])
	require.Contains(t, schema["properties"], "time_range")
	require.Contains(t, schema["$defs"], "TimeRange")
}

func TestCompleteSetsCacheBreakpoints(t *testing.T) {
	// Tools, system prompt and history are resent on every step of the tool loop.
	h, srv := newTestHandle(t, nil, textResponse)
	opts := &drivers.CompleteOptions{
		Messages: []*aiv1.CompletionMessage{
			textMessage("system", "You are an analyst."),
			textMessage("user", "hi"),
		},
		Tools: []*aiv1.Tool{{Name: "a"}, {Name: "b"}},
	}
	_, err := h.Complete(t.Context(), opts)
	require.NoError(t, err)

	body := srv.lastBody(t)
	tools := body["tools"].([]any)
	require.NotContains(t, tools[0], "cache_control")
	require.Contains(t, tools[1], "cache_control")
	system := body["system"].([]any)
	require.Contains(t, system[len(system)-1], "cache_control")
	messages := body["messages"].([]any)
	content := messages[len(messages)-1].(map[string]any)["content"].([]any)
	require.Contains(t, content[len(content)-1], "cache_control")
}

func TestCompleteDropsThinkingBlocks(t *testing.T) {
	// Current models return the thinking block empty by default; as a text block it is rejected on replay.
	resp := `{"id":"msg_1","type":"message","role":"assistant","model":"claude-test","content":[{"type":"thinking","thinking":"","signature":"sig"},{"type":"text","text":"hello"}],"stop_reason":"end_turn","usage":{"input_tokens":10,"output_tokens":5}}`
	h, _ := newTestHandle(t, nil, resp)
	res, err := h.Complete(t.Context(), userPrompt("hi"))
	require.NoError(t, err)
	require.Len(t, res.Message.Content, 1)
	require.Equal(t, "hello", res.Message.Content[0].GetText())
}

const textResponse = `{"id":"msg_1","type":"message","role":"assistant","model":"claude-test","content":[{"type":"text","text":"hello"}],"stop_reason":"end_turn","usage":{"input_tokens":10,"output_tokens":5}}`

type testServer struct {
	bodies [][]byte
}

// newTestHandle opens the driver against a fake Messages API that answers every request with response.
func newTestHandle(t *testing.T, config map[string]any, response string) (*handle, *testServer) {
	ts := &testServer{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		ts.bodies = append(ts.bodies, b)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(response))
	}))
	t.Cleanup(srv.Close)

	conf := map[string]any{"api_key": "test", "base_url": srv.URL, "model": "claude-test"}
	for k, v := range config {
		conf[k] = v
	}
	h, err := driver{}.Open(context.Background(), "", "default", conf, nil, nil, zap.NewNop())
	require.NoError(t, err)
	return h.(*handle), ts
}

func (ts *testServer) lastBody(t *testing.T) map[string]any {
	require.NotEmpty(t, ts.bodies)
	var body map[string]any
	require.NoError(t, json.Unmarshal(ts.bodies[len(ts.bodies)-1], &body))
	return body
}

func userPrompt(text string) *drivers.CompleteOptions {
	return &drivers.CompleteOptions{Messages: []*aiv1.CompletionMessage{textMessage("user", text)}}
}

func textMessage(role, text string) *aiv1.CompletionMessage {
	return &aiv1.CompletionMessage{
		Role:    role,
		Content: []*aiv1.ContentBlock{{BlockType: &aiv1.ContentBlock_Text{Text: text}}},
	}
}
