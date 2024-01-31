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
			cfg := ch.Config
			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()

			res, err := client.ListServices(cmd.Context(), &adminv1.ListServicesRequest{
				OrganizationName: cfg.Org,
			})
			if err != nil {
				return err
			}

			return ch.Printer.PrintResource(toTable(res.Services))
		},
	}

	return listCmd
}
