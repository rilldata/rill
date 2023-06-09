package project

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func EditCmd(cfg *config.Config) *cobra.Command {
	var name, description, prodBranch, path, region string
	var public bool
	var slots int

	editCmd := &cobra.Command{
		Use:   "edit [<project-name>]",
		Args:  cobra.MaximumNArgs(1),
		Short: "Edit the project details",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()

			if len(args) > 0 {
				name = args[0]
			}

			if !cmd.Flags().Changed("project") && len(args) == 0 && cfg.Interactive {
				names, err := cmdutil.ProjectNamesByOrg(ctx, client, cfg.Org)
				if err != nil {
					return err
				}

				// prompt for name from user
				name = cmdutil.SelectPrompt("Select project", names, "")
			}
			if name == "" {
				return fmt.Errorf("pass project name as argument or with --project flag")
			}

			req := &adminv1.UpdateProjectRequest{
				OrganizationName: cfg.Org,
				Name:             name,
			}
			promptFlagValues := true
			if cmd.Flags().Changed("prod-slots") {
				promptFlagValues = false
				prodSlots := int64(slots)
				req.ProdSlots = &prodSlots
			}
			if cmd.Flags().Changed("region") {
				promptFlagValues = false
				req.Region = &region
			}
			if cmd.Flags().Changed("description") {
				promptFlagValues = false
				req.Description = &description
			}
			if cmd.Flags().Changed("prod-branch") {
				promptFlagValues = false
				req.ProdBranch = &prodBranch
			}
			if cmd.Flags().Changed("public") {
				promptFlagValues = false
				req.Public = &public
			}

			if promptFlagValues {
				resp, err := client.GetProject(ctx, &adminv1.GetProjectRequest{OrganizationName: cfg.Org, Name: name})
				if err != nil {
					return err
				}
				proj := resp.Project

				description, err = cmdutil.InputPrompt("Enter the description", proj.Description)
				if err != nil {
					return err
				}
				req.Description = &description

				prodBranch, err = cmdutil.InputPrompt("Enter the production branch", proj.ProdBranch)
				if err != nil {
					return err
				}
				req.ProdBranch = &prodBranch

				public = cmdutil.ConfirmPrompt("Is project public", "", proj.Public)
				req.Public = &public
			}

			// Todo: Need to add prompt for repo_path <path_for_monorepo>

			updatedProj, err := client.UpdateProject(ctx, req)
			if err != nil {
				return err
			}

			cmdutil.PrintlnSuccess("Updated project")
			cmdutil.TablePrinter(toRow(updatedProj.Project))
			return nil
		},
	}

	editCmd.Flags().SortFlags = false
	editCmd.Flags().StringVar(&name, "project", "", "Project Name")
	editCmd.Flags().StringVar(&description, "description", "", "Project Description")
	editCmd.Flags().StringVar(&prodBranch, "prod-branch", "", "Production branch name")
	editCmd.Flags().BoolVar(&public, "public", false, "Make dashboards publicly accessible")
	editCmd.Flags().StringVar(&path, "path", ".", "Project directory")
	editCmd.Flags().IntVar(&slots, "prod-slots", 0, "Slots to allocate for production deployments (default: current slots)")
	editCmd.Flags().StringVar(&region, "region", "", "Deployment region (default: current region)")
	if !cfg.IsDev() {
		if err := editCmd.Flags().MarkHidden("prod-slots"); err != nil {
			panic(err)
		}
	}

	return editCmd
}
