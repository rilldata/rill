package org

import (
	"context"

	"github.com/rilldata/rill/admin/client"
	"github.com/rilldata/rill/cli/cmd/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	"github.com/rilldata/rill/cli/pkg/dotrill"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func CreateCmd(cfg *config.Config) *cobra.Command {
	var description string

	createCmd := &cobra.Command{
		Use:   "create <org-name>",
		Short: "Create",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()

			org, err := Create(client, args[0], description)
			if err != nil {
				return err
			}

			cmdutil.SuccessPrinter("Created organization \n")
			cmdutil.TablePrinter(toRow(org))
			return nil
		},
	}
	createCmd.Flags().SortFlags = false
	createCmd.Flags().StringVar(&description, "description", "", "Description")

	return createCmd
}

// Create org and run any post creation steps
func Create(adminClient *client.Client, name, description string) (*adminv1.Organization, error) {
	resp, err := adminClient.CreateOrganization(context.Background(), &adminv1.CreateOrganizationRequest{
		Name:        name,
		Description: description,
	})
	if err != nil {
		return nil, err
	}

	// Switching to the created org
	err = dotrill.SetDefaultOrg(resp.Organization.Name)
	if err != nil {
		return nil, err
	}

	return resp.Organization, nil
}
