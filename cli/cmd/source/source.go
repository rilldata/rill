package source

import (
	"github.com/rilldata/rill/cli/pkg/version"
	"github.com/spf13/cobra"
)

// sourceCmd represents the source command
func SourceCmd(ver version.Version) *cobra.Command {
	sourceCmd := &cobra.Command{
		Use:   "source",
		Short: "Create or drop a source",
		// RunE: func(cmd *cobra.Command, args []string) error {
		// 	return fmt.Errorf("must specify a sub command")
		// },
	}
	sourceCmd.AddCommand(AddCmd(ver))
	sourceCmd.AddCommand(DropCmd(ver))

	return sourceCmd
}
