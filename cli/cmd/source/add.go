package source

import (
	"fmt"

	"github.com/spf13/cobra"
)

// addCmd represents the add command
func AddCmd() *cobra.Command {
	var sourcePath string
	var addCmd = &cobra.Command{
		Use:   "add",
		Short: "Add source to the project",
		Long:  `Add source to the project`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("add called")
			return nil
		},
	}

	addCmd.Flags().StringVarP(&sourcePath, "source-path", "p", ".", "Source path for Rill")
	return addCmd
}
