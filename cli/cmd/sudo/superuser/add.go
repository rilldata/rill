package superuser

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func AddCmd(cfg *config.Config) *cobra.Command {
	var email string
	addCmd := &cobra.Command{
		Use:   "add <email>",
		Args:  cobra.MaximumNArgs(1),
		Short: "invite users as superuser",
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

			cmdutil.WarnPrinter(fmt.Sprintf("Warn: Super user inviting will give all access to the %q", email))
			if !cmdutil.ConfirmPrompt("Do you want to continue", "", false) {
				cmdutil.WarnPrinter("Aborted")
				return nil
			}

			_, err = client.AddSuperUser(ctx, &adminv1.AddSuperUserRequest{
				Email: email,
			})
			if err != nil {
				return err
			}

			cmdutil.SuccessPrinter(fmt.Sprintf("Updated role of user %q as superuser", email))

			return nil
		},
	}

	addCmd.Flags().SortFlags = false
	addCmd.Flags().StringVar(&email, "email", "", "Superuser Email")

	return addCmd
}
