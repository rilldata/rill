package autoinvite

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func AddCmd(cfg *config.Config) *cobra.Command {
	addCmd := &cobra.Command{
		Use:   "add <org> <domain> <role>",
		Args:  cobra.ExactArgs(3),
		Short: "Autoinvite users from a domain",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()

			org := args[0]
			domain := args[1]
			role := args[2]

			cmdutil.WarnPrinter(fmt.Sprintf("Warn: Auto inviting will give all users from domain %q access to the organization %q as %s", domain, org, role))
			if !cmdutil.ConfirmPrompt("Do you want to continue", "", false) {
				cmdutil.WarnPrinter("Aborted")
				return nil
			}

			_, err = client.CreateAutoinviteDomain(ctx, &adminv1.CreateAutoinviteDomainRequest{
				Organization: org,
				Domain:       domain,
				Role:         role,
			})
			if err != nil {
				return err
			}

			cmdutil.SuccessPrinter("Success")

			return nil
		},
	}

	return addCmd
}
