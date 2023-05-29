package user

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	"github.com/rilldata/rill/cli/pkg/dotrill"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func AssumeCmd(cfg *config.Config) *cobra.Command {
	var ttlMinutes int
	assumeCmd := &cobra.Command{
		Use:   "assume <email>",
		Args:  cobra.ExactArgs(1),
		Short: "Assume users by email",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()

			token, err := client.RequestRepresentativeAuthToken(ctx, &adminv1.RequestRepresentativeAuthTokenRequest{
				Email: args[0],
				Ttl:   int64(ttlMinutes),
			})
			if err != nil {
				return err
			}

			// Backup current token as original_token
			originalToken, err := dotrill.GetAccessToken()
			if err != nil {
				return err
			}
			err = dotrill.BackupOriginalToken(originalToken)
			if err != nil {
				return err
			}

			// Set new access token
			err = dotrill.SetAccessToken(token.TokenId)
			if err != nil {
				return err
			}

			// Set representing user email
			err = dotrill.SetRepresentingUserEmail(args[0])
			if err != nil {
				return err
			}

			// set the default token to the one we just got
			cfg.AdminTokenDefault = token.TokenId
			return nil
		},
	}

	assumeCmd.Flags().IntVar(&ttlMinutes, "ttl-minutes", 60, "Expiration Timestamp (in Minutes)")
	return assumeCmd
}
