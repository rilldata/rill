package auth

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/config"
	"github.com/spf13/cobra"
)

func LoginCmd(cfg *config.Config) *cobra.Command {
	var token string

	loginCmd := &cobra.Command{
		Use:   "login",
		Short: "Login",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Logging in")
		},
	}
	loginCmd.Flags().SortFlags = false
	loginCmd.Flags().StringVar(&token, "token", "", "Authentication token")

	return loginCmd
}
