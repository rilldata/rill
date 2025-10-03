package subscription

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func EditCmd(ch *cmdutil.Helper) *cobra.Command {
	var plan string
	var force bool

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
				Org:                  ch.Org,
				SuperuserForceAccess: force,
			})
			if err != nil {
				return err
			}

			if subResp.Subscription == nil {
				ch.PrintfWarn("Organization %q has no subscription\n", ch.Org)
			} else {
				ch.PrintfBold("Organization %q has the following subscription\n", ch.Org)
				ch.PrintSubscriptions([]*adminv1.Subscription{subResp.Subscription})
			}

			ch.PrintfWarn("\nEditing plan for organization %q. Plan change will take place immediately.\n", ch.Org)
			ch.PrintfWarn("\nTo renew a cancelled subscription, please use `rill billing subscription renew` command.\n")
			ok, err := cmdutil.ConfirmPrompt("Do you want to continue?", "", false)
			if err != nil {
				return err
			}
			if !ok {
				ch.PrintfWarn("Aborted\n")
				return nil
			}

			resp, err := client.UpdateBillingSubscription(cmd.Context(), &adminv1.UpdateBillingSubscriptionRequest{
				Org:                  ch.Org,
				PlanName:             plan,
				SuperuserForceAccess: force,
			})
			if err != nil {
				return err
			}

			ch.PrintfSuccess("Successfully subscribed to plan %q for org %q\n", plan, ch.Org)
			ch.PrintSubscriptions([]*adminv1.Subscription{resp.Subscription})
			return nil
		},
	}
	editCmd.Flags().StringVar(&plan, "plan", "", "Plan Name to change subscription to")
	editCmd.Flags().BoolVar(&force, "force", false, "Allows superusers to bypass certain checks")
	_ = editCmd.Flags().MarkHidden("force")

	return editCmd
}
