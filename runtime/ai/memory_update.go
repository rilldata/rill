package ai

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
)

// UpdateMemoryName is the name of the tool that lets the agents remember or forget something about the user.
const UpdateMemoryName = "update_memory"

// UpdateMemory is the tool through which the AI writes to the user's memory.
// The agents call it when the user states a durable preference or explicitly asks the AI to remember or forget something.
// A memory is scoped to the agent that learned it, so the analyst and the developer agent keep separate memories.
type UpdateMemory struct {
	Runtime *runtime.Runtime
}

var _ Tool[*UpdateMemoryArgs, *MemoryUpdateResult] = (*UpdateMemory)(nil)

type UpdateMemoryArgs struct {
	Action   string `json:"action" jsonschema:"The operation to perform.,enum=add,enum=update,enum=delete"`
	MemoryID string `json:"memory_id,omitempty" jsonschema:"ID of an existing memory. Required for update and delete."`
	Content  string `json:"content,omitempty" jsonschema:"The memory as a short standalone statement about the user (max 300 characters). Required for add and update."`
}

// MemoryUpdateResult is the outcome of update_memory: the operation that was applied, or a noop with the reason the operation was rejected.
// The frontend renders applied operations as a "Memory updated" notice.
type MemoryUpdateResult struct {
	Op       string `json:"op" jsonschema:"The operation that was applied: add, update or delete; noop if nothing changed."`
	MemoryID string `json:"memory_id,omitempty" jsonschema:"ID of the memory that was added, updated or deleted."`
	Content  string `json:"content,omitempty" jsonschema:"The memory's content after the operation, or the content that was deleted."`
	Reason   string `json:"reason,omitempty" jsonschema:"Why the operation was rejected."`
}

func (t *UpdateMemory) Spec() *mcp.Tool {
	return &mcp.Tool{
		Name:  UpdateMemoryName,
		Title: "Update Memory",
		Description: `Remember or forget something about the user for future conversations in this project.
Use it when the user explicitly asks you to remember or forget something, or states a durable preference: how they like results formatted, what a business term means to them, which metrics views they care about, or a correction to your behavior.
Do not use it for data values or query results, secrets or credentials, sensitive personal details, one-off requests that only apply to the current question, or anything the project instructions already say.
To update or delete, use the memory ID shown in the "user_memory" section of your instructions.`,
		Annotations: &mcp.ToolAnnotations{
			DestructiveHint: boolPtr(true),
			IdempotentHint:  false,
			OpenWorldHint:   boolPtr(false),
			ReadOnlyHint:    false,
		},
		Meta: map[string]any{
			"openai/toolInvocation/invoking": "Updating memory...",
			"openai/toolInvocation/invoked":  "Updated memory",
		},
	}
}

func (t *UpdateMemory) CheckAccess(ctx context.Context) (bool, error) {
	s := GetSession(ctx)
	if !s.Claims().Can(runtime.UseAI) {
		return false, nil
	}
	// MemoryEnabled already covers the feature flag, the Rill user agent, anonymous users, and the user's own memory setting.
	return s.MemoryEnabled(), nil
}

// Handler validates and applies the operation for the session's user. It is deterministic and never calls an LLM.
// A rejected operation is returned as a noop with a reason rather than as an error, so the agent can tell the user what happened.
func (t *UpdateMemory) Handler(ctx context.Context, args *UpdateMemoryArgs) (*MemoryUpdateResult, error) {
	s := GetSession(ctx)
	if !s.MemoryEnabled() {
		return nil, errors.New("memory is not enabled for this session")
	}
	ownerID := s.Claims().UserID

	catalog, release, err := s.acquireCatalog(ctx)
	if err != nil {
		return nil, err
	}
	defer release()

	existing, err := catalog.FindAIMemories(ctx, ownerID)
	if err != nil {
		return nil, err
	}
	byID := make(map[string]*drivers.AIMemory, len(existing))
	for _, m := range existing {
		byID[m.ID] = m
	}

	content := SanitizeMemoryContent(args.Content)
	res := &MemoryUpdateResult{Op: args.Action, MemoryID: args.MemoryID, Content: content}

	switch args.Action {
	case MemoryOpAdd:
		if reason := validateMemoryContent(content); reason != "" {
			return rejectMemoryOp(res, reason), nil
		}
		for _, m := range existing {
			if normalizeMemoryContent(m.Content) == normalizeMemoryContent(content) {
				res.MemoryID = m.ID
				return rejectMemoryOp(res, "already remembered"), nil
			}
		}
		if len(existing) >= MaxMemories {
			return rejectMemoryOp(res, fmt.Sprintf("memory limit of %d reached; ask the user to remove memories they no longer need", MaxMemories)), nil
		}
		now := time.Now()
		m := &drivers.AIMemory{
			ID:              uuid.NewString(),
			InstanceID:      s.InstanceID(),
			OwnerID:         ownerID,
			Agent:           s.agentName(),
			Content:         content,
			SourceSessionID: s.ID(),
			CreatedOn:       now,
			UpdatedOn:       now,
		}
		if err := catalog.InsertAIMemory(ctx, m); err != nil {
			return nil, err
		}
		res.MemoryID = m.ID

	case MemoryOpUpdate:
		m, ok := byID[args.MemoryID]
		if !ok {
			return rejectMemoryOp(res, "memory not found"), nil
		}
		if reason := validateMemoryContent(content); reason != "" {
			return rejectMemoryOp(res, reason), nil
		}
		if m.Content == content {
			return rejectMemoryOp(res, "no change"), nil
		}
		m.Content = content
		m.SourceSessionID = s.ID()
		if err := catalog.UpdateAIMemory(ctx, m); err != nil {
			return nil, err
		}

	case MemoryOpDelete:
		m, ok := byID[args.MemoryID]
		if !ok {
			return rejectMemoryOp(res, "memory not found"), nil
		}
		if err := catalog.DeleteAIMemory(ctx, m.ID); err != nil {
			return nil, err
		}
		// Echo what was removed so the notice in the chat can show it.
		res.Content = m.Content

	default:
		return rejectMemoryOp(res, fmt.Sprintf("unknown action %q", args.Action)), nil
	}

	// Refresh the session's copy so later turns in this request see the change.
	memories, err := catalog.FindAIMemories(ctx, ownerID)
	if err != nil {
		return nil, err
	}
	s.setUserMemories(memories, true)

	return res, nil
}

func rejectMemoryOp(res *MemoryUpdateResult, reason string) *MemoryUpdateResult {
	res.Op = MemoryOpNoop
	res.Reason = reason
	return res
}
