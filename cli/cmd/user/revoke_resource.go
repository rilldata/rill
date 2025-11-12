package user

import (
	"fmt"
	"strings"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func RevokeResourceCmd(ch *cmdutil.Helper) *cobra.Command {
	var projectName string
	var email string
	var resourceFlags []string
	var resourceInput string

	cmd := &cobra.Command{
		Use:   "revoke-resource",
		Short: "Remove resource-level access previously granted to a user",
		RunE: func(cmd *cobra.Command, args []string) error {
			if projectName == "" && ch.Interactive {
				if err := cmdutil.StringPromptIfEmpty(&projectName, "Enter project name"); err != nil {
					return err
				}
			}
			if projectName == "" {
				return fmt.Errorf("--project is required")
			}

			if email == "" && ch.Interactive {
				if err := cmdutil.StringPromptIfEmpty(&email, "Enter user email"); err != nil {
					return err
				}
			}
			if email == "" {
				return fmt.Errorf("--email is required")
			}

			if len(resourceFlags) == 0 && ch.Interactive {
				if err := cmdutil.StringPromptIfEmpty(&resourceInput, "Enter resources to remove (kind/name, comma-separated)"); err != nil {
					return err
				}
				for _, part := range strings.Split(resourceInput, ",") {
					part = strings.TrimSpace(part)
					if part != "" {
						resourceFlags = append(resourceFlags, part)
					}
				}
			}
			if len(resourceFlags) == 0 {
				return fmt.Errorf("at least one --resource kind/name must be provided")
			}

			resources, err := cmdutil.ParseResourceStrings(resourceFlags)
			if err != nil {
				return err
			}

			client, err := ch.Client()
			if err != nil {
				return err
			}

			_, err = client.RemoveProjectMemberUserResources(cmd.Context(), &adminv1.RemoveProjectMemberUserResourcesRequest{
				Org:       ch.Org,
				Project:   projectName,
				Email:     email,
				Resources: resources,
			})
			if err != nil {
				return err
			}

			resourceList := cmdutil.FormatResourceNames(resources)
			ch.PrintfSuccess("Removed resource access (%s) for user %q in project \"%s/%s\"\n", resourceList, email, ch.Org, projectName)
			return nil
		},
	}

	cmd.Flags().StringVar(&ch.Org, "org", ch.Org, "Organization")
	cmd.Flags().StringVar(&projectName, "project", "", "Project (required)")
	cmd.Flags().StringVar(&email, "email", "", "Email of the user (required)")
	cmd.Flags().StringArrayVar(&resourceFlags, "resource", nil, "Resource to revoke in the format kind/name (repeat for multiple)")

	return cmd
}
