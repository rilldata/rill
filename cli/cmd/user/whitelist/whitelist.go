package whitelist

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func WhitelistCmd(ch *cmdutil.Helper) *cobra.Command {
	whitelistCmd := &cobra.Command{
		Use:               "whitelist",
		Short:             "Whitelist access by email domain",
		PersistentPreRunE: cmdutil.CheckChain(cmdutil.CheckAuth(ch.Config), cmdutil.CheckOrganization(ch.Config)),
	}

	whitelistCmd.AddCommand(SetupCmd(ch))
	whitelistCmd.AddCommand(RemoveCmd(ch))
	whitelistCmd.AddCommand(ListCmd(ch))

	return whitelistCmd
}
