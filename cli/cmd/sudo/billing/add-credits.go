package billing

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func AddCreditsCmd(ch *cmdutil.Helper) *cobra.Command {
	var org string
	var amount float64
	var expiryDays int32
	var description string

	cmd := &cobra.Command{
		Use:   "add-credits",
		Short: "Add billing credits to an organization",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			if org == "" {
				return fmt.Errorf("please set --org")
			}

			if amount <= 0 {
				return fmt.Errorf("please set --amount to a positive value")
			}

			client, err := ch.Client()
			if err != nil {
				return err
			}

			res, err := client.SudoAddCredits(ctx, &adminv1.SudoAddCreditsRequest{
				Org:         org,
				Amount:      amount,
				ExpiryDays:  expiryDays,
				Description: description,
			})
			if err != nil {
				return err
			}

			if res.CreditInfo != nil {
				ch.PrintfSuccess("Added $%.0f credits to organization %q\n", amount, org)
				ch.PrintfSuccess("  Total:     $%.0f\n", res.CreditInfo.TotalCredit)
				ch.PrintfSuccess("  Used:      $%.0f\n", res.CreditInfo.UsedCredit)
				ch.PrintfSuccess("  Remaining: $%.0f\n", res.CreditInfo.RemainingCredit)
				if res.CreditInfo.CreditExpiry != nil {
					ch.PrintfSuccess("  Expires:   %s\n", res.CreditInfo.CreditExpiry.AsTime().Format("2006-01-02"))
				}
			} else {
				ch.PrintfSuccess("Added $%.0f credits to organization %q\n", amount, org)
			}

			return nil
		},
	}

	cmd.Flags().SortFlags = false
	cmd.Flags().StringVar(&org, "org", "", "Organization name")
	cmd.Flags().Float64Var(&amount, "amount", 250, "Credit amount in dollars")
	cmd.Flags().Int32Var(&expiryDays, "expiry-days", 365, "Days until credits expire")
	cmd.Flags().StringVar(&description, "description", "", "Description for the credit grant")
	return cmd
}
