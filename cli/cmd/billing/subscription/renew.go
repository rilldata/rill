package subscription

import (
	"fmt"
	"strings"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func RenewCmd(ch *cmdutil.Helper) *cobra.Command {
	var plan string
	var force bool

	renewCmd := &cobra.Command{
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
				ch.PrintfBold("Organization %q has the following subscription\n", ch.Org)
				ch.PrintSubscriptions([]*adminv1.Subscription{subResp.Subscription})

				ch.PrintfWarn("\nSubscription renewal for %q will take place immediately.\n", ch.Org)
				ch.PrintfWarn("\nTo edit the plan of non-cancelled subscription, run `rill billing subscription edit`.\n")
				ok, err := cmdutil.ConfirmPrompt("Do you want to continue?", "", false)
				if err != nil {
					return err
				}
				if !ok {
					ch.PrintfWarn("Aborted\n")
					return nil
				}
			}

			if cmd.Flags().Changed("plan") {
				valid := IsValidPlan(plan)
				if !valid {
					return fmt.Errorf("invalid plan %q, must be one of: %s", plan, strings.Join(AllPlans, ", "))
				}
			} else {
				// Only show plan selection prompt if in interactive mode
				if !ch.Interactive {
					return fmt.Errorf("--plan flag is required when running in non-interactive mode. Valid plans: %s",
						strings.Join(AllPlans, ", "))
				}
				// Prompt for plan if not provided via flag
				selectedPlan, err := cmdutil.SelectPrompt("Select plan:", AllPlans, "")
				if err != nil {
					return err
				}
				plan = selectedPlan
			}

			resp, err := client.RenewBillingSubscription(cmd.Context(), &adminv1.RenewBillingSubscriptionRequest{
				Organization:         ch.Org,
				PlanName:             plan,
				SuperuserForceAccess: force,
			})
			if err != nil {
				return err
			}

			ch.PrintfSuccess("Successfully renewed subscription to plan %q for org %q.\n", plan, ch.Org)
			ch.PrintSubscriptions([]*adminv1.Subscription{resp.Subscription})
			return nil
		},
	}
	renewCmd.Flags().StringVar(&plan, "plan", "", "Plan name to renew subscription to")
	renewCmd.Flags().BoolVar(&force, "force", false, "Allows superusers to bypass certain checks")
	_ = renewCmd.Flags().MarkHidden("force")

	return renewCmd
}
