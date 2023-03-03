package project

import (
	"github.com/rilldata/rill/cli/cmd/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	"github.com/spf13/cobra"
)

func ProjectCmd(cfg *config.Config) *cobra.Command {
	projectCmd := &cobra.Command{
		Use:               "project",
		Hidden:            !cfg.IsDev(),
		Short:             "Manage projects",
		PersistentPreRunE: cmdutil.CheckAuth(cfg),
	}
	projectCmd.AddCommand(ShowCmd(cfg))
	projectCmd.AddCommand(StatusCmd(cfg))
	projectCmd.AddCommand(ConnectCmd(cfg))
	projectCmd.AddCommand(EditCmd(cfg))
	projectCmd.AddCommand(DeleteCmd(cfg))
	projectCmd.AddCommand(ListCmd(cfg))

	return projectCmd
}
