package org

import (
	"context"
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	"github.com/rilldata/rill/cli/pkg/dotrill"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func DeleteCmd(cfg *config.Config) *cobra.Command {
	var force bool
	var name string

	deleteCmd := &cobra.Command{
		Use:   "delete <org-name>",
		Short: "Delete organization",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()

			if len(args) > 0 {
				name = args[0]
			}

			if !cmd.Flags().Changed("name") && len(args) == 0 {
				// Get the new org name from user if not provided in the flag
				name, err = cmdutil.InputPrompt("Enter the org name", "")
				if err != nil {
					return err
				}
			}

			// Find all the projects for the given org
			res, err := client.ListProjectsForOrganization(context.Background(), &adminv1.ListProjectsForOrganizationRequest{OrganizationName: args[0]})
			if err != nil {
				return err
			}

			var projects []string
			for _, proj := range res.Projects {
				projects = append(projects, proj.Name)
			}

			if len(projects) > 0 {
				fmt.Printf("Deleting %q will also delete these projects:\n", args[0])
				for _, proj := range projects {
					fmt.Printf("\t%s/%s\n", args[0], proj)
				}
			}

			if !force {
				fmt.Printf("Warn: Deleting the org %q will remove all metadata associated with the org\n", args[0])
				msg := fmt.Sprintf("Enter %q to confirm deletion", args[0])
				org, err := cmdutil.InputPrompt(msg, "")
				if err != nil {
					return err
				}

				if org != args[0] {
					return fmt.Errorf("Entered incorrect name : %s", org)
				}
			}

			for _, proj := range projects {
				_, err := client.DeleteProject(context.Background(), &adminv1.DeleteProjectRequest{OrganizationName: args[0], Name: proj})
				if err != nil {
					return err
				}

				fmt.Printf("Deleted project %s/%s\n", args[0], proj)
			}

			_, err = client.DeleteOrganization(context.Background(), &adminv1.DeleteOrganizationRequest{Name: args[0]})
			if err != nil {
				return err
			}

			// If deleting the default org, set the default org to empty
			if args[0] == cfg.Org {
				err = dotrill.SetDefaultOrg("")
				if err != nil {
					return err
				}
			}

			cmdutil.SuccessPrinter(fmt.Sprintf("Deleted organization: %v", args[0]))
			return nil
		},
	}
	deleteCmd.Flags().SortFlags = false
	deleteCmd.Flags().StringVar(&name, "name", "", "Name")
	deleteCmd.Flags().BoolVar(&force, "force", false, "Delete forcefully, skips the confirmation")

	return deleteCmd
}
