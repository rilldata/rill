package auth

import (
	"context"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

// LogoutCmd is the command for logging out of a Rill account.
func LogoutCmd(ch *cmdutil.Helper) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logout",
		Short: "Logout of the Rill API",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			token := ch.AdminToken()
			if token == "" {
				ch.PrintfWarn("You are already logged out.\n")
				return nil
			}

			err := Logout(ctx, ch)
			if err != nil {
				return err
			}

			ch.PrintfSuccess("Successfully logged out.\n")
			return nil
		},
	}
	return cmd
}

func Logout(ctx context.Context, ch *cmdutil.Helper) error {
	if !ch.IsAuthenticated() {
		return nil
	}

	client, err := ch.Client()
	if err != nil {
		return err
	}

	_, err = client.RevokeUserAuthToken(ctx, &adminv1.RevokeUserAuthTokenRequest{TokenId: "current"})
	if err != nil {
		ch.Printf("Failed to revoke token (did you revoke it manually?). Clearing local token anyway.\n")
	}

	err = ch.DotRill.SetAccessToken("")
	if err != nil {
		return err
	}

	// Set original_token as empty
	err = ch.DotRill.SetBackupToken("")
	if err != nil {
		return err
	}

	// Set representing user email as empty
	err = ch.DotRill.SetRepresentingUser("")
	if err != nil {
		return err
	}

	// Clear the state during logout
	err = ch.DotRill.SetDefaultOrg("")
	if err != nil {
		return err
	}

	err = ch.ReloadAdminConfig()
	if err != nil {
		return err
	}

	return nil
}
