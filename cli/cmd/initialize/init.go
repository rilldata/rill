package initialize

import (
	"fmt"

	"github.com/spf13/cobra"
)

// InitCmd represents the init command
func InitCmd() *cobra.Command {
	var exampleProject string
	var initCmd = &cobra.Command{
		Use:   "init",
		Short: "Initializing the example project",
		Long:  `Initializing the example project`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("init called with example project:'%s'", exampleProject)
			return nil
		},
	}

	initCmd.Flags().StringVarP(&exampleProject, "example", "p", ".", "Example project directory")
	return initCmd
}
