package auth

import (
	"context"
	"fmt"

	"github.com/fatih/color"
	"github.com/rilldata/rill/cli/pkg/cmdutil"
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
			ctx := cmd.Context()

			token := cfg.AdminToken()
			if token == "" {
				cmdutil.WarnPrinter("You are already logged out.")
				return nil
			}

			err := Logout(ctx, cfg)
			if err != nil {
				return err
			}

			color.New(color.FgGreen).Println("Successfully logged out.")
			return nil
		},
	}
	return cmd
}

func Logout(ctx context.Context, cfg *config.Config) error {
	client, err := cmdutil.Client(cfg)
	if err != nil {
		return err
	}
	defer client.Close()

	_, err = client.RevokeCurrentAuthToken(ctx, &adminv1.RevokeCurrentAuthTokenRequest{})
	if err != nil {
		fmt.Printf("Failed to revoke token (did you revoke it manually?). Clearing local token anyway.\n")
	}

	err = dotrill.SetAccessToken("")
	if err != nil {
		return err
	}

	// Set original token backup as empty
	err = dotrill.BackupOriginalToken("")
	if err != nil {
		return err
	}

	// Set email for representing user as empty
	err = dotrill.SetRepresentingUserEmail("")
	if err != nil {
		return err
	}

	// Clear the state during logout
	err = dotrill.SetDefaultOrg("")
	if err != nil {
		return err
	}

	cfg.AdminTokenDefault = ""

	return nil
}
