package ai_test

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/uuid"
	aiv1 "github.com/rilldata/rill/proto/gen/rill/ai/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/ai"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	_ "github.com/rilldata/rill/runtime/resolvers"
)

// newSession sets up a new AI session for testing.
// It is suitable for testing tool calls that do not require LLM completions.
// If LLM completions are needed, use newEval instead.
func newSession(t *testing.T, rt *runtime.Runtime, instanceID string) *ai.Session {
	// Create test AI session
	claims := &runtime.SecurityClaims{UserID: uuid.NewString(), SkipChecks: true}
	r := ai.NewRunner(rt, activity.NewNoopClient())
	s, err := r.Session(t.Context(), &ai.SessionOptions{
		InstanceID: instanceID,
		Claims:     claims,
		UserAgent:  "rill-evals",
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		err := s.Flush(t.Context())
		require.NoError(t, err)
	})

	return s
}

// newEval sets up a new eval AI session for a runtime instance.
// It differs from newSession in a two ways:
// - It records messages and LLM calls to the ./evals directory for later inspection.
// - It ensures the test is skipped in CI (short mode).
func newEval(t *testing.T, rt *runtime.Runtime, instanceID string) *ai.Session {
	// Eval tests are expensive, but we don't use testmode.Expensive here because it should have already been called before ingesting data.
	// Most of the time, the test should have been marked expensive when EnableLLM was set, and we can check it wasn't forgotten by checking testing.Short().
	if testing.Short() {
		t.Fatal("eval test was not marked expensive; did you forget to set EnableLLM in testruntime.InstanceOptions?")
	}

	// Create test runtime instance and AI session
	s := newSession(t, rt, instanceID)

	// Wrap the session's LLM with a recordingAIService to capture every LLM call.
	ai, release, err := rt.AI(t.Context(), instanceID)
	require.NoError(t, err)
	t.Cleanup(release)
	wrappedAI := &recordingAIService{ai: ai}
	s.SetLLM(func(ctx context.Context) (drivers.AIService, func(), error) {
		return wrappedAI, func() {}, nil
	})

	// When the test is done, save the transcripts to ./evals
	t.Cleanup(func() {
		// Setup the output destination
		dir := "evals"
		name := strings.TrimPrefix(t.Name(), "Test")
		name = strings.ReplaceAll(name, "/", "_") // Replace slashes in sub-test names
		err = os.MkdirAll(dir, 0755)
		require.NoError(t, err)

		// Save the session messages to ./testdata/<test name>.messages.yaml.
		buf := &bytes.Buffer{}
		yamlEncoder := yaml.NewEncoder(buf)
		yamlEncoder.SetIndent(2)
		err := yamlEncoder.Encode(s.Messages())
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(dir, name+".messages.yaml"), buf.Bytes(), 0644)
		require.NoError(t, err)

		// Save the LLM invocations recorded in wrappedAI to ./testdata/<test name>.completions.yaml.
		buf = &bytes.Buffer{}
		yamlEncoder = yaml.NewEncoder(buf)
		yamlEncoder.SetIndent(2)
		err = yamlEncoder.Encode(wrappedAI.calls)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(dir, name+".completions.yaml"), buf.Bytes(), 0644)
		require.NoError(t, err)
	})

	return s
}

// recordingAIService wraps a drivers.AIService and records all interactions with it.
type recordingAIService struct {
	ai    drivers.AIService
	calls []*recordingAICall
}

// recordingAICall represents a recorded call in a recordingAIService.
type recordingAICall struct {
	Index    int                  `yaml:"index"`
	Input    []recordingAIMessage `yaml:"input"`
	Error    string               `yaml:"error,omitempty"`
	Response []recordingAIMessage `yaml:"response,omitempty"`
}

// recordingAIMessage represents a recorded message in a recordingAIService.
type recordingAIMessage struct {
	Role        string `yaml:"role"`
	ContentType string `yaml:"content_type"`
	ID          string `yaml:"id,omitempty"`
	ToolName    string `yaml:"tool_name,omitempty"`
	IsError     bool   `yaml:"is_error,omitempty"`
	Content     string `yaml:"content"`
}

var _ drivers.AIService = &recordingAIService{}

// Complete(ctx context.Context, msgs []*aiv1.CompletionMessage, tools []*aiv1.Tool, outputSchema *jsonschema.Schema) (*aiv1.CompletionMessage, error)
func (r *recordingAIService) Complete(ctx context.Context, opts *drivers.CompleteOptions) (*drivers.CompleteResult, error) {
	// Create a recorded call
	call := &recordingAICall{Index: len(r.calls) + 1}
	for _, m := range opts.Messages {
		call.Input = append(call.Input, newRecordingAIMessages(m)...)
	}
	r.calls = append(r.calls, call)

	// Forward to the underlying AI service
	res, err := r.ai.Complete(ctx, opts)
	if err != nil {
		call.Error = err.Error()
		return nil, err
	}
	call.Response = newRecordingAIMessages(res.Message)
	return res, nil
}

// newRecordingAIMessages creates new recordingAIMessages from a CompletionMessage.
func newRecordingAIMessages(msg *aiv1.CompletionMessage) []recordingAIMessage {
	var res []recordingAIMessage
	for _, b := range msg.Content {
		resMsg := recordingAIMessage{
			Role: msg.Role,
		}
		switch b := b.BlockType.(type) {
		case *aiv1.ContentBlock_Text:
			resMsg.ContentType = "text"
			resMsg.Content = b.Text
		case *aiv1.ContentBlock_ToolCall:
			resMsg.ContentType = "tool_call"
			resMsg.ID = b.ToolCall.Id
			resMsg.ToolName = b.ToolCall.Name
			data, _ := json.Marshal(b.ToolCall.Input.AsMap())
			resMsg.Content = string(data)
		case *aiv1.ContentBlock_ToolResult:
			resMsg.ContentType = "tool_response"
			resMsg.ID = b.ToolResult.Id
			resMsg.IsError = b.ToolResult.IsError
			resMsg.Content = b.ToolResult.Content
		default:
			resMsg.ContentType = "unknown"
			resMsg.Content = ""
		}
		res = append(res, resMsg)
	}

	return res
}
