package billing

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func BillingCmd(ch *cmdutil.Helper) *cobra.Command {
	billingCmd := &cobra.Command{
		Use:   "billing",
		Short: "Billing update for customers",
	}

	billingCmd.AddCommand(SetCmd(ch))
	billingCmd.AddCommand(DeleteIssueCmd(ch))
	billingCmd.AddCommand(SetMessageCmd(ch))
	billingCmd.AddCommand(DeleteMessageCmd(ch))
	billingCmd.AddCommand(ExtendTrialCmd(ch))
	billingCmd.AddCommand(GrantTrialCreditsCmd(ch))
	billingCmd.AddCommand(RepairCmd(ch))
	billingCmd.AddCommand(SetupCmd(ch))
	billingCmd.AddCommand(MockUsageCmd(ch))

	return billingCmd
}
