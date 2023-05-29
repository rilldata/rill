package user

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	"github.com/rilldata/rill/cli/pkg/dotrill"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func UnAssumeCmd(cfg *config.Config) *cobra.Command {
	unAssumeCmd := &cobra.Command{
		Use:   "unassume",
		Args:  cobra.NoArgs,
		Short: "Unassume users by email",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()

			originalToken, err := dotrill.GetBackupOriginalToken()
			if err != nil {
				return err
			}

			if originalToken == "" {
				return fmt.Errorf("Original token is not available, you are not assuming any user")
			}

			currentToken, err := dotrill.GetAccessToken()
			if err != nil {
				return err
			}

			// Need to check if we require this as we have background job for this
			if originalToken != currentToken {
				_, err = client.RevokeCurrentAuthToken(ctx, &adminv1.RevokeCurrentAuthTokenRequest{})
				if err != nil {
					fmt.Printf("Failed to revoke token (did you revoke it manually?). Clearing local token anyway.\n")
				}
			}

			err = dotrill.SetAccessToken(originalToken)
			if err != nil {
				return err
			}

			// Set original token as empty
			err = dotrill.BackupOriginalToken("")
			if err != nil {
				return err
			}

			// Set email for representing user as empty
			err = dotrill.SetRepresentingUserEmail("")
			if err != nil {
				return err
			}

			return nil
		},
	}
	return unAssumeCmd
}
