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

			ch.Printer.Println("Warn: Renaming an project would invalidate dashboard URLs")

			if !cmd.Flags().Changed("project") && ch.Interactive {
				projectNames, err := projectNames(ctx, ch)
				if err != nil {
					return err
				}

				name = cmdutil.SelectPrompt("Select project to rename", projectNames, "")
			}

			if ch.Interactive {
				err = cmdutil.SetFlagsByInputPrompts(*cmd, "new-name")
				if err != nil {
					return err
				}
			}

			msg := fmt.Sprintf("Do you want to rename the project \"%s\" to \"%s\"?", color.YellowString(name), color.YellowString(newName))
			if !cmdutil.ConfirmPrompt(msg, "", false) {
				return nil
			}

			updatedProj, err := client.UpdateProject(ctx, &adminv1.UpdateProjectRequest{
				OrganizationName: ch.Org,
				Name:             name,
				NewName:          &newName,
			})
			if err != nil {
				return err
			}

			ch.Printer.PrintlnSuccess("Renamed project")
			ch.Printer.PrintlnSuccess(fmt.Sprintf("New web url is: %s\n", updatedProj.Project.FrontendUrl))
			return ch.Printer.PrintResource(toRow(updatedProj.Project))
		},
	}

	renameCmd.Flags().SortFlags = false
	renameCmd.Flags().StringVar(&name, "project", "", "Current Project Name")
	renameCmd.Flags().StringVar(&newName, "new-name", "", "New Project Name")

	return renameCmd
}
