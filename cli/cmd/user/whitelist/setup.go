package whitelist

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func SetupCmd(ch *cmdutil.Helper) *cobra.Command {
	var role string
	var project string

	setupCmd := &cobra.Command{
		Use:   "setup <email-domain>",
		Short: "Whitelist an email domain for the org or project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := ch.Client()
			if err != nil {
				return err
			}

			domain := args[0]

			if project != "" {
				ch.PrintfWarn("If you confirm, new and existing users with email addresses ending in %q will automatically be added to project %q of %q with role %q.\n\nTo whitelist another email domain than your own, reach out to support: https://rilldata.com/support\n", domain, project, ch.Org, role)
			} else {
				ch.PrintfWarn("If you confirm, new and existing users with email addresses ending in %q will automatically be added to organization %q with role %q.\n\nTo whitelist another email domain than your own, reach out to support: https://rilldata.com/support\n", domain, ch.Org, role)
			}
			ok, err := cmdutil.ConfirmPrompt("Do you confirm?", "", false)
			if err != nil {
				return err
			}
			if !ok {
				ch.PrintfWarn("Aborted\n")
				return nil
			}

			if project != "" {
				_, err = client.CreateProjectWhitelistedDomain(cmd.Context(), &adminv1.CreateProjectWhitelistedDomainRequest{
					Org:     ch.Org,
					Project: project,
					Domain:  domain,
					Role:    role,
				})
				if err != nil {
					return err
				}
				ch.PrintfSuccess("Whitelisted %q for project %q of organization %q (to remove it, use `rill user whitelist remove`).\n", domain, project, ch.Org)
			} else {
				_, err = client.CreateWhitelistedDomain(cmd.Context(), &adminv1.CreateWhitelistedDomainRequest{
					Org:    ch.Org,
					Domain: domain,
					Role:   role,
				})
				if err != nil {
					return err
				}
				ch.PrintfSuccess("Whitelisted %q for organization %q (to remove it, use `rill user whitelist remove`).\n", domain, ch.Org)
			}

			return nil
		},
	}

	setupCmd.Flags().StringVar(&ch.Org, "org", ch.Org, "Organization")
	setupCmd.Flags().StringVar(&project, "project", "", "Project name")
	setupCmd.Flags().StringVar(&role, "role", "viewer", "Role of the user")

	return setupCmd
}
