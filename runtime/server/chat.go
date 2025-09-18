package server

import (
	"context"
	"errors"
	"time"

	aiv1 "github.com/rilldata/rill/proto/gen/rill/ai/v1"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/ai"
	"github.com/rilldata/rill/runtime/server/auth"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ListConversations returns a list of conversations for an instance.
func (s *Server) ListConversations(ctx context.Context, req *runtimev1.ListConversationsRequest) (*runtimev1.ListConversationsResponse, error) {
	if !auth.GetClaims(ctx).CanInstance(req.InstanceId, auth.UseAI) {
		return nil, ErrForbidden
	}

	ownerID := auth.GetClaims(ctx).Subject()

	catalog, release, err := s.runtime.Catalog(ctx, req.InstanceId)
	if err != nil {
		return nil, err
	}
	defer release()

	catalogConversations, err := catalog.FindConversations(ctx, ownerID)
	if err != nil {
		return nil, err
	}

	// Convert catalog conversations to protobuf conversations
	conversations := make([]*runtimev1.Conversation, len(catalogConversations))
	for i, conv := range catalogConversations {
		conversations[i] = runtime.ConversationToPB(conv)
	}

	return &runtimev1.ListConversationsResponse{
		Conversations: conversations,
	}, nil
}

// GetConversation returns a conversation and its messages.
func (s *Server) GetConversation(ctx context.Context, req *runtimev1.GetConversationRequest) (*runtimev1.GetConversationResponse, error) {
	if !auth.GetClaims(ctx).CanInstance(req.InstanceId, auth.UseAI) {
		return nil, ErrForbidden
	}

	catalog, release, err := s.runtime.Catalog(ctx, req.InstanceId)
	if err != nil {
		return nil, err
	}
	defer release()

	catalogConversation, err := catalog.FindConversation(ctx, req.ConversationId)
	if err != nil {
		return nil, err
	}

	// For now, we only allow users to access their own conversations.
	if catalogConversation.OwnerID != auth.GetClaims(ctx).Subject() {
		return nil, status.Error(codes.NotFound, "conversation not found")
	}

	// Convert catalog conversation to protobuf and fetch messages
	conversation := runtime.ConversationToPB(catalogConversation)

	// Fetch messages separately and convert them
	catalogMessages, err := catalog.FindMessages(ctx, req.ConversationId)
	if err != nil {
		return nil, err
	}

	messages := make([]*runtimev1.Message, 0, len(catalogMessages))
	for _, msg := range catalogMessages {
		pbMessage, err := runtime.MessageToPB(msg)
		if err != nil {
			return nil, err
		}

		// Filter out system messages unless explicitly requested
		if msg.Role == "system" && !req.IncludeSystemMessages {
			continue
		}

		messages = append(messages, pbMessage)
	}
	conversation.Messages = messages

	return &runtimev1.GetConversationResponse{
		Conversation: conversation,
	}, nil
}

// serverToolService implements runtime.ToolService using the server's MCP functionality
type serverToolService struct {
	server     *Server
	instanceID string
}

// ListTools implements runtime.ToolService
func (s *serverToolService) ListTools(ctx context.Context) ([]*aiv1.Tool, error) {
	return s.server.mcpListTools(ctx, s.instanceID)
}

// ExecuteTool implements runtime.ToolService
func (s *serverToolService) ExecuteTool(ctx context.Context, toolName string, toolArgs map[string]any) (any, error) {
	return s.server.mcpExecuteTool(ctx, s.instanceID, toolName, toolArgs)
}

// Complete runs a conversational AI completion with tool calling support.
func (s *Server) Complete(ctx context.Context, req *runtimev1.CompleteRequest) (resp *runtimev1.CompleteResponse, err error) {
	if !auth.GetClaims(ctx).CanInstance(req.InstanceId, auth.UseAI) {
		return nil, ErrForbidden
	}

	// Add basic validation - fail fast for invalid requests
	if len(req.Messages) == 0 {
		return nil, status.Error(codes.InvalidArgument, "messages cannot be empty")
	}

	ownerID := auth.GetClaims(ctx).Subject()

	// Handle conversation ID: nil or empty string means create new conversation
	var conversationID string
	if req.ConversationId != nil {
		conversationID = *req.ConversationId
	}

	// Create tool service for this server
	toolService := &serverToolService{server: s, instanceID: req.InstanceId}

	// Delegate to runtime business logic
	result, err := s.runtime.CompleteWithTools(ctx, &runtime.CompleteWithToolsOptions{
		OwnerID:        ownerID,
		InstanceID:     req.InstanceId,
		ConversationID: conversationID,
		AppContext:     req.AppContext,
		Messages:       req.Messages,
		ToolService:    toolService,
	})
	if err != nil {
		return nil, err
	}

	// Transform runtime result to gRPC response
	return &runtimev1.CompleteResponse{
		ConversationId: result.ConversationID,
		Messages:       result.Messages,
	}, nil
}

// CompleteStreaming implements RuntimeService
func (s *Server) CompleteStreaming(req *runtimev1.CompleteStreamingRequest, stream runtimev1.RuntimeService_CompleteStreamingServer) error {
	// Access check
	claims := auth.GetClaims(stream.Context())
	if !claims.CanInstance(req.InstanceId, auth.UseAI) {
		return ErrForbidden
	}

	// Add basic validation - fail fast for invalid requests
	if req.Prompt == "" {
		return status.Error(codes.InvalidArgument, "prompt cannot be empty")
	}

	// Open the AI session
	runner := ai.NewRunner(s.runtime)
	session, err := runner.Session(stream.Context(), req.InstanceId, req.ConversationId, claims.Subject(), claims.SecurityClaims())
	if err != nil {
		return err
	}
	defer session.Close()

	// Context
	ctx, cancel := context.WithCancel(stream.Context())
	defer cancel()

	// Make the call
	callErrCh := make(chan error)
	go func() {
		time.Sleep(time.Millisecond * 10) // Allow the subscribe to happen. TODO: Find a non-hacky solution here.
		var res *ai.RouterAgentResult
		_, err := session.CallTool(ctx, ai.RoleUser, "router_agent", &res, ai.RouterAgentArgs{
			Prompt: req.Prompt,
		})
		time.Sleep(time.Millisecond * 50) // Allow the last message to be sent. TODO: Find a non-hacky solution here.
		cancel()
		callErrCh <- err
	}()

	// Subscribe to session messages and stream them to the client
	subErr := session.Subscribe(ctx, func(msg *ai.Message) {
		pb, err := msg.ToProto()
		if err != nil {
			s.logger.Error("failed to convert AI message to protobuf", zap.Error(err))
			return
		}
		err = stream.Send(&runtimev1.CompleteStreamingResponse{
			ConversationId: msg.SessionID,
			Message: &runtimev1.Message{
				Id:        msg.ID,
				Role:      pb.Role,
				Content:   pb.Content,
				CreatedOn: timestamppb.New(msg.Time),
				UpdatedOn: timestamppb.New(msg.Time),
			},
		})
		if err != nil {
			s.logger.Warn("failed to send AI message to stream", zap.Error(err))
		}
	})

	// Wait for call to finish
	cancel()
	callErr := <-callErrCh
	if callErr != nil && !errors.Is(callErr, context.Canceled) {
		return callErr
	}
	if subErr != nil && !errors.Is(subErr, context.Canceled) {
		return subErr
	}
	return nil
}
