package server_test

import (
	"context"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/pkg/ratelimit"
	"github.com/rilldata/rill/runtime/server"
	"github.com/rilldata/rill/runtime/server/auth"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/rilldata/rill/runtime/testruntime/testmode"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestConversations(t *testing.T) {
	// Skip in CI since we make real LLM calls.
	testmode.Expensive(t)

	// Setup test runtime and server with an LLM configured.
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		TestConnectors: []string{"openai"},
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
	require.Len(t, res2.Messages, 6)

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

	// Check that an anonymous user cannot list the conversations
	ctx = auth.WithClaims(t.Context(), &runtime.SecurityClaims{
		UserID:      "",
		Permissions: []runtime.Permission{runtime.ReadObjects, runtime.ReadMetrics, runtime.UseAI}, // Sufficient for analyst_agent, excludes developer agents
	})
	list3, err := srv.ListConversations(ctx, &runtimev1.ListConversationsRequest{
		InstanceId: instanceID,
	})
	require.NoError(t, err)
	require.Len(t, list3.Conversations, 0)

	// Check user agent pattern filtering works correctly
	ctx = auth.WithClaims(t.Context(), &runtime.SecurityClaims{
		UserID:      "foo",
		Permissions: []runtime.Permission{runtime.ReadObjects, runtime.ReadMetrics, runtime.UseAI},
	})

	// Filter for "rill" conversations only (prefix match)
	list4, err := srv.ListConversations(ctx, &runtimev1.ListConversationsRequest{
		InstanceId:       instanceID,
		UserAgentPattern: "rill%",
	})
	require.NoError(t, err)
	require.Len(t, list4.Conversations, 2)

	// Filter for "mcp" conversations (should be none since all conversations are rill)
	list5, err := srv.ListConversations(ctx, &runtimev1.ListConversationsRequest{
		InstanceId:       instanceID,
		UserAgentPattern: "mcp%",
	})
	require.NoError(t, err)
	require.Len(t, list5.Conversations, 0)

	// Filter for specific version (should match both)
	list6, err := srv.ListConversations(ctx, &runtimev1.ListConversationsRequest{
		InstanceId:       instanceID,
		UserAgentPattern: "rill/%",
	})
	require.NoError(t, err)
	require.Len(t, list6.Conversations, 2)

	// No filter returns all conversations
	list7, err := srv.ListConversations(ctx, &runtimev1.ListConversationsRequest{
		InstanceId: instanceID,
	})
	require.NoError(t, err)
	require.Len(t, list7.Conversations, 2)
}
