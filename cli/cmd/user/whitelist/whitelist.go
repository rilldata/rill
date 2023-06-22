package whitelist

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	"github.com/spf13/cobra"
)

func WhitelistCmd(cfg *config.Config) *cobra.Command {
	whitelistCmd := &cobra.Command{
		Use:               "whitelist",
		Short:             "Whitelist access by email domain",
		PersistentPreRunE: cmdutil.CheckChain(cmdutil.CheckAuth(cfg), cmdutil.CheckOrganization(cfg)),
	}

	whitelistCmd.AddCommand(SetupCmd(cfg))
	whitelistCmd.AddCommand(RemoveCmd(cfg))
	whitelistCmd.AddCommand(ListCmd(cfg))

	return whitelistCmd
}
