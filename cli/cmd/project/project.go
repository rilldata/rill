package project

import (
	"github.com/rilldata/rill/cli/pkg/config"
	"github.com/spf13/cobra"
)

func ProjectCmd(cfg *config.Config) *cobra.Command {
	projectCmd := &cobra.Command{
		Use:    "project",
		Hidden: !cfg.IsDev(),
		Short:  "Manage projects",
	}
	projectCmd.AddCommand(ShowCmd(cfg))
	projectCmd.AddCommand(StatusCmd(cfg))
	projectCmd.AddCommand(ConnectCmd(cfg))
	projectCmd.AddCommand(EditCmd(cfg))
	projectCmd.AddCommand(DeleteCmd(cfg))

	return projectCmd
}
