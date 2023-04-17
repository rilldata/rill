package org

import (
	"github.com/rilldata/rill/cli/cmd/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func EditCmd(cfg *config.Config) *cobra.Command {
	var description string

	editCmd := &cobra.Command{
		Use:   "edit <org-name>",
		Args:  cobra.ExactArgs(1),
		Short: "Edit",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()

			resp, err := client.GetOrganization(ctx, &adminv1.GetOrganizationRequest{Name: args[0]})
			if err != nil {
				return err
			}

			org := resp.Organization

			updatedOrg, err := client.UpdateOrganization(ctx, &adminv1.UpdateOrganizationRequest{
				Id:          org.Id,
				Name:        org.Name,
				Description: description,
			})
			if err != nil {
				return err
			}

			cmdutil.SuccessPrinter("Updated organization \n")
			cmdutil.TablePrinter(toRow(updatedOrg.Organization))
			return nil
		},
	}
	editCmd.Flags().SortFlags = false
	editCmd.Flags().StringVar(&description, "description", "", "Description")

	return editCmd
}
