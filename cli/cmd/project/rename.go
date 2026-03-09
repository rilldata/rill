package project

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func RenameCmd(ch *cmdutil.Helper) *cobra.Command {
	var name, newName string

	renameCmd := &cobra.Command{
		Use:   "rename",
		Args:  cobra.NoArgs,
		Short: "Rename project",
		Long: `Rename project

Warning: Renaming a project will invalidate all dashboard URLs.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			client, err := ch.Client()
			if err != nil {
				return err
			}

			ch.PrintfWarn("Warning: Renaming a project will invalidate all dashboard URLs.\n")

			if name == "" {
				if !ch.Interactive {
					return fmt.Errorf("project name must be specified in non-interactive mode")
				}
				projectNames, err := ProjectNames(ctx, ch)
				if err != nil {
					return err
				}
				name, err = cmdutil.SelectPrompt("Select project to rename", projectNames, "")
				if err != nil {
					return err
				}
			}

			if newName == "" {
				if !ch.Interactive {
					return fmt.Errorf("new project name must be specified in non-interactive mode")
				}
				newName, err = cmdutil.InputPrompt("Enter new name", "")
				if err != nil {
					return err
				}
			}

			if ch.Interactive {
				ok, err := cmdutil.ConfirmPrompt(fmt.Sprintf("Do you want to rename the project %q to %q?", name, newName), "", false)
				if err != nil {
					return err
				}
				if !ok {
					return nil
				}
			}

			updatedProj, err := client.UpdateProject(ctx, &adminv1.UpdateProjectRequest{
				Org:     ch.Org,
				Project: name,
				NewName: &newName,
			})
			if err != nil {
				return err
			}

			ch.PrintfSuccess("Renamed project\n")
			ch.PrintfSuccess("New web url is: %s\n", updatedProj.Project.FrontendUrl)
			ch.PrintProjects([]*adminv1.Project{updatedProj.Project})

			return nil
		},
	}

	renameCmd.Flags().SortFlags = false
	renameCmd.Flags().StringVar(&name, "project", "", "Current Project Name")
	renameCmd.Flags().StringVar(&newName, "new-name", "", "New Project Name")

	return renameCmd
}
