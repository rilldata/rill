package billing

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func DeleteMessageCmd(ch *cmdutil.Helper) *cobra.Command {
	var org string
	cmd := &cobra.Command{
		Use:   "delete-message",
		Short: "Remove the message banner for an organization",
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

			_, err = client.SudoDeleteOrganizationBillingMessage(ctx, &adminv1.SudoDeleteOrganizationBillingMessageRequest{
				Org: org,
			})
			if err != nil {
				return err
			}

			ch.PrintfSuccess("Message banner removed for organization %q\n", org)

			return nil
		},
	}

	cmd.Flags().StringVar(&org, "org", "", "Organization Name")
	return cmd
}
