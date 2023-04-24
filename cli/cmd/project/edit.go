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

func EditCmd(cfg *config.Config) *cobra.Command {
	var name, description, prodBranch string
	var public bool

	editCmd := &cobra.Command{
		Use:   "edit",
		Args:  cobra.NoArgs,
		Short: "Edit",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()

			if !cmd.Flags().Changed("project") {
				err := cmdutil.PromptIfUnset(&name, "Project Name", "")
				if err != nil {
					return err
				}
			}

			resp, err := client.GetProject(ctx, &adminv1.GetProjectRequest{OrganizationName: cfg.Org, Name: name})
			if err != nil {
				return err
			}

			proj := resp.Project

			if !cmd.Flags().Changed("description") {
				err := cmdutil.PromptIfUnset(&description, "Project Description", proj.Description)
				if err != nil {
					return err
				}
			}

			if !cmd.Flags().Changed("prod-branch") {
				err := cmdutil.PromptIfUnset(&prodBranch, "Project Production Branch", proj.ProdBranch)
				if err != nil {
					return err
				}
			}

			if !cmd.Flags().Changed("public") {
				prompt := &survey.Confirm{
					Message: fmt.Sprintf("Do you want project \"%s\" to public?", color.YellowString(name)),
				}

				err = survey.AskOne(prompt, &public)
				if err != nil {
					return err
				}
			}

			// Todo: Need to add prompt for repo_path <path_for_monorepo>

			updatedProj, err := client.UpdateProject(ctx, &adminv1.UpdateProjectRequest{
				Id:               proj.Id,
				OrganizationName: cfg.Org,
				Name:             proj.Name,
				Description:      description,
				Public:           public,
				ProdBranch:       prodBranch,
				GithubUrl:        proj.GithubUrl,
			})
			if err != nil {
				return err
			}

			cmdutil.SuccessPrinter("Updated project \n")
			cmdutil.TablePrinter(toRow(updatedProj.Project))
			return nil
		},
	}

	editCmd.Flags().SortFlags = false

	editCmd.Flags().StringVar(&name, "project", "noname", "Name")
	editCmd.Flags().StringVar(&description, "description", "", "Description")
	editCmd.Flags().StringVar(&prodBranch, "prod-branch", "noname", "Production branch name")
	editCmd.Flags().BoolVar(&public, "public", false, "Public")

	return editCmd
}
