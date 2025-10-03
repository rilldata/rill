package billing

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func ListIssuesCmd(ch *cmdutil.Helper) *cobra.Command {
	var force bool

	listCmd := &cobra.Command{
		Use:   "list-issues",
		Short: "List billing issues for an organization",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := ch.Client()
			if err != nil {
				return err
			}

			resp, err := client.ListOrganizationBillingIssues(cmd.Context(), &adminv1.ListOrganizationBillingIssuesRequest{
				Org:                  ch.Org,
				SuperuserForceAccess: force,
			})
			if err != nil {
				return err
			}

			if len(resp.Issues) == 0 {
				ch.PrintfSuccess("No billing issues for organization %q.\n", ch.Org)
				return nil
			}

			ch.PrintBillingIssues(resp.Issues)
			return nil
		},
	}

	listCmd.Flags().BoolVar(&force, "force", false, "Allows superusers to bypass certain checks")
	_ = listCmd.Flags().MarkHidden("force")

	listCmd.Flags().StringVar(&ch.Org, "org", ch.Org, "Organization Name")
	return listCmd
}
