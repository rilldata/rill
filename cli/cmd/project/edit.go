package project

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func EditCmd(ch *cmdutil.Helper) *cobra.Command {
	var name, description, prodVersion, prodBranch, subpath, path, provisioner string
	var public bool
	var prodTTL int64

	editCmd := &cobra.Command{
		Use:   "edit [<project-name>]",
		Args:  cobra.MaximumNArgs(1),
		Short: "Edit the project details",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client, err := ch.Client()
			if err != nil {
				return fmt.Errorf("failed to initialize client: %w", err)
			}

			if len(args) > 0 {
				name = args[0]
			}

			if name == "" {
				return fmt.Errorf("pass project name as argument or with --project flag")
			}

			req := &adminv1.UpdateProjectRequest{
				OrganizationName: ch.Org,
				Name:             name,
			}

			// Set values from flags if they were changed
			if cmd.Flags().Changed("prod-ttl-seconds") && prodTTL < 0 {
				return fmt.Errorf("prod-ttl-seconds must be non-negative")
			}
			if cmd.Flags().Changed("provisioner") {
				req.Provisioner = &provisioner
			}
			if cmd.Flags().Changed("description") {
				req.Description = &description
			}
			if cmd.Flags().Changed("prod-version") {
				req.ProdVersion = &prodVersion
			}
			if cmd.Flags().Changed("prod-branch") {
				req.ProdBranch = &prodBranch
			}
			if cmd.Flags().Changed("subpath") {
				req.Subpath = &subpath
			}
			if cmd.Flags().Changed("public") {
				req.Public = &public
			}
			if cmd.Flags().Changed("prod-ttl-seconds") {
				req.ProdTtlSeconds = &prodTTL
			}

			updatedProj, err := client.UpdateProject(ctx, req)
			if err != nil {
				return fmt.Errorf("failed to update project: %w", err)
			}

			ch.PrintfSuccess("Updated project\n")
			ch.PrintProjects([]*adminv1.Project{updatedProj.Project})

			return nil
		},
	}

	editCmd.Flags().SortFlags = false
	editCmd.Flags().StringVar(&name, "project", "", "Project Name")
	editCmd.Flags().StringVar(&description, "description", "", "Project Description")
	editCmd.Flags().StringVar(&prodBranch, "prod-branch", "", "Production branch name")
	editCmd.Flags().BoolVar(&public, "public", false, "Make dashboards publicly accessible")
	editCmd.Flags().StringVar(&path, "path", ".", "Project directory")
	editCmd.Flags().StringVar(&subpath, "subpath", "", "Relative path to project in the repository (for monorepos)")
	editCmd.Flags().StringVar(&provisioner, "provisioner", "", "Project provisioner (default: current provisioner)")
	editCmd.Flags().Int64Var(&prodTTL, "prod-ttl-seconds", 0, "Prod deployment TTL in seconds")
	editCmd.Flags().StringVar(&prodVersion, "prod-version", "", "Rill version (default: current version)")

	return editCmd
}
