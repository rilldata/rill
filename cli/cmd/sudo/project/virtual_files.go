package project

import (
	"github.com/rilldata/rill/cli/cmd/sudo/project/virtual_files"
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func VirtualFilesCmd(ch *cmdutil.Helper) *cobra.Command {
	virtualFilesCmd := &cobra.Command{
		Use:   "virtual-files",
		Short: "Manage virtual files in a project",
	}

	virtualFilesCmd.AddCommand(virtual_files.ListCmd(ch))
	virtualFilesCmd.AddCommand(virtual_files.GetCmd(ch))
	virtualFilesCmd.AddCommand(virtual_files.DeleteCmd(ch))

	return virtualFilesCmd
}
