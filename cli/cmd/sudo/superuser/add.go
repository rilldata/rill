package superuser

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func AddCmd(ch *cmdutil.Helper) *cobra.Command {
	addCmd := &cobra.Command{
		Use:   "add <email>",
		Args:  cobra.ExactArgs(1),
		Short: "Add new superuser",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			client, err := ch.Client()
			if err != nil {
				return err
			}

			_, err = client.SetSuperuser(ctx, &adminv1.SetSuperuserRequest{
				Email:     args[0],
				Superuser: true,
			})
			if err != nil {
				return err
			}

			ch.Printer.PrintlnSuccess(fmt.Sprintf("Granted superuser to %q", args[0]))

			return nil
		},
	}

	return addCmd
}
