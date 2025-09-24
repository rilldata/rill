package subscription

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func CancelCmd(ch *cmdutil.Helper) *cobra.Command {
	var force bool

	cancelCmd := &cobra.Command{
		Use:   "cancel",
		Short: "Cancel subscription for an organization",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := ch.Client()
			if err != nil {
				return err
			}

			subResp, err := client.GetBillingSubscription(cmd.Context(), &adminv1.GetBillingSubscriptionRequest{
				Org:                  ch.Org,
				SuperuserForceAccess: force,
			})
			if err != nil {
				return err
			}

			if subResp.Subscription == nil {
				ch.PrintfWarn("No subscriptions found for organization %q\n", ch.Org)
				return nil
			}

			ch.PrintfBold("Organization %q has following subscription\n", ch.Org)
			ch.PrintSubscriptions([]*adminv1.Subscription{subResp.Subscription})

			ch.PrintfWarn("\nAt the end of the current billing cycle, you will lose access to %q and all its projects.\n", ch.Org)
			ch.PrintfWarn("\nIf you want to change the plan, please use `rill billing subscription edit` command.\n")
			ok, err := cmdutil.ConfirmPrompt("Do you want to continue?", "", false)
			if err != nil {
				return err
			}
			if !ok {
				ch.PrintfWarn("Aborted\n")
				return nil
			}

			_, err = client.CancelBillingSubscription(cmd.Context(), &adminv1.CancelBillingSubscriptionRequest{
				Org:                  ch.Org,
				SuperuserForceAccess: force,
			})
			if err != nil {
				return err
			}

			ch.PrintfWarn("Subscription cancelled successfully\n")
			return nil
		},
	}

	cancelCmd.Flags().BoolVar(&force, "force", false, "Allows superusers to bypass certain checks")
	_ = cancelCmd.Flags().MarkHidden("force")

	return cancelCmd
}
