package token

import (
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func IssueCmd(ch *cmdutil.Helper) *cobra.Command {
	var displayName string
	var ttlMinutes int

	issueCmd := &cobra.Command{
		Use:   "issue",
		Args:  cobra.NoArgs,
		Short: "Issue personal access token",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := ch.Client()
			if err != nil {
				return err
			}

			if ch.Interactive && displayName == "" {
				displayName, err = cmdutil.InputPrompt("Please enter a display name for the token", "")
				if err != nil {
					return err
				}
			}

			res, err := client.IssueUserAuthToken(cmd.Context(), &adminv1.IssueUserAuthTokenRequest{
				UserId:      "current",
				ClientId:    database.AuthClientIDRillManual,
				DisplayName: displayName,
				TtlMinutes:  int64(ttlMinutes),
			})
			if err != nil {
				return err
			}

			ch.PrintfSuccess("Token: %v\n", res.Token)

			return nil
		},
	}

	issueCmd.Flags().SortFlags = false
	issueCmd.Flags().StringVar(&displayName, "display-name", "", "Display name for the token")
	issueCmd.Flags().IntVar(&ttlMinutes, "ttl-minutes", 0, "Optional minutes until the token should expire")

	return issueCmd
}
