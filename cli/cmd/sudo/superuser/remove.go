package superuser

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func RemoveCmd(cfg *config.Config) *cobra.Command {
	removeCmd := &cobra.Command{
		Use:   "remove <email>",
		Args:  cobra.ExactArgs(1),
		Short: "Remove a superuser",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()

			_, err = client.SetSuperuser(ctx, &adminv1.SetSuperuserRequest{
				Email:     args[0],
				Superuser: false,
			})
			if err != nil {
				return err
			}

			cmdutil.SuccessPrinter(fmt.Sprintf("Removed superuser from %q", args[0]))

			return nil
		},
	}

	return removeCmd
}
