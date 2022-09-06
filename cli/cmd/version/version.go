package version

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// VersionCmd represents the version command
func VersionCmd(ver string) *cobra.Command {
	var versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Show rill version",
		Long:  `A longer description`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Print(Format(ver))
		},
	}

	return versionCmd
}

func init() {
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// versionCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// versionCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func Format(ver string) string {
	if ver == "" {
		return "rill version (built from source)"
	}
	ver = strings.TrimPrefix(ver, "v")
	return fmt.Sprintf("rill version %s\n", ver)
}
