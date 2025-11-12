package user

import (
	"fmt"
	"strings"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func GrantResourceCmd(ch *cmdutil.Helper) *cobra.Command {
	var projectName string
	var email string
	var resourceFlags []string
	var resourceInput string

	cmd := &cobra.Command{
		Use:   "grant-resource",
		Short: "Grant a user access to specific project resources (viewer scoped)",
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
				if err := cmdutil.StringPromptIfEmpty(&resourceInput, "Enter resources (kind/name, comma-separated)"); err != nil {
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

			resp, err := client.AddProjectMemberUserResources(cmd.Context(), &adminv1.AddProjectMemberUserResourcesRequest{
				Org:       ch.Org,
				Project:   projectName,
				Email:     email,
				Resources: resources,
			})
			if err != nil {
				return err
			}

			resourceList := cmdutil.FormatResourceNames(resources)
			if resp.PendingSignup {
				ch.PrintfSuccess("Invitation sent to %q for project \"%s/%s\" with viewer access to %s\n", email, ch.Org, projectName, resourceList)
			} else {
				ch.PrintfSuccess("Granted %q viewer access to %s in project \"%s/%s\"\n", email, resourceList, ch.Org, projectName)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&ch.Org, "org", ch.Org, "Organization")
	cmd.Flags().StringVar(&projectName, "project", "", "Project (required)")
	cmd.Flags().StringVar(&email, "email", "", "Email of the user (required)")
	cmd.Flags().StringArrayVar(&resourceFlags, "resource", nil, "Resource to grant in the format kind/name (repeat for multiple)")

	return cmd
}
