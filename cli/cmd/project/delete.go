package project

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
		Args:  cobra.ExactArgs(1),
		Short: "Delete",
		RunE: func(cmd *cobra.Command, args []string) error {
			sp := cmdutil.Spinner("Deleting project...")
			sp.Start()

			client, err := client.New(cfg.AdminURL, cfg.AdminToken())
			if err != nil {
				return err
			}
			defer client.Close()

			proj, err := client.DeleteProject(context.Background(), &adminv1.DeleteProjectRequest{
				OrganizationName: cfg.Org(),
				Name:             args[0],
			})
			if err != nil {
				return err
			}

			sp.Stop()
			cmdutil.TextPrinter(fmt.Sprintf("Deleted project: %v\n", proj))
			return nil
		},
	}
	return deleteCmd
}
