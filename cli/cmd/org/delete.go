package org

import (
	"context"
	"fmt"

	"github.com/rilldata/rill/cli/cmd/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	"github.com/rilldata/rill/cli/pkg/dotrill"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func DeleteCmd(cfg *config.Config) *cobra.Command {
	deleteCmd := &cobra.Command{
		Use:   "delete <org-name>",
		Short: "Delete",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()

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

			msg := fmt.Sprintf("Enter %q to confirm deletion", args[0])
			org := cmdutil.InputPrompt(msg, "")
			if org != args[0] {
				return fmt.Errorf("Entered incorrect name : %s", org)
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

			cmdutil.SuccessPrinter(fmt.Sprintf("Deleted organization: %v\n", args[0]))
			return nil
		},
	}
	deleteCmd.Flags().SortFlags = false

	return deleteCmd
}
