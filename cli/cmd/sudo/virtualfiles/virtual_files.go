package virtualfiles

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func VirtualFilesCmd(ch *cmdutil.Helper) *cobra.Command {
	virtualFilesCmd := &cobra.Command{
		Use:   "virtual-files",
		Short: "Manage virtual files across projects",
	}

	virtualFilesCmd.AddCommand(ListCmd(ch))
	virtualFilesCmd.AddCommand(GetCmd(ch))
	virtualFilesCmd.AddCommand(DeleteCmd(ch))

	return virtualFilesCmd
}
