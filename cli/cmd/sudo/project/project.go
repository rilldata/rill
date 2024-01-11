package project

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func ProjectCmd(ch *cmdutil.Helper) *cobra.Command {
	projectCmd := &cobra.Command{
		Use:   "project",
		Short: "Project search for support users",
	}

	projectCmd.AddCommand(GetCmd(ch))
	projectCmd.AddCommand(SearchCmd(ch))

	return projectCmd
}
