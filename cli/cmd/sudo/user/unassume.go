package user

import (
	"fmt"

	"github.com/rilldata/rill/cli/cmd/auth"
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	"github.com/rilldata/rill/cli/pkg/dotrill"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func UnassumeCmd(cfg *config.Config) *cobra.Command {
	unassumeCmd := &cobra.Command{
		Use:   "unassume",
		Args:  cobra.NoArgs,
		Short: "Revert a call to `assume`",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()

			originalToken, err := dotrill.GetBackupToken()
			if err != nil {
				return err
			}

			if originalToken == "" {
				return fmt.Errorf("Original token is not available, you are not assuming any user")
			}

			// Revoke current token if have original token
			_, err = client.RevokeCurrentAuthToken(ctx, &adminv1.RevokeCurrentAuthTokenRequest{})
			if err != nil {
				fmt.Printf("Failed to revoke token (it may have expired). Clearing local token anyway.\n")
			}

			err = dotrill.SetAccessToken(originalToken)
			if err != nil {
				return err
			}
			cfg.AdminTokenDefault = originalToken

			// Set original_token as empty
			err = dotrill.SetBackupToken("")
			if err != nil {
				return err
			}

			// Set email for representing user as empty
			err = dotrill.SetRepresentingUser("")
			if err != nil {
				return err
			}

			// Select org again for original user
			err = auth.SelectOrgFlow(ctx, cfg)
			if err != nil {
				return err
			}

			return nil
		},
	}
	return unassumeCmd
}
