package ai

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rilldata/rill/runtime"
)

const DeleteReportName = "delete_report"

// DeleteReport deletes one of the current user's own scheduled reports, once they confirmed it.
type DeleteReport struct {
	Runtime *runtime.Runtime
}

var _ Tool[*DeleteResourceArgs, *DeleteResourceResult] = (*DeleteReport)(nil)

func (t *DeleteReport) Spec() *mcp.Tool {
	return &mcp.Tool{
		Name:  DeleteReportName,
		Title: "Delete scheduled report",
		Description: "Deletes one of the current user's own scheduled reports, as listed by list_reports. " +
			"The first call only returns 'confirmation_required': ask the user to confirm that they want to delete the report, and end your turn. " +
			"Once they confirm in their next message, call it again with confirmed set to true to delete it.",
		Annotations: &mcp.ToolAnnotations{
			DestructiveHint: boolPtr(true),
			IdempotentHint:  false,
			OpenWorldHint:   boolPtr(false),
			ReadOnlyHint:    false,
		},
		Meta: map[string]any{
			"openai/toolInvocation/invoking": "Deleting report...",
			"openai/toolInvocation/invoked":  "Deleted report",
		},
	}
}

func (t *DeleteReport) CheckAccess(ctx context.Context) (bool, error) {
	return checkAlertReportAccess(ctx, t.Runtime)
}

func (t *DeleteReport) Handler(ctx context.Context, args *DeleteResourceArgs) (*DeleteResourceResult, error) {
	admin, release, err := t.Runtime.Admin(ctx, GetSession(ctx).InstanceID())
	if err != nil {
		return nil, err
	}
	defer release()
	return deleteOwnedResource(ctx, t.Runtime, DeleteReportName, runtime.ResourceKindReport, args, func(userID, name string) error {
		if err := admin.DeleteReportForUser(ctx, userID, name); err != nil {
			return fmt.Errorf("failed to delete report: %w", err)
		}
		return nil
	})
}
