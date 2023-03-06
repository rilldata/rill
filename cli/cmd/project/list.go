package project

import (
	"context"
	"fmt"

	"github.com/rilldata/rill/admin/client"
	"github.com/rilldata/rill/cli/pkg/config"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func ListCmd(cfg *config.Config) *cobra.Command {
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := client.New(cfg.AdminURL, cfg.GetAdminToken())
			if err != nil {
				return err
			}
			defer client.Close()

			proj, err := client.FindProjects(context.Background(), &adminv1.FindProjectsRequest{
				Organization: cfg.DefaultOrg,
			})
			if err != nil {
				return err
			}

			fmt.Printf("Projects list: %v\n", proj)
			return nil
		},
	}
	return listCmd
}
