package billing

import (
	"fmt"
	"time"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func SetCmd(ch *cmdutil.Helper) *cobra.Command {
	var org, customerID string
	setCmd := &cobra.Command{
		Use:   "set",
		Short: "Set billing information for customers",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			client, err := ch.Client()
			if err != nil {
				return err
			}

			if org == "" {
				return fmt.Errorf("Please set --org")
			}

			res, err := client.SudoUpdateOrganizationBillingCustomer(ctx, &adminv1.SudoUpdateOrganizationBillingCustomerRequest{
				OrgName:           org,
				BillingCustomerId: customerID,
			})
			if err != nil {
				return err
			}

			ch.PrintfSuccess("Updated billing information for organization %s\n", org)
			fmt.Printf("Billing Customer Id: %s\n", res.Organization.BillingCustomerId)
			if len(res.Subscriptions) > 0 {
				fmt.Printf("Subscriptions:\n")
				for _, subscription := range res.Subscriptions {
					fmt.Printf("  Subscription ID: %s\n", subscription.Id)
					fmt.Printf("  Plan ID: %s\n", subscription.PlanId)
					fmt.Printf("  Plan Name: %s\n", subscription.PlanName)
					fmt.Printf("  Subscription Start Date: %s\n", subscription.StartDate.AsTime().Format(time.DateTime))
					fmt.Printf("  Subscription End Date: %s\n", subscription.EndDate.AsTime().Format(time.DateTime))
					fmt.Printf("  Subscription Current Billing Cycle Start Date: %s\n", subscription.CurrentBillingCycleStartDate.AsTime().Format(time.DateTime))
					fmt.Printf("  Subscription Current Billing Cycle End Date: %s\n", subscription.CurrentBillingCycleEndDate.AsTime().Format(time.DateTime))
					fmt.Printf("  Subscription Trial End Date: %s\n", subscription.TrialEndDate.AsTime().Format(time.DateTime))
					fmt.Printf("\n")
				}
			} else {
				fmt.Printf("No existing subscriptions\n")
			}

			return nil
		},
	}

	setCmd.Flags().StringVar(&org, "org", "", "Organization Name")
	setCmd.Flags().StringVar(&customerID, "customer-id", "", "Billing Customer Id")
	return setCmd
}
