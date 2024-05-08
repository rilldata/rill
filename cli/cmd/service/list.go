package service

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func ListCmd(ch *cmdutil.Helper) *cobra.Command {
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List service",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := ch.Client()
			if err != nil {
				return err
			}

			res, err := client.ListServices(cmd.Context(), &adminv1.ListServicesRequest{
				OrganizationName: ch.Org,
			})
			if err != nil {
				return err
			}

			ch.PrintServices(res.Services)

			return nil
		},
	}

	return listCmd
}
