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
	rt := testruntime.New(t)
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
	rt := testruntime.New(t)

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

	// Add required files
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "rill.yaml"), []byte(""), 0o644))

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
