package service

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func CreateCmd(cfg *config.Config) *cobra.Command {
	createCmd := &cobra.Command{
		Use:   "create <service-name>",
		Args:  cobra.ExactArgs(1),
		Short: "Create service",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()

			res, err := client.CreateService(cmd.Context(), &adminv1.CreateServiceRequest{
				Name:             args[0],
				OrganizationName: cfg.Org,
			})
			if err != nil {
				return err
			}

			cmdutil.PrintlnSuccess("Created service")
			cmdutil.TablePrinter(toRow(res.Service))

			return nil
		},
	}

	return createCmd
}
