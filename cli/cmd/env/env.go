package env

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func EnvCmd(ch *cmdutil.Helper) *cobra.Command {
	cfg := ch.Config

	envCmd := &cobra.Command{
		Use:               "env",
		Short:             "Manage variables for a project",
		PersistentPreRunE: cmdutil.CheckChain(cmdutil.CheckAuth(cfg), cmdutil.CheckOrganization(cfg)),
	}
	envCmd.AddCommand(ConfigureCmd(ch))
	envCmd.AddCommand(SetCmd(ch))
	envCmd.AddCommand(RmCmd(ch))
	envCmd.AddCommand(ShowEnvCmd(ch))
	return envCmd
}
