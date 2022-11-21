package drop

import (
	"fmt"

	"github.com/spf13/cobra"
)

// sourceCmd represents the source command
func SourceCmd() *cobra.Command {
	var sourcePath string
	var sourceCmd = &cobra.Command{
		Use:   "source",
		Short: "Sources to be dropped from the project",
		Long:  `Sources to be dropped from the project.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Drop source called")
			return nil
		},
	}

	sourceCmd.Flags().StringVarP(&sourcePath, "source-path", "p", ".", "Source path to be added")
	return sourceCmd
}
