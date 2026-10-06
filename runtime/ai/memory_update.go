package ai

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
)

// UpdateMemoryName is the name of the tool that lets the agents explicitly remember or forget something about the user.
const UpdateMemoryName = "update_memory"

// UpdateMemory is the hot-path memory tool.
// The agents call it when the user states a durable preference or explicitly asks the AI to remember or forget something.
// Background extraction (extract_memories) covers everything the user did not say explicitly.
type UpdateMemory struct {
	Runtime *runtime.Runtime
}

var _ Tool[*UpdateMemoryArgs, *MemoryUpdateResult] = (*UpdateMemory)(nil)

type UpdateMemoryArgs struct {
	Action   string `json:"action" jsonschema:"The operation to perform.,enum=add,enum=update,enum=delete"`
	MemoryID string `json:"memory_id,omitempty" jsonschema:"ID of an existing memory. Required for update and delete."`
	Category string `json:"category,omitempty" jsonschema:"Category of the memory. Required for add.,enum=preference,enum=definition,enum=context,enum=feedback"`
	Content  string `json:"content,omitempty" jsonschema:"The memory as a short standalone statement about the user (max 300 characters). Required for add and update."`
}

// MemoryUpdateResult is the result of update_memory and of background extraction.
// The frontend renders the applied ops as a "Memory updated" notice and uses the previous values to undo them.
type MemoryUpdateResult struct {
	Ops []MemoryOp `json:"ops"`
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
	// MemoryFormationEnabled already covers the feature flag, the Rill user agent, anonymous users, pause, and the per-conversation opt-out.
	return s.MemoryFormationEnabled(), nil
}

func (t *UpdateMemory) Handler(ctx context.Context, args *UpdateMemoryArgs) (*MemoryUpdateResult, error) {
	s := GetSession(ctx)

	ops, err := applyMemoryOps(ctx, s, []MemoryOp{{
		Op:       args.Action,
		MemoryID: args.MemoryID,
		Category: args.Category,
		Content:  args.Content,
	}}, drivers.AIMemorySourceExplicit, s.ParentID)
	if err != nil {
		return nil, err
	}

	return &MemoryUpdateResult{Ops: ops}, nil
}
