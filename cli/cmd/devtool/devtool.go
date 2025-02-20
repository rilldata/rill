package devtool

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func DevtoolCmd(ch *cmdutil.Helper) *cobra.Command {
	internalGroupID := ""
	devtoolCmd := &cobra.Command{
		Use:   "devtool",
		Short: "Utilities for developing Rill",
		Example: `  rill devtool start cloud
  rill devtool seed cloud
  rill devtool start cloud --reset
  rill devtool start cloud --except runtime
  rill devtool start cloud --only admin,deps
  rill devtool start local
  rill devtool start local --reset
  rill devtool switch-env stage
  rill devtool dotenv upload cloud`,
		Hidden:  true,
		GroupID: internalGroupID,
	}

	devtoolCmd.AddCommand(StartCmd(ch))
	devtoolCmd.AddCommand(SeedCmd(ch))
	devtoolCmd.AddCommand(DotenvCmd(ch))
	devtoolCmd.AddCommand(SwitchEnvCmd(ch))
	devtoolCmd.AddCommand(SubscriptionCmd(ch))

	return devtoolCmd
}
