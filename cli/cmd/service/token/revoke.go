package token

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func RevokeCmd(cfg *config.Config) *cobra.Command {
	revokeCmd := &cobra.Command{
		Use:   "revoke <token-id>",
		Args:  cobra.ExactArgs(1),
		Short: "Revoke token",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()

			_, err = client.RevokeServiceAuthToken(cmd.Context(), &adminv1.RevokeServiceAuthTokenRequest{
				TokenId: args[0],
			})
			if err != nil {
				return err
			}

			cmdutil.PrintlnSuccess("Revoked token")

			return nil
		},
	}

	return revokeCmd
}
