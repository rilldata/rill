package drop

import (
	"fmt"

	"github.com/spf13/cobra"
)

// dropCmd represents the drop command
func DropCmd() *cobra.Command {
	var dropCmd = &cobra.Command{
		Use:   "drop",
		Short: "Drop source, models to the project",
		Long:  `Drop source, models to the project.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("Error: must also specify a sub commands like source or model")
		},
	}

	dropCmd.AddCommand(SourceCmd())
	return dropCmd
}
