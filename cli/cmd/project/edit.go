package project

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/config"
	"github.com/spf13/cobra"
)

func EditCmd(cfg *config.Config) *cobra.Command {
	var name, displayName, prodBranch string
	var public bool

	editCmd := &cobra.Command{
		Use:   "edit",
		Args:  cobra.ExactArgs(1),
		Short: "Edit",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("not implemented")
		},
	}

	editCmd.Flags().SortFlags = false

	editCmd.Flags().StringVar(&name, "name", "noname", "Name")
	editCmd.Flags().StringVar(&displayName, "display-name", "noname", "Display name")
	editCmd.Flags().StringVar(&prodBranch, "prod-branch", "noname", "Production branch name")
	editCmd.Flags().BoolVar(&public, "public", false, "Public")

	return editCmd
}
