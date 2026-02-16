package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	aiv1 "github.com/rilldata/rill/proto/gen/rill/ai/v1"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/ai"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/metricsview"
	"github.com/rilldata/rill/runtime/pkg/mapstructureutil"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/rilldata/rill/runtime/pkg/pbutil"
	"github.com/rilldata/rill/runtime/server/auth"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
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

	if claims.UserID == "" && !claims.SkipChecks {
		// This case matches anonymous users on runtimes with auth enabled (i.e. on Rill Cloud).
		// This prevents anonymous users from seeing previous/other anonymous users' conversations.
		// (In Rill Developer, auth is disabled so SkipChecks is true for anonymous users.)
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
		IsOwner:      session.CatalogSession().OwnerID == claims.UserID,
	}, nil
}

func (s *Server) ShareConversation(ctx context.Context, req *runtimev1.ShareConversationRequest) (*runtimev1.ShareConversationResponse, error) {
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

	// This check prevents changing sharing boundaries on already shared conversations by non-owners.
	if session.CatalogSession().OwnerID != claims.UserID && !claims.SkipChecks {
		return nil, ErrForbidden
	}

	// unshare conversation
	if req.UntilMessageId == "none" {
		err = session.UpdateSharedUntilMessageID(ctx, "")
		if err != nil {
			return nil, err
		}
		err = session.Flush(ctx)
		if err != nil {
			return nil, err
		}
		return &runtimev1.ShareConversationResponse{}, nil
	}

	var preds []ai.Predicate
	if req.UntilMessageId == "" {
		preds = []ai.Predicate{ai.FilterByTool(ai.RouterAgentName), ai.FilterByType(ai.MessageTypeResult)}
	} else {
		preds = []ai.Predicate{ai.FilterByID(req.UntilMessageId)}
	}
	msg, ok := session.LatestMessage(preds...)
	if !ok {
		return nil, status.Errorf(codes.InvalidArgument, "message with id %q not found in conversation %q", req.UntilMessageId, req.ConversationId)
	}
	if req.UntilMessageId != "" && !(msg.Tool == ai.RouterAgentName && msg.Type == ai.MessageTypeResult) {
		return nil, status.Errorf(codes.InvalidArgument, "cannot share incomplete conversation as message with id %q is not a router agent result message", req.UntilMessageId)
	}

	// now save the session with the shared until message id and flush immediately
	err = session.UpdateSharedUntilMessageID(ctx, msg.ID)
	if err != nil {
		return nil, err
	}
	err = session.Flush(ctx)
	if err != nil {
		return nil, err
	}

	return &runtimev1.ShareConversationResponse{}, nil
}

func (s *Server) ForkConversation(ctx context.Context, req *runtimev1.ForkConversationRequest) (*runtimev1.ForkConversationResponse, error) {
	claims := auth.GetClaims(ctx, req.InstanceId)
	if !claims.Can(runtime.UseAI) {
		return nil, ErrForbidden
	}

	// Setup user agent
	version := s.runtime.Version().Number
	if version == "" {
		version = "unknown"
	}
	userAgent := fmt.Sprintf("rill/%s", version)

	// Open the existing AI session, this will only contain messages the user has access to
	id, err := s.ai.ForkSession(ctx, &ai.SessionOptions{
		InstanceID: req.InstanceId,
		SessionID:  req.ConversationId,
		Claims:     claims,
		UserAgent:  userAgent,
	})
	if err != nil {
		return nil, err
	}

	return &runtimev1.ForkConversationResponse{
		ConversationId: id,
	}, nil
}

func (s *Server) ListTools(ctx context.Context, req *runtimev1.ListToolsRequest) (*runtimev1.ListToolsResponse, error) {
	// Access check
	claims := auth.GetClaims(ctx, req.InstanceId)
	if !claims.Can(runtime.UseAI) {
		return nil, ErrForbidden
	}

	// List all registered tools
	var pbs []*aiv1.Tool
	for _, tool := range s.ai.Tools {
		pb, err := tool.AsProto()
		if err != nil {
			return nil, fmt.Errorf("failed to convert tool %q to proto: %w", tool.Name, err)
		}
		pbs = append(pbs, pb)
	}
	return &runtimev1.ListToolsResponse{
		Tools: pbs,
	}, nil
}

// Complete runs a conversational AI completion with tool calling support.
func (s *Server) Complete(ctx context.Context, req *runtimev1.CompleteRequest) (resp *runtimev1.CompleteResponse, resErr error) {
	// Access check
	claims := auth.GetClaims(ctx, req.InstanceId)
	if !claims.Can(runtime.UseAI) {
		return nil, ErrForbidden
	}

	// Validate request - either prompt or feedback context must be provided
	if req.Prompt == "" && req.FeedbackAgentContext == nil {
		return nil, status.Error(codes.InvalidArgument, "prompt or feedback_agent_context must be provided")
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

	// Prepare agent args if provided
	var analystAgentArgs *ai.AnalystAgentArgs
	if req.AnalystAgentContext != nil {
		wherePerMetricsView := map[string]*metricsview.Expression{}
		for m, e := range req.AnalystAgentContext.WherePerMetricsView {
			wherePerMetricsView[m] = metricsview.NewExpressionFromProto(e)
		}

		analystAgentArgs = &ai.AnalystAgentArgs{
			Explore:             req.AnalystAgentContext.Explore,
			Canvas:              req.AnalystAgentContext.Canvas,
			CanvasComponent:     req.AnalystAgentContext.CanvasComponent,
			WherePerMetricsView: wherePerMetricsView,
			Dimensions:          req.AnalystAgentContext.Dimensions,
			Measures:            req.AnalystAgentContext.Measures,
			Where:               metricsview.NewExpressionFromProto(req.AnalystAgentContext.Where),
			TimeStart:           req.AnalystAgentContext.TimeStart.AsTime(),
			TimeEnd:             req.AnalystAgentContext.TimeEnd.AsTime(),
		}
	}
	var developerAgentArgs *ai.DeveloperAgentArgs
	if req.DeveloperAgentContext != nil {
		developerAgentArgs = &ai.DeveloperAgentArgs{
			InitProject:     req.DeveloperAgentContext.InitProject,
			CurrentFilePath: req.DeveloperAgentContext.CurrentFilePath,
		}
	}
	var feedbackAgentArgs *ai.FeedbackAgentArgs
	if req.FeedbackAgentContext != nil {
		feedbackAgentArgs = &ai.FeedbackAgentArgs{
			TargetMessageID: req.FeedbackAgentContext.TargetMessageId,
			Sentiment:       req.FeedbackAgentContext.Sentiment,
			Categories:      req.FeedbackAgentContext.Categories,
			Comment:         req.FeedbackAgentContext.Comment,
		}
	}

	// Make the call
	var res *ai.RouterAgentResult
	msg, err := session.CallTool(ctx, ai.RoleUser, ai.RouterAgentName, &res, ai.RouterAgentArgs{
		Prompt:             req.Prompt,
		Agent:              req.Agent,
		AnalystAgentArgs:   analystAgentArgs,
		DeveloperAgentArgs: developerAgentArgs,
		FeedbackAgentArgs:  feedbackAgentArgs,
	})
	if err != nil && msg == nil {
		// We only return errors when msg == nil. When msg != nil, the error was a tool call error, which will be captured in the messages.
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

	// Validate request - either prompt or feedback context must be provided
	if req.Prompt == "" && req.FeedbackAgentContext == nil {
		return status.Error(codes.InvalidArgument, "prompt or feedback_agent_context must be provided")
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

	// Prepare optional context args
	var analystAgentArgs *ai.AnalystAgentArgs
	if req.AnalystAgentContext != nil {
		wherePerMetricsView := map[string]*metricsview.Expression{}
		for m, e := range req.AnalystAgentContext.WherePerMetricsView {
			wherePerMetricsView[m] = metricsview.NewExpressionFromProto(e)
		}

		analystAgentArgs = &ai.AnalystAgentArgs{
			Explore:             req.AnalystAgentContext.Explore,
			Canvas:              req.AnalystAgentContext.Canvas,
			CanvasComponent:     req.AnalystAgentContext.CanvasComponent,
			WherePerMetricsView: wherePerMetricsView,
			Dimensions:          req.AnalystAgentContext.Dimensions,
			Measures:            req.AnalystAgentContext.Measures,
			Where:               metricsview.NewExpressionFromProto(req.AnalystAgentContext.Where),
			TimeStart:           req.AnalystAgentContext.TimeStart.AsTime(),
			TimeEnd:             req.AnalystAgentContext.TimeEnd.AsTime(),
		}
	}
	var developerAgentArgs *ai.DeveloperAgentArgs
	if req.DeveloperAgentContext != nil {
		developerAgentArgs = &ai.DeveloperAgentArgs{
			InitProject:     req.DeveloperAgentContext.InitProject,
			CurrentFilePath: req.DeveloperAgentContext.CurrentFilePath,
		}
	}
	var feedbackAgentArgs *ai.FeedbackAgentArgs
	if req.FeedbackAgentContext != nil {
		feedbackAgentArgs = &ai.FeedbackAgentArgs{
			TargetMessageID: req.FeedbackAgentContext.TargetMessageId,
			Sentiment:       req.FeedbackAgentContext.Sentiment,
			Categories:      req.FeedbackAgentContext.Categories,
			Comment:         req.FeedbackAgentContext.Comment,
		}
	}

	// Make the call
	var res *ai.RouterAgentResult
	msg, err := session.CallTool(ctx, ai.RoleUser, ai.RouterAgentName, &res, ai.RouterAgentArgs{
		Prompt:             req.Prompt,
		Agent:              req.Agent,
		AnalystAgentArgs:   analystAgentArgs,
		DeveloperAgentArgs: developerAgentArgs,
		FeedbackAgentArgs:  feedbackAgentArgs,
	})
	if err != nil && !errors.Is(err, context.Canceled) && msg == nil {
		// We only return errors when msg == nil. When msg != nil, the error was a tool call error, which will be captured in the messages.
		return err
	}
	return nil
}

// CompleteStreamingHandler is a HTTP handler that wraps CompleteStreaming and maps it to SSE.
// This is required as vanguard doesn't currently map streaming RPCs to SSE, so we register this handler manually override the behavior
func (s *Server) CompleteStreamingHandler(w http.ResponseWriter, req *http.Request) {
	// Add timeout for AI completion
	ctx, cancel := context.WithTimeout(req.Context(), time.Minute*5)
	defer cancel()
	req = req.WithContext(ctx) // Replace request context with the timed context

	// Observability
	instanceID := req.PathValue("instance_id")
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

	// Start goroutine that calls CompleteStreaming and publishes responses to a channel
	events := make(chan *sseEvent)
	go func() {
		// We must close the events channel when done to make sure the SSE handler returns
		defer close(events)

		// Create the shim that implements RuntimeService_CompleteStreamingServer
		shim := &grpcStreamingShim[*runtimev1.CompleteStreamingResponse]{
			ctx: ctx,
			fn: func(data []byte) error {
				events <- &sseEvent{Data: data}
				return nil
			},
		}

		// Call the existing CompleteStreaming implementation with our shim
		err := s.CompleteStreaming(completeReq, shim)
		if err != nil {
			code := codes.Unknown
			msg := err.Error()
			if s, ok := status.FromError(err); ok {
				code = s.Code()
				msg = s.Message()
			}

			errJSON, err := json.Marshal(map[string]string{"code": code.String(), "error": msg})
			if err != nil {
				s.logger.Error("failed to marshal error as json", zap.Error(err))
			}

			events <- &sseEvent{
				Event: "error",
				Data:  errJSON,
			}
		}
	}()

	// Serve the SSE stream.
	// This will only return when the background goroutine calls close(events).
	serveSSEUntilClose(w, events)
}

func (s *Server) GetAIToolCall(ctx context.Context, req *runtimev1.GetAIToolCallRequest) (*runtimev1.GetAIToolCallResponse, error) {
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
		return nil, status.Errorf(codes.NotFound, "failed to find the conversaion %q", req.ConversationId)
	}

	callMsg, ok := session.Message(ai.FilterByID(req.CallId))
	if !ok {
		return nil, status.Errorf(codes.NotFound, "call message with ID %q not found", req.CallId)
	}

	rawReq, err := session.UnmarshalMessageContent(callMsg)
	if err != nil {
		return nil, err
	}
	var queryRes ai.QueryMetricsViewArgs
	err = mapstructureutil.WeakDecode(rawReq, &queryRes)
	if err != nil {
		return nil, err
	}

	queryPb, err := pbutil.ToStruct(queryRes, nil)
	if err != nil {
		return nil, err
	}

	return &runtimev1.GetAIToolCallResponse{
		Query: queryPb,
	}, nil
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
