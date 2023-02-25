package auth

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/version"
	"github.com/spf13/cobra"
)

func LogoutCmd(ver version.Version) *cobra.Command {
	loginCmd := &cobra.Command{
		Use:   "logout",
		Short: "Logout",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Logging out")
		},
	}

	return loginCmd
}
