package whitelist

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func ListCmd(ch *cmdutil.Helper) *cobra.Command {
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List whitelisted email domains for the org",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			client, err := ch.Client()
			if err != nil {
				return err
			}

			whitelistedDomains, err := client.ListWhitelistedDomains(ctx, &adminv1.ListWhitelistedDomainsRequest{Organization: ch.Org})
			if err != nil {
				return err
			}

			if len(whitelistedDomains.Domains) > 0 {
				ch.PrintfSuccess("Whitelisted email domains for %q:\n", ch.Org)
				for _, d := range whitelistedDomains.Domains {
					ch.PrintfSuccess("%q (%q)\n", d.Domain, d.Role)
				}
			} else {
				ch.PrintfSuccess("No whitelisted email domains for %q\n", ch.Org)
			}
			return nil
		},
	}

	listCmd.Flags().StringVar(&ch.Org, "org", ch.Org, "Organization")

	return listCmd
}
