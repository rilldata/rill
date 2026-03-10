package python

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func PythonCmd(ch *cmdutil.Helper) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "python",
		Short: "Manage Python environment for data sources",
		Long:  "Set up and manage Python environments for running Python scripts as Rill data sources.",
	}

	cmd.AddCommand(SetupCmd(ch))

	return cmd
}
