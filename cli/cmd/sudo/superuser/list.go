package superuser

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func ListCmd(cfg *config.Config) *cobra.Command {
	addCmd := &cobra.Command{
		Use:   "list",
		Args:  cobra.NoArgs,
		Short: "List all superusers",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()

			res, err := client.ListSuperUsers(ctx, &adminv1.ListSuperUsersRequest{})
			if err != nil {
				return err
			}

			cmdutil.PrintSuperUsers(res.Users)

			return nil
		},
	}
	return addCmd
}
