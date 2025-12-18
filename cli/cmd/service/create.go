package service

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/rilldata/rill/cli/cmd/project"
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/structpb"
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

			var attrs map[string]any
			if attributes != "" {
				if err := json.Unmarshal([]byte(attributes), &attrs); err != nil {
					return fmt.Errorf("failed to parse --attributes as JSON: %w", err)
				}
			}

			req := &adminv1.CreateServiceRequest{
				Name: args[0],
				Org:  ch.Org,
			}

			if ch.Interactive && orgRole == "" {
				ok, err := cmdutil.ConfirmPrompt("Do you want to assign an organization role to the service?", "", false)
				if err != nil {
					return err
				}
				if ok {
					err = cmdutil.SelectPromptIfEmpty(&orgRole, "Select role", orgRoles, "")
					if err != nil {
						return err
					}
				}
			}

			if ch.Interactive && projectRole == "" && !cmd.Flags().Changed("org-role") {
				ok, err := cmdutil.ConfirmPrompt("Do you want to assign a project role to the service?", "", false)
				if err != nil {
					return err
				}
				if ok {
					projectNames, err := project.ProjectNames(cmd.Context(), ch)
					if err != nil {
						return fmt.Errorf("failed to fetch project names: %w", err)
					}
					err = cmdutil.SelectPromptIfEmpty(&projectName, "Select project", projectNames, "")
					if err != nil {
						return err
					}
					err = cmdutil.SelectPromptIfEmpty(&projectRole, "Select role", projectRoles, "")
					if err != nil {
						return err
					}
				}
			}

			if orgRole == "" && projectRole == "" {
				return fmt.Errorf("either --org-role or --project-role must be specified")
			}

			req.OrgRoleName = orgRole
			req.Project = projectName
			req.ProjectRoleName = projectRole

			// Set attributes if provided
			if attrs != nil {
				req.Attributes, err = structpb.NewStruct(attrs)
				if err != nil {
					return fmt.Errorf("failed to convert attributes to struct: %w", err)
				}
			}

			res1, err := client.CreateService(cmd.Context(), req)
			if err != nil {
				return err
			}

			res2, err := client.IssueServiceAuthToken(cmd.Context(), &adminv1.IssueServiceAuthTokenRequest{
				Org:         ch.Org,
				ServiceName: res1.Service.Name,
			})
			if err != nil {
				return err
			}

			ch.Printf("Created service %q in org %q.\n", res1.Service.Name, ch.Org)
			ch.Printf("Access token: %s\n", res2.Token)

			return nil
		},
	}

	createCmd.Flags().StringVar(&orgRole, "org-role", "", fmt.Sprintf("Organization role to assign to the service (%s)", strings.Join(orgRoles, ", ")))
	createCmd.Flags().StringVar(&projectRole, "project-role", "", fmt.Sprintf("Project role to assign to the service (%s)", strings.Join(projectRoles, ", ")))
	createCmd.Flags().StringVar(&projectName, "project", "", "Project to assign the role to (required if project-role is set)")
	createCmd.Flags().StringVar(&attributes, "attributes", "", "JSON object of key-value pairs for service attributes")

	return createCmd
}
