package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/rilldata/rill/cli/pkg/browser"
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/deviceauth"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/spf13/cobra"
)

// LoginCmd is the command for logging into a Rill account.
func LoginCmd(ch *cmdutil.Helper) *cobra.Command {
	var orgName string

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Authenticate with the Rill API",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			// Logout if already logged in
			if ch.AdminTokenDefault != "" {
				err := Logout(ctx, ch)
				if err != nil {
					return err
				}
			}

			// Login user
			err := Login(ctx, ch, "")
			if err != nil {
				return err
			}

			// Set default org after login
			interactive := orgName == ""
			err = SelectOrgFlow(ctx, ch, interactive, orgName)
			if err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&orgName, "org", "o", "", "Organization to use")
	return cmd
}

func Login(ctx context.Context, ch *cmdutil.Helper, redirectURL string) error {
	authURL := ch.AdminURL()

	authenticator, err := deviceauth.New(authURL)
	if err != nil {
		return err
	}

	deviceVerification, err := authenticator.VerifyDevice(ctx, redirectURL)
	if err != nil {
		return err
	}

	ch.PrintfBold("\nConfirmation Code: ")
	ch.PrintfSuccess("%s\n", deviceVerification.UserCode)

	ch.PrintfBold("\nOpen this URL in your browser to confirm the login: %s\n\n", deviceVerification.VerificationCompleteURL)

	if ch.Interactive {
		_ = browser.Open(deviceVerification.VerificationCompleteURL)
	}

	res1, err := authenticator.GetAccessTokenForDevice(ctx, deviceVerification)
	if err != nil {
		return err
	}

	err = ch.DotRill.SetAccessToken(res1.AccessToken)
	if err != nil {
		return err
	}

	err = ch.ReloadAdminConfig()
	if err != nil {
		return err
	}

	ch.PrintfBold("Successfully logged in. Welcome to Rill!\n")
	return nil
}

func LoginWithTelemetry(ctx context.Context, ch *cmdutil.Helper, redirectURL string) error {
	ch.PrintfBold("Please log in or sign up for Rill. Opening browser...\n")
	select {
	case <-time.After(2 * time.Second):
	case <-ctx.Done():
		return ctx.Err()
	}

	ch.Telemetry(ctx).RecordBehavioralLegacy(activity.BehavioralEventLoginStart)

	if err := Login(ctx, ch, redirectURL); err != nil {
		if errors.Is(err, deviceauth.ErrAuthenticationTimedout) {
			ch.PrintfWarn("Rill login has timed out as the code was not confirmed in the browser.\n")
			ch.PrintfWarn("Run the command again.\n")
			return nil
		} else if errors.Is(err, deviceauth.ErrCodeRejected) {
			ch.PrintfError("Login failed: Confirmation code rejected\n")
			return nil
		}
		return fmt.Errorf("login failed: %w", err)
	}

	// The cmdutil.Helper automatically detects the login and will add the user's ID to the telemetry.
	ch.Telemetry(ctx).RecordBehavioralLegacy(activity.BehavioralEventLoginSuccess)

	return nil
}

func SelectOrgFlow(ctx context.Context, ch *cmdutil.Helper, interactive bool, requestedOrg string) error {
	client, err := ch.Client()
	if err != nil {
		return err
	}

	res, err := client.ListOrganizations(ctx, &adminv1.ListOrganizationsRequest{
		PageSize: 1000,
	})
	if err != nil {
		return err
	}

	if len(res.Organizations) == 0 {
		if interactive {
			ch.PrintfWarn("You are not part of an org. Run `rill org create` to create one.\n")
		}
		return nil
	}

	var orgNames []string
	for _, org := range res.Organizations {
		orgNames = append(orgNames, org.Name)
	}

	defaultOrg := orgNames[0]
	if requestedOrg != "" {
		// Verify the requested org exists
		found := false
		for _, name := range orgNames {
			if name == requestedOrg {
				defaultOrg = requestedOrg
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("organization %q not found", requestedOrg)
		}
	} else if interactive && len(orgNames) > 1 {
		defaultOrg, err = cmdutil.SelectPrompt("Select default org (to change later, run `rill org switch`).", orgNames, defaultOrg)
		if err != nil {
			return err
		}
	}

	err = ch.DotRill.SetDefaultOrg(defaultOrg)
	if err != nil {
		return err
	}
	ch.Org = defaultOrg

	if interactive {
		ch.Printf("Set default organization to %q. Change using `rill org switch`.\n", defaultOrg)
	}
	return nil
}
