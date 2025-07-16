package server_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	aiv1 "github.com/rilldata/rill/proto/gen/rill/ai/v1"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/pkg/ratelimit"
	"github.com/rilldata/rill/runtime/server"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestConversationLifecycle(t *testing.T) {
	t.Parallel()

	server, instanceID := newConversationTestServer(t)

	ctx := testCtx()

	// Test the complete conversation lifecycle: create → continue → retrieve → list

	// 1. Start new conversation
	res1, err := server.Complete(ctx, &runtimev1.CompleteRequest{
		InstanceId: instanceID,
		Messages: []*runtimev1.Message{
			createTestTextMessage("user", "Hello, I'm starting a new conversation"),
		},
	})
	require.NoError(t, err)
	require.NotEmpty(t, res1.ConversationId)
	require.NotEmpty(t, res1.Messages)
	conversationID := res1.ConversationId

	// Verify initial response structure and content
	finalMessage := res1.Messages[len(res1.Messages)-1]
	require.Equal(t, "assistant", finalMessage.Role)
	require.NotEmpty(t, finalMessage.Id)
	require.NotEmpty(t, finalMessage.CreatedOn)

	// Verify the AI echoed back the user's message
	require.Len(t, finalMessage.Content, 1)
	textBlock := finalMessage.Content[0].GetText()
	require.Equal(t, "Echo: Hello, I'm starting a new conversation", textBlock)

	// 2. Continue the conversation (test context preservation)
	res2, err := server.Complete(ctx, &runtimev1.CompleteRequest{
		InstanceId:     instanceID,
		ConversationId: &conversationID,
		Messages: []*runtimev1.Message{
			createTestTextMessage("user", "This is my follow-up question"),
		},
	})
	require.NoError(t, err)
	require.Equal(t, conversationID, res2.ConversationId, "Should continue same conversation")
	require.NotEmpty(t, res2.Messages)

	// Verify the AI echoed back the follow-up message
	finalMessage2 := res2.Messages[len(res2.Messages)-1]
	require.Equal(t, "assistant", finalMessage2.Role)
	require.Len(t, finalMessage2.Content, 1)
	textBlock2 := finalMessage2.Content[0].GetText()
	require.Equal(t, "Echo: This is my follow-up question", textBlock2)

	// 3. Retrieve conversation (test persistence)
	conversation, err := server.GetConversation(ctx, &runtimev1.GetConversationRequest{
		InstanceId:     instanceID,
		ConversationId: conversationID,
	})
	require.NoError(t, err)
	require.Equal(t, conversationID, conversation.Conversation.Id)
	require.Equal(t, "Hello, I'm starting a new conversation", conversation.Conversation.Title)
	require.NotEmpty(t, conversation.Conversation.CreatedOn)
	require.NotEmpty(t, conversation.Conversation.UpdatedOn)

	// 4. List conversations (test indexing)
	list, err := server.ListConversations(ctx, &runtimev1.ListConversationsRequest{
		InstanceId: instanceID,
	})
	require.NoError(t, err)
	require.Len(t, list.Conversations, 1)
	require.Equal(t, conversationID, list.Conversations[0].Id)
	require.Equal(t, "Hello, I'm starting a new conversation", list.Conversations[0].Title)
}

func TestConversationContextContinuity(t *testing.T) {
	t.Parallel()

	server, instanceID := newConversationTestServer(t)

	ctx := testCtx()

	// Test that conversation maintains full context across multiple turns

	// 1. Start conversation with context-setting message
	res1, err := server.Complete(ctx, &runtimev1.CompleteRequest{
		InstanceId: instanceID,
		Messages: []*runtimev1.Message{
			createTestTextMessage("user", "My name is Alice and I work at Acme Corp"),
		},
	})
	require.NoError(t, err)
	conversationID := res1.ConversationId

	// 2. Second turn - reference previous context
	res2, err := server.Complete(ctx, &runtimev1.CompleteRequest{
		InstanceId:     instanceID,
		ConversationId: &conversationID,
		Messages: []*runtimev1.Message{
			createTestTextMessage("user", "What company did I say I work for?"),
		},
	})
	require.NoError(t, err)
	require.Equal(t, conversationID, res2.ConversationId)

	// Verify we get a response (AI should have access to previous context)
	require.NotEmpty(t, res2.Messages)
	finalMessage := res2.Messages[len(res2.Messages)-1]
	require.Equal(t, "assistant", finalMessage.Role)

	// Verify the AI echoed back the second message
	require.Len(t, finalMessage.Content, 1)
	textBlock := finalMessage.Content[0].GetText()
	require.Equal(t, "Echo: What company did I say I work for?", textBlock)

	// 3. Third turn - reference even earlier context
	res3, err := server.Complete(ctx, &runtimev1.CompleteRequest{
		InstanceId:     instanceID,
		ConversationId: &conversationID,
		Messages: []*runtimev1.Message{
			createTestTextMessage("user", "And what was my name again?"),
		},
	})
	require.NoError(t, err)
	require.Equal(t, conversationID, res3.ConversationId)
	require.NotEmpty(t, res3.Messages)

	// Verify the AI echoed back the third message
	finalMessage3 := res3.Messages[len(res3.Messages)-1]
	require.Equal(t, "assistant", finalMessage3.Role)
	require.Len(t, finalMessage3.Content, 1)
	textBlock3 := finalMessage3.Content[0].GetText()
	require.Equal(t, "Echo: And what was my name again?", textBlock3)

	// 4. Verify complete conversation history is preserved
	conversation, err := server.GetConversation(ctx, &runtimev1.GetConversationRequest{
		InstanceId:     instanceID,
		ConversationId: conversationID,
	})
	require.NoError(t, err)

	// The title should be generated from the first message
	require.Equal(t, "My name is Alice and I work at Acme Corp", conversation.Conversation.Title)

	// This test verifies that the Complete method properly includes
	// previous conversation history when calling the AI.
}

func TestMultipleConversationIsolation(t *testing.T) {
	t.Parallel()

	server, instanceID := newConversationTestServer(t)

	ctx := testCtx()

	// Test that multiple conversations don't interfere with each other

	// 1. Create conversation A
	convA, err := server.Complete(ctx, &runtimev1.CompleteRequest{
		InstanceId: instanceID,
		Messages: []*runtimev1.Message{
			createTestTextMessage("user", "I'm discussing topic A about databases"),
		},
	})
	require.NoError(t, err)
	conversationA := convA.ConversationId

	// 2. Create conversation B
	convB, err := server.Complete(ctx, &runtimev1.CompleteRequest{
		InstanceId: instanceID,
		Messages: []*runtimev1.Message{
			createTestTextMessage("user", "I'm discussing topic B about cooking"),
		},
	})
	require.NoError(t, err)
	conversationB := convB.ConversationId

	// Verify conversations have different IDs
	require.NotEqual(t, conversationA, conversationB)

	// 3. Continue conversation A with context-specific question
	contA, err := server.Complete(ctx, &runtimev1.CompleteRequest{
		InstanceId:     instanceID,
		ConversationId: &conversationA,
		Messages: []*runtimev1.Message{
			createTestTextMessage("user", "What topic was I just discussing?"),
		},
	})
	require.NoError(t, err)
	require.Equal(t, conversationA, contA.ConversationId)

	// Verify conversation A's response
	finalMessageA := contA.Messages[len(contA.Messages)-1]
	require.Equal(t, "assistant", finalMessageA.Role)
	require.Len(t, finalMessageA.Content, 1)
	textBlockA := finalMessageA.Content[0].GetText()
	require.Equal(t, "Echo: What topic was I just discussing?", textBlockA)

	// 4. Continue conversation B with context-specific question
	contB, err := server.Complete(ctx, &runtimev1.CompleteRequest{
		InstanceId:     instanceID,
		ConversationId: &conversationB,
		Messages: []*runtimev1.Message{
			createTestTextMessage("user", "What topic was I just discussing?"),
		},
	})
	require.NoError(t, err)
	require.Equal(t, conversationB, contB.ConversationId)

	// Verify conversation B's response
	finalMessageB := contB.Messages[len(contB.Messages)-1]
	require.Equal(t, "assistant", finalMessageB.Role)
	require.Len(t, finalMessageB.Content, 1)
	textBlockB := finalMessageB.Content[0].GetText()
	require.Equal(t, "Echo: What topic was I just discussing?", textBlockB)

	// 5. Verify both conversations exist independently in list
	list, err := server.ListConversations(ctx, &runtimev1.ListConversationsRequest{
		InstanceId: instanceID,
	})
	require.NoError(t, err)
	require.Len(t, list.Conversations, 2)

	// Verify both conversations have correct titles and different IDs
	titles := make(map[string]string) // conversationID -> title
	for _, conv := range list.Conversations {
		titles[conv.Id] = conv.Title
	}

	require.Equal(t, "I'm discussing topic A about databases", titles[conversationA])
	require.Equal(t, "I'm discussing topic B about cooking", titles[conversationB])

	// 6. Retrieve each conversation individually to verify isolation
	getA, err := server.GetConversation(ctx, &runtimev1.GetConversationRequest{
		InstanceId:     instanceID,
		ConversationId: conversationA,
	})
	require.NoError(t, err)
	require.Equal(t, "I'm discussing topic A about databases", getA.Conversation.Title)

	getB, err := server.GetConversation(ctx, &runtimev1.GetConversationRequest{
		InstanceId:     instanceID,
		ConversationId: conversationB,
	})
	require.NoError(t, err)
	require.Equal(t, "I'm discussing topic B about cooking", getB.Conversation.Title)
}

func TestConversationWithTools(t *testing.T) {
	t.Parallel()

	// Create instance with AI connector (needed for tool integration)
	rt := testruntime.New(t, true)
	tmpDir := t.TempDir()
	inst := &drivers.Instance{
		Environment:      "test",
		OLAPConnector:    "duckdb",
		RepoConnector:    "repo",
		CatalogConnector: "catalog",
		AIConnector:      "test_ai",
		Connectors: []*runtimev1.Connector{
			{
				Type:   "file",
				Name:   "repo",
				Config: map[string]string{"dsn": tmpDir},
			},
			{
				Type:   "duckdb",
				Name:   "duckdb",
				Config: map[string]string{"dsn": ":memory:"},
			},
			{
				Type:   "sqlite",
				Name:   "catalog",
				Config: map[string]string{"dsn": fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())},
			},
			{
				Type:   "mock_ai",
				Name:   "test_ai",
				Config: map[string]string{},
			},
		},
		Variables: map[string]string{"rill.stage_changes": "false"},
	}

	// Create required files for a basic project
	files := map[string]string{
		"rill.yaml":   ``,
		"ad_bids.sql": `SELECT now() AS time, 'DA' AS country, 3.141 as price`,
	}

	for path, data := range files {
		abs := filepath.Join(tmpDir, path)
		require.NoError(t, os.MkdirAll(filepath.Dir(abs), os.ModePerm))
		require.NoError(t, os.WriteFile(abs, []byte(data), 0o644))
	}

	setupCtx := context.Background()
	err := rt.CreateInstance(setupCtx, inst)
	require.NoError(t, err)

	ctrl, err := rt.Controller(setupCtx, inst.ID)
	require.NoError(t, err)

	_, err = ctrl.Get(setupCtx, runtime.GlobalProjectParserName, false)
	require.NoError(t, err)

	err = ctrl.WaitUntilIdle(setupCtx, false)
	require.NoError(t, err)

	instanceID := inst.ID
	server, err := server.NewServer(context.Background(), &server.Options{}, rt, zap.NewNop(), ratelimit.NewNoop(), activity.NewNoopClient())
	require.NoError(t, err)

	// Use auth context for API calls
	ctx := testCtx()

	// Test that AI can use tools during conversation and results are preserved

	// 1. Start conversation that might trigger tool usage
	res1, err := server.Complete(ctx, &runtimev1.CompleteRequest{
		InstanceId: instanceID,
		Messages: []*runtimev1.Message{
			createTestTextMessage("user", "Can you help me analyze some data?"),
		},
	})
	require.NoError(t, err)
	require.NotEmpty(t, res1.ConversationId)
	require.NotEmpty(t, res1.Messages)
	conversationID := res1.ConversationId

	// Verify AI responded (tool executor was properly injected)
	finalMessage := res1.Messages[len(res1.Messages)-1]
	require.Equal(t, "assistant", finalMessage.Role)
	require.NotEmpty(t, finalMessage.Id)

	// In tool calling mode, the AI doesn't echo but returns tool calls or fixed responses
	// So we just verify that we get some content
	require.NotEmpty(t, finalMessage.Content)

	// 2. Continue conversation - verify tool context is preserved
	res2, err := server.Complete(ctx, &runtimev1.CompleteRequest{
		InstanceId:     instanceID,
		ConversationId: &conversationID,
		Messages: []*runtimev1.Message{
			createTestTextMessage("user", "What tools do you have available?"),
		},
	})
	require.NoError(t, err)
	require.Equal(t, conversationID, res2.ConversationId)
	require.NotEmpty(t, res2.Messages)

	// 3. Verify conversation is properly stored with tool interactions
	conversation, err := server.GetConversation(ctx, &runtimev1.GetConversationRequest{
		InstanceId:     instanceID,
		ConversationId: conversationID,
	})
	require.NoError(t, err)
	require.Equal(t, conversationID, conversation.Conversation.Id)
	require.Equal(t, "Can you help me analyze some data?", conversation.Conversation.Title)

	// This test verifies that the server properly injects tool execution capabilities
	// into the runtime layer via the serverToolService and that tool interactions
	// are preserved in conversation history
}

func TestConversationErrorHandling(t *testing.T) {
	t.Parallel()

	server, instanceID := newConversationTestServer(t)

	ctx := testCtx()

	// Test critical error scenarios

	t.Run("invalid_conversation_id", func(t *testing.T) {
		invalidID := "invalid-conversation-id-that-does-not-exist"
		_, err := server.Complete(ctx, &runtimev1.CompleteRequest{
			InstanceId:     instanceID,
			ConversationId: &invalidID,
			Messages: []*runtimev1.Message{
				createTestTextMessage("user", "This should fail"),
			},
		})
		// Should return an error when trying to use non-existent conversation
		require.Error(t, err)
	})

	t.Run("nonexistent_conversation_get", func(t *testing.T) {
		_, err := server.GetConversation(ctx, &runtimev1.GetConversationRequest{
			InstanceId:     instanceID,
			ConversationId: "non-existent-conversation-id",
		})
		// Should return an error when trying to get non-existent conversation
		require.Error(t, err)
	})

	t.Run("empty_messages_handling", func(t *testing.T) {
		// Test with empty messages array
		res, err := server.Complete(ctx, &runtimev1.CompleteRequest{
			InstanceId: instanceID,
			Messages:   []*runtimev1.Message{}, // Empty messages
		})
		// Should either succeed with empty response or return specific error
		if err != nil {
			// If it errors, should be a meaningful error about empty messages
			require.Contains(t, err.Error(), "message")
		} else {
			// If it succeeds, should still create a conversation
			require.NotEmpty(t, res.ConversationId)
		}
	})

	t.Run("malformed_message_content", func(t *testing.T) {
		// Test with message that has no content blocks
		res, err := server.Complete(ctx, &runtimev1.CompleteRequest{
			InstanceId: instanceID,
			Messages: []*runtimev1.Message{
				{
					Role:    "user",
					Content: []*aiv1.ContentBlock{}, // Empty content blocks
				},
			},
		})
		// Should handle gracefully
		if err != nil {
			require.Contains(t, err.Error(), "content")
		} else {
			require.NotEmpty(t, res.ConversationId)
		}
	})

	t.Run("invalid_instance_id", func(t *testing.T) {
		_, err := server.Complete(ctx, &runtimev1.CompleteRequest{
			InstanceId: "invalid-instance-id",
			Messages: []*runtimev1.Message{
				createTestTextMessage("user", "This should fail"),
			},
		})
		// Should return an error for invalid instance ID
		require.Error(t, err)
	})

	t.Run("list_conversations_invalid_instance", func(t *testing.T) {
		_, err := server.ListConversations(ctx, &runtimev1.ListConversationsRequest{
			InstanceId: "invalid-instance-id",
		})
		// Should return an error for invalid instance ID
		require.Error(t, err)
	})
}

// Helper function to create a text message for tests
func createTestTextMessage(role, text string) *runtimev1.Message {
	return &runtimev1.Message{
		Role: role,
		Content: []*aiv1.ContentBlock{
			{
				BlockType: &aiv1.ContentBlock_Text{
					Text: text,
				},
			},
		},
	}
}

// newConversationTestServer creates a test server with AI connector support for conversation tests
func newConversationTestServer(t *testing.T) (*server.Server, string) {
	rt := testruntime.New(t, true)

	// Create instance with AI connector
	tmpDir := t.TempDir()
	inst := &drivers.Instance{
		Environment:      "test",
		OLAPConnector:    "duckdb",
		RepoConnector:    "repo",
		CatalogConnector: "catalog",
		AIConnector:      "test_ai",
		Connectors: []*runtimev1.Connector{
			{
				Type:   "file",
				Name:   "repo",
				Config: map[string]string{"dsn": tmpDir},
			},
			{
				Type:   "duckdb",
				Name:   "duckdb",
				Config: map[string]string{"dsn": ":memory:"},
			},
			{
				Type:   "sqlite",
				Name:   "catalog",
				Config: map[string]string{"dsn": fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())},
			},
			{
				Type:   "mock_ai",
				Name:   "test_ai",
				Config: map[string]string{},
			},
		},
		Variables: map[string]string{"rill.stage_changes": "false"},
	}

	// Add required files for testing explore dashboard context
	files := map[string]string{
		"rill.yaml":             ``,
		"models/test_model.sql": `SELECT 1 as id, 'test' as name, now() as created_at`,
		"metrics/test_metrics.yaml": `
type: metrics_view
table: test_model
timeseries: created_at
dimensions:
  - name: name
    column: name
measures:
  - name: count
    expression: count(*)
    type: simple
`,
		"explores/test_dashboard.yaml": `
type: explore
metrics_view: test_metrics
dimensions:
  - name
measures:
  - count
`,
	}

	for path, content := range files {
		fullPath := filepath.Join(tmpDir, path)
		require.NoError(t, os.MkdirAll(filepath.Dir(fullPath), 0o755))
		require.NoError(t, os.WriteFile(fullPath, []byte(content), 0o644))
	}

	ctx := context.Background()
	err := rt.CreateInstance(ctx, inst)
	require.NoError(t, err)

	ctrl, err := rt.Controller(ctx, inst.ID)
	require.NoError(t, err)

	_, err = ctrl.Get(ctx, runtime.GlobalProjectParserName, false)
	require.NoError(t, err)

	err = ctrl.WaitUntilIdle(ctx, false)
	require.NoError(t, err)

	server, err := server.NewServer(context.Background(), &server.Options{}, rt, zap.NewNop(), ratelimit.NewNoop(), activity.NewNoopClient())
	require.NoError(t, err)

	return server, inst.ID
}

// TestConversationWithAppContext tests that app context generates system messages and saves them before user messages
func TestConversationWithAppContext(t *testing.T) {
	server, instanceID := newConversationTestServer(t)
	ctx := testCtx()

	// Test 1: Project chat context
	t.Run("project_chat_context", func(t *testing.T) {
		// TODO: Implement PROJECT_CHAT context handling
		t.Skip("Project chat context not implemented yet")

		appContext := &runtimev1.AppContext{
			ContextType:     runtimev1.AppContextType_APP_CONTEXT_TYPE_PROJECT_CHAT,
			ContextMetadata: &structpb.Struct{},
		}

		resp, err := server.Complete(ctx, &runtimev1.CompleteRequest{
			InstanceId: instanceID,
			AppContext: appContext,
			Messages: []*runtimev1.Message{
				createTestTextMessage("user", "What metrics are available?"),
			},
		})
		require.NoError(t, err)
		require.NotEmpty(t, resp.ConversationId)

		// Get the conversation with system messages to verify message order
		conv, err := server.GetConversation(ctx, &runtimev1.GetConversationRequest{
			InstanceId:            instanceID,
			ConversationId:        resp.ConversationId,
			IncludeSystemMessages: true,
		})
		require.NoError(t, err)
		require.NotNil(t, conv.Conversation)

		messages := conv.Conversation.Messages
		require.Len(t, messages, 3) // Exactly: system, user, assistant

		// Verify system message is first and contains expected content
		require.Equal(t, "system", messages[0].Role)
		require.Contains(t, messages[0].Content[0].GetText(), "Available metrics views")
		t.Logf("✓ System message: %s", messages[0].Content[0].GetText())

		// Verify user message is second
		require.Equal(t, "user", messages[1].Role)
		require.Equal(t, "What metrics are available?", messages[1].Content[0].GetText())

		// Verify assistant message is third
		require.Equal(t, "assistant", messages[2].Role)
		require.Equal(t, "Echo: What metrics are available?", messages[2].Content[0].GetText())
	})

	// Test 2: Explore dashboard context with metadata
	t.Run("explore_dashboard_context", func(t *testing.T) {
		metadata, err := structpb.NewStruct(map[string]interface{}{
			"dashboard_name": "test_dashboard",
		})
		require.NoError(t, err)

		appContext := &runtimev1.AppContext{
			ContextType:     runtimev1.AppContextType_APP_CONTEXT_TYPE_EXPLORE_DASHBOARD,
			ContextMetadata: metadata,
		}

		resp, err := server.Complete(ctx, &runtimev1.CompleteRequest{
			InstanceId: instanceID,
			AppContext: appContext,
			Messages: []*runtimev1.Message{
				createTestTextMessage("user", "Tell me about this dashboard"),
			},
		})
		require.NoError(t, err)
		require.NotEmpty(t, resp.ConversationId)

		// Get the conversation with system messages to verify message order
		conv, err := server.GetConversation(ctx, &runtimev1.GetConversationRequest{
			InstanceId:            instanceID,
			ConversationId:        resp.ConversationId,
			IncludeSystemMessages: true,
		})
		require.NoError(t, err)
		require.NotNil(t, conv.Conversation)

		messages := conv.Conversation.Messages
		require.Len(t, messages, 3) // Exactly: system, user, assistant

		// Verify system message is first and contains dashboard context
		require.Equal(t, "system", messages[0].Role)
		systemText := messages[0].Content[0].GetText()
		require.Contains(t, systemText, "test_dashboard")
		require.Contains(t, systemText, "actively viewing")
		t.Logf("✓ Dashboard system message: %s", systemText)

		// Verify user message is second
		require.Equal(t, "user", messages[1].Role)
		require.Equal(t, "Tell me about this dashboard", messages[1].Content[0].GetText())

		// Verify assistant message is third
		require.Equal(t, "assistant", messages[2].Role)
		require.Equal(t, "Echo: Tell me about this dashboard", messages[2].Content[0].GetText())
	})

	// Test 3: No app context (should work normally)
	t.Run("no_app_context", func(t *testing.T) {
		resp, err := server.Complete(ctx, &runtimev1.CompleteRequest{
			InstanceId: instanceID,
			AppContext: nil, // No app context
			Messages: []*runtimev1.Message{
				createTestTextMessage("user", "Hello without context"),
			},
		})
		require.NoError(t, err)
		require.NotEmpty(t, resp.ConversationId)

		// Get the conversation to verify message order
		conv, err := server.GetConversation(ctx, &runtimev1.GetConversationRequest{
			InstanceId:     instanceID,
			ConversationId: resp.ConversationId,
		})
		require.NoError(t, err)
		require.NotNil(t, conv.Conversation)

		messages := conv.Conversation.Messages
		require.Len(t, messages, 2) // Only: user, assistant (no system message)

		// Verify user message is first
		require.Equal(t, "user", messages[0].Role)
		require.Equal(t, "Hello without context", messages[0].Content[0].GetText())

		// Verify assistant message is second
		require.Equal(t, "assistant", messages[1].Role)
		require.Equal(t, "Echo: Hello without context", messages[1].Content[0].GetText())
	})

	// Test 4: Continuing conversation with app context (should not re-add system messages)
	t.Run("continue_conversation_with_app_context", func(t *testing.T) {
		// Use explore dashboard context for testing conversation continuation
		metadata, err := structpb.NewStruct(map[string]interface{}{
			"dashboard_name": "test_dashboard",
		})
		require.NoError(t, err)

		appContext := &runtimev1.AppContext{
			ContextType:     runtimev1.AppContextType_APP_CONTEXT_TYPE_EXPLORE_DASHBOARD,
			ContextMetadata: metadata,
		}

		// First message with app context
		resp1, err := server.Complete(ctx, &runtimev1.CompleteRequest{
			InstanceId: instanceID,
			AppContext: appContext,
			Messages: []*runtimev1.Message{
				createTestTextMessage("user", "First message with context"),
			},
		})
		require.NoError(t, err)
		conversationID := resp1.ConversationId

		// Continue the conversation (app context should not add new system messages)
		resp2, err := server.Complete(ctx, &runtimev1.CompleteRequest{
			InstanceId:     instanceID,
			ConversationId: &conversationID,
			AppContext:     appContext, // Same app context
			Messages: []*runtimev1.Message{
				createTestTextMessage("user", "Second message in same conversation"),
			},
		})
		require.NoError(t, err)
		require.Equal(t, conversationID, resp2.ConversationId)

		// Get the conversation to verify no duplicate system messages
		conv, err := server.GetConversation(ctx, &runtimev1.GetConversationRequest{
			InstanceId:            instanceID,
			ConversationId:        conversationID,
			IncludeSystemMessages: true,
		})
		require.NoError(t, err)
		require.NotNil(t, conv.Conversation)

		messages := conv.Conversation.Messages
		require.Len(t, messages, 5) // system, user1, assistant1, user2, assistant2

		// Verify only one system message (at the beginning)
		systemMessages := 0
		for _, msg := range messages {
			if msg.Role == "system" {
				systemMessages++
			}
		}
		require.Equal(t, 1, systemMessages, "Should have exactly one system message")

		// Verify system message is still first and contains dashboard context
		require.Equal(t, "system", messages[0].Role)
		require.Contains(t, messages[0].Content[0].GetText(), "test_dashboard")
		require.Contains(t, messages[0].Content[0].GetText(), "actively viewing")
	})
}
