package service

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func CreateCmd(ch *cmdutil.Helper) *cobra.Command {
	createCmd := &cobra.Command{
		Use:   "create <service-name>",
		Args:  cobra.ExactArgs(1),
		Short: "Create service",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := ch.Config
			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()

			res1, err := client.CreateService(cmd.Context(), &adminv1.CreateServiceRequest{
				Name:             args[0],
				OrganizationName: cfg.Org,
			})
			if err != nil {
				return err
			}

			res2, err := client.IssueServiceAuthToken(cmd.Context(), &adminv1.IssueServiceAuthTokenRequest{
				OrganizationName: cfg.Org,
				ServiceName:      res1.Service.Name,
			})
			if err != nil {
				return err
			}

			ch.Printer.Printf("Created service %q in org %q.\n", res1.Service.Name, res1.Service.OrgName)
			ch.Printer.Printf("Access token: %s\n", res2.Token)

			return nil
		},
	}

	return createCmd
}
