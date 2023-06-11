package whitelist

import (
	"context"
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func SetupCmd(cfg *config.Config) *cobra.Command {
	var role string

	setupCmd := &cobra.Command{
		Use:   "setup <email-domain>",
		Short: "Whitelist an email domain for the org",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()

			domain := args[0]

			cmdutil.WarnPrinter(fmt.Sprintf("If you confirm, new and existing users with email addresses ending in %q will automatically be added to %q with role %q."+
				"\n\nTo whitelist another email domain than your own, reach out to support: https://rilldata.com/support", domain, cfg.Org, role))
			if !cmdutil.ConfirmPrompt("Do you confirm?", "", false) {
				cmdutil.WarnPrinter("Aborted")
				return nil
			}

			_, err = client.CreateWhitelistedDomain(context.Background(), &adminv1.CreateWhitelistedDomainRequest{
				Organization: cfg.Org,
				Domain:       domain,
				Role:         role,
			})
			if err != nil {
				return err
			}

			cmdutil.SuccessPrinter(fmt.Sprintf("Whitelisted %q for %q (to remove it, use `rill user whitelist remove`).", domain, cfg.Org))

			return nil
		},
	}

	setupCmd.Flags().StringVar(&cfg.Org, "org", cfg.Org, "Organization")
	setupCmd.Flags().StringVar(&role, "role", "viewer", fmt.Sprintf("Role of the user [%v]", "admin, viewer"))

	return setupCmd
}
