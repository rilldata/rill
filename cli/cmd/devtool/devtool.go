package devtool

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func DevtoolCmd(ch *cmdutil.Helper) *cobra.Command {
	devtoolCmd := &cobra.Command{
		Use:    "devtool",
		Short:  "Utilities for developing Rill",
		Hidden: true,
	}

	devtoolCmd.AddCommand(StartCmd(ch))
	devtoolCmd.AddCommand(SeedCmd(ch))
	devtoolCmd.AddCommand(DotenvCmd(ch))
	devtoolCmd.AddCommand(SwitchEnvCmd(ch))

	return devtoolCmd
}
