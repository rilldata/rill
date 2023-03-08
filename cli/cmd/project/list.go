package project

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

func ListCmd(cfg *config.Config) *cobra.Command {
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List",
		RunE: func(cmd *cobra.Command, args []string) error {
			sp := cmdutil.Spinner("Listing project...")
			sp.Start()

			client, err := client.New(cfg.AdminURL, cfg.GetAdminToken())
			if err != nil {
				return err
			}
			defer client.Close()

			proj, err := client.ListProjects(context.Background(), &adminv1.ListProjectsRequest{
				Organization: cfg.DefaultOrg,
			})
			if err != nil {
				return err
			}

			sp.Stop()
			fmt.Printf("Projects list: %v\n", proj)
			return nil
		},
	}
	return listCmd
}
