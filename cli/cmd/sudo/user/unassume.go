package user

import (
	"fmt"

	"github.com/rilldata/rill/cli/cmd/auth"
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/dotrill"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func UnassumeCmd(ch *cmdutil.Helper) *cobra.Command {
	unassumeCmd := &cobra.Command{
		Use:   "unassume",
		Args:  cobra.NoArgs,
		Short: "Revert a call to `assume`",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			client, err := ch.Client()
			if err != nil {
				return err
			}

			// Fetch the original token
			originalToken, err := dotrill.GetBackupToken()
			if err != nil {
				return err
			}
			if originalToken == "" {
				return fmt.Errorf("original token is not available, you are not assuming any user")
			}

			// Revoke current token
			_, err = client.RevokeCurrentAuthToken(ctx, &adminv1.RevokeCurrentAuthTokenRequest{})
			if err != nil {
				fmt.Printf("Failed to revoke token (it may have expired). Clearing local token anyway.\n")
			}

			// Restore the original token as the access token
			err = dotrill.SetAccessToken(originalToken)
			if err != nil {
				return err
			}

			// Clear backup token
			err = dotrill.SetBackupToken("")
			if err != nil {
				return err
			}

			// Set email for representing user as empty
			err = dotrill.SetRepresentingUser("")
			if err != nil {
				return err
			}

			// Reload access tokens
			err = ch.ReloadAdminConfig()
			if err != nil {
				return err
			}

			// Select org again for original user
			err = auth.SelectOrgFlow(ctx, ch, true, "")
			if err != nil {
				return err
			}

			return nil
		},
	}
	return unassumeCmd
}
