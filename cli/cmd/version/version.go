package version

import (
	"github.com/spf13/cobra"
)

// VersionCmd represents the version command
func VersionCmd() *cobra.Command {
	var versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Show rill version",
		Long:  `A longer description`,
		Run: func(cmd *cobra.Command, args []string) {
			root := cmd.Root()
			root.SetArgs([]string{"--version"})
			root.Execute()
		},
	}

	return versionCmd
}
