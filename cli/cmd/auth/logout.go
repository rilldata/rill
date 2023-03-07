package auth

import (
	"github.com/fatih/color"
	"github.com/rilldata/rill/admin/client"
	"github.com/rilldata/rill/cli/pkg/config"
	"github.com/rilldata/rill/cli/pkg/dotrill"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

// LogoutCmd is the command for logging out of a Rill account.
func LogoutCmd(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logout",
		Short: "Logout of the Rill API",
		RunE: func(cmd *cobra.Command, args []string) error {
			warn := color.New(color.Bold).Add(color.FgYellow)
			token := cfg.AdminToken
			if token == "" {
				warn.Println("You are already logged out.")
				return nil
			}

			client, err := client.New(cfg.AdminURL, cfg.AdminToken)
			if err != nil {
				return err
			}
			defer client.Close()
			_, err = client.RevokeCurrentAuthToken(cmd.Context(), &adminv1.RevokeCurrentAuthTokenRequest{})
			if err != nil {
				return err
			}

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
