package billing

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func GetCreditsCmd(ch *cmdutil.Helper) *cobra.Command {
	var org string

	cmd := &cobra.Command{
		Use:   "get-credits",
		Short: "Show billing credit balance for an organization",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			if org == "" {
				return fmt.Errorf("please set --org")
			}

			client, err := ch.Client()
			if err != nil {
				return err
			}

			res, err := client.GetBillingSubscription(ctx, &adminv1.GetBillingSubscriptionRequest{
				Org:                 org,
				SuperuserForceAccess: true,
			})
			if err != nil {
				return err
			}

			planName := ""
			if res.Subscription != nil && res.Subscription.Plan != nil {
				planName = res.Subscription.Plan.DisplayName
			}
			fmt.Printf("Organization: %s\n", org)
			fmt.Printf("Plan:         %s\n", planName)

			ci := res.CreditInfo
			if ci == nil {
				fmt.Println("Credits:      n/a (no credit balance)")
				return nil
			}

			fmt.Printf("Credits:\n")
			fmt.Printf("  Total:      $%.2f\n", ci.TotalCredit)
			fmt.Printf("  Used:       $%.2f\n", ci.UsedCredit)
			fmt.Printf("  Remaining:  $%.2f\n", ci.RemainingCredit)
			if ci.BurnRatePerDay > 0 {
				fmt.Printf("  Burn rate:  $%.2f/day\n", ci.BurnRatePerDay)
				daysLeft := int(ci.RemainingCredit / ci.BurnRatePerDay)
				fmt.Printf("  Est. days:  ~%d\n", daysLeft)
			}
			if ci.CreditExpiry != nil {
				fmt.Printf("  Expires:    %s\n", ci.CreditExpiry.AsTime().Format("2006-01-02"))
			}

			return nil
		},
	}

	cmd.Flags().SortFlags = false
	cmd.Flags().StringVar(&org, "org", "", "Organization name")
	return cmd
}
