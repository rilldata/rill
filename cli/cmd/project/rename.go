package project

import (
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"github.com/rilldata/rill/cli/cmd/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func RenameCmd(cfg *config.Config) *cobra.Command {
	renameCmd := &cobra.Command{
		Use:   "rename <current-project-name> <new-project-name>",
		Args:  cobra.MaximumNArgs(2),
		Short: "Rename",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			var currentName string
			var newName string

			if len(args) == 1 {
				return fmt.Errorf("Invalid args provided, required 0 or 2 args")
			}

			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()

			if len(args) > 1 {
				currentName = args[0]
				newName = args[1]
			} else {
				resp, err := client.ListProjectsForOrganization(ctx, &adminv1.ListProjectsForOrganizationRequest{OrganizationName: cfg.Org})
				if err != nil {
					return err
				}

				var projectNames []string
				for _, proj := range resp.Projects {
					projectNames = append(projectNames, proj.Name)
				}

				currentName = cmdutil.SelectPrompt("Select the project for rename", projectNames, "")

				// Get the new project name from user if not provided in the args
				question := []*survey.Question{
					{
						Name: "new",
						Prompt: &survey.Input{
							Message: "New project name",
						},
						Validate: func(any interface{}) error {
							name := any.(string)
							if name == "" {
								return fmt.Errorf("empty name")
							}

							return nil
						},
					},
				}

				if err := survey.Ask(question, newName); err != nil {
					return err
				}
			}

			confirm := false
			prompt := &survey.Confirm{
				Message: fmt.Sprintf("Do you want to rename project %s to %s?", color.YellowString(currentName), color.YellowString(newName)),
			}

			err = survey.AskOne(prompt, &confirm)
			if err != nil {
				return err
			}

			if !confirm {
				return nil
			}

			resp, err := client.GetProject(ctx, &adminv1.GetProjectRequest{OrganizationName: cfg.Org, Name: currentName})
			if err != nil {
				return err
			}

			proj := resp.Project

			updatedProj, err := client.UpdateProject(ctx, &adminv1.UpdateProjectRequest{
				Id:               proj.Id,
				OrganizationName: cfg.Org,
				Name:             newName,
				Description:      proj.Description,
				Public:           proj.Public,
				ProductionBranch: proj.ProductionBranch,
				GithubUrl:        proj.GithubUrl,
			})
			if err != nil {
				return err
			}

			cmdutil.SuccessPrinter("Renamed project \n")
			cmdutil.TablePrinter(toRow(updatedProj.Project))

			return nil
		},
	}

	return renameCmd
}
