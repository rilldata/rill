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
			fmt.Println("drop called")
			return nil
		},
	}

	dropCmd.AddCommand(SourceCmd())
	return dropCmd
}
