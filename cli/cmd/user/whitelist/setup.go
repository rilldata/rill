package whitelist

import (
	"context"
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func SetupCmd(ch *cmdutil.Helper) *cobra.Command {
	var role string

	setupCmd := &cobra.Command{
		Use:   "setup <email-domain>",
		Short: "Whitelist an email domain for the org",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := ch.Client()
			if err != nil {
				return err
			}

			domain := args[0]

			ch.PrintfWarn("If you confirm, new and existing users with email addresses ending in %q will automatically be added to %q with role %q.\n\nTo whitelist another email domain than your own, reach out to support: https://rilldata.com/support\n", domain, ch.Org, role)
			if !cmdutil.ConfirmPrompt("Do you confirm?", "", false) {
				ch.PrintfWarn("Aborted\n")
				return nil
			}

			_, err = client.CreateWhitelistedDomain(context.Background(), &adminv1.CreateWhitelistedDomainRequest{
				Organization: ch.Org,
				Domain:       domain,
				Role:         role,
			})
			if err != nil {
				return err
			}
			ch.PrintfSuccess("Whitelisted %q for %q (to remove it, use `rill user whitelist remove`).\n", domain, ch.Org)

			return nil
		},
	}

	setupCmd.Flags().StringVar(&ch.Org, "org", ch.Org, "Organization")
	setupCmd.Flags().StringVar(&role, "role", "viewer", fmt.Sprintf("Role of the user [%v]", "admin, viewer"))

	return setupCmd
}
