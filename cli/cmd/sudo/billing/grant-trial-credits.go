package billing

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func GrantTrialCreditsCmd(ch *cmdutil.Helper) *cobra.Command {
	var org, description string
	var amount float64
	cmd := &cobra.Command{
		Use:   "grant-trial-credits",
		Short: "Grant additional trial credits to an organization on the credit-based trial",
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

			res, err := client.SudoGrantTrialCredits(ctx, &adminv1.SudoGrantTrialCreditsRequest{
				Org:         org,
				Amount:      amount,
				Description: description,
			})
			if err != nil {
				return err
			}

			ch.PrintfSuccess("Granted %g trial credits to organization %q\n", res.Granted, org)
			return nil
		},
	}

	cmd.Flags().SortFlags = false
	cmd.Flags().StringVar(&org, "org", "", "Organization Name")
	cmd.Flags().Float64Var(&amount, "amount", 0, "Amount of trial credits to grant")
	cmd.Flags().StringVar(&description, "description", "", "Optional description for the Orb ledger entry")
	return cmd
}
