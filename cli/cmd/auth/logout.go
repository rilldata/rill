package auth

import (
	"github.com/fatih/color"
	"github.com/rilldata/rill/cli/pkg/config"
	"github.com/rilldata/rill/cli/pkg/dotrill"
	"github.com/spf13/cobra"
)

// LogoutCmd is the command for logging out of a Rill account.
func LogoutCmd(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logout",
		Short: "Logout of the Rill API",
		RunE: func(cmd *cobra.Command, args []string) error {
			warn := color.New(color.Bold).Add(color.FgYellow)
			token, err := dotrill.GetAccessToken()
			if err != nil {
				return err
			}
			if token == "" {
				warn.Println("You are already logged out.")
				return nil
			}
			// TODO actually revoke the token from admin server

			err = dotrill.SetAccessToken("")
			if err != nil {
				return err
			}
			color.New(color.FgGreen).Println("Successfully logged out.")
			return nil
		},
	}
	return cmd
}
