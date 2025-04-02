package user

import (
	"context"
	"fmt"

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
			return UnassumeUser(ctx, ch)
		},
	}
	return unassumeCmd
}

func UnassumeUser(ctx context.Context, ch *cmdutil.Helper) error {
	client, err := ch.Client()
	if err != nil {
		return err
	}
	// Revoke current token
	_, err = client.RevokeCurrentAuthToken(ctx, &adminv1.RevokeCurrentAuthTokenRequest{})
	if err != nil {
		fmt.Printf("Failed to revoke token (it may have expired). Clearing local token anyway.\n")
	}
	return RestoreOriginalUserState(ctx, ch)
}

func RestoreOriginalUserState(ctx context.Context, ch *cmdutil.Helper) error {
	// Fetch the original token
	originalToken, err := dotrill.GetBackupToken()
	if err != nil {
		return err
	}
	if originalToken == "" {
		return fmt.Errorf("original token is not available, you are not assuming any user")
	}

	// Restore the original token as the access token
	err = dotrill.SetAccessToken(originalToken)
	if err != nil {
		return err
	}

	// Fetch the original defualt org
	originalDefaultOrg, err := dotrill.GetBackupDefaultOrg()
	if err != nil {
		return err
	}

	// Restore the original defualt org as defualt org
	err = dotrill.SetDefaultOrg(originalDefaultOrg)
	if err != nil {
		return err
	}
	ch.Org = originalDefaultOrg

	// Fetch the original token expiry
	originalTokenExpiry, err := dotrill.GetBackupTokenExpiry()
	if err != nil {
		return err
	}

	// Restore the original token expiry as the access token
	err = dotrill.SetAccessTokenExpiry(originalTokenExpiry)
	if err != nil {
		return err
	}

	// Clear backup token
	err = dotrill.SetBackupToken("")
	if err != nil {
		return err
	}

	// Clear backup token expiry
	err = dotrill.SetBackupTokenExpiry("")
	if err != nil {
		return err
	}

	// Set email for representing user as empty
	err = dotrill.SetRepresentingUser("")
	if err != nil {
		return err
	}

	// Clear backup default org
	err = dotrill.SetBackupDefaultOrg("")
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
