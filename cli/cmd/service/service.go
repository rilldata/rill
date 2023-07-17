package service

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	"github.com/spf13/cobra"
)

func ServiceCmd(cfg *config.Config) *cobra.Command {
	serviceCmd := &cobra.Command{
		Use:               "service",
		Short:             "Manage service accounts",
		PersistentPreRunE: cmdutil.CheckChain(cmdutil.CheckAuth(cfg), cmdutil.CheckOrganization(cfg)),
	}

	serviceCmd.AddCommand(RenameCmd(cfg))
	// serviceCmd.AddCommand(AddCmd(cfg))
	// serviceCmd.AddCommand(RemoveCmd(cfg))

	return serviceCmd
}
