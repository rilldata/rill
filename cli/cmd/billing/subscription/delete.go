package subscription

import (
	"fmt"

	"github.com/rilldata/rill/cli/cmd/org"
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func DeleteCmd(ch *cmdutil.Helper) *cobra.Command {
	var orgName, subID, effective string
	effectiveOptions := []string{"end-of-billing-cycle", "immediate"}

	deleteCmd := &cobra.Command{
		Use:   "delete",
		Args:  cobra.NoArgs,
		Short: "delete organization subscription",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			if subID == "" {
				return fmt.Errorf("subscription-id is required")
			}

			client, err := ch.Client()
			if err != nil {
				return err
			}

			if !cmd.Flags().Changed("org") && ch.Interactive {
				orgNames, err := org.OrgNames(ctx, ch)
				if err != nil {
					return err
				}

				orgName, err = cmdutil.SelectPrompt("Select org to delete subscription", orgNames, ch.Org)
				if err != nil {
					return err
				}
			}

			if !cmd.Flags().Changed("effective") && ch.Interactive {
				effective, err = cmdutil.SelectPrompt("Select effective time of cancellation", effectiveOptions, effectiveOptions[0])
				if err != nil {
					return err
				}
			}

			var cancelEffective adminv1.SubscriptionCancelEffective
			if effective == effectiveOptions[0] {
				cancelEffective = adminv1.SubscriptionCancelEffective_SUBSCRIPTION_CANCEL_EFFECTIVE_END_OF_BILLING_CYCLE
			} else {
				cancelEffective = adminv1.SubscriptionCancelEffective_SUBSCRIPTION_CANCEL_EFFECTIVE_NOW
			}

			_, err = client.DeleteOrganizationSubscription(ctx, &adminv1.DeleteOrganizationSubscriptionRequest{
				OrgName:                     orgName,
				SubscriptionId:              subID,
				SubscriptionCancelEffective: cancelEffective,
			})
			if err != nil {
				return err
			}

			ch.PrintfSuccess("Successfully deleted subscription %q for org %q\n", subID, orgName)

			return nil
		},
	}
	deleteCmd.Flags().SortFlags = false
	deleteCmd.Flags().StringVar(&orgName, "org", ch.Org, "Organization name")
	deleteCmd.Flags().StringVar(&subID, "subscription-id", "", "Subscription ID to cancel")
	deleteCmd.Flags().StringVar(&effective, "effective", effectiveOptions[0], "Effective time of cancellation")

	return deleteCmd
}
