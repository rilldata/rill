package subscription

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func SubscriptionCmd(ch *cmdutil.Helper) *cobra.Command {
	subsCmd := &cobra.Command{
		Use:               "subscription",
		Short:             "Manage organisation subscription",
		PersistentPreRunE: cmdutil.CheckAuth(ch),
	}

	subsCmd.PersistentFlags().StringVar(&ch.Org, "org", ch.Org, "Organization Name")
	subsCmd.AddCommand(EditCmd(ch))
	subsCmd.AddCommand(ListCmd(ch))
	subsCmd.AddCommand(CancelCmd(ch))
	subsCmd.AddCommand(RenewCmd(ch))

	return subsCmd
}
