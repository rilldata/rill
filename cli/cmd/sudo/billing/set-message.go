package billing

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func SetMessageCmd(ch *cmdutil.Helper) *cobra.Command {
	var org, level, message string
	levels := []string{"warning", "error"}
	cmd := &cobra.Command{
		Use:   "set-message",
		Short: "Set (overriding any existing) a message banner for an organization",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			client, err := ch.Client()
			if err != nil {
				return err
			}

			if org == "" {
				return fmt.Errorf("please set --org")
			}

			if message == "" {
				return fmt.Errorf("please set --message")
			}

			if level == "" {
				if !ch.Interactive {
					return fmt.Errorf("--level flag is required in non-interactive mode")
				}
				level, err = cmdutil.SelectPrompt("Select message level", levels, "warning")
				if err != nil {
					return err
				}
			}

			var l adminv1.BillingIssueLevel
			switch level {
			case "warning":
				l = adminv1.BillingIssueLevel_BILLING_ISSUE_LEVEL_WARNING
			case "error":
				l = adminv1.BillingIssueLevel_BILLING_ISSUE_LEVEL_ERROR
			default:
				return fmt.Errorf("invalid level %q, must be one of: warning, error", level)
			}

			_, err = client.SudoUpdateOrganizationBillingMessage(ctx, &adminv1.SudoUpdateOrganizationBillingMessageRequest{
				Org:     org,
				Level:   l,
				Message: message,
			})
			if err != nil {
				return err
			}

			ch.PrintfSuccess("Message banner set for organization %q\n", org)

			return nil
		},
	}

	cmd.Flags().StringVar(&org, "org", "", "Organization Name")
	cmd.Flags().StringVar(&level, "level", "", "Message level (warning or error)")
	cmd.Flags().StringVar(&message, "message", "", "Message to display in the banner")
	return cmd
}
