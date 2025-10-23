package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/r3labs/sse/v2"
	aiv1 "github.com/rilldata/rill/proto/gen/rill/ai/v1"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/rilldata/rill/runtime/server/auth"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
)

// ListConversations returns a list of conversations for an instance.
func (s *Server) ListConversations(ctx context.Context, req *runtimev1.ListConversationsRequest) (*runtimev1.ListConversationsResponse, error) {
	claims := auth.GetClaims(ctx, req.InstanceId)
	if !claims.Can(runtime.UseAI) {
		return nil, ErrForbidden
	}

	catalog, release, err := s.runtime.Catalog(ctx, req.InstanceId)
	if err != nil {
		return nil, err
	}
	defer release()

	catalogConversations, err := catalog.FindConversations(ctx, claims.UserID)
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
	claims := auth.GetClaims(ctx, req.InstanceId)
	if !claims.Can(runtime.UseAI) {
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
	if catalogConversation.OwnerID != claims.UserID {
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

// Complete runs a conversational AI completion with tool calling support.
func (s *Server) Complete(ctx context.Context, req *runtimev1.CompleteRequest) (resp *runtimev1.CompleteResponse, err error) {
	claims := auth.GetClaims(ctx, req.InstanceId)
	if !claims.Can(runtime.UseAI) {
		return nil, ErrForbidden
	}

	// Add basic validation - fail fast for invalid requests
	if len(req.Messages) == 0 {
		return nil, status.Error(codes.InvalidArgument, "messages cannot be empty")
	}

	// Handle conversation ID: nil or empty string means create new conversation
	var conversationID string
	if req.ConversationId != nil {
		conversationID = *req.ConversationId
	}

	// Create tool service for this server
	mcpServer, err := s.newMCPServer(ctx, req.InstanceId, false)
	if err != nil {
		return nil, fmt.Errorf("failed to create MCP server: %w", err)
	}
	mcpClient, err := newMCPClient(mcpServer)
	if err != nil {
		return nil, fmt.Errorf("failed to create MCP client: %w", err)
	}
	defer mcpClient.Close()
	toolService := &serverToolService{instanceID: req.InstanceId, mcpClient: mcpClient}

	// Delegate to runtime business logic
	result, err := s.runtime.CompleteWithTools(ctx, &runtime.CompleteWithToolsOptions{
		OwnerID:        claims.UserID,
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
	claims := auth.GetClaims(stream.Context(), req.InstanceId)
	if !claims.Can(runtime.UseAI) {
		return ErrForbidden
	}

	// Add basic validation - fail fast for invalid requests
	if req.Prompt == "" {
		return status.Error(codes.InvalidArgument, "prompt cannot be empty")
	}

	// Create tool service for this server
	mcpServer, err := s.newMCPServer(stream.Context(), req.InstanceId, false)
	if err != nil {
		return fmt.Errorf("failed to create MCP server: %w", err)
	}
	mcpClient, err := newMCPClient(mcpServer)
	if err != nil {
		return fmt.Errorf("failed to create MCP client: %w", err)
	}
	defer mcpClient.Close()
	toolService := &serverToolService{instanceID: req.InstanceId, mcpClient: mcpClient}

	// Delegate to runtime business logic
	_, err = s.runtime.CompleteWithTools(stream.Context(), &runtime.CompleteWithToolsOptions{
		OwnerID:        claims.UserID,
		InstanceID:     req.InstanceId,
		ConversationID: req.ConversationId,
		Messages: []*runtimev1.Message{{Role: "user", Content: []*aiv1.ContentBlock{{
			BlockType: &aiv1.ContentBlock_Text{
				Text: req.Prompt,
			},
		}}}},
		ToolService: toolService,
		OnMessage: func(conversationID string, msg *runtimev1.Message) error {
			// Emit one message for each content block.
			// In a future refactor, we'll try to apply this in the internal interfaces as well.
			for _, block := range msg.Content {
				err := stream.Send(&runtimev1.CompleteStreamingResponse{
					ConversationId: conversationID,
					Message: &runtimev1.Message{
						Id:        msg.Id,
						Role:      msg.Role,
						Content:   []*aiv1.ContentBlock{block},
						CreatedOn: msg.CreatedOn,
						UpdatedOn: msg.UpdatedOn,
					},
				})
				if err != nil {
					return err
				}
			}
			return nil
		},
	})
	if err != nil {
		return err
	}

	return nil
}

// CompleteStreamingHandler is a HTTP handler that wraps CompleteStreaming and maps it to SSE.
// This is required as vanguard doesn't currently map streaming RPCs to SSE, so we register this handler manually override the behavior
func (s *Server) CompleteStreamingHandler(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	instanceID := req.PathValue("instance_id")

	// Add timeout matching the completionTimeout from runtime/completion.go
	ctx, cancel := context.WithTimeout(ctx, time.Minute*2)
	defer cancel()

	// Replace request context with the timed context
	req = req.WithContext(ctx)

	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", instanceID),
	)

	// Access check
	if !auth.GetClaims(ctx, instanceID).Can(runtime.UseAI) {
		http.Error(w, "action not allowed", http.StatusUnauthorized)
		return
	}

	// Build request. Note we try to support both GET and POST.
	completeReq := &runtimev1.CompleteStreamingRequest{}
	if req.Method == http.MethodGet {
		// Parse from query parameters
		completeReq.ConversationId = req.URL.Query().Get("conversationId")
		completeReq.Prompt = req.URL.Query().Get("prompt")
	} else {
		// Parse from JSON body
		body, err := io.ReadAll(req.Body)
		if err != nil {
			http.Error(w, "failed to read request body", http.StatusBadRequest)
			return
		}
		if err := protojson.Unmarshal(body, completeReq); err != nil {
			http.Error(w, "failed to parse request body", http.StatusBadRequest)
			return
		}
	}
	completeReq.InstanceId = instanceID // Set instance ID from path

	// Initialize SSE server
	eventServer := sse.New()
	eventServer.CreateStream("messages")
	eventServer.Headers = map[string]string{
		"Content-Type":  "text/event-stream",
		"Cache-Control": "no-cache",
		"Connection":    "keep-alive",
	}

	// Create the shim that implements RuntimeService_CompleteStreamingServer
	shim := &completeStreamingServerShim{
		r: req,
		s: eventServer,
	}

	// Create a goroutine to handle the streaming
	go func() {
		// Call the existing CompleteStreaming implementation with our shim
		err := s.CompleteStreaming(completeReq, shim)
		if err != nil {
			if !errors.Is(err, context.Canceled) {
				s.logger.Warn("complete streaming error", zap.String("instance_id", instanceID), zap.Error(err))
			}

			errJSON, err := json.Marshal(map[string]string{"error": err.Error()})
			if err != nil {
				s.logger.Error("failed to marshal error as json", zap.Error(err))
			}

			eventServer.Publish("messages", &sse.Event{
				Data:  errJSON,
				Event: []byte("error"),
			})
		}
		eventServer.Close()
	}()

	// Serve the SSE stream
	eventServer.ServeHTTP(w, req)
}

// completeStreamingServerShim is a shim for runtimev1.RuntimeService_CompleteStreamingServer
type completeStreamingServerShim struct {
	r *http.Request
	s *sse.Server
}

func (ss *completeStreamingServerShim) Context() context.Context {
	return ss.r.Context()
}

func (ss *completeStreamingServerShim) Send(e *runtimev1.CompleteStreamingResponse) error {
	data, err := protojson.Marshal(e)
	if err != nil {
		return err
	}

	ss.s.Publish("messages", &sse.Event{Data: data})
	return nil
}

func (ss *completeStreamingServerShim) SetHeader(metadata.MD) error {
	return errors.New("not implemented")
}

func (ss *completeStreamingServerShim) SendHeader(metadata.MD) error {
	return errors.New("not implemented")
}

func (ss *completeStreamingServerShim) SetTrailer(metadata.MD) {}

func (ss *completeStreamingServerShim) SendMsg(m any) error {
	return errors.New("not implemented")
}

func (ss *completeStreamingServerShim) RecvMsg(m any) error {
	return errors.New("not implemented")
}

// serverToolService implements runtime.ToolService with the provided MCP client.
type serverToolService struct {
	instanceID string
	mcpClient  *client.Client
}

// ListTools implements runtime.ToolService
func (s *serverToolService) ListTools(ctx context.Context) ([]*aiv1.Tool, error) {
	// Add instance ID to context for internal MCP server tools
	ctxWithInstance := context.WithValue(ctx, mcpInstanceIDKey{}, s.instanceID)

	tools, err := s.mcpClient.ListTools(ctxWithInstance, mcp.ListToolsRequest{})
	if err != nil {
		return nil, err
	}

	aiTools := make([]*aiv1.Tool, len(tools.Tools))
	for i := range tools.Tools {
		tool := &tools.Tools[i]
		aiTool := &aiv1.Tool{
			Name:        tool.Name,
			Description: tool.Description,
		}

		// Convert InputSchema to JSON string if present
		if schemaBytes, err := json.Marshal(tool.InputSchema); err == nil && string(schemaBytes) != "{}" && string(schemaBytes) != "null" {
			aiTool.InputSchema = string(schemaBytes)
		}

		aiTools[i] = aiTool
	}

	return aiTools, nil
}

// ExecuteTool implements runtime.ToolService
func (s *serverToolService) ExecuteTool(ctx context.Context, toolName string, toolArgs map[string]any) (any, error) {
	// Add instance ID to context for internal MCP server tools
	ctxWithInstance := context.WithValue(ctx, mcpInstanceIDKey{}, s.instanceID)

	resp, err := s.mcpClient.CallTool(ctxWithInstance, mcp.CallToolRequest{
		Params: struct {
			Name      string    `json:"name"`
			Arguments any       `json:"arguments,omitempty"`
			Meta      *mcp.Meta `json:"_meta,omitempty"`
		}{
			Name:      toolName,
			Arguments: toolArgs,
		},
	})

	// Handle errors
	if err != nil {
		return "", err
	} else if len(resp.Content) == 0 {
		return "", nil
	} else if len(resp.Content) > 1 {
		return "", fmt.Errorf("multiple content items not supported, got %d items", len(resp.Content))
	}

	// Extract text content from MCP response
	switch content := resp.Content[0].(type) {
	case mcp.TextContent:
		return content.Text, nil
	default:
		return "", fmt.Errorf("unsupported content type: %T", content) // Future work: support other content types
	}
}

func newMCPClient(mcpServer *server.MCPServer) (*client.Client, error) {
	client, err := client.NewInProcessClient(mcpServer)
	if err != nil {
		return nil, err
	}

	// Start the client with a timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := client.Start(ctx); err != nil {
		return nil, fmt.Errorf("failed to start MCP client: %w", err)
	}

	// Try to initialize the client
	initRequest := mcp.InitializeRequest{}
	initRequest.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initRequest.Params.ClientInfo = mcp.Implementation{
		Name:    "rill",
		Version: "0.0.1",
	}

	if _, err := client.Initialize(ctx, initRequest); err != nil {
		return nil, fmt.Errorf("failed to initialize MCP client: %w", err)
	}

	return client, nil
}
