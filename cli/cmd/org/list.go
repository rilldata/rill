package org

import (
	"context"
	"net/http"

	"github.com/bufbuild/connect-go"
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/rilldata/rill/proto/gen/rill/admin/v1/adminv1connect"
	"github.com/spf13/cobra"
)

func ListCmd(cfg *config.Config) *cobra.Command {
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all organizations",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()

			// result, err := client.Say(context.Background(), connect.NewRequest(&elizav1.SayRequest{
			// 	Sentence: "Hello",
			// }))

			// res, err := client.ListOrganizations(context.Background(), &adminv1.ListOrganizationsRequest{})
			// if err != nil {
			// 	return err
			// }

			// 	connectClient := elizav1connect.NewElizaServiceClient(
			// server.Client(),
			// server.URL,
			// )

			client1 := adminv1connect.NewAdminServiceClient(http.DefaultClient, cfg.AdminURL)

			res1, err := client1.ListOrganizations(context.Background(), connect.NewRequest(&adminv1.ListOrganizationsRequest{}))
			if err != nil {
				return err
			}

			if len(res1.Msg.Organizations) == 0 {
				cmdutil.PrintlnWarn("No orgs found")
				return nil
			}

			cmdutil.PrintlnSuccess("Organizations list")
			cmdutil.TablePrinter(toTable(res1.Msg.Organizations, cfg.Org))
			return nil
		},
	}

	return listCmd
}
