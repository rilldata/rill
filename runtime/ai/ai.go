package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	aiv1 "github.com/rilldata/rill/proto/gen/rill/ai/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
)

// Runner tracks available tools and manages the lifecycle of AI sessions.
type Runner struct {
	Runtime *runtime.Runtime
	Tools   map[string]*wrappedTool
}

// NewRunner creates a new Runner.
func NewRunner(rt *runtime.Runtime) *Runner {
	r := &Runner{
		Runtime: rt,
		Tools:   make(map[string]*wrappedTool),
	}

	RegisterTool(r, &RouterAgent{Runtime: rt})
	RegisterTool(r, &AnalystAgent{Runtime: rt})
	RegisterTool(r, &DeveloperAgent{Runtime: rt})

	RegisterTool(r, &ListMetricsViews{Runtime: rt})
	RegisterTool(r, &GetMetricsView{Runtime: rt})
	RegisterTool(r, &QueryMetricsViewTimeRange{Runtime: rt})
	RegisterTool(r, &QueryMetricsView{Runtime: rt})

	RegisterTool(r, &DevelopModel{Runtime: rt})
	RegisterTool(r, &DevelopMetricsView{Runtime: rt})
	RegisterTool(r, &ListFiles{Runtime: rt})
	RegisterTool(r, &ReadFile{Runtime: rt})
	RegisterTool(r, &WriteFile{Runtime: rt})

	return r
}

// SessionOptions provides options for initializing a new session.
type SessionOptions struct {
	InstanceID string
	SessionID  string
	Claims     *runtime.SecurityClaims
	UserAgent  string
}

// Session creates or loads an AI session.
func (r *Runner) Session(ctx context.Context, opts *SessionOptions) (*Session, error) {
	// TODO: Load from database or create in database.
	if opts.SessionID == "" {
		opts.SessionID = uuid.NewString()
	}

	ai, release, err := r.Runtime.AI(ctx, opts.InstanceID)
	if err != nil {
		return nil, err
	}

	instance, err := r.Runtime.Instance(ctx, opts.InstanceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get instance %q: %w", opts.InstanceID, err)
	}

	logger := r.Runtime.Logger.Named("ai").With(
		zap.String("instance_id", opts.InstanceID),
		zap.String("session_id", opts.SessionID),
		zap.String("user_id", opts.Claims.UserID),
	)

	base := &BaseSession{
		id:         opts.SessionID,
		instanceID: opts.InstanceID,
		title:      "", // TODO: Load from database
		userID:     opts.Claims.UserID,
		claims:     opts.Claims,

		runner:              r,
		logger:              logger,
		llm:                 ai,
		llmRelease:          release,
		projectInstructions: instance.AIInstructions,

		subscribers: make(map[chan *Message]struct{}),
	}

	return &Session{
		BaseSession: base,
	}, nil
}

// Tool is an interface for an AI tool.
type Tool[In, Out any] interface {
	Spec() *mcp.Tool
	CheckAccess(claims *runtime.SecurityClaims) bool
	Handler(ctx context.Context, args In) (Out, error)
}

// wrappedTool is the internal representation of a registered tool.
type wrappedTool struct {
	spec                  *mcp.Tool
	checkAccess           func(claims *runtime.SecurityClaims) bool
	jsonHandler           func(ctx context.Context, input json.RawMessage) (json.RawMessage, error)
	registerWithMCPServer func(srv *mcp.Server)
}

// RegisterTool registers a new tool with the Runner.
func RegisterTool[In, Out any](s *Runner, t Tool[In, Out]) {
	spec := t.Spec()
	if spec.InputSchema == nil {
		spec.InputSchema, _ = schemaFor[In](false)
	}
	if spec.OutputSchema == nil {
		spec.OutputSchema, _ = schemaFor[Out](true)
	}

	s.Tools[spec.Name] = &wrappedTool{
		spec:        spec,
		checkAccess: t.CheckAccess,
		jsonHandler: func(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
			var args In
			if err := json.Unmarshal(input, &args); err != nil {
				return nil, err
			}
			result, err := t.Handler(ctx, args)
			if err != nil {
				return nil, err
			}
			data, err := json.Marshal(result)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal result for tool %q: %w", spec.Name, err)
			}
			return data, nil
		},
		registerWithMCPServer: func(srv *mcp.Server) {
			mcp.AddTool(srv, spec, func(ctx context.Context, req *mcp.CallToolRequest, args In) (*mcp.CallToolResult, Out, error) {
				res, err := t.Handler(ctx, args)
				if err != nil {
					var emptyOut Out
					return nil, emptyOut, err
				}
				return nil, res, nil
			})
		},
	}
}

// schemaFor generates a JSON schema for a given type.
// If ignoreIfAny is true, it will return a nil schema if T has type any (use for MCP output schema, where no schema means unstructured result).
// It is loosely derived from similar logic in github.com/modelcontextprotocol/go-sdk.
func schemaFor[T any](ignoreIfAny bool) (*jsonschema.Schema, error) {
	if reflect.TypeFor[T]() == reflect.TypeFor[any]() {
		if ignoreIfAny {
			return nil, nil
		}
		return &jsonschema.Schema{Type: "object"}, nil
	}

	tt := reflect.TypeFor[T]()
	if tt.Kind() == reflect.Pointer {
		tt = tt.Elem()
	}

	schema, err := jsonschema.ForType(tt, &jsonschema.ForOptions{})
	if err != nil {
		return nil, err
	}

	return schema, nil
}

// Role is the role of the actor that created a message.
type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool"
)

// MessageType is the type of message being sent.
type MessageType string

const (
	MessageTypePrompt   MessageType = "prompt"
	MessageTypeCall     MessageType = "call"
	MessageTypeProgress MessageType = "progress"
	MessageTypeResult   MessageType = "result"
	MessageTypeError    MessageType = "error"
)

// MessageContentType is the type of content contained in a message.
type MessageContentType string

const (
	MessageContentTypeText MessageContentType = "text"
	MessageContentTypeJSON MessageContentType = "json"
)

// Message represents a message in an AI session.
// Unlike lower-level LLM messages, the messages here include a call hierarchy, enabling tracking of calls and results inside tool calls.
//
// Mental model:
// - Messages represent user input, tool calls/results, LLM thinking, LLM responses.
// - Messages can be called by users, deterministic code, or LLMs.
// - LLM invocations retrieve messages from current scope for context.
type Message struct {
	// ID is unique for each message.
	ID string `json:"id" yaml:"id"`
	// ParentID is the ID of the parent message, usually the current tool call.
	ParentID string `json:"parent_id" yaml:"parent_id"`
	// SessionID is the ID of the session this message belongs to.
	SessionID string `json:"session_id" yaml:"session_id"`
	// Time the message was created.
	Time time.Time `json:"time" yaml:"time"`
	// Index of the message in the session. Used to order messages returned at the same time.
	Index int `json:"index" yaml:"index"`
	// Role is the actor that created the message.
	Role Role `json:"role" yaml:"role"`
	// Type is the type of the message.
	// For any given call, there will be only one "result" or "error" message.
	Type MessageType `json:"type" yaml:"type"`
	// Tool is the name of the tool that emitted the message, if any.
	Tool string `json:"tool"`
	// ContentType is the type of the Content string.
	ContentType MessageContentType `json:"content_type" yaml:"content_type"`
	// Content is the content of the message.
	Content string `json:"content" yaml:"content"`
}

// ToProto converts the message to an aiv1.CompletionMessage
func (m *Message) ToProto() (*aiv1.CompletionMessage, error) {
	// As an exception, rewrite results from the root tool call to a plain-text assistant message.
	if m.Tool == "router_agent" { // TODO: Make generic
		return &aiv1.CompletionMessage{
			Role: string(RoleAssistant),
			Content: []*aiv1.ContentBlock{{
				BlockType: &aiv1.ContentBlock_Text{
					Text: m.Content,
				},
			}},
		}, nil
	}

	var block *aiv1.ContentBlock
	switch m.Type {
	case MessageTypePrompt:
		block = &aiv1.ContentBlock{
			BlockType: &aiv1.ContentBlock_Text{
				Text: m.Content,
			},
		}
	case MessageTypeCall:
		var args map[string]any
		if m.ContentType == MessageContentTypeJSON && m.Content != "" {
			err := json.Unmarshal([]byte(m.Content), &args)
			if err != nil {
				return nil, fmt.Errorf("failed to unmarshal JSON args: %w", err)
			}
		} else {
			args = map[string]any{"content": m.Content}
		}

		input, err := structpb.NewStruct(args)
		if err != nil {
			return nil, fmt.Errorf("failed to convert args to structpb: %w", err)
		}

		block = &aiv1.ContentBlock{
			BlockType: &aiv1.ContentBlock_ToolCall{
				ToolCall: &aiv1.ToolCall{
					Id:    m.ID,
					Name:  m.Tool,
					Input: input,
				},
			},
		}
	case MessageTypeResult, MessageTypeError:
		block = &aiv1.ContentBlock{
			BlockType: &aiv1.ContentBlock_ToolResult{
				ToolResult: &aiv1.ToolResult{
					Id:      m.ParentID,
					Content: m.Content,
					IsError: m.Type == MessageTypeError,
				},
			},
		}
	}

	return &aiv1.CompletionMessage{
		Role:    string(m.Role),
		Content: []*aiv1.ContentBlock{block},
	}, nil
}

// sessionCtxKey is used for saving a session in a context.
type sessionCtxKey struct{}

// GetSession retrieves a session from a context.
func GetSession(ctx context.Context) *Session {
	return ctx.Value(sessionCtxKey{}).(*Session)
}

// WithSession adds a session to a context.
func WithSession(ctx context.Context, s *Session) context.Context {
	return context.WithValue(ctx, sessionCtxKey{}, s)
}

// BaseSession contains the session implementation that is not specific to the current call.
type BaseSession struct {
	id         string
	instanceID string
	title      string
	userAgent  string
	userID     string
	claims     *runtime.SecurityClaims

	runner              *Runner
	logger              *zap.Logger
	llm                 drivers.AIService
	llmRelease          func()
	projectInstructions string

	mu          sync.RWMutex
	messages    []*Message
	subscribers map[chan *Message]struct{}
}

func (s *BaseSession) Close() error {
	// TODO: Flush messages and title to DB
	s.llmRelease()
	return nil
}

func (s *BaseSession) InstanceID() string {
	return s.instanceID
}

func (s *BaseSession) UserID() string {
	return s.userID
}

func (s *BaseSession) Claims() *runtime.SecurityClaims {
	return s.claims
}

func (s *BaseSession) Title() string {
	return s.title
}

func (s *BaseSession) UpdateTitle(ctx context.Context, title string) error {
	s.title = title
	return nil
}

func (s *BaseSession) UpdateUserAgent(ctx context.Context, userAgent string) error {
	s.userAgent = userAgent
	return nil
}

func (s *BaseSession) Subscribe() chan *Message {
	ch := make(chan *Message)
	s.mu.Lock()
	s.subscribers[ch] = struct{}{}
	s.mu.Unlock()
	return ch
}

func (s *BaseSession) Unsubscribe(ch chan *Message) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.subscribers, ch)
	close(ch)
}

func (s *BaseSession) WithParent(messageID string) *Session {
	return &Session{
		BaseSession: s,
		ParentID:    messageID,
	}
}

func (s *BaseSession) ProjectInstructions() string {
	return s.projectInstructions
}

func (s *BaseSession) SetLLM(llm drivers.AIService, release func()) {
	if s.llmRelease != nil {
		s.llmRelease()
	}
	s.llm = llm
	s.llmRelease = release
}

func (s *BaseSession) NextIndex() int {
	return len(s.messages)
}

func (s *BaseSession) Messages() []*Message {
	return s.messages
}

func (s *BaseSession) MessageByID(id string) (*Message, bool) {
	for _, msg := range s.messages {
		if msg.ID == id {
			return msg, true
		}
	}
	return nil, false
}

func (s *BaseSession) MessagesByCall(id string, nested bool) []*Message {
	var res []*Message
	var callBegun bool
	for _, msg := range s.messages {
		if msg.ID == id {
			callBegun = true
		}
		if !callBegun {
			continue
		}
		if msg.ID != id && msg.ParentID == "" {
			// Next call starts here
			break
		}

		if nested {
			res = append(res, msg)
		} else if msg.ID == id || msg.ParentID == id {
			res = append(res, msg)
		}
	}

	return res
}

func (s *BaseSession) Calls() []*Message {
	var calls []*Message
	for _, msg := range s.messages {
		if msg.ParentID == "" {
			calls = append(calls, msg)
		}
	}
	return calls
}

func (s *BaseSession) LatestCall() *Message {
	calls := s.Calls()
	if len(calls) == 0 {
		return nil
	}
	return calls[len(calls)-1]
}

func (s *BaseSession) FilterMessages() []*Message {
	// TODO: Implement predicates (filter by type, tool, actor, union or intersection, latest system message)
	return s.messages
}

// type Predicate func(*Message) bool

// func TypeFilter(t MessageType) Predicate {
// 	return func(m *Message) bool {
// 		return m.Type == t
// 	}
// }

// func ToolFilter(tool string) Predicate {
// 	return func(m *Message) bool {
// 		return m.Tool == tool
// 	}
// }

// func ActorFilter(actor string) Predicate {
// 	return func(m *Message) bool {
// 		return m.Role == actor
// 	}
// }

// func OrFilter(predicates ...Predicate) Predicate {
// 	return func(m *Message) bool {
// 		for _, p := range predicates {
// 			if p(m) {
// 				return true
// 			}
// 		}
// 		return false
// 	}
// }

// func AndFilter(predicates ...Predicate) Predicate {
// 	return func(m *Message) bool {
// 		for _, p := range predicates {
// 			if !p(m) {
// 				return false
// 			}
// 		}
// 		return true
// 	}
// }

// Session wraps a BaseSession with a reference to the current call's parent message.
type Session struct {
	*BaseSession
	ParentID string
}

// AddMessageOptions provides options for Session.AddMessage.
type AddMessageOptions struct {
	Role        Role
	Type        MessageType
	Tool        string
	ContentType MessageContentType
	Content     string
}

// AddMessage adds a message linked to the current session's parent call.
func (s *Session) AddMessage(opts *AddMessageOptions) *Message {
	msg := &Message{
		ID:          uuid.NewString(),
		ParentID:    s.ParentID,
		SessionID:   s.id,
		Time:        time.Now(),
		Index:       s.NextIndex(),
		Role:        opts.Role,
		Type:        opts.Type,
		Tool:        opts.Tool,
		ContentType: opts.ContentType,
		Content:     opts.Content,
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.messages = append(s.messages, msg)
	for sub := range s.subscribers {
		sub <- msg
	}

	return msg
}

func (s *Session) RootID() string {
	root := s.ParentID
	if root == "" {
		panic("no parent ID set")
	}
	for range 100 { // Fail-safe, not expected to reach the limit
		msg, ok := s.MessageByID(root)
		if !ok {
			panic(fmt.Errorf("failed to find referenced message with ID %q", root))
		}
		if msg.ParentID == "" {
			break
		}
		root = msg.ParentID
	}
	return root
}

func (s *Session) DefaultCompletionMessages() []*Message {
	if s.ParentID == "" {
		return nil
	}

	// Identify the current call and root call
	currentCall := s.ParentID
	rootCall := s.RootID()

	// Find the previous root calls, and their user messages and responses
	var previousRootCalls []*Message
	var callID string
	for _, msg := range s.messages {
		if msg.ID == rootCall {
			break
		}
		if msg.ParentID == "" {
			callID = msg.ID
		} else if msg.Type == MessageTypePrompt && msg.Role == RoleUser {
			previousRootCalls = append(previousRootCalls, msg)
		} else if msg.ParentID == callID && (msg.Type == MessageTypeResult || msg.Type == MessageTypeError) {
			previousRootCalls = append(previousRootCalls, msg)
		}
	}

	// Find relevant messages in the current call stack
	rootCallMessages := s.MessagesByCall(rootCall, true)

	// Find the latest system prompt
	var systemPrompt *Message
	for _, msg := range rootCallMessages {
		if msg.Role == RoleSystem && msg.Type == MessageTypePrompt {
			systemPrompt = msg
		}
	}

	// Find all user messages in the root call stack
	var userMessages []*Message
	for _, msg := range rootCallMessages {
		if msg.Role == RoleUser && msg.Type == MessageTypePrompt {
			userMessages = append(userMessages, msg)
		}
	}

	// Find all calls in the current call
	var currentCallMessages []*Message
	callID = ""
	for _, msg := range rootCallMessages {
		if msg.ParentID == currentCall && msg.Type == MessageTypeCall {
			callID = msg.ID
			currentCallMessages = append(currentCallMessages, msg)
		} else if callID != "" && msg.ParentID == callID && (msg.Type == MessageTypeResult || msg.Type == MessageTypeError) {
			currentCallMessages = append(currentCallMessages, msg)
		}
	}

	// Build the final message list
	res := []*Message{systemPrompt}
	res = append(res, previousRootCalls...)
	res = append(res, userMessages...)
	res = append(res, currentCallMessages...)
	return res
}

// CallToolResult contains the messages created during a tool call.
type CallToolResult struct {
	Call   *Message
	Result *Message
}

// CallTool runs a tool call in the current session and adds it, its result, and all messages from nested calls to the session.
func (s *Session) CallTool(ctx context.Context, role Role, toolName string, out, args any) (*CallToolResult, error) {
	var argsJSON json.RawMessage
	if args != nil {
		var err error
		argsJSON, err = json.Marshal(args)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal args: %w", err)
		}
	}

	return s.call(ctx, role, toolName, out, argsJSON, func(ctx context.Context) (any, error) {
		tool, ok := s.runner.Tools[toolName]
		if !ok {
			return nil, fmt.Errorf("unknown tool %q", toolName)
		}
		if tool.checkAccess != nil && !tool.checkAccess(s.claims) {
			return nil, fmt.Errorf("access denied to tool %q", toolName)
		}
		return tool.jsonHandler(ctx, argsJSON)
	})
}

// CallLambda runs a function call and adds it, its result, and all messages from nested calls to the session.
func (s *Session) CallLambda(ctx context.Context, role Role, anonToolName string, out any, fn func(context.Context) (any, error)) (*CallToolResult, error) {
	return s.call(ctx, role, anonToolName, out, nil, fn)
}

// call is the internal implementation for durable execution of tool/function calls.
// TODO: Implement resume where if there's a matching tool call, return immediately.
// TODO: Implement awaiting a human input, where it returns ErrAwaitInput.
func (s *Session) call(ctx context.Context, role Role, name string, out, args any, handler func(context.Context) (any, error)) (*CallToolResult, error) {
	var argsJSON json.RawMessage
	if args != nil {
		var err error
		argsJSON, err = json.Marshal(args)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal args: %w", err)
		}
	}

	callMsg := s.AddMessage(&AddMessageOptions{
		Role:        role,
		Type:        MessageTypeCall,
		Tool:        name,
		ContentType: MessageContentTypeJSON,
		Content:     string(argsJSON),
	})
	callSession := s.WithParent(callMsg.ID)
	callCtx := WithSession(ctx, callSession)

	handlerOut, handlerErr := handler(callCtx)
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	var outJSON json.RawMessage
	if handlerOut != nil {
		var err error
		outJSON, err = json.Marshal(handlerOut)
		if err != nil {
			handlerErr = fmt.Errorf("failed to marshal result: %w (out: %v)", err, handlerOut)
		}
	}

	var resultMsg *Message
	if handlerErr == nil {
		resultMsg = callSession.AddMessage(&AddMessageOptions{
			Role:        RoleTool,
			Type:        MessageTypeResult,
			Tool:        name,
			ContentType: MessageContentTypeJSON,
			Content:     string(outJSON),
		})
	} else {
		resultMsg = callSession.AddMessage(&AddMessageOptions{
			Role:        RoleTool,
			Type:        MessageTypeError,
			Tool:        name,
			ContentType: MessageContentTypeText,
			Content:     handlerErr.Error(),
		})
	}

	if out != nil && outJSON != nil {
		err := json.Unmarshal(outJSON, out)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal result: %w", err)
		}
	}

	res := &CallToolResult{
		Call:   callMsg,
		Result: resultMsg,
	}

	return res, handlerErr
}

// CompleteOptions provides options for Session.Complete.
type CompleteOptions struct {
	Messages      []*Message
	Tools         []string
	MaxIterations int
	UnwrapCall    bool
}

// Complete runs LLM completions.
// If tools are provided, it runs a completion loop involving multiple LLM invocations.
// If the output pointer is not a string and not nil, it infers the output schema using reflection and instructs the LLM to produce structured output.
func (s *Session) Complete(ctx context.Context, name string, out any, opts *CompleteOptions) error {
	// Validate max iterations and apply defaults.
	if opts.MaxIterations == 0 {
		if len(opts.Tools) > 0 {
			opts.MaxIterations = 10
		} else {
			opts.MaxIterations = 1
		}
	}
	opts.MaxIterations = 5 // TODO: Temporary

	// Prepare tool definitions.
	tools := make([]*aiv1.Tool, 0, len(opts.Tools))
	for _, toolName := range opts.Tools {
		tool, ok := s.runner.Tools[toolName]
		if !ok {
			return fmt.Errorf("unknown tool %q", toolName)
		}
		var inputSchema string
		if tool.spec.InputSchema != nil {
			inputSchemaBytes, err := json.Marshal(tool.spec.InputSchema)
			if err != nil {
				return fmt.Errorf("failed to marshal input schema for tool %q: %w", toolName, err)
			}
			inputSchema = string(inputSchemaBytes)

			// OpenAI currently does not accept object schemas without explicit properties.
			// So for now, we skip such schemas.
			if s, ok := tool.spec.InputSchema.(*jsonschema.Schema); ok && s.Properties == nil {
				inputSchema = ""
			}
		}
		tools = append(tools, &aiv1.Tool{
			Name:        tool.spec.Name,
			Description: tool.spec.Description,
			InputSchema: inputSchema,
		})
	}

	// Prepare output schema.
	var outputText bool
	var outputSchema *jsonschema.Schema
	if out != nil {
		_, isStr := out.(*string)
		if isStr {
			outputText = true
		} else {
			outType := reflect.TypeOf(out)
			if outType.Kind() != reflect.Pointer {
				return fmt.Errorf("completion output must be a pointer, got %T", out)
			}
			if outType.Elem().Kind() != reflect.Struct && outType.Elem().Kind() != reflect.Map {
				return fmt.Errorf("completion output must be a string, struct or map, got %T", out)
			}
			var err error
			outputSchema, err = jsonschema.ForType(outType.Elem(), &jsonschema.ForOptions{})
			if err != nil {
				return err
			}
		}
	}

	// Create a lambda that runs the complete loop.
	completeLoop := func(ctx context.Context) (any, error) {
		// Reload session from the new context (may be a sub-call if CallLambda was used.)
		s := GetSession(ctx)

		// Setup input messages.
		messages := make([]*aiv1.CompletionMessage, 0, len(opts.Messages))
		for _, msg := range opts.Messages {
			aimsg, err := msg.ToProto()
			if err != nil {
				return nil, err
			}
			messages = append(messages, aimsg)
		}

		// Filter out messages if there are too many
		messages = maybeTruncateMessages(messages)

		// TODO: For durable execution, add messages from current scope.

		// Complete and execute tool calls in a loop.
		var result *aiv1.CompletionMessage
		for i := range opts.MaxIterations {
			// Disable tool calls in the last iteration
			if i+1 == opts.MaxIterations {
				tools = nil
			}

			// Call the LLM to complete the messages.
			res, err := s.llm.Complete(ctx, messages, tools, outputSchema)
			if err != nil {
				return nil, fmt.Errorf("completion failed: %w", err)
			}

			// Break the tool call loop if no tool calls were requested.
			var hasCall bool
			for _, block := range res.Content {
				if call := block.GetToolCall(); call != nil {
					hasCall = true
					break
				}
			}
			if !hasCall {
				result = res
				break
			}

			// Add returned blocks as messages.
			// Run the requested tool calls.
			// TODO: How to do durable execution here?
			for _, block := range res.Content {
				switch block := block.BlockType.(type) {
				case *aiv1.ContentBlock_Text:
					s.AddMessage(&AddMessageOptions{
						Role:        RoleAssistant,
						Type:        MessageTypeProgress,
						ContentType: MessageContentTypeText,
						Content:     block.Text,
					})
				case *aiv1.ContentBlock_ToolCall:
					toolResult, _ := s.CallTool(ctx, RoleAssistant, block.ToolCall.Name, nil, block.ToolCall.Input.AsMap())
					// TODO: Err handling?
					if ctx.Err() != nil {
						return nil, ctx.Err()
					}
					callMessage, err := toolResult.Call.ToProto()
					if err != nil {
						return nil, err
					}
					resultMsg, err := toolResult.Result.ToProto()
					if err != nil {
						return nil, err
					}
					messages = append(messages, callMessage, resultMsg)
				default:
					return nil, fmt.Errorf("unexpected progress block type: %T", block)
				}
			}
		}

		// Handle the final complete result
		if result == nil {
			return nil, fmt.Errorf("completion loop did not produce a final result")
		}
		var outVal any
		for _, block := range result.Content {
			switch block := block.BlockType.(type) {
			case *aiv1.ContentBlock_Text:
				// Capture the output value as a type that we can deserialize into `out`.
				if outputText {
					outVal = block.Text
				} else if outputSchema != nil {
					outVal = json.RawMessage(block.Text)
				}

				// Add the final result message.
				contentType := MessageContentTypeText
				if outputSchema != nil {
					contentType = MessageContentTypeJSON
				}
				s.AddMessage(&AddMessageOptions{
					Role:        RoleAssistant,
					Type:        MessageTypeResult,
					ContentType: contentType,
					Content:     block.Text,
				})
			default:
				return nil, fmt.Errorf("unexpected result block type: %T", block)
			}
		}

		return outVal, nil
	}

	// The complete loop will add intermediate messages for LLM thinking and tool calls to session under the current call.
	// In some cases, it's desirable to capture these intermediate messages in the parent call's context, in other cases it's better to isolate them and only expose the final result to the parent context.
	// When UnwrapCall is true, we run the completion loop within the current call, otherwise we wrap the complete loop in a new call to isolate internal messages.
	if opts.UnwrapCall {
		outVal, err := completeLoop(ctx)
		if err != nil {
			return err
		}

		// Shim to copy `outVal` into `out` in a way consistent with how CallLambda does it.
		if out != nil {
			outJSON, err := json.Marshal(outVal)
			if err != nil {
				return fmt.Errorf("failed to marshal complete loop result: %w", err)
			}
			err = json.Unmarshal(outJSON, out)
			if err != nil {
				return fmt.Errorf("failed to unmarshal complete loop result: %w", err)
			}
		}
	} else {
		_, err := s.CallLambda(ctx, RoleSystem, name, out, completeLoop)
		if err != nil {
			return err
		}
	}

	return nil
}

// maybeTruncateMessages keeps recent messages and a few early ones for context.
// It's a simple placeholder strategy. In the future, we'll enhance this with AI summarization.
func maybeTruncateMessages(messages []*aiv1.CompletionMessage) []*aiv1.CompletionMessage {
	const (
		maxMessages = 20 // Keep up to 20 messages total
		keepFirst   = 3  // Always keep first 3 messages for context
		keepLast    = 16 // Keep last 16 messages
	)

	if len(messages) <= maxMessages {
		return messages
	}

	var result []*aiv1.CompletionMessage

	// Keep first messages
	result = append(result, messages[:keepFirst]...)

	// Add truncation indicator
	skipped := len(messages) - keepFirst - keepLast
	result = append(result, &aiv1.CompletionMessage{
		Role: "system",
		Content: []*aiv1.ContentBlock{
			{
				BlockType: &aiv1.ContentBlock_Text{
					Text: fmt.Sprintf("... [%d messages omitted for brevity] ...", skipped),
				},
			},
		},
	})

	// Keep last messages
	start := len(messages) - keepLast
	result = append(result, messages[start:]...)

	return result
}
