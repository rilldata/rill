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

func DeleteCmd(cfg *config.Config) *cobra.Command {
	deleteCmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sp := cmdutil.GetSpinner(4, "Deleting org...")
			sp.Start()
			// Just for spinner, will have to remove it
			time.Sleep(2 * time.Second)

			client, err := client.New(cfg.AdminURL, cfg.GetAdminToken())
			if err != nil {
				return err
			}
			defer client.Close()

			org, err := client.DeleteOrganization(context.Background(), &adminv1.DeleteOrganizationRequest{
				Name: args[0],
			})
			if err != nil {
				return err
			}

			sp.Stop()
			fmt.Printf("Deleted organization: %v\n", org)
			return nil
		},
	}
	deleteCmd.Flags().SortFlags = false

	return deleteCmd
}
