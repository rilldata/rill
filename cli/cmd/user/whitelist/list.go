package whitelist

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func ListCmd(cfg *config.Config) *cobra.Command {
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List whitelisted email domains for the org",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()

			whitelistedDomains, err := client.ListWhitelistedDomains(ctx, &adminv1.ListWhitelistedDomainsRequest{Organization: cfg.Org})
			if err != nil {
				return err
			}

			if len(whitelistedDomains.Domains) > 0 {
				cmdutil.PrintlnSuccess(fmt.Sprintf("Whitelisted email domains for %q:", cfg.Org))
				for _, d := range whitelistedDomains.Domains {
					cmdutil.PrintlnSuccess(fmt.Sprintf("%q (%q)", d.Domain, d.Role))
				}
			} else {
				cmdutil.PrintlnSuccess(fmt.Sprintf("No whitelisted email domains for %q", cfg.Org))
			}
			return nil
		},
	}

	listCmd.Flags().StringVar(&cfg.Org, "org", cfg.Org, "Organization")

	return listCmd
}
