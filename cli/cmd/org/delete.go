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

func DeleteCmd(cfg *config.Config) *cobra.Command {
	deleteCmd := &cobra.Command{
		Use:   "delete <org-name>",
		Short: "Delete",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := cmdutil.Client(cfg)
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

			if cfg.Org == args[0] {
				if err := dotrill.SetDefaultOrg(""); err != nil {
					return err
				}
			}

			cmdutil.SuccessPrinter(fmt.Sprintf("Deleted organization: %v\n", args[0]))
			return nil
		},
	}
	deleteCmd.Flags().SortFlags = false

	return deleteCmd
}
