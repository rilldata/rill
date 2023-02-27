package project

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/version"
	"github.com/spf13/cobra"
)

func ConnectCmd(ver version.Version) *cobra.Command {
	var name, displayName, prodBranch string
	var public bool

	connectCmd := &cobra.Command{
		Use:   "connect",
		Args:  cobra.ExactArgs(1),
		Short: "Connect",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("not implemented")
		},
	}

	connectCmd.Flags().SortFlags = false

	connectCmd.Flags().StringVar(&name, "name", "noname", "Name")
	connectCmd.Flags().StringVar(&displayName, "display-name", "noname", "Display name")
	connectCmd.Flags().StringVar(&prodBranch, "prod-branch", "noname", "Production branch name")
	connectCmd.Flags().BoolVar(&public, "public", false, "Public")

	return connectCmd
}
