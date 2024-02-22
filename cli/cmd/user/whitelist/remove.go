package whitelist

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func RemoveCmd(ch *cmdutil.Helper) *cobra.Command {
	removeCmd := &cobra.Command{
		Use:   "remove <email-domain>",
		Args:  cobra.ExactArgs(1),
		Short: "Remove whitelisted email domain for the org",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			client, err := ch.Client()
			if err != nil {
				return err
			}

			domain := args[0]

			_, err = client.RemoveWhitelistedDomain(ctx, &adminv1.RemoveWhitelistedDomainRequest{
				Organization: ch.Org,
				Domain:       domain,
			})
			if err != nil {
				return err
			}

			ch.PrintfWarn("New users with email addresses ending in %q will no longer automatically be added to %q. Existing users previously added through this policy will keep their access. (To remove users, use `rill user remove`.)\n", domain, ch.Org)
			return nil
		},
	}

	removeCmd.Flags().StringVar(&ch.Org, "org", ch.Org, "Organization")

	return removeCmd
}
