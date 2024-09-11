package subscription

import (
	"github.com/rilldata/rill/cli/cmd/org"
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func CancelCmd(ch *cmdutil.Helper) *cobra.Command {
	var orgName string

	cancelCmd := &cobra.Command{
		Use:   "cancel",
		Short: "Cancel subscription for an organization",
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

				orgName, err = cmdutil.SelectPrompt("Select org to cancel subscription", orgNames, ch.Org)
				if err != nil {
					return err
				}
			}

			subResp, err := client.GetBillingSubscription(cmd.Context(), &adminv1.GetBillingSubscriptionRequest{
				OrgName: orgName,
			})
			if err != nil {
				return err
			}

			if subResp.Subscription == nil {
				ch.PrintfWarn("No subscriptions found for organization %q\n", orgName)
				return nil
			}

			ch.PrintfBold("Organization has following subscription\n")
			ch.PrintSubscriptions([]*adminv1.Subscription{subResp.Subscription})

			ch.PrintfWarn("\nAt the end of the current billing cycle, you will lose access to %q and all its projects.", ch.Org)
			ok, err := cmdutil.ConfirmPrompt("Do you want to Continue ?", "", false)
			if err != nil {
				return err
			}
			if !ok {
				ch.PrintfWarn("Aborted\n")
				return nil
			}

			_, err = client.CancelBillingSubscription(cmd.Context(), &adminv1.CancelBillingSubscriptionRequest{
				OrgName: orgName,
			})
			if err != nil {
				return err
			}

			ch.PrintfWarn("Subscription cancelled successfully\n")
			return nil
		},
	}
	cancelCmd.Flags().SortFlags = false
	cancelCmd.Flags().StringVar(&orgName, "org", ch.Org, "Organization Name")
	return cancelCmd
}
