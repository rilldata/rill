package user

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	"github.com/rilldata/rill/cli/pkg/dotrill"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func AssumeCmd(cfg *config.Config) *cobra.Command {
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
				Ttl:   2,
			})
			if err != nil {
				return err
			}

			// Backup original token
			originalToken, err := dotrill.GetAccessToken()
			if err != nil {
				return err
			}
			err = dotrill.BackupOriginalToken(originalToken)
			if err != nil {
				return err
			}

			// Set token new token
			err = dotrill.SetAccessToken(token.TokenId)
			if err != nil {
				return err
			}

			// Set email for representing user
			err = dotrill.SetRepresentingUserEmail(args[0])
			if err != nil {
				return err
			}

			// set the default token to the one we just got
			cfg.AdminTokenDefault = token.TokenId
			return nil
		},
	}
	return assumeCmd
}
