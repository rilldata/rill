package add

import (
	"fmt"

	"github.com/spf13/cobra"
)

// addCmd represents the add command
func AddCmd() *cobra.Command {
	var addCmd = &cobra.Command{
		Use:   "add",
		Short: "Add source, models to the project",
		Long:  `Add source, models to the project`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("Error: must also specify a sub commands like source or model")
		},
	}

	addCmd.AddCommand(SourceCmd())
	return addCmd
}
