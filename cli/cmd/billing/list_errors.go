package billing

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func ListErrorsCmd(ch *cmdutil.Helper) *cobra.Command {
	listCmd := &cobra.Command{
		Use:   "list-errors",
		Short: "List billing errors for an organization",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := ch.Client()
			if err != nil {
				return err
			}

			resp, err := client.ListOrganizationBillingErrors(cmd.Context(), &adminv1.ListOrganizationBillingErrorsRequest{
				Organization: ch.Org,
			})
			if err != nil {
				return err
			}

			if len(resp.Errors) == 0 {
				ch.PrintfSuccess("No billing errors for organization %q.\n", ch.Org)
				return nil
			}

			ch.PrintBillingErrors(resp.Errors)
			return nil
		},
	}

	listCmd.Flags().StringVar(&ch.Org, "org", ch.Org, "Organization Name")
	return listCmd
}
