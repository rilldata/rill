package project

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func HibernateCmd(ch *cmdutil.Helper) *cobra.Command {
	hibernateCmd := &cobra.Command{
		Use:   "hibernate <org> <project>",
		Args:  cobra.ExactArgs(2),
		Short: "Hibernate project",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			client, err := ch.Client()
			if err != nil {
				return err
			}

			_, err = client.HibernateProject(ctx, &adminv1.HibernateProjectRequest{
				Org:                  args[0],
				Project:              args[1],
				SuperuserForceAccess: true,
			})
			if err != nil {
				return err
			}

			return nil
		},
	}

	return hibernateCmd
}
