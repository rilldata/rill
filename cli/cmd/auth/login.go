package auth

import (
	"github.com/fatih/color"
	"github.com/rilldata/rill/cli/pkg/config"
	"github.com/rilldata/rill/cli/pkg/deviceauth"
	"github.com/rilldata/rill/cli/pkg/dotrill"
	"github.com/spf13/cobra"
)

// LoginCmd is the command for logging into a Rill account.
func LoginCmd(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Authenticate with the Rill API",
		RunE: func(cmd *cobra.Command, args []string) error {
			warn := color.New(color.Bold).Add(color.FgYellow)
			if cfg.AdminToken != "" {
				warn.Println("You are already logged in. To log in again, run `rill auth logout` first.")
				return nil
			}

			authenticator, err := deviceauth.New(cfg.AdminURL)
			if err != nil {
				return err
			}

			ctx := cmd.Context()
			deviceVerification, err := authenticator.VerifyDevice(ctx)
			if err != nil {
				return err
			}

			bold := color.New(color.Bold)
			bold.Printf("\nConfirmation Code: ")
			boldGreen := color.New(color.FgGreen).Add(color.Bold)
			boldGreen.Fprintln(color.Output, deviceVerification.UserCode)

			bold.Printf("\nOpen this URL in your browser to confirm the login: %s\n\n", deviceVerification.VerificationCompleteURL)

			OAuthTokenResponse, err := authenticator.GetAccessTokenForDevice(ctx, *deviceVerification)
			if err != nil {
				return err
			}

			bold.Print("Successfully logged in. Access Token: ")
			boldGreen.Fprintln(color.Output, OAuthTokenResponse.AccessToken)

			err = dotrill.SetAccessToken(OAuthTokenResponse.AccessToken)
			if err != nil {
				return err
			}
			return nil
		},
	}

	return cmd
}
