package org

import (
	"context"

	"github.com/rilldata/rill/admin/client"
	"github.com/rilldata/rill/cli/cmd/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func CreateCmd(cfg *config.Config) *cobra.Command {
	var description string

	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sp := cmdutil.Spinner("Creating org...")
			sp.Start()

			client, err := client.New(cfg.AdminURL, cfg.AdminToken())
			if err != nil {
				return err
			}
			defer client.Close()

			org, err := client.CreateOrganization(context.Background(), &adminv1.CreateOrganizationRequest{
				Name:        args[0],
				Description: description,
			})
			if err != nil {
				return err
			}

			sp.Stop()
			cmdutil.TextPrinter("Created organization \n")
			cmdutil.TablePrinter(toOrg(org.Organization))
			return nil
		},
	}
	createCmd.Flags().SortFlags = false
	createCmd.Flags().StringVar(&description, "description", "", "Description")

	return createCmd
}
