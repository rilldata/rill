package user

import (
	"github.com/rilldata/rill/cli/cmd/auth"
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/dotrill"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func AssumeCmd(ch *cmdutil.Helper) *cobra.Command {
	var ttlMinutes int

	assumeCmd := &cobra.Command{
		Use:   "assume <email>",
		Args:  cobra.ExactArgs(1),
		Short: "Temporarily act as another user",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			client, err := ch.Client()
			if err != nil {
				return err
			}

			res, err := client.IssueRepresentativeAuthToken(ctx, &adminv1.IssueRepresentativeAuthTokenRequest{
				Email:      args[0],
				TtlMinutes: int64(ttlMinutes),
			})
			if err != nil {
				return err
			}

			// Backup current token as original_token
			originalToken, err := dotrill.GetAccessToken()
			if err != nil {
				return err
			}
			err = dotrill.SetBackupToken(originalToken)
			if err != nil {
				return err
			}

			// Set new access token
			err = dotrill.SetAccessToken(res.Token)
			if err != nil {
				return err
			}

			// Set representing user email
			err = dotrill.SetRepresentingUser(args[0])
			if err != nil {
				return err
			}

			// Load new token
			err = ch.ReloadAdminConfig()
			if err != nil {
				return err
			}

			// Select org for new user
			err = auth.SelectOrgFlow(ctx, ch, true, "")
			if err != nil {
				return err
			}

			return nil
		},
	}

	assumeCmd.Flags().IntVar(&ttlMinutes, "ttl-minutes", 60, "Minutes until the token should expire")

	return assumeCmd
}
