package user

import (
	"context"
	"fmt"
	"time"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
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
			return UnassumeUser(ctx, ch)
		},
	}
	return unassumeCmd
}

func UnassumeUser(ctx context.Context, ch *cmdutil.Helper) error {
	// we reverted to original user first because we want to call RevokeRepresentativeAuthTokens api with Original User
	// Fetch the original token
	originalToken, err := ch.DotRill.GetBackupToken()
	if err != nil {
		return err
	}
	representingUser, err := ch.DotRill.GetRepresentingUser()
	if err != nil {
		return err
	}
	if originalToken == "" || representingUser == "" {
		return fmt.Errorf("you are not assuming any user")
	}

	// Restore the original token as the access token
	err = ch.DotRill.SetAccessToken(originalToken)
	if err != nil {
		return err
	}

	client, err := ch.Client()
	if err != nil {
		return err
	}

	// Revoke all tokens issued by the current user for acting as the given representing user.
	_, err = client.RevokeRepresentativeAuthTokens(ctx, &adminv1.RevokeRepresentativeAuthTokensRequest{Email: representingUser})
	if err != nil {
		ch.Printf("Failed to revoke token. Clearing local token anyway.\n")
	}

	// Clear local token and expiry
	err = ch.DotRill.SetRepresentingUserAccessTokenExpiry(time.Time{})
	if err != nil {
		return err
	}

	// Fetch the original default org
	originalDefaultOrg, err := ch.DotRill.GetBackupDefaultOrg()
	if err != nil {
		return err
	}

	// Restore the original default org as default org
	err = ch.DotRill.SetDefaultOrg(originalDefaultOrg)
	if err != nil {
		return err
	}
	ch.Org = originalDefaultOrg

	// Clear backup token
	err = ch.DotRill.SetBackupToken("")
	if err != nil {
		return err
	}

	// Set email for representing user as empty
	err = ch.DotRill.SetRepresentingUser("")
	if err != nil {
		return err
	}

	// Clear backup default org
	err = ch.DotRill.SetBackupDefaultOrg("")
	if err != nil {
		return err
	}

	// Reload access tokens
	err = ch.ReloadAdminConfig()
	if err != nil {
		return err
	}

	return nil
}
