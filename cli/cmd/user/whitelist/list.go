package whitelist

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func ListCmd(ch *cmdutil.Helper) *cobra.Command {
	var project string

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List whitelisted email domains for the org or project",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			client, err := ch.Client()
			if err != nil {
				return err
			}

			if project != "" {
				whitelistedDomains, err := client.ListProjectWhitelistedDomains(ctx, &adminv1.ListProjectWhitelistedDomainsRequest{Org: ch.Org, Project: project})
				if err != nil {
					return err
				}

				if len(whitelistedDomains.Domains) > 0 {
					ch.PrintfSuccess("Whitelisted email domains for project %q of %q:\n", project, ch.Org)
					for _, d := range whitelistedDomains.Domains {
						ch.PrintfSuccess("%q (%q)\n", d.Domain, d.Role)
					}
				} else {
					ch.PrintfSuccess("No whitelisted email domains for project %q of %q\n", project, ch.Org)
				}
				return nil
			}

			whitelistedDomains, err := client.ListWhitelistedDomains(ctx, &adminv1.ListWhitelistedDomainsRequest{Org: ch.Org})
			if err != nil {
				return err
			}

			if len(whitelistedDomains.Domains) > 0 {
				ch.PrintfSuccess("Whitelisted email domains for organization %q:\n", ch.Org)
				for _, d := range whitelistedDomains.Domains {
					ch.PrintfSuccess("%q (%q)\n", d.Domain, d.Role)
				}
			} else {
				ch.PrintfSuccess("No whitelisted email domains for organization %q\n", ch.Org)
			}
			return nil
		},
	}

	listCmd.Flags().StringVar(&ch.Org, "org", ch.Org, "Organization")
	listCmd.Flags().StringVar(&project, "project", "", "Project")

	return listCmd
}
