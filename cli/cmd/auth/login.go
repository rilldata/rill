package auth

import (
	"context"
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/rilldata/rill/cli/pkg/browser"
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	"github.com/rilldata/rill/cli/pkg/deviceauth"
	"github.com/rilldata/rill/cli/pkg/dotrill"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

// LoginCmd is the command for logging into a Rill account.
func LoginCmd(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Authenticate with the Rill API",
		RunE: func(cmd *cobra.Command, args []string) error {
			warn := color.New(color.Bold).Add(color.FgYellow)
			ctx := cmd.Context()

			// updating this as its not required to logout first and login again
			if cfg.AdminTokenDefault != "" {
				err := Logout(ctx, cfg)
				if err != nil {
					return err
				}

				cfg.AdminTokenDefault = ""
			}

			// Login user
			if err := Login(ctx, cfg, ""); err != nil {
				return err
			}

			// Set default org after login
			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()

			res2, err := client.ListOrganizations(context.Background(), &adminv1.ListOrganizationsRequest{})
			if err != nil {
				return err
			}

			if len(res2.Organizations) > 0 {
				var orgNames []string
				for _, org := range res2.Organizations {
					orgNames = append(orgNames, org.Name)
				}

				defaultOrg := orgNames[0]
				if len(orgNames) > 1 {
					defaultOrg = cmdutil.SelectPrompt("Select default org (to change later, run `rill org switch`).", orgNames, defaultOrg)
				}

				err = dotrill.SetDefaultOrg(defaultOrg)
				if err != nil {
					return err
				}

				fmt.Printf("Set default organization to %q. Change using `rill org switch`.\n", defaultOrg)
			} else {
				warn.Println("You are not part of any org. Run `rill org create` or `rill deploy` to create one.")
			}
			return nil
		},
	}

	return cmd
}

func Login(ctx context.Context, cfg *config.Config, redirectURL string) error {
	// In production, the REST and gRPC endpoints are the same, but in development, they're served on different ports.
	// We plan to move to connect.build for gRPC, which will allow us to serve both on the same port in development as well.
	// Until we make that change, this is a convenient hack for local development (assumes gRPC on port 9090 and REST on port 8080).
	authURL := cfg.AdminURL
	if strings.Contains(authURL, "http://localhost:9090") {
		authURL = "http://localhost:8080"
	}

	authenticator, err := deviceauth.New(authURL)
	if err != nil {
		return err
	}

	deviceVerification, err := authenticator.VerifyDevice(ctx, redirectURL)
	if err != nil {
		return err
	}

	bold := color.New(color.Bold)
	bold.Printf("\nConfirmation Code: ")
	boldGreen := color.New(color.FgGreen).Add(color.Bold)
	boldGreen.Fprintln(color.Output, deviceVerification.UserCode)

	bold.Printf("\nOpen this URL in your browser to confirm the login: %s\n\n", deviceVerification.VerificationCompleteURL)

	_ = browser.Open(deviceVerification.VerificationCompleteURL)

	res1, err := authenticator.GetAccessTokenForDevice(ctx, deviceVerification)
	if err != nil {
		return err
	}

	err = dotrill.SetAccessToken(res1.AccessToken)
	if err != nil {
		return err
	}
	// set the default token to the one we just got
	cfg.AdminTokenDefault = res1.AccessToken
	bold.Print("Successfully logged in. Welcome to Rill!\n")
	return nil
}
