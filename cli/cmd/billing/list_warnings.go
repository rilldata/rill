package billing

import (
	"github.com/rilldata/rill/cli/cmd/org"
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func ListWarningsCmd(ch *cmdutil.Helper) *cobra.Command {
	var orgName string

	listCmd := &cobra.Command{
		Use:   "list-warnings",
		Short: "List billing warnings for an organization",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			client, err := ch.Client()
			if err != nil {
				return err
			}

			if !cmd.Flags().Changed("org") && ch.Interactive {
				orgNames, err := org.OrgNames(ctx, ch)
				if err != nil {
					return err
				}

				orgName, err = cmdutil.SelectPrompt("Select org to list billing warnings", orgNames, ch.Org)
				if err != nil {
					return err
				}
			}

			resp, err := client.ListOrganizationBillingWarnings(cmd.Context(), &adminv1.ListOrganizationBillingWarningsRequest{
				Organization: orgName,
			})
			if err != nil {
				return err
			}

			if len(resp.Warnings) == 0 {
				ch.PrintfSuccess("No billing warnings for organization %q.\n", orgName)
				return nil
			}

			ch.PrintBillingWarnings(resp.Warnings)
			return nil
		},
	}
	listCmd.Flags().SortFlags = false
	listCmd.Flags().StringVar(&orgName, "org", ch.Org, "Organization Name")
	return listCmd
}
