package source

import (
	"fmt"

	"github.com/spf13/cobra"
)

// dropCmd represents the drop command
func DropCmd() *cobra.Command {
	var sourcePath string
	var dropCmd = &cobra.Command{
		Use:   "drop",
		Short: "Drop source to the project",
		Long:  `Add source to the project.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("drop called")
			return nil
		},
	}

	dropCmd.Flags().StringVarP(&sourcePath, "source-path", "p", ".", "Source path for Rill")
	return dropCmd
}
