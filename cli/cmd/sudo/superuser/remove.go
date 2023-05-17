package superuser

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func RemoveCmd(cfg *config.Config) *cobra.Command {
	var email string
	removeCmd := &cobra.Command{
		Use:   "remove <email>",
		Args:  cobra.MaximumNArgs(1),
		Short: "remove access as superuser",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()

			if len(args) > 0 {
				email = args[0]
			}

			cmdutil.WarnPrinter(fmt.Sprintf("Warn: Super user removing will revoke all superuser access for %q", email))
			if !cmdutil.ConfirmPrompt("Do you want to continue", "", false) {
				cmdutil.WarnPrinter("Aborted")
				return nil
			}

			_, err = client.RemoveSuperUser(ctx, &adminv1.RemoveSuperUserRequest{
				Email: email,
			})
			if err != nil {
				return err
			}

			cmdutil.SuccessPrinter(fmt.Sprintf("Removed superuser role for %q", email))

			return nil
		},
	}

	removeCmd.Flags().SortFlags = false
	removeCmd.Flags().StringVar(&email, "email", "", "Superuser Email")

	return removeCmd
}
