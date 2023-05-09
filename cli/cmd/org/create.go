package org

import (
	"context"
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	"github.com/rilldata/rill/cli/pkg/dotrill"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func CreateCmd(cfg *config.Config) *cobra.Command {
	var name, description string

	createCmd := &cobra.Command{
		Use:   "create <org-name>",
		Short: "Create organization",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()

			if len(args) > 0 {
				name = args[0]
			}

			if !cmd.Flags().Changed("name") && len(args) == 0 {
				// Get the new org name from user if not provided in the flag
				name, err = cmdutil.InputPrompt("Enter the org name", "")
				if err != nil {
					return err
				}
			}

			res, err := client.CreateOrganization(context.Background(), &adminv1.CreateOrganizationRequest{
				Name:        name,
				Description: description,
			})
			if err != nil {
				if !cmdutil.IsNameExistsErr(err) {
					return err
				}

				fmt.Printf("Org name %q already exists\n", name)
				return nil
			}

			// Switching to the created org
			err = dotrill.SetDefaultOrg(res.Organization.Name)
			if err != nil {
				return err
			}

			cmdutil.SuccessPrinter("Created organization")
			cmdutil.TablePrinter(toRow(res.Organization))
			return nil
		},
	}
	createCmd.Flags().SortFlags = false
	createCmd.Flags().StringVar(&name, "name", "", "Name")
	createCmd.Flags().StringVar(&description, "description", "", "Description")
	return createCmd
}
