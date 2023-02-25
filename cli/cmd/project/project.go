package project

import (
	"github.com/rilldata/rill/cli/pkg/version"
	"github.com/spf13/cobra"
)

func ProjectCmd(ver version.Version) *cobra.Command {
	projectCmd := &cobra.Command{
		Use:    "project",
		Hidden: !ver.IsDev(),
		Short:  "Manage projects",
	}
	projectCmd.AddCommand(ShowCmd(ver))
	projectCmd.AddCommand(StatusCmd(ver))
	projectCmd.AddCommand(ConnectCmd(ver))
	projectCmd.AddCommand(EditCmd(ver))
	projectCmd.AddCommand(DeleteCmd(ver))

	return projectCmd
}
