package org

import (
	"context"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func ListCmd(cfg *config.Config) *cobra.Command {
	var pageSize uint32
	var pageToken string

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

			res, err := client.ListOrganizations(context.Background(), &adminv1.ListOrganizationsRequest{
				PageSize:  pageSize,
				PageToken: pageToken,
			})
			if err != nil {
				return err
			}

			if len(res.Organizations) == 0 {
				cmdutil.PrintlnWarn("No orgs found")
				return nil
			}

			cmdutil.PrintlnSuccess("Organizations list")
			cmdutil.TablePrinter(toTable(res.Organizations, cfg.Org))
			if res.NextPageToken != "" {
				cmd.Println()
				cmd.Printf("Next page token: %s\n", res.NextPageToken)
			}
			return nil
		},
	}

	listCmd.Flags().Uint32Var(&pageSize, "page-size", 50, "Number of orgs to return per page")
	listCmd.Flags().StringVar(&pageToken, "page-token", "", "Pagination token")

	return listCmd
}
