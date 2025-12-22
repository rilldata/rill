package server_test

import (
	"context"
	"testing"
	"time"

	aiv1 "github.com/rilldata/rill/proto/gen/rill/ai/v1"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/ai"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/pkg/ratelimit"
	"github.com/rilldata/rill/runtime/server"
	"github.com/rilldata/rill/runtime/server/auth"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/rilldata/rill/runtime/testruntime/testmode"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestConversations(t *testing.T) {
	// Skip in CI since we make real LLM calls.
	testmode.Expensive(t)

	// Setup test runtime and server with an LLM configured.
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		EnableLLM: true,
		Files: map[string]string{
			"models/orders.yaml": `
type: model
materialize: true
sql: |
  SELECT '2025-01-01T00:00:00Z'::TIMESTAMP AS event_time, 'United States' AS country, 100 AS revenue
  UNION ALL
  SELECT '2025-01-01T00:00:00Z'::TIMESTAMP AS event_time, 'Denmark' AS country, 10 AS revenue
  UNION ALL
  SELECT '2025-01-02T00:00:00Z'::TIMESTAMP AS event_time, 'United States' AS country, 100 AS revenue
  UNION ALL
  SELECT '2025-01-02T00:00:00Z'::TIMESTAMP AS event_time, 'Denmark' AS country, 10 AS revenue
`,
			"metrics/orders.yaml": `
type: metrics_view
model: orders
timeseries: event_time
dimensions:
- column: country
measures:
- name: count
  expression: COUNT(*)
- name: revenue
  expression: SUM(revenue)
`,
		},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 4, 0, 0)

	// Create test server
	srv, err := server.NewServer(context.Background(), &server.Options{}, rt, zap.NewNop(), ratelimit.NewNoop(), activity.NewNoopClient())
	require.NoError(t, err)

	// Create test context with claims (to test conversation listings, which filter by user ID)
	ctx := auth.WithClaims(t.Context(), &runtime.SecurityClaims{
		UserID:      "foo",
		Permissions: []runtime.Permission{runtime.ReadObjects, runtime.ReadMetrics, runtime.UseAI}, // Sufficient for analyst_agent, excludes developer agents
	})

	// Ask a question
	res1, err := srv.Complete(ctx, &runtimev1.CompleteRequest{
		InstanceId: instanceID,
		Prompt:     "What are the names of the available metrics views?",
	})
	require.NoError(t, err)
	require.NotEmpty(t, res1.ConversationId)
	require.NotEmpty(t, res1.Messages)
	require.Len(t, res1.Messages, 6)

	// Ask another question in the same conversation
	res2, err := srv.Complete(ctx, &runtimev1.CompleteRequest{
		InstanceId:     instanceID,
		ConversationId: res1.ConversationId,
		Prompt:         "Can you simply repeat your previous answer? Don't make any tool calls.",
	})
	require.NoError(t, err)
	require.Equal(t, res2.ConversationId, res1.ConversationId)
	require.NotEmpty(t, res2.Messages)
	require.Len(t, res2.Messages, 4)

	// Ask a question in a new conversation
	res3, err := srv.Complete(ctx, &runtimev1.CompleteRequest{
		InstanceId: instanceID,
		Prompt:     "What are the names of the available metrics views?",
	})
	require.NoError(t, err)
	require.NotEmpty(t, res3.ConversationId)
	require.NotEqual(t, res3.ConversationId, res1.ConversationId)
	require.NotEmpty(t, res3.Messages)

	// Check it persisted the messages in the first conversation
	get1, err := srv.GetConversation(ctx, &runtimev1.GetConversationRequest{
		InstanceId:     instanceID,
		ConversationId: res1.ConversationId,
	})
	require.NoError(t, err)
	require.Len(t, get1.Messages, len(res1.Messages)+len(res2.Messages))

	// Check it persisted the messages in the second conversation
	get2, err := srv.GetConversation(ctx, &runtimev1.GetConversationRequest{
		InstanceId:     instanceID,
		ConversationId: res3.ConversationId,
	})
	require.NoError(t, err)
	require.Len(t, get2.Messages, len(res3.Messages))

	// Check it lists the conversations
	list1, err := srv.ListConversations(ctx, &runtimev1.ListConversationsRequest{
		InstanceId: instanceID,
	})
	require.NoError(t, err)
	require.Len(t, list1.Conversations, 2)

	// Check user agent pattern filtering works correctly.
	// Filter for "rill" conversations only (prefix match).
	list4, err := srv.ListConversations(ctx, &runtimev1.ListConversationsRequest{
		InstanceId:       instanceID,
		UserAgentPattern: "rill/%",
	})
	require.NoError(t, err)
	require.Len(t, list4.Conversations, 2)

	// Filter for "mcp" conversations (should be none since all conversations are rill).
	list5, err := srv.ListConversations(ctx, &runtimev1.ListConversationsRequest{
		InstanceId:       instanceID,
		UserAgentPattern: "mcp%",
	})
	require.NoError(t, err)
	require.Len(t, list5.Conversations, 0)

	// Check it errors if completing a conversation that doesn't exist
	_, err = srv.Complete(ctx, &runtimev1.CompleteRequest{
		InstanceId:     instanceID,
		ConversationId: "doesntexist",
		Prompt:         "What is 2 + 2?",
	})
	require.ErrorContains(t, err, "failed to find")

	// Check that another user cannot list the conversations
	ctx = auth.WithClaims(t.Context(), &runtime.SecurityClaims{
		UserID:      "bar",
		Permissions: []runtime.Permission{runtime.ReadObjects, runtime.ReadMetrics, runtime.UseAI}, // Sufficient for analyst_agent, excludes developer agents
	})
	list2, err := srv.ListConversations(ctx, &runtimev1.ListConversationsRequest{
		InstanceId: instanceID,
	})
	require.NoError(t, err)
	require.Len(t, list2.Conversations, 0)

	// Check that an anonymous user can create a conversation
	ctx = auth.WithClaims(t.Context(), &runtime.SecurityClaims{
		UserID:      "",
		Permissions: []runtime.Permission{runtime.ReadObjects, runtime.ReadMetrics, runtime.UseAI}, // Sufficient for analyst_agent, excludes developer agents
	})
	res4, err := srv.Complete(ctx, &runtimev1.CompleteRequest{
		InstanceId: instanceID,
		Prompt:     "What are the names of the available metrics views?",
	})
	require.NoError(t, err)
	require.NotEmpty(t, res1.ConversationId)
	require.NotEmpty(t, res1.Messages)
	require.Len(t, res1.Messages, 6)

	// Check that an anonymous user cannot list conversations, even their own (since we don't know if it's the same or a different anonymous user).
	list3, err := srv.ListConversations(ctx, &runtimev1.ListConversationsRequest{
		InstanceId: instanceID,
	})
	require.NoError(t, err)
	require.Len(t, list3.Conversations, 0)

	// Check that an anonymous user with SkipChecks can list and get conversations.
	// (This matches Rill Developer behavior where auth is disabled.)
	ctx = auth.WithClaims(t.Context(), &runtime.SecurityClaims{
		UserID:      "",
		Permissions: runtime.AllPermissions,
		SkipChecks:  true,
	})
	list6, err := srv.ListConversations(ctx, &runtimev1.ListConversationsRequest{
		InstanceId: instanceID,
	})
	require.NoError(t, err)
	require.Len(t, list6.Conversations, 1)
	require.Equal(t, res4.ConversationId, list6.Conversations[0].Id)
	get3, err := srv.GetConversation(ctx, &runtimev1.GetConversationRequest{
		InstanceId:     instanceID,
		ConversationId: res4.ConversationId,
	})
	require.NoError(t, err)
	require.Len(t, get3.Messages, len(res4.Messages))
}

func TestAgentChoiceAndContext(t *testing.T) {
	// Skip in CI since we make real LLM calls.
	testmode.Expensive(t)

	// Setup test runtime and server with an LLM configured.
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		EnableLLM: true,
	})
	srv, err := server.NewServer(context.Background(), &server.Options{}, rt, zap.NewNop(), ratelimit.NewNoop(), activity.NewNoopClient())
	require.NoError(t, err)

	// Ask a question for the analyst agent
	res1, err := srv.Complete(testCtx(), &runtimev1.CompleteRequest{
		InstanceId: instanceID,
		Prompt:     "What are the names of the available metrics views?",
		Agent:      ai.AnalystAgentName,
		AnalystAgentContext: &runtimev1.AnalystAgentContext{ // NOTE: This is incoherent, but for this test, we just want to verify that its passed through correctly.
			Explore:    "foo",
			Dimensions: []string{"bar"},
			Measures:   []string{"baz"},
			Where:      &runtimev1.Expression{Expression: &runtimev1.Expression_Ident{Ident: "is_true"}},
			TimeStart:  timestamppb.New(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)),
			TimeEnd:    timestamppb.New(time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)),
		},
	})
	require.NoError(t, err)
	require.NotEmpty(t, res1.ConversationId)
	require.NotEmpty(t, res1.Messages)
	var found bool
	for _, msg := range res1.Messages {
		if msg.Tool == ai.AnalystAgentName {
			found = true
			break
		}
	}
	require.True(t, found)

	// Ask a question for the developer agent
	res2, err := srv.Complete(testCtx(), &runtimev1.CompleteRequest{
		InstanceId: instanceID,
		Prompt:     "Generate a single model that just does `SELECT 1 AS one`.",
		Agent:      ai.DeveloperAgentName,
		DeveloperAgentContext: &runtimev1.DeveloperAgentContext{
			InitProject: true,
		},
	})
	require.NoError(t, err)
	require.NotEmpty(t, res2.ConversationId)
	require.NotEmpty(t, res2.Messages)
	found = false
	for _, msg := range res2.Messages {
		if msg.Tool == ai.DeveloperAgentName {
			found = true
			break
		}
	}
	require.True(t, found)
}

func TestListTools(t *testing.T) {
	// Create test server
	srv, instanceID := getTestServer(t)

	// List tools
	res, err := srv.ListTools(testCtx(), &runtimev1.ListToolsRequest{InstanceId: instanceID})
	require.NoError(t, err)
	require.Greater(t, len(res.Tools), 0)

	// Check tool info is populated
	tools := make(map[string]*aiv1.Tool)
	for _, tool := range res.Tools {
		require.NotEmpty(t, tool.Name)
		require.NotEmpty(t, tool.Description)
		tools[tool.Name] = tool
	}

	// Extra checks for some specific expected tools
	names := []string{
		ai.RouterAgentName,
		ai.AnalystAgentName,
		ai.QueryMetricsViewName,
	}
	for _, name := range names {
		tool, ok := tools[name]
		require.Truef(t, ok, "expected tool %q to be present", name)
		require.NotEmpty(t, tool.Meta.AsMap())
		require.NotEmpty(t, tool.InputSchema)
		require.NotEmpty(t, tool.OutputSchema)
	}
}
