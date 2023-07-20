package token

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	"github.com/rilldata/rill/cli/pkg/dotrill"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func IssueCmd(cfg *config.Config) *cobra.Command {
	issueCmd := &cobra.Command{
		Use:   "issue",
		Args:  cobra.MaximumNArgs(1),
		Short: "Issue tokens",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()

			res, err := client.IssueServiceAuthToken(cmd.Context(), &adminv1.IssueServiceAuthTokenRequest{
				OrganizationName: cfg.Org,
				ServiceName:      args[0],
			})
			if err != nil {
				return err
			}

			// Set new access token
			err = dotrill.SetAccessToken(res.Token)
			if err != nil {
				return err
			}

			// set the default token to the one we just got
			cfg.AdminTokenDefault = res.Token
			cmdutil.PrintlnSuccess("Issued token")

			return nil
		},
	}

	return issueCmd
}
