package whitelist

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func RemoveCmd(ch *cmdutil.Helper) *cobra.Command {
	var project string

	removeCmd := &cobra.Command{
		Use:   "remove <email-domain>",
		Args:  cobra.ExactArgs(1),
		Short: "Remove whitelisted email domain for the org or project",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			client, err := ch.Client()
			if err != nil {
				return err
			}

			domain := args[0]

			if project != "" {
				_, err = client.RemoveProjectWhitelistedDomain(ctx, &adminv1.RemoveProjectWhitelistedDomainRequest{
					Org:     ch.Org,
					Project: project,
					Domain:  domain,
				})
				if err != nil {
					return err
				}

				ch.PrintfWarn("New users with email addresses ending in %q will no longer automatically be added to project %q of %q. Existing users previously added through this policy will keep their access. (To remove users, use `rill user remove`.)\n", domain, project, ch.Org)
				return nil
			}

			_, err = client.RemoveWhitelistedDomain(ctx, &adminv1.RemoveWhitelistedDomainRequest{
				Org:    ch.Org,
				Domain: domain,
			})
			if err != nil {
				return err
			}

			ch.PrintfWarn("New users with email addresses ending in %q will no longer automatically be added to organization %q. Existing users previously added through this policy will keep their access. (To remove users, use `rill user remove`.)\n", domain, ch.Org)
			return nil
		},
	}

	removeCmd.Flags().StringVar(&ch.Org, "org", ch.Org, "Organization")
	removeCmd.Flags().StringVar(&project, "project", "", "Project")

	return removeCmd
}
