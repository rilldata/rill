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

type ProjectRenameInfo struct {
	current string
	new     string
}

func RenameCmd(cfg *config.Config) *cobra.Command {
	renameCmd := &cobra.Command{
		Use:   "rename <current-project-name> <new-project-name>",
		Args:  cobra.MaximumNArgs(2),
		Short: "Rename",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			projectRenameInfo := ProjectRenameInfo{}

			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()

			if len(args) > 1 {
				projectRenameInfo.current = args[0]
				projectRenameInfo.new = args[1]
			} else {
				// Get the new project current name and new name from user if not provided in the args
				currentUserQuestion := []*survey.Question{
					{
						Name: "current",
						Prompt: &survey.Input{
							Message: "Current project name",
						},
						Validate: func(any interface{}) error {
							name := any.(string)
							if name == "" {
								return fmt.Errorf("empty name")
							}
							exists, err := cmdutil.ProjectExists(ctx, client, cfg.Org, name)
							if err != nil {
								return err
							}
							if !exists {
								return fmt.Errorf("project with name %v not exists in the org", name)
							}
							return nil
						},
					},
				}

				if err := survey.Ask(currentUserQuestion, &projectRenameInfo.current); err != nil {
					return err
				}

				newUserQuestion := []*survey.Question{
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
							exists, err := cmdutil.ProjectExists(ctx, client, cfg.Org, name)
							if err != nil {
								return err
							}
							if exists {
								return fmt.Errorf("project with name %v already exists in the org, please try with different name", name)
							}
							if projectRenameInfo.current == name {
								return fmt.Errorf("current project name %v same as new project name %v", projectRenameInfo.current, name)
							}
							return nil
						},
					},
				}

				if err := survey.Ask(newUserQuestion, &projectRenameInfo.new); err != nil {
					return err
				}
			}

			confirm := false
			prompt := &survey.Confirm{
				Message: fmt.Sprintf("Do you want to rename project %s to %s?", color.YellowString(projectRenameInfo.current), color.YellowString(projectRenameInfo.new)),
			}

			err = survey.AskOne(prompt, &confirm)
			if err != nil {
				return err
			}

			if !confirm {
				return nil
			}

			resp, err := client.GetProject(ctx, &adminv1.GetProjectRequest{OrganizationName: cfg.Org, Name: projectRenameInfo.current})
			if err != nil {
				return err
			}

			proj := resp.Project

			updatedProj, err := client.UpdateProject(ctx, &adminv1.UpdateProjectRequest{
				Id:               proj.Id,
				OrganizationName: cfg.Org,
				Name:             projectRenameInfo.new,
				Description:      proj.Description,
				Public:           proj.Public,
				ProductionBranch: proj.ProductionBranch,
				GithubUrl:        proj.GithubUrl,
				Variables:        proj.Variables,
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
