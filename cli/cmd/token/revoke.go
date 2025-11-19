package token

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func RevokeCmd(ch *cmdutil.Helper) *cobra.Command {
	var all bool

	revokeCmd := &cobra.Command{
		Use:   "revoke [token-id]",
		Args:  cobra.MaximumNArgs(1),
		Short: "Revoke personal access token(s)",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := ch.Client()
			if err != nil {
				return err
			}

			if len(args) > 0 && all {
				return fmt.Errorf("cannot specify a token ID when using --all flag\n")
			}

			if all {
				confirm, err := cmdutil.ConfirmPrompt("Are you sure you want to revoke all access and refresh tokens for the current user? This action cannot be undone.", "", false)
				if err != nil {
					return err
				}
				if !confirm {
					ch.PrintfWarn("Operation cancelled\n")
					return nil
				}
				// Revoke all access and refresh tokens
				resp, err := client.RevokeAllUserAuthTokens(cmd.Context(), &adminv1.RevokeAllUserAuthTokensRequest{
					UserId: "current",
				})
				if err != nil {
					return err
				}

				if resp.TokensRevoked == 0 {
					ch.PrintfWarn("No tokens found to revoke\n")
				} else {
					ch.Printf("Successfully revoked %d token(s)\n", resp.TokensRevoked)
				}
				return nil
			}

			// Single token revocation
			if len(args) == 0 {
				return fmt.Errorf("Please specify a token ID to revoke or use the --all flag to revoke all tokens\n")
			}

			_, err = client.RevokeUserAuthToken(cmd.Context(), &adminv1.RevokeUserAuthTokenRequest{
				TokenId: args[0],
			})
			if err != nil {
				return err
			}

			ch.Printf("Token revoked successfully\n")
			return nil
		},
	}

	revokeCmd.Flags().BoolVar(&all, "all", false, "Revoke all access and refresh tokens for the current user")

	return revokeCmd
}
