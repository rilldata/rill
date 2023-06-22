package whitelist

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func RemoveCmd(cfg *config.Config) *cobra.Command {
	removeCmd := &cobra.Command{
		Use:   "remove <email-domain>",
		Args:  cobra.ExactArgs(1),
		Short: "Remove whitelisted email domain for the org",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()

			domain := args[0]

			_, err = client.RemoveWhitelistedDomain(ctx, &adminv1.RemoveWhitelistedDomainRequest{
				Organization: cfg.Org,
				Domain:       domain,
			})
			if err != nil {
				return err
			}

			cmdutil.PrintlnWarn(fmt.Sprintf("New users with email addresses ending in %q will no longer automatically be added to %q. "+
				"Existing users previously added through this policy will keep their access. (To remove users, use `rill user remove`.)", domain, cfg.Org))

			return nil
		},
	}

	removeCmd.Flags().StringVar(&cfg.Org, "org", cfg.Org, "Organization")

	return removeCmd
}
