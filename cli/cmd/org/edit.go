package org

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func EditCmd(cfg *config.Config) *cobra.Command {
	var orgName, description string

	editCmd := &cobra.Command{
		Use:   "edit [<org-name>]",
		Args:  cobra.MaximumNArgs(1),
		Short: "Edit organization details",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()

			if len(args) > 0 {
				orgName = args[0]
			}
			if !cmd.Flags().Changed("org") && len(args) == 0 && cfg.Interactive {
				orgNames, err := cmdutil.OrgNames(ctx, client)
				if err != nil {
					return err
				}

				orgName = cmdutil.SelectPrompt("Select org to edit", orgNames, cfg.Org)
			}

			if cfg.Interactive {
				err = cmdutil.SetFlagsByInputPrompts(*cmd, "description")
				if err != nil {
					return err
				}
			}

			resp, err := client.GetOrganization(ctx, &adminv1.GetOrganizationRequest{Name: orgName})
			if err != nil {
				if st, ok := status.FromError(err); ok {
					if st.Code() != codes.NotFound {
						return err
					}
				}

				fmt.Printf("Org name %q doesn't exist, please run `rill org list` to list available orgs\n", orgName)
				return nil
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

			cmdutil.SuccessPrinter("Updated organization")
			cmdutil.TablePrinter(toRow(updatedOrg.Organization))
			return nil
		},
	}
	editCmd.Flags().SortFlags = false
	editCmd.Flags().StringVar(&orgName, "org", cfg.Org, "Organization name")
	editCmd.Flags().StringVar(&description, "description", "", "Description")

	return editCmd
}
