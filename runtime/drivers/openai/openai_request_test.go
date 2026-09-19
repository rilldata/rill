package openai

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/google/jsonschema-go/jsonschema"
	aiv1 "github.com/rilldata/rill/proto/gen/rill/ai/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/stretchr/testify/require"
)

func TestCompleteAppliesConnectorRequestBehavior(t *testing.T) {
	fake := newFakeChatCompletionsServer([]string{completionResponse(`{"answer":"ok"}`)})
	defer fake.Close()

	ai := openTestAI(t, fake.URL, map[string]any{
		"structured_output_mode": structuredOutputModeJSONObject,
		"extra_body": map[string]any{
			"thinking": map[string]any{"type": "disabled"},
		},
	})

	_, err := ai.Complete(t.Context(), &drivers.CompleteOptions{
		Messages: []*aiv1.CompletionMessage{textMessage("user", "answer as JSON")},
		OutputSchema: &jsonschema.Schema{
			Type: "object",
			Properties: map[string]*jsonschema.Schema{
				"answer": {Type: "string"},
			},
			Required: []string{"answer"},
		},
	})
	require.NoError(t, err)

	body := fake.request(t, 0)
	require.Equal(t, map[string]any{"type": "disabled"}, body["thinking"])
	require.Equal(t, map[string]any{"type": "json_object"}, body["response_format"])

	messages := requireJSONArray(t, body["messages"])
	require.Len(t, messages, 2)
	schemaInstruction := requireJSONObject(t, messages[0])
	require.Equal(t, "system", schemaInstruction["role"])
	require.Contains(t, schemaInstruction["content"], "Return ONLY a single valid JSON object")
	require.Contains(t, schemaInstruction["content"], `"required":["answer"]`)
	require.Equal(t, "user", requireJSONObject(t, messages[1])["role"])
}

func TestCompletePassesNestedExtraBody(t *testing.T) {
	fake := newFakeChatCompletionsServer([]string{completionResponse("ok")})
	defer fake.Close()

	ai := openTestAI(t, fake.URL, map[string]any{
		"extra_body": map[string]any{
			"chat_template_kwargs": map[string]any{"enable_thinking": false},
		},
	})

	_, err := ai.Complete(t.Context(), &drivers.CompleteOptions{
		Messages: []*aiv1.CompletionMessage{textMessage("user", "hello")},
	})
	require.NoError(t, err)

	body := fake.request(t, 0)
	require.Equal(t, map[string]any{"enable_thinking": false}, body["chat_template_kwargs"])
	require.NotContains(t, body, "response_format")
}

func TestCompleteDefaultsToJSONSchema(t *testing.T) {
	fake := newFakeChatCompletionsServer([]string{completionResponse(`{"answer":"ok"}`)})
	defer fake.Close()

	ai := openTestAI(t, fake.URL, nil)
	_, err := ai.Complete(t.Context(), &drivers.CompleteOptions{
		Messages:     []*aiv1.CompletionMessage{textMessage("user", "answer as JSON")},
		OutputSchema: &jsonschema.Schema{Type: "object"},
	})
	require.NoError(t, err)

	body := fake.request(t, 0)
	require.Equal(t, "json_schema", requireJSONObject(t, body["response_format"])["type"])
	require.Len(t, requireJSONArray(t, body["messages"]), 1)
}

func TestOpenValidatesProviderRequestBehavior(t *testing.T) {
	t.Run("structured output mode", func(t *testing.T) {
		_, err := (driver{}).Open("", "", map[string]any{
			"api_key":                "test-key",
			"structured_output_mode": "xml",
		}, nil, nil, nil)
		require.ErrorContains(t, err, `invalid structured_output_mode "xml"`)
	})

	t.Run("reserved extra body fields", func(t *testing.T) {
		_, err := (driver{}).Open("", "", map[string]any{
			"api_key": "test-key",
			"extra_body": map[string]any{
				"audio":      map[string]any{"format": "wav"},
				"modalities": []any{"text", "audio"},
				"Tools":      []any{},
				"model":      "other-model",
				"stream":     true,
			},
		}, nil, nil, nil)
		require.EqualError(t, err, "extra_body cannot override core request fields: Tools, audio, modalities, model, stream")
	})

	t.Run("non JSON extra body", func(t *testing.T) {
		_, err := (driver{}).Open("", "", map[string]any{
			"api_key": "test-key",
			"extra_body": map[string]any{
				"extension": make(chan int),
			},
		}, nil, nil, nil)
		require.ErrorContains(t, err, "extra_body must contain JSON-serializable values")
	})
}

func textMessage(role, text string) *aiv1.CompletionMessage {
	return &aiv1.CompletionMessage{
		Role: role,
		Content: []*aiv1.ContentBlock{{
			BlockType: &aiv1.ContentBlock_Text{Text: text},
		}},
	}
}

func openTestAI(t *testing.T, serverURL string, config map[string]any) drivers.AIService {
	t.Helper()
	if config == nil {
		config = make(map[string]any)
	}
	config["api_key"] = "test-key"
	config["base_url"] = serverURL + "/v1"
	config["model"] = "test-model"

	handle, err := (driver{}).Open("", "", config, nil, nil, nil)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, handle.Close()) })
	ai, ok := handle.AsAI("")
	require.True(t, ok)
	return ai
}

type fakeChatCompletionsServer struct {
	*httptest.Server
	mu        sync.Mutex
	requests  []map[string]any
	responses []string
}

func newFakeChatCompletionsServer(responses []string) *fakeChatCompletionsServer {
	fake := &fakeChatCompletionsServer{responses: responses}
	fake.Server = httptest.NewServer(http.HandlerFunc(fake.serveHTTP))
	return fake
}

func (f *fakeChatCompletionsServer) serveHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "unexpected method "+r.Method, http.StatusMethodNotAllowed)
		return
	}
	if r.URL.Path != "/v1/chat/completions" {
		http.Error(w, "unexpected path "+r.URL.Path, http.StatusNotFound)
		return
	}

	var body map[string]any
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid JSON request: "+err.Error(), http.StatusBadRequest)
		return
	}

	f.mu.Lock()
	idx := len(f.requests)
	f.requests = append(f.requests, body)
	if idx >= len(f.responses) {
		f.mu.Unlock()
		http.Error(w, "unexpected request", http.StatusInternalServerError)
		return
	}
	response := f.responses[idx]
	f.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(response))
}

func (f *fakeChatCompletionsServer) request(t *testing.T, idx int) map[string]any {
	t.Helper()
	f.mu.Lock()
	defer f.mu.Unlock()
	require.Greater(t, len(f.requests), idx)
	return f.requests[idx]
}

func completionResponse(content string) string {
	encoded, _ := json.Marshal(content)
	return `{"id":"chatcmpl-1","object":"chat.completion","created":1,"model":"test-model","choices":[{"index":0,"message":{"role":"assistant","content":` +
		string(encoded) +
		`},"finish_reason":"stop"}],"usage":{"prompt_tokens":1,"completion_tokens":1,"total_tokens":2}}`
}

func requireJSONObject(t *testing.T, value any) map[string]any {
	t.Helper()
	result, ok := value.(map[string]any)
	require.True(t, ok, "expected JSON object, got %T", value)
	return result
}

func requireJSONArray(t *testing.T, value any) []any {
	t.Helper()
	result, ok := value.([]any)
	require.True(t, ok, "expected JSON array, got %T", value)
	return result
}
