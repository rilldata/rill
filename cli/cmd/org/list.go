package org

import (
	"context"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func ListCmd(cfg *config.Config) *cobra.Command {
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all organizations",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()

			res, err := client.ListOrganizations(context.Background(), &adminv1.ListOrganizationsRequest{})
			if err != nil {
				return err
			}

			if len(res.Organizations) == 0 {
				cmdutil.WarnPrinter("No orgs found")
				return nil
			}

			cmdutil.SuccessPrinter("Organizations list")
			cmdutil.TablePrinter(toTable(res.Organizations, cfg.Org))
			return nil
		},
	}

	return listCmd
}
