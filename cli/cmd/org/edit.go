package org

import (
	"fmt"

	"github.com/rilldata/rill/cli/cmd/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func EditCmd(cfg *config.Config) *cobra.Command {
	var org, description string

	editCmd := &cobra.Command{
		Use:   "edit",
		Args:  cobra.NoArgs,
		Short: "Edit",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()

			if !cmd.Flags().Changed("org") {
				// Get the new org name from user if not provided in the flag
				err := cmdutil.PromptIfUnset(&org, "Org Name", org)
				if err != nil {
					return err
				}
			}

			if !cmd.Flags().Changed("description") {
				// Get the new org description from user if not provided in the flag
				err := cmdutil.PromptIfUnset(&description, "Org Description", description)
				if err != nil {
					return err
				}
			}

			exists, err := cmdutil.OrgExists(ctx, client, org)
			if err != nil {
				return err
			}

			if !exists {
				return fmt.Errorf("Org name %q not exists, please run `rill org list` to list available orgs", org)
			}

			resp, err := client.GetOrganization(ctx, &adminv1.GetOrganizationRequest{Name: org})
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
	editCmd.Flags().StringVar(&org, "org", cfg.Org, "Organization name")
	editCmd.Flags().StringVar(&description, "description", "Unknown", "Description")

	return editCmd
}
