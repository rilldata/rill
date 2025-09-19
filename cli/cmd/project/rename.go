package project

import (
	"fmt"

	"github.com/fatih/color"
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
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			client, err := ch.Client()
			if err != nil {
				return err
			}

			ch.PrintfWarn("Warn: Renaming a project will invalidate dashboard URLs\n")

			if !cmd.Flags().Changed("project") && ch.Interactive {
				projectNames, err := ProjectNames(ctx, ch)
				if err != nil {
					return err
				}

				name, err = cmdutil.SelectPrompt("Select project to rename", projectNames, "")
				if err != nil {
					return err
				}
			}

			if !cmd.Flags().Changed("new-name") && ch.Interactive {
				err = cmdutil.SetFlagsByInputPrompts(*cmd, "new-name")
				if err != nil {
					return err
				}
			}

			msg := fmt.Sprintf("Do you want to rename the project \"%s\" to \"%s\"?", color.YellowString(name), color.YellowString(newName)) // nolint:gocritic // Because it uses colors
			ok, err := cmdutil.ConfirmPrompt(msg, "", false)
			if err != nil {
				return err
			}
			if !ok {
				return nil
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
