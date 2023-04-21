package env

import (
	"github.com/rilldata/rill/cli/cmd/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	"github.com/spf13/cobra"
)

func EnvCmd(cfg *config.Config) *cobra.Command {
	envCmd := &cobra.Command{
		Use:               "env",
		Short:             "Manage variables for a project",
		Hidden:            !cfg.IsDev(),
		PersistentPreRunE: cmdutil.CheckChain(cmdutil.CheckAuth(cfg), cmdutil.CheckOrganization(cfg)),
	}
	envCmd.AddCommand(ConfigureCmd(cfg))
	envCmd.AddCommand(SetCmd(cfg))
	envCmd.AddCommand(RmCmd(cfg))
	envCmd.AddCommand(ShowEnvCmd(cfg))
	return envCmd
}
