package auth

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/config"
	"github.com/spf13/cobra"
)

func LogoutCmd(cfg *config.Config) *cobra.Command {
	loginCmd := &cobra.Command{
		Use:   "logout",
		Short: "Logout",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Logging out")
		},
	}

	return loginCmd
}
