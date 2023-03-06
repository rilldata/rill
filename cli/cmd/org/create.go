package org

import (
	"context"
	"fmt"
	"time"

	"github.com/rilldata/rill/admin/client"
	"github.com/rilldata/rill/cli/cmd/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func CreateCmd(cfg *config.Config) *cobra.Command {
	var displayName string

	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sp := cmdutil.GetSpinner(4, "Creating org...")
			sp.Start()
			// Just for spinner, will have to remove it
			time.Sleep(2 * time.Second)

			client, err := client.New(cfg.AdminURL, cfg.GetAdminToken())
			if err != nil {
				return err
			}
			defer client.Close()

			org, err := client.CreateOrganization(context.Background(), &adminv1.CreateOrganizationRequest{
				Name:        args[0],
				Description: displayName,
			})
			if err != nil {
				return err
			}

			sp.Stop()
			fmt.Printf("Created organization: %v\n", org)
			return nil
		},
	}
	createCmd.Flags().SortFlags = false
	createCmd.Flags().StringVar(&displayName, "display-name", "", "Display name")

	return createCmd
}
