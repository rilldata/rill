package project

import (
	"fmt"
	"time"

	"github.com/rilldata/rill/cli/cmd/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	"github.com/spf13/cobra"
)

func ConnectCmd(cfg *config.Config) *cobra.Command {
	var name, displayName, prodBranch string
	var public bool

	connectCmd := &cobra.Command{
		Use:   "connect",
		Args:  cobra.ExactArgs(1),
		Short: "Connect",
		Run: func(cmd *cobra.Command, args []string) {
			sp := cmdutil.GetSpinner(4, "Connecting project...")
			sp.Start()
			// Just for spinner, will have to remove it
			time.Sleep(1 * time.Second)
			sp.Stop()
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
