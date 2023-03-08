package auth

import (
	"fmt"

	"github.com/rilldata/rill/cli/cmd/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	"github.com/rilldata/rill/cli/pkg/dotrill"
	"github.com/spf13/cobra"
)

func LoginCmd(cfg *config.Config) *cobra.Command {
	var token string

	loginCmd := &cobra.Command{
		Use:   "login",
		Short: "Login",
		RunE: func(cmd *cobra.Command, args []string) error {
			sp := cmdutil.Spinner("Login in...")
			sp.Start()

			if token != "" {
				err := dotrill.SetAccessToken(token)
				if err != nil {
					return err
				}

				fmt.Println("Saved access token to ~/.rill.")
				return nil
			}

			// TODO: Start browser-based login flow
			sp.Stop()
			fmt.Println("Logging in")
			return nil
		},
	}

	loginCmd.Flags().SortFlags = false
	loginCmd.Flags().StringVar(&token, "token", "", "Authentication token")

	return loginCmd
}
