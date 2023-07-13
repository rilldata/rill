package project

import (
	"github.com/rilldata/rill/cli/pkg/config"
	"github.com/spf13/cobra"
)

func ProjectCmd(cfg *config.Config) *cobra.Command {
	projectCmd := &cobra.Command{
		Use:   "project",
		Short: "Project search for support users",
	}

	projectCmd.AddCommand(GetCmd(cfg))
	projectCmd.AddCommand(SearchCmd(cfg))

	return projectCmd
}
