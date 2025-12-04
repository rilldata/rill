package org

import (
	"fmt"
	"strings"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func DeleteCmd(ch *cmdutil.Helper) *cobra.Command {
	deleteCmd := &cobra.Command{
		Use:   "delete [<org-name>]",
		Short: "Delete organization",
		Long: `Delete an organization and all its associated projects.
This operation cannot be undone. Use --force to skip confirmation.`,
		Example: `  rill org delete myorg
  rill org delete myorg --force`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := ch.Client()
			if err != nil {
				return err
			}

			name := args[0]
			if name = strings.TrimSpace(name); name == "" {
				return fmt.Errorf("organization name cannot be empty")
			}

			// Find all the projects for the given org
			res, err := client.ListProjectsForOrganization(cmd.Context(), &adminv1.ListProjectsForOrganizationRequest{Org: name})
			if err != nil {
				return err
			}

			var projects []string
			for _, proj := range res.Projects {
				projects = append(projects, proj.Name)
			}

			if len(projects) > 0 {
				fmt.Printf("Deleting %q will also delete these projects:\n", name)
				for _, proj := range projects {
					fmt.Printf("\t%s/%s\n", name, proj)
				}
			}

			if ch.Interactive {
				ch.Printf("Warn: Deleting the org %q will remove all metadata associated with the org\n", name)
				msg := fmt.Sprintf("Type %q to confirm deletion", name)
				org, err := cmdutil.InputPrompt(msg, "")
				if err != nil {
					return err
				}

				if org != name {
					return fmt.Errorf("confirmation failed: entered %q but expected %q", org, name)
				}
			}

			totalProjects := len(projects)
			for i, proj := range projects {
				fmt.Printf("Deleting project %d/%d: %s/%s\n", i+1, totalProjects, name, proj)
				_, err := client.DeleteProject(cmd.Context(), &adminv1.DeleteProjectRequest{Org: name, Project: proj})
				if err != nil {
					return err
				}

				fmt.Printf("Deleted project %s/%s\n", name, proj)
			}

			_, err = client.DeleteOrganization(cmd.Context(), &adminv1.DeleteOrganizationRequest{Org: name})
			if err != nil {
				return err
			}

			// If deleting the default org, set the default org to empty
			if name == ch.Org {
				err = ch.DotRill.SetDefaultOrg("")
				if err != nil {
					return err
				}
			}

			ch.PrintfSuccess("Deleted organization: %v\n", name)
			return nil
		},
	}

	return deleteCmd
}
