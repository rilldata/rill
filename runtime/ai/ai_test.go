package ai_test

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/google/uuid"
	aiv1 "github.com/rilldata/rill/proto/gen/rill/ai/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/ai"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	_ "github.com/rilldata/rill/runtime/resolvers"
)

// newTest sets up a new runtime instance and AI chat session for it.
func newTest(t *testing.T, opts testruntime.InstanceOptions) (*runtime.Runtime, string, *ai.Session) {
	// Create test runtime instance
	rt, instanceID := testruntime.NewInstanceWithOptions(t, opts)

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

	return rt, instanceID, s
}

// newEval sets up a new eval and returns a runtime instance and AI chat session for it.
// It differs from newTest in that a) it's disabled in CI/short mode, and b) the messages and LLM calls are recorded to ./evals for later inspection.
func newEval(t *testing.T, opts testruntime.InstanceOptions) (*runtime.Runtime, string, *ai.Session) {
	// Skip AI tests in short mode since they're comparatively expensive.
	if testing.Short() {
		t.SkipNow()
	}

	// Add openai to the test connectors if not already present.
	// This enables LLM completions.
	if !slices.Contains(opts.TestConnectors, "openai") {
		opts.TestConnectors = append(opts.TestConnectors, "openai")
	}

	// Create test runtime instance and AI session
	rt, instanceID, s := newTest(t, opts)

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

	return rt, instanceID, s
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
	Response recordingAIMessage   `yaml:"response,omitempty"`
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
func (r *recordingAIService) Complete(ctx context.Context, msgs []*aiv1.CompletionMessage, tools []*aiv1.Tool, outputSchema *jsonschema.Schema) (*aiv1.CompletionMessage, error) {
	// Create a recorded call
	call := &recordingAICall{Index: len(r.calls) + 1}
	for _, m := range msgs {
		call.Input = append(call.Input, newRecordingAIMessage(m))
	}
	r.calls = append(r.calls, call)

	// Forward to the underlying AI service
	res, err := r.ai.Complete(ctx, msgs, tools, outputSchema)
	if err != nil {
		call.Error = err.Error()
		return nil, err
	}
	call.Response = newRecordingAIMessage(res)
	return res, nil
}

// newRecordingAIMessage creates a new recordingAIMessage from a CompletionMessage.
func newRecordingAIMessage(msg *aiv1.CompletionMessage) recordingAIMessage {
	res := recordingAIMessage{
		Role: msg.Role,
	}
	switch b := msg.Content[0].BlockType.(type) {
	case *aiv1.ContentBlock_Text:
		res.ContentType = "text"
		res.Content = b.Text
	case *aiv1.ContentBlock_ToolCall:
		res.ContentType = "tool_call"
		res.ID = b.ToolCall.Id
		res.ToolName = b.ToolCall.Name
		data, _ := json.Marshal(b.ToolCall.Input.AsMap())
		res.Content = string(data)
	case *aiv1.ContentBlock_ToolResult:
		res.ContentType = "tool_response"
		res.ID = b.ToolResult.Id
		res.IsError = b.ToolResult.IsError
		res.Content = b.ToolResult.Content
	default:
		res.ContentType = "unknown"
		res.Content = ""
	}
	return res
}
