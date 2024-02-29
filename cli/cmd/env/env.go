package env

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func EnvCmd(ch *cmdutil.Helper) *cobra.Command {
	envCmd := &cobra.Command{
		Use:               "env",
		Short:             "Manage variables for a project",
		PersistentPreRunE: cmdutil.CheckChain(cmdutil.CheckAuth(ch), cmdutil.CheckOrganization(ch)),
	}
	envCmd.AddCommand(ConfigureCmd(ch))
	envCmd.AddCommand(SetCmd(ch))
	envCmd.AddCommand(RmCmd(ch))
	envCmd.AddCommand(ShowEnvCmd(ch))
	return envCmd
}
