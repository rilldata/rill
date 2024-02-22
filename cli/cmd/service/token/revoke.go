package token

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func RevokeCmd(ch *cmdutil.Helper) *cobra.Command {
	revokeCmd := &cobra.Command{
		Use:   "revoke <token-id>",
		Args:  cobra.ExactArgs(1),
		Short: "Revoke token",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := ch.Client()
			if err != nil {
				return err
			}

			_, err = client.RevokeServiceAuthToken(cmd.Context(), &adminv1.RevokeServiceAuthTokenRequest{
				TokenId: args[0],
			})
			if err != nil {
				return err
			}

			ch.PrintfSuccess("Revoked token\n")

			return nil
		},
	}

	return revokeCmd
}
