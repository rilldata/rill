package whitelist

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func AddCmd(ch *cmdutil.Helper) *cobra.Command {
	addCmd := &cobra.Command{
		Use:   "add <org> <domain> <role>",
		Args:  cobra.ExactArgs(3),
		Short: "Whitelist users from a domain",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			client, err := ch.Client()
			if err != nil {
				return err
			}

			org := args[0]
			domain := args[1]
			role := args[2]

			ch.PrintfWarn("Warn: Whitelisting will give all users from domain %q access to the organization %q as %s\n", domain, org, role)
			ok, err := cmdutil.ConfirmPrompt("Do you want to continue", "", false)
			if err != nil {
				return err
			}
			if !ok {
				ch.PrintfWarn("Aborted\n")
				return nil
			}

			_, err = client.CreateWhitelistedDomain(ctx, &adminv1.CreateWhitelistedDomainRequest{
				Organization: org,
				Domain:       domain,
				Role:         role,
			})
			if err != nil {
				return err
			}

			ch.PrintfSuccess("Success\n")

			return nil
		},
	}

	return addCmd
}
