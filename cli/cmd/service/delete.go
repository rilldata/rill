package service

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func DeleteCmd(cfg *config.Config) *cobra.Command {
	deleteCmd := &cobra.Command{
		Use:   "delete <service-name>",
		Args:  cobra.ExactArgs(1),
		Short: "Delete service",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()

			_, err = client.DeleteService(cmd.Context(), &adminv1.DeleteServiceRequest{
				Name:             args[0],
				OrganizationName: cfg.Org,
			})

			if err != nil {
				return err
			}

			cmdutil.PrintlnSuccess(fmt.Sprintf("Deleted service: %s", args[0]))

			return nil
		},
	}

	return deleteCmd
}
