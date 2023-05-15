package project

import (
	"context"
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func DeleteCmd(cfg *config.Config) *cobra.Command {
	var name, path string
	var force bool

	deleteCmd := &cobra.Command{
		Use:   "delete <project-name>",
		Args:  cobra.MaximumNArgs(1),
		Short: "Delete the project",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()

			if len(args) > 0 {
				name = args[0]
			}

			if !cmd.Flags().Changed("project") && len(args) == 0 && cfg.Interactive {
				name, err = inferProjectName(cmd.Context(), client, cfg.Org, path)
				if err != nil {
					return err
				}
			}

			if name == "" {
				return fmt.Errorf("please provide valid project name, Run `rill project list` for available projects")
			}

			if !force {
				fmt.Printf("Warn: Deleting the project %q will remove all metadata associated with the project\n", name)

				msg := fmt.Sprintf("Type %q to confirm deletion", name)
				project, err := cmdutil.InputPrompt(msg, "")
				if err != nil {
					return err
				}

				if project != name {
					return fmt.Errorf("Entered incorrect name : %q, expected value is %q", project, name)
				}
			}

			_, err = client.DeleteProject(context.Background(), &adminv1.DeleteProjectRequest{
				OrganizationName: cfg.Org,
				Name:             name,
			})
			if err != nil {
				return err
			}

			cmdutil.SuccessPrinter(fmt.Sprintf("Deleted project: %v", name))
			return nil
		},
	}

	deleteCmd.Flags().SortFlags = false
	deleteCmd.Flags().BoolVar(&force, "force", false, "Delete forcefully, skips the confirmation")
	deleteCmd.Flags().StringVar(&name, "project", "", "Project Name")
	deleteCmd.Flags().StringVar(&path, "path", ".", "Project directory")

	return deleteCmd
}
