package whitelist

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func RemoveCmd(ch *cmdutil.Helper) *cobra.Command {
	removeCmd := &cobra.Command{
		Use:   "remove <org> <domain>",
		Args:  cobra.ExactArgs(2),
		Short: "Remove whitelist for an org and domain",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			client, err := ch.Client()
			if err != nil {
				return err
			}

			org := args[0]
			domain := args[1]

			_, err = client.RemoveWhitelistedDomain(ctx, &adminv1.RemoveWhitelistedDomainRequest{
				Organization: org,
				Domain:       domain,
			})
			if err != nil {
				return err
			}

			ch.PrintfSuccess("Removed whitelist for org %q and domain %q\n", org, domain)

			return nil
		},
	}

	return removeCmd
}
