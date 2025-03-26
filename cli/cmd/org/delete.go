package org

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func DeleteCmd(ch *cmdutil.Helper) *cobra.Command {
	var force bool
	var name string

	deleteCmd := &cobra.Command{
		Use:   "delete [<org-name>]",
		Short: "Delete organization",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := ch.Client()
			if err != nil {
				return err
			}

			if len(args) > 0 {
				name = args[0]
			}

			if len(args) == 0 && ch.Interactive {
				err = cmdutil.SetFlagsByInputPrompts(*cmd, "org")
				if err != nil {
					return err
				}
			}

			// Find all the projects for the given org
			res, err := client.ListProjectsForOrganization(cmd.Context(), &adminv1.ListProjectsForOrganizationRequest{OrganizationName: name})
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

			if !force {
				fmt.Printf("Warn: Deleting the org %q will remove all metadata associated with the org\n", name)
				msg := fmt.Sprintf("Type %q to confirm deletion", name)
				org, err := cmdutil.InputPrompt(msg, "")
				if err != nil {
					return err
				}

				if org != name {
					return fmt.Errorf("Entered incorrect name: %q, expected value is %q", org, name)
				}
			}

			for _, proj := range projects {
				_, err := client.DeleteProject(cmd.Context(), &adminv1.DeleteProjectRequest{OrganizationName: name, Name: proj})
				if err != nil {
					return err
				}

				fmt.Printf("Deleted project %s/%s\n", name, proj)
			}

			_, err = client.DeleteOrganization(cmd.Context(), &adminv1.DeleteOrganizationRequest{Name: name})
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
	deleteCmd.Flags().SortFlags = false
	deleteCmd.Flags().StringVar(&name, "org", ch.Org, "Organization Name")
	deleteCmd.Flags().BoolVar(&force, "force", false, "Delete forcefully, skips the confirmation")

	return deleteCmd
}
