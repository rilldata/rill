package auth

import (
	"context"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/dotrill"
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
				ch.Printer.PrintlnWarn("You are already logged out.")
				return nil
			}

			err := Logout(ctx, ch)
			if err != nil {
				return err
			}

			ch.Printer.PrintlnSuccess("Successfully logged out.")
			return nil
		},
	}
	return cmd
}

func Logout(ctx context.Context, ch *cmdutil.Helper) error {
	client, err := ch.Client()
	if err != nil {
		return err
	}

	_, err = client.RevokeCurrentAuthToken(ctx, &adminv1.RevokeCurrentAuthTokenRequest{})
	if err != nil {
		ch.Printer.Printf("Failed to revoke token (did you revoke it manually?). Clearing local token anyway.\n")
	}

	err = dotrill.SetAccessToken("")
	if err != nil {
		return err
	}

	// Set original_token as empty
	err = dotrill.SetBackupToken("")
	if err != nil {
		return err
	}

	// Set representing user email as empty
	err = dotrill.SetRepresentingUser("")
	if err != nil {
		return err
	}

	// Clear the state during logout
	err = dotrill.SetDefaultOrg("")
	if err != nil {
		return err
	}

	ch.AdminTokenDefault = ""

	return nil
}
