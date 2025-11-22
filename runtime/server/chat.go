package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/r3labs/sse/v2"
	aiv1 "github.com/rilldata/rill/proto/gen/rill/ai/v1"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/ai"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/metricsview"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/rilldata/rill/runtime/server/auth"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Server) ListConversations(ctx context.Context, req *runtimev1.ListConversationsRequest) (*runtimev1.ListConversationsResponse, error) {
	claims := auth.GetClaims(ctx, req.InstanceId)
	if !claims.Can(runtime.UseAI) {
		return nil, ErrForbidden
	}

	if claims.UserID == "" {
		return &runtimev1.ListConversationsResponse{}, nil
	}

	catalog, release, err := s.runtime.Catalog(ctx, req.InstanceId)
	if err != nil {
		return nil, err
	}
	defer release()

	sessions, err := catalog.FindAISessions(ctx, claims.UserID, req.UserAgentPattern)
	if err != nil {
		return nil, err
	}

	res := make([]*runtimev1.Conversation, len(sessions))
	for i, s := range sessions {
		res[i] = sessionToPB(s, nil)
	}
	return &runtimev1.ListConversationsResponse{
		Conversations: res,
	}, nil
}

func (s *Server) GetConversation(ctx context.Context, req *runtimev1.GetConversationRequest) (*runtimev1.GetConversationResponse, error) {
	claims := auth.GetClaims(ctx, req.InstanceId)
	if !claims.Can(runtime.UseAI) {
		return nil, ErrForbidden
	}

	session, err := s.ai.Session(ctx, &ai.SessionOptions{
		InstanceID: req.InstanceId,
		SessionID:  req.ConversationId,
		Claims:     claims,
	})
	if err != nil {
		return nil, err
	}

	messages := session.Messages()
	messagePBs := make([]*runtimev1.Message, 0, len(messages))
	for _, msg := range messages {
		pb, err := messageToPB(session, msg)
		if err != nil {
			return nil, err
		}
		messagePBs = append(messagePBs, pb)
	}

	return &runtimev1.GetConversationResponse{
		Conversation: sessionToPB(session.CatalogSession(), messagePBs),
		Messages:     messagePBs,
	}, nil
}

// Complete runs a conversational AI completion with tool calling support.
func (s *Server) Complete(ctx context.Context, req *runtimev1.CompleteRequest) (resp *runtimev1.CompleteResponse, resErr error) {
	// Access check
	claims := auth.GetClaims(ctx, req.InstanceId)
	if !claims.Can(runtime.UseAI) {
		return nil, ErrForbidden
	}

	// Add basic validation - fail fast for invalid requests
	if req.Prompt == "" {
		return nil, status.Error(codes.InvalidArgument, "prompt cannot be empty")
	}

	// Setup user agent
	version := s.runtime.Version().Number
	if version == "" {
		version = "unknown"
	}
	userAgent := fmt.Sprintf("rill/%s", version)

	// Open the AI session
	session, err := s.ai.Session(ctx, &ai.SessionOptions{
		InstanceID: req.InstanceId,
		SessionID:  req.ConversationId,
		Claims:     claims,
		UserAgent:  userAgent,
	})
	if err != nil {
		return nil, err
	}
	defer func() {
		err := session.Flush(ctx)
		if err != nil {
			resErr = errors.Join(resErr, err)
		}
	}()

	// Context
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var analystAgentArgs *ai.AnalystAgentArgs
	if req.Explore != "" {
		analystAgentArgs = &ai.AnalystAgentArgs{
			Explore:    req.Explore,
			Dimensions: req.Dimensions,
			Measures:   req.Measures,
			Where:      metricsview.NewExpressionFromProto(req.Where),
			TimeStart:  req.TimeStart.AsTime(),
			TimeEnd:    req.TimeEnd.AsTime(),
		}
	}

	// Make the call
	var res *ai.RouterAgentResult
	msg, err := session.CallTool(ctx, ai.RoleUser, "router_agent", &res, ai.RouterAgentArgs{
		Prompt:           req.Prompt,
		AnalystAgentArgs: analystAgentArgs,
	})
	if err != nil {
		return nil, err
	}

	// Lookup the result message and all its descendents
	msgs := session.MessagesWithDescendents(ai.FilterByID(msg.Call.ID))

	// Build result
	pbs := make([]*runtimev1.Message, 0, len(msgs))
	for _, msg := range msgs {
		pb, err := messageToPB(session, msg)
		if err != nil {
			return nil, err
		}
		pbs = append(pbs, pb)
	}

	return &runtimev1.CompleteResponse{
		ConversationId: session.ID(),
		Messages:       pbs,
	}, nil
}

// CompleteStreaming implements RuntimeService
func (s *Server) CompleteStreaming(req *runtimev1.CompleteStreamingRequest, stream runtimev1.RuntimeService_CompleteStreamingServer) (resErr error) {
	// Access check
	claims := auth.GetClaims(stream.Context(), req.InstanceId)
	if !claims.Can(runtime.UseAI) {
		return ErrForbidden
	}

	// Add basic validation - fail fast for invalid requests
	if req.Prompt == "" {
		return status.Error(codes.InvalidArgument, "prompt cannot be empty")
	}

	// Setup user agent
	version := s.runtime.Version().Number
	if version == "" {
		version = "unknown"
	}
	userAgent := fmt.Sprintf("rill/%s", version)

	// Open the AI session
	session, err := s.ai.Session(stream.Context(), &ai.SessionOptions{
		InstanceID: req.InstanceId,
		SessionID:  req.ConversationId,
		Claims:     claims,
		UserAgent:  userAgent,
	})
	if err != nil {
		return err
	}
	defer func() {
		err := session.Flush(stream.Context())
		if err != nil {
			resErr = errors.Join(resErr, err)
		}
	}()

	// Context
	ctx, cancel := context.WithCancel(stream.Context())
	defer cancel()

	// Open subscription for session messages and stream them to the client in the background
	subCh := session.Subscribe()
	defer session.Unsubscribe(subCh)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-subCh:
				if !ok {
					return
				}
				pb, err := messageToPB(session, msg)
				if err != nil {
					s.logger.Error("failed to convert AI message to protobuf", zap.Error(err))
					continue
				}
				err = stream.Send(&runtimev1.CompleteStreamingResponse{
					ConversationId: msg.SessionID,
					Message:        pb,
				})
				if err != nil {
					s.logger.Warn("failed to send AI message to stream", zap.Error(err))
				}
			}
		}
	}()

	var analystAgentArgs *ai.AnalystAgentArgs
	if req.Explore != "" {
		analystAgentArgs = &ai.AnalystAgentArgs{
			Explore:    req.Explore,
			Dimensions: req.Dimensions,
			Measures:   req.Measures,
			Where:      metricsview.NewExpressionFromProto(req.Where),
			TimeStart:  req.TimeStart.AsTime(),
			TimeEnd:    req.TimeEnd.AsTime(),
		}
	}

	// Make the call
	var res *ai.RouterAgentResult
	_, err = session.CallTool(ctx, ai.RoleUser, "router_agent", &res, ai.RouterAgentArgs{
		Prompt:           req.Prompt,
		AnalystAgentArgs: analystAgentArgs,
	})
	if err != nil && !errors.Is(err, context.Canceled) {
		return err
	}
	return nil
}

// CompleteStreamingHandler is a HTTP handler that wraps CompleteStreaming and maps it to SSE.
// This is required as vanguard doesn't currently map streaming RPCs to SSE, so we register this handler manually override the behavior
func (s *Server) CompleteStreamingHandler(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	instanceID := req.PathValue("instance_id")

	// Add timeout for AI completion
	ctx, cancel := context.WithTimeout(ctx, time.Minute*5)
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

// sessionToPB converts a drivers.AISession to a runtimev1.Conversation.
func sessionToPB(s *drivers.AISession, messages []*runtimev1.Message) *runtimev1.Conversation {
	return &runtimev1.Conversation{
		Id:        s.ID,
		OwnerId:   s.OwnerID,
		Title:     s.Title,
		UserAgent: s.UserAgent,
		CreatedOn: timestamppb.New(s.CreatedOn),
		UpdatedOn: timestamppb.New(s.UpdatedOn),
		Messages:  messages,
	}
}

// messageToPB converts an ai.Message to a runtimev1.Message.
func messageToPB(s *ai.Session, msg *ai.Message) (*runtimev1.Message, error) {
	// If it's the top-level router_agent tool call, parse its content and return a plain text block with the prompt/response.
	// In other cases, handle it as a normal block.
	var block *aiv1.ContentBlock
	if msg.Tool == ai.RouterAgentName {
		var text string
		switch msg.Type {
		case ai.MessageTypeCall:
			args, err := s.UnmarshalMessageContent(msg)
			if err != nil {
				return nil, err
			}
			text = args.(*ai.RouterAgentArgs).Prompt
		case ai.MessageTypeResult:
			switch msg.ContentType {
			case ai.MessageContentTypeJSON:
				res, err := s.UnmarshalMessageContent(msg)
				if err != nil {
					return nil, err
				}
				text = res.(*ai.RouterAgentResult).Response
			case ai.MessageContentTypeError:
				text = fmt.Sprintf("Error: %s", msg.Content)
			default:
				text = msg.Content
			}
		default:
			text = msg.Content
		}

		block = &aiv1.ContentBlock{
			BlockType: &aiv1.ContentBlock_Text{
				Text: text,
			},
		}
	} else {
		var err error
		block, err = messageContentToPB(msg)
		if err != nil {
			return nil, err
		}
	}

	// The roles used by the `ai` package do not map to conventional LLM roles, so we change them here.
	// TODO: Refactor such that this is not needed.
	var role string
	switch msg.Type {
	case ai.MessageTypeCall:
		if msg.Tool == ai.RouterAgentName {
			role = "user"
		} else {
			role = "assistant"
		}
	case ai.MessageTypeResult:
		if msg.Tool == ai.RouterAgentName {
			role = "assistant"
		} else {
			role = "tool"
		}
	default:
		if msg.Role == ai.RoleSystem {
			role = "system"
		} else {
			role = "assistant"
		}
	}

	return &runtimev1.Message{
		Id:          msg.ID,
		ParentId:    msg.ParentID,
		CreatedOn:   timestamppb.New(msg.Time),
		UpdatedOn:   timestamppb.New(msg.Time),
		Index:       uint32(msg.Index),
		Role:        role,
		Type:        string(msg.Type),
		Tool:        msg.Tool,
		ContentType: string(msg.ContentType),
		ContentData: msg.Content,
		Content:     []*aiv1.ContentBlock{block},
	}, nil
}

// messageContentToPB converts an ai.Message Content to a aiv1.ContentBlock.
func messageContentToPB(msg *ai.Message) (*aiv1.ContentBlock, error) {
	switch msg.Type {
	case ai.MessageTypeProgress:
		return &aiv1.ContentBlock{
			BlockType: &aiv1.ContentBlock_Text{
				Text: msg.Content,
			},
		}, nil
	case ai.MessageTypeCall:
		if msg.ContentType != ai.MessageContentTypeJSON {
			return nil, fmt.Errorf("unexpected content type %q for tool call message %q", msg.ContentType, msg.ID)
		}
		var input map[string]any
		err := json.Unmarshal([]byte(msg.Content), &input)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal JSON content: %w", err)
		}
		inputPB, err := structpb.NewStruct(input)
		if err != nil {
			return nil, fmt.Errorf("failed to convert tool call input to StructPB: %w", err)
		}

		return &aiv1.ContentBlock{
			BlockType: &aiv1.ContentBlock_ToolCall{
				ToolCall: &aiv1.ToolCall{
					Id:    msg.ID,
					Name:  msg.Tool,
					Input: inputPB,
				},
			},
		}, nil
	case ai.MessageTypeResult:
		return &aiv1.ContentBlock{
			BlockType: &aiv1.ContentBlock_ToolResult{
				ToolResult: &aiv1.ToolResult{
					Id:      msg.ParentID,
					Content: msg.Content,
					IsError: msg.ContentType == ai.MessageContentTypeError,
				},
			},
		}, nil
	default:
		return nil, fmt.Errorf("unexpected message type %q for message %q", msg.Type, msg.ID)
	}
}
