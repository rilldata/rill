package source

import (
	"github.com/rilldata/rill/cli/pkg/config"
	"github.com/spf13/cobra"
)

// sourceCmd represents the source command
func SourceCmd(cfg *config.Config) *cobra.Command {
	sourceCmd := &cobra.Command{
		Use:   "source",
		Short: "Create or drop a source",
		// RunE: func(cmd *cobra.Command, args []string) error {
		// 	return fmt.Errorf("must specify a sub command")
		// },
	}
	sourceCmd.AddCommand(AddCmd(cfg))
	sourceCmd.AddCommand(DropCmd(cfg))

	return sourceCmd
}
