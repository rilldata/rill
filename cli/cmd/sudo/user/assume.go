package user

import (
	"fmt"
	"time"

	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/cli/cmd/auth"
	"github.com/rilldata/rill/cli/pkg/cmdutil"
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

			// If a user is already assumed, silently unassume and revert to the original user before assuming another one.
			representingUser, err := ch.DotRill.GetRepresentingUser()
			if err != nil {
				ch.PrintfWarn("Could not parse representing user email\n\n")
			}
			if representingUser != "" {
				err = UnassumeUser(ctx, ch)
				if err != nil {
					return err
				}
			}

			// Store expiryTime before requesting the token.
			// It could be fetched from the server, but that may not be needed.
			expiry := time.Now().Add(time.Duration(ttlMinutes) * time.Minute)

			client, err := ch.Client()
			if err != nil {
				return err
			}

			// Issue a new token for the *current* user that *represents* the user we want to assume.
			// The token will still show up in the current user's token listings, but will be consumed as if it were the user we are assuming.
			res, err := client.IssueUserAuthToken(ctx, &adminv1.IssueUserAuthTokenRequest{
				UserId:               "current",
				ClientId:             database.AuthClientIDRillSupport,
				DisplayName:          fmt.Sprintf("Support for %s", args[0]),
				TtlMinutes:           int64(ttlMinutes),
				RepresentEmail:       args[0],
				SuperuserForceAccess: true,
			})
			if err != nil {
				return err
			}

			// Backup current token as original_token
			originalToken, err := ch.DotRill.GetAccessToken()
			if err != nil {
				return err
			}
			err = ch.DotRill.SetBackupToken(originalToken)
			if err != nil {
				return err
			}

			// Set new access token
			err = ch.DotRill.SetAccessToken(res.Token)
			if err != nil {
				return err
			}

			// Backup current org as backup org
			defaultOrg, err := ch.DotRill.GetDefaultOrg()
			if err != nil {
				return err
			}
			err = ch.DotRill.SetBackupDefaultOrg(defaultOrg)
			if err != nil {
				return err
			}

			// Set representing user email
			err = ch.DotRill.SetRepresentingUser(args[0])
			if err != nil {
				return err
			}

			// Set the representing user token expiry
			err = ch.DotRill.SetRepresentingUserAccessTokenExpiry(expiry)
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
