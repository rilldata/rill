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

	subsCmd.AddCommand(CreateCmd(ch))
	subsCmd.AddCommand(EditCmd(ch))
	subsCmd.AddCommand(ListCmd(ch))

	return subsCmd
}
