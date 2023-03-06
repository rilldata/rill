package auth

import (
	"fmt"
	"time"

	"github.com/rilldata/rill/cli/cmd/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	"github.com/spf13/cobra"
)

func LogoutCmd(cfg *config.Config) *cobra.Command {
	loginCmd := &cobra.Command{
		Use:   "logout",
		Short: "Logout",
		Run: func(cmd *cobra.Command, args []string) {
			sp := cmdutil.GetSpinner(4, "Login in...")
			sp.Start()
			// Just for spinner, will have to remove it
			time.Sleep(1 * time.Second)

			fmt.Println("Logging out")
			sp.Stop()
		},
	}

	return loginCmd
}
