package billing

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func SetupCmd(ch *cmdutil.Helper) *cobra.Command {
	var org string
	var update bool
	setupCmd := &cobra.Command{
		Use:   "setup",
		Short: "Setup billing information returns a Stripe setup page to collect billing information like payment method and billing address for the organization",
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

			res, err := client.GetPaymentsPortalURL(ctx, &adminv1.GetPaymentsPortalURLRequest{
				Org:                  org,
				SuperuserForceAccess: true,
			})
			if err != nil {
				return err
			}

			ch.PrintfSuccess("Setup URL: %s\n", res.Url)

			return nil
		},
	}

	setupCmd.Flags().StringVar(&org, "org", "", "Organization Name")
	setupCmd.Flags().BoolVar(&update, "update", false, "url for updating the billing information")
	return setupCmd
}
