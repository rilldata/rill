package subscription

import (
	"github.com/rilldata/rill/cli/cmd/org"
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func CreateCmd(ch *cmdutil.Helper) *cobra.Command {
	var orgName, plan string

	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create subscription for an organization",
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

				orgName, err = cmdutil.SelectPrompt("Select org to add subscription", orgNames, ch.Org)
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

			if subResp.Subscription != nil {
				ch.PrintfWarn("Organization already has following subscription, use `rill billing subscription edit` to update\n")
				ch.PrintSubscriptions([]*adminv1.Subscription{subResp.Subscription})
				return nil
			}

			resp, err := client.UpdateBillingSubscription(cmd.Context(), &adminv1.UpdateBillingSubscriptionRequest{
				OrgName:  orgName,
				PlanName: plan,
			})
			if err != nil {
				return err
			}

			ch.PrintfSuccess("Successfully subscribed to plan %q for org %q with billing customer ID %q\n", plan, orgName, resp.Organization.BillingCustomerId)
			ch.PrintSubscriptions(resp.Subscriptions)
			return nil
		},
	}
	createCmd.Flags().SortFlags = false
	createCmd.Flags().StringVar(&orgName, "org", ch.Org, "Organization Name")
	createCmd.Flags().StringVar(&plan, "plan", "", "Plan Name to subscribe to")
	return createCmd
}
