package subscription

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func EditCmd(ch *cmdutil.Helper) *cobra.Command {
	var plan string

	editCmd := &cobra.Command{
		Use:   "edit",
		Args:  cobra.NoArgs,
		Short: "Edit organization subscription",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			client, err := ch.Client()
			if err != nil {
				return err
			}

			subResp, err := client.GetBillingSubscription(ctx, &adminv1.GetBillingSubscriptionRequest{
				OrgName: ch.Org,
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

			ch.PrintfWarn("Editing plan for organization %q. Plan change will take place immediately.", ch.Org)
			ok, err := cmdutil.ConfirmPrompt("Do you want to Continue ?", "", false)
			if err != nil {
				return err
			}
			if !ok {
				ch.PrintfWarn("Aborted\n")
				return nil
			}

			resp, err := client.UpdateBillingSubscription(cmd.Context(), &adminv1.UpdateBillingSubscriptionRequest{
				OrgName:  ch.Org,
				PlanName: plan,
			})
			if err != nil {
				return err
			}

			ch.PrintfSuccess("Successfully subscribed to plan %q for org %q\n", plan, ch.Org)
			ch.PrintSubscriptions(resp.Subscriptions)
			return nil
		},
	}
	editCmd.Flags().StringVar(&plan, "plan", "", "Plan Name to change subscription to")

	return editCmd
}
