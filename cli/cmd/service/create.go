package service

import (
	"encoding/json"
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func CreateCmd(ch *cmdutil.Helper) *cobra.Command {
	var orgRole, projectRole, projectName string
	var attributes string

	createCmd := &cobra.Command{
		Use:   "create <service-name>",
		Args:  cobra.ExactArgs(1),
		Short: "Create service",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := ch.Client()
			if err != nil {
				return err
			}

			// Parse attributes if provided
			var attrs map[string]string
			if attributes != "" {
				if err := json.Unmarshal([]byte(attributes), &attrs); err != nil {
					return err
				}
			}

			req := &adminv1.CreateServiceRequest{
				Name:             args[0],
				OrganizationName: ch.Org,
			}

			// Set org role if provided
			if orgRole != "" {
				req.OrgRoleName = orgRole
			}

			// Set project role if provided
			if projectRole != "" {
				if projectName == "" {
					return fmt.Errorf("project name is required when project role is set")
				}
				req.ProjectName = projectName
				req.ProjectRoleName = projectRole
			}

			// Set attributes if provided
			if attrs != nil {
				req.Attributes = attrs
			}

			res1, err := client.CreateService(cmd.Context(), req)
			if err != nil {
				return err
			}

			res2, err := client.IssueServiceAuthToken(cmd.Context(), &adminv1.IssueServiceAuthTokenRequest{
				OrganizationName: ch.Org,
				ServiceName:      res1.Service.Name,
			})
			if err != nil {
				return err
			}

			ch.Printf("Created service %q in org %q.\n", res1.Service.Name, ch.Org)
			ch.Printf("Access token: %s\n", res2.Token)

			return nil
		},
	}

	createCmd.Flags().StringVar(&orgRole, "org-role", "", "Organization role to assign to the service")
	createCmd.Flags().StringVar(&projectRole, "project-role", "", "Project role to assign to the service")
	createCmd.Flags().StringVar(&projectName, "project", "", "Project to assign the role to (required if project-role is set)")
	createCmd.Flags().StringVar(&attributes, "attributes", "", "JSON object of key-value pairs for service attributes")

	return createCmd
}
