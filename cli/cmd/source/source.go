package source

import (
	"fmt"

	"github.com/spf13/cobra"
)

// sourceCmd represents the source command
func SourceCmd() *cobra.Command {
	var sourceCmd = &cobra.Command{
		Use:   "source",
		Short: "Create, drop sources to the project",
		Long:  `Create, drop sources to the project.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("Error: must also specify a sub commands like add or drop")
		},
	}
	sourceCmd.AddCommand(AddCmd())
	sourceCmd.AddCommand(DropCmd())

	return sourceCmd
}
