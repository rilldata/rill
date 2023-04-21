package org

import (
	"context"
	"fmt"

	"github.com/rilldata/rill/cli/cmd/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	"github.com/rilldata/rill/cli/pkg/dotrill"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func CreateCmd(cfg *config.Config) *cobra.Command {
	var name, description string

	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()

			fmt.Println("Value flag name set", cmd.Flags().Changed("name"))

			if !cmd.Flags().Changed("name") {
				// Get the new org name from user if not provided in the flag
				err := cmdutil.PromptIfUnset(&name, "Org Name", "")
				if err != nil {
					return err
				}
			}

			exist, err := cmdutil.OrgExists(ctx, client, name)
			if err != nil {
				return err
			}

			if exist {
				return fmt.Errorf("Org name %q already exists", name)
			}

			res, err := client.CreateOrganization(context.Background(), &adminv1.CreateOrganizationRequest{
				Name:        name,
				Description: description,
			})
			if err != nil {
				return err
			}

			// Switching to the created org
			err = dotrill.SetDefaultOrg(res.Organization.Name)
			if err != nil {
				return err
			}

			cmdutil.SuccessPrinter("Created organization \n")
			cmdutil.TablePrinter(toRow(res.Organization))
			return nil
		},
	}
	createCmd.Flags().SortFlags = false
	createCmd.Flags().StringVar(&name, "name", "", "Name")
	createCmd.Flags().StringVar(&description, "description", "", "Description")
	return createCmd
}
