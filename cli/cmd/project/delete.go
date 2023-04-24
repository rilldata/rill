package project

import (
	"context"
	"fmt"

	"github.com/rilldata/rill/cli/cmd/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func DeleteCmd(cfg *config.Config) *cobra.Command {
	var name, path string
	var force bool

	deleteCmd := &cobra.Command{
		Use:   "delete",
		Args:  cobra.NoArgs,
		Short: "Delete",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()

			if !cmd.Flags().Changed("project") {
				name, err = inferProjectName(cmd.Context(), client, cfg.Org, path)
				if err != nil {
					return err
				}
			}

			if !force {
				fmt.Printf("Warn: Deleting the project %q will remove all metadata associated with the project\n", name)

				msg := fmt.Sprintf("Enter %q to confirm deletion", name)
				project := cmdutil.InputPrompt(msg, "")
				if project != name {
					return fmt.Errorf("Entered incorrect name : %s", name)
				}
			}

			_, err = client.DeleteProject(context.Background(), &adminv1.DeleteProjectRequest{
				OrganizationName: cfg.Org,
				Name:             name,
			})
			if err != nil {
				return err
			}

			cmdutil.SuccessPrinter(fmt.Sprintf("Deleted project: %v\n", name))
			return nil
		},
	}

	deleteCmd.Flags().SortFlags = false
	deleteCmd.Flags().BoolVar(&force, "force", false, "Delete forcefully, skips the confirmation")
	deleteCmd.Flags().StringVar(&name, "project", "", "Name")
	deleteCmd.Flags().StringVar(&path, "path", ".", "Project directory")

	return deleteCmd
}
