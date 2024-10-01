package subscription

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func RenewCmd(ch *cmdutil.Helper) *cobra.Command {
	var plan string

	editCmd := &cobra.Command{
		Use:   "renew",
		Args:  cobra.NoArgs,
		Short: "Renew cancelled organization subscription",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			client, err := ch.Client()
			if err != nil {
				return err
			}

			subResp, err := client.GetBillingSubscription(ctx, &adminv1.GetBillingSubscriptionRequest{
				Organization: ch.Org,
			})
			if err != nil {
				return err
			}

			if subResp.Subscription != nil {
				ch.PrintfBold("Organization %q has following subscription\n", ch.Org)
				ch.PrintSubscriptions([]*adminv1.Subscription{subResp.Subscription})

				ch.PrintfWarn("\nSubscription renewal for %q will take place immediately.\n", ch.Org)
				ch.PrintfWarn("\nTo edit plan for non cancelled subscription, please use `rill billing subscription edit` command.\n")
				ok, err := cmdutil.ConfirmPrompt("Do you want to Continue ?", "", false)
				if err != nil {
					return err
				}
				if !ok {
					ch.PrintfWarn("Aborted\n")
					return nil
				}
			}

			resp, err := client.RenewBillingSubscription(cmd.Context(), &adminv1.RenewBillingSubscriptionRequest{
				Organization: ch.Org,
				PlanName:     plan,
			})
			if err != nil {
				return err
			}

			ch.PrintfSuccess("Successfully renewed subscription with plan %q for org %q\n", plan, ch.Org)
			ch.PrintSubscriptions(resp.Subscriptions)
			return nil
		},
	}
	editCmd.Flags().StringVar(&plan, "plan", "", "Plan Name to renew subscription to")

	return editCmd
}
