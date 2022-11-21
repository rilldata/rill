package add

import (
	"github.com/spf13/cobra"
)

// addCmd represents the add command
func AddCmd() *cobra.Command {
	var addCmd = &cobra.Command{
		Use:   "add",
		Short: "Add source, models to the project",
		Long:  `Add source, models to the project`,
		// RunE: func(cmd *cobra.Command, args []string) error {
		// 	fmt.Println("add called")
		// 	return nil
		// },
	}

	addCmd.AddCommand(SourceCmd())
	return addCmd
}
