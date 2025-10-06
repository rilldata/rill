package whitelist

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func RemoveCmd(ch *cmdutil.Helper) *cobra.Command {
	removeCmd := &cobra.Command{
		Use:   "remove <org> [project] <domain>",
		Args:  cobra.RangeArgs(2, 3),
		Short: "Remove whitelist for an org and domain",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			client, err := ch.Client()
			if err != nil {
				return err
			}

			// check the number of args to determine if the project is provided
			if len(args) == 2 {
				// project is not provided, this is an organization whitelist
				org := args[0]
				domain := args[1]

				_, err = client.RemoveWhitelistedDomain(ctx, &adminv1.RemoveWhitelistedDomainRequest{
					Org:    org,
					Domain: domain,
				})
				if err != nil {
					return err
				}

				ch.PrintfSuccess("Removed whitelist for org %q and domain %q\n", org, domain)

				return nil
			}

			// project is provided, this is a project whitelist
			org := args[0]
			project := args[1]
			domain := args[2]

			_, err = client.RemoveProjectWhitelistedDomain(ctx, &adminv1.RemoveProjectWhitelistedDomainRequest{
				Org:     org,
				Project: project,
				Domain:  domain,
			})
			if err != nil {
				return err
			}

			ch.PrintfSuccess("Removed whitelist for project %q of %q and domain %q\n", project, org, domain)

			return nil
		},
	}

	return removeCmd
}
