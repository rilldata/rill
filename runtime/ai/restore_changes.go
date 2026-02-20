package ai

import (
	"context"
	"fmt"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rilldata/rill/runtime"
)

const RestoreChangesName = "restore_changes"

type RestoreChanges struct {
	Runtime *runtime.Runtime
}

var _ Tool[*RestoreChangesArgs, *RestoreChangesResult] = (*RestoreChanges)(nil)

type RestoreChangesArgs struct {
	RevertTillWriteCallID string `json:"revert_till_write_call_id" jsonschema:"Revert changes until the given write call ID."`
}

type RestoreChangesResult struct{}

func (t *RestoreChanges) Spec() *mcp.Tool {
	return &mcp.Tool{
		Name:        RestoreChangesName,
		Title:       "Restore Changes",
		Description: "Restore changes made by AI until a given checkpoint.",
		Meta: map[string]any{
			"openai/toolInvocation/invoking": "Restoring changes...",
			"openai/toolInvocation/invoked":  "Restored changes",
		},
	}
}

func (t *RestoreChanges) CheckAccess(ctx context.Context) (bool, error) {
	// Must be allowed to use AI features
	s := GetSession(ctx)
	if !s.Claims().Can(runtime.UseAI) {
		return false, nil
	}

	// Only allow for rill user agents since it's not functional in MCP contexts.
	if !strings.HasPrefix(s.CatalogSession().UserAgent, "rill") {
		return false, nil
	}
	return true, nil
}

func (t *RestoreChanges) Handler(ctx context.Context, args *RestoreChangesArgs) (*RestoreChangesResult, error) {
	s := GetSession(ctx)

	fileWriteResMsg, ok := s.Message(FilterByParent(args.RevertTillWriteCallID))
	if !ok {
		return nil, fmt.Errorf("write call ID %q not found", args.RevertTillWriteCallID)
	}

	if fileWriteResMsg.Tool != WriteFileName || fileWriteResMsg.Type != MessageTypeResult {
		return nil, fmt.Errorf("write call ID %q refers to invalid tool call", args.RevertTillWriteCallID)
	}

	rawFileWriteRes, err := s.UnmarshalMessageContent(fileWriteResMsg)
	if err != nil {
		return nil, fmt.Errorf("failed to parse write call result: %w", err)
	}
	var fileWriteRes WriteFileResult
	err = mapstructure.WeakDecode(rawFileWriteRes, &fileWriteRes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse write call result: %w", err)
	}
	if fileWriteRes.CheckpointCommitHash == "" {
		return nil, fmt.Errorf("write call ID %q did create a checkpoint", args.RevertTillWriteCallID)
	}

	repo, release, err := t.Runtime.Repo(ctx, s.InstanceID())
	if err != nil {
		return nil, err
	}
	defer release()

	_, err = repo.RestoreCommit(ctx, fileWriteRes.CheckpointCommitHash, true)
	if err != nil {
		return nil, err
	}

	return &RestoreChangesResult{}, nil
}
