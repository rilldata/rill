package source

import (
	"fmt"

	"github.com/spf13/cobra"
)

// sourceCmd represents the source command
func SourceCmd() *cobra.Command {
	var sourceCmd = &cobra.Command{
		Use:   "source",
		Short: "Create or drop a source",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("must specify a sub commands")
		},
	}
	sourceCmd.AddCommand(AddCmd())
	sourceCmd.AddCommand(DropCmd())

	return sourceCmd
}
