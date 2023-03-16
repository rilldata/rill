package org

import (
	"context"
	"fmt"

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
			client, err := client.New(cfg.AdminURL, cfg.AdminToken(), cfg.Version.String())
			if err != nil {
				return err
			}
			defer client.Close()

			_, err = client.DeleteOrganization(context.Background(), &adminv1.DeleteOrganizationRequest{
				Name: args[0],
			})
			if err != nil {
				return err
			}

			cmdutil.TextPrinter(fmt.Sprintf("Deleted organization: %v\n", args[0]))
			return nil
		},
	}
	deleteCmd.Flags().SortFlags = false

	return deleteCmd
}
