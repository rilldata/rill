package ai

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rilldata/rill/runtime"
)

const DeleteAlertName = "delete_alert"

// DeleteAlert deletes one of the current user's own alerts, once they confirmed it.
type DeleteAlert struct {
	Runtime *runtime.Runtime
}

var _ Tool[*DeleteResourceArgs, *DeleteResourceResult] = (*DeleteAlert)(nil)

func (t *DeleteAlert) Spec() *mcp.Tool {
	return &mcp.Tool{
		Name:  DeleteAlertName,
		Title: "Delete alert",
		Description: "Deletes one of the current user's own alerts, as listed by list_alerts. " +
			"The first call only returns 'confirmation_required': ask the user to confirm that they want to delete the alert, and end your turn. " +
			"Once they confirm in their next message, call it again with confirmed set to true to delete it.",
		Annotations: &mcp.ToolAnnotations{
			DestructiveHint: boolPtr(true),
			IdempotentHint:  false,
			OpenWorldHint:   boolPtr(false),
			ReadOnlyHint:    false,
		},
		Meta: map[string]any{
			"openai/toolInvocation/invoking": "Deleting alert...",
			"openai/toolInvocation/invoked":  "Deleted alert",
		},
	}
}

func (t *DeleteAlert) CheckAccess(ctx context.Context) (bool, error) {
	return checkAlertReportAccess(ctx, t.Runtime)
}

func (t *DeleteAlert) Handler(ctx context.Context, args *DeleteResourceArgs) (*DeleteResourceResult, error) {
	admin, release, err := t.Runtime.Admin(ctx, GetSession(ctx).InstanceID())
	if err != nil {
		return nil, err
	}
	defer release()
	return deleteOwnedResource(ctx, t.Runtime, DeleteAlertName, runtime.ResourceKindAlert, args, func(userID, name string) error {
		if err := admin.DeleteAlertForUser(ctx, userID, name); err != nil {
			return fmt.Errorf("failed to delete alert: %w", err)
		}
		return nil
	})
}
