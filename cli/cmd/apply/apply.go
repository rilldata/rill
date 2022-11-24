package apply

import (
	"fmt"

	"github.com/spf13/cobra"
)

// applyCmd represents the apply command
func ApplyCmd() *cobra.Command {
	var repoDSN string
	var olapDriver string
	var olapDSN string
	var validate bool

	var applyCmd = &cobra.Command{
		Use:   "apply",
		Short: "Apply the available artifacts and apply them Rill",
		Long:  `loads a folder of artifacts and apply them to local project and reconciles the available sources, models and dashboards`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("apply called")
			return nil
		},
	}
	applyCmd.Flags().StringVar(&repoDSN, "dir", ".", "Project directory")
	applyCmd.Flags().StringVar(&olapDriver, "db-driver", "duckdb", "OLAP database driver")
	applyCmd.Flags().StringVar(&olapDSN, "db", "stage.db", "OLAP database DSN")
	applyCmd.Flags().BoolVar(&validate, "validate", false, "Validate and print actions")

	return applyCmd
}
