package billing

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func ListWarningsCmd(ch *cmdutil.Helper) *cobra.Command {
	listCmd := &cobra.Command{
		Use:   "list-warnings",
		Short: "List billing warnings for an organization",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := ch.Client()
			if err != nil {
				return err
			}

			resp, err := client.ListOrganizationBillingWarnings(cmd.Context(), &adminv1.ListOrganizationBillingWarningsRequest{
				Organization: ch.Org,
			})
			if err != nil {
				return err
			}

			if len(resp.Warnings) == 0 {
				ch.PrintfSuccess("No billing warnings for organization %q.\n", ch.Org)
				return nil
			}

			ch.PrintBillingWarnings(resp.Warnings)
			return nil
		},
	}

	listCmd.Flags().StringVar(&ch.Org, "org", ch.Org, "Organization Name")
	return listCmd
}
