package org

import (
	"context"
	"fmt"

	"github.com/rilldata/rill/admin/client"
	"github.com/rilldata/rill/cli/pkg/config"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func EditCmd(cfg *config.Config) *cobra.Command {
	var displayName string

	editCmd := &cobra.Command{
		Use:   "edit",
		Args:  cobra.ExactArgs(1),
		Short: "Edit",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := client.New(cfg.AdminURL, cfg.AdminToken)
			if err != nil {
				return err
			}
			defer client.Close()

			org, err := client.UpdateOrganization(context.Background(), &adminv1.UpdateOrganizationRequest{
				Name:        args[0],
				Description: displayName,
			})
			if err != nil {
				return err
			}

			fmt.Printf("Updated organization: %v\n", org)
			return nil
		},
	}
	editCmd.Flags().SortFlags = false
	editCmd.Flags().StringVar(&displayName, "display-name", "noname", "Display name")

	return editCmd
}
