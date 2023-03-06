package auth

import (
	"fmt"
	"time"

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
			sp := cmdutil.GetSpinner(4, "Login in...")
			sp.Start()
			// Just for spinner, will have to remove it
			time.Sleep(1 * time.Second)

			if token != "" {
				err := dotrill.SetAccessToken(token)
				if err != nil {
					return err
				}

				fmt.Println("Saved access token to ~/.rill.")
				return nil
			}

			// TODO: Start browser-based login flow
			fmt.Println("Logging in")
			sp.Stop()
			return nil
		},
	}

	loginCmd.Flags().SortFlags = false
	loginCmd.Flags().StringVar(&token, "token", "", "Authentication token")

	return loginCmd
}
