package server

import (
	"context"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/server/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

	conversations, err := catalog.ListConversations(ctx, ownerID)
	if err != nil {
		return nil, err
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

	conversation, err := catalog.GetConversation(ctx, req.ConversationId)
	if err != nil {
		return nil, err
	}

	// For now, we only allow users to access their own conversations.
	if conversation.OwnerId != auth.GetClaims(ctx).Subject() {
		return nil, status.Error(codes.NotFound, "conversation not found")
	}

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
func (s *serverToolService) ListTools(ctx context.Context) ([]runtime.Tool, error) {
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
	if req.Messages == nil {
		return nil, status.Error(codes.InvalidArgument, "messages cannot be nil")
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
	result, err := s.runtime.CompleteWithTools(ctx, ownerID, &runtime.CompletionRequest{
		InstanceID:     req.InstanceId,
		ConversationID: conversationID,
		Messages:       req.Messages,
	}, toolService)
	if err != nil {
		return nil, err
	}

	// Transform runtime result to gRPC response
	return &runtimev1.CompleteResponse{
		ConversationId: result.ConversationID,
		Messages:       result.Messages,
	}, nil
}
