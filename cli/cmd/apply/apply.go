package apply

import (
	"fmt"

	"github.com/spf13/cobra"
)

// applyCmd represents the apply command
func ApplyCmd() *cobra.Command {
	var applyCmd = &cobra.Command{
		Use:   "apply",
		Short: "Apply the available artifacts and apply them Rill",
		Long: `loads a folder of artifacts and apply them to local project and migrate the available sources, models
	and dashboards`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("apply called")
		},
	}
	return applyCmd
}
