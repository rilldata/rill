package project

import (
	"context"
	"fmt"

	"github.com/rilldata/rill/admin/client"
	"github.com/rilldata/rill/cli/pkg/config"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func EditCmd(cfg *config.Config) *cobra.Command {
	var name, displayName, prodBranch string
	var public bool

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

			// Todo how will get the org name? will it be flag with cmd.
			proj, err := client.UpdateProject(context.Background(), &adminv1.UpdateProjectRequest{
				Name:        args[0],
				Description: displayName,
			})
			if err != nil {
				return err
			}

			fmt.Printf("Updated project: %v\n", proj)
			return nil
		},
	}

	editCmd.Flags().SortFlags = false

	editCmd.Flags().StringVar(&name, "name", "noname", "Name")
	editCmd.Flags().StringVar(&displayName, "display-name", "noname", "Display name")
	editCmd.Flags().StringVar(&prodBranch, "prod-branch", "noname", "Production branch name")
	editCmd.Flags().BoolVar(&public, "public", false, "Public")

	return editCmd
}
