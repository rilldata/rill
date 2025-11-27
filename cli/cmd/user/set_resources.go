package user

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func SetResourceCmd(ch *cmdutil.Helper) *cobra.Command {
	var projectName string
	var email string
	var role string
	var explores []string
	var canvases []string
	var restrict bool

	cmd := &cobra.Command{
		Use:   "set-resources",
		Short: "Set a user's project resources and restriction flag (overwrites existing list)",
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

			resources, err := cmdutil.ParseResourceStrings(explores, canvases)
			if err != nil {
				return err
			}

			if len(resources) == 0 && !restrict {
				// confirm with user
				if ch.Interactive {
					confirm, err := cmdutil.ConfirmPrompt("No resources specified and --restrict-resources is false. This will clear all resource restrictions for the user. Are you sure?", "", false)
					if err != nil {
						return err
					}
					if !confirm {
						ch.PrintfWarn("Operation cancelled.\n")
						return nil
					}
				}
			}

			if len(resources) > 0 {
				restrict = true
			}

			client, err := ch.Client()
			if err != nil {
				return err
			}

			_, err = client.SetProjectMemberUserRole(cmd.Context(), &adminv1.SetProjectMemberUserRoleRequest{
				Org:               ch.Org,
				Project:           projectName,
				Email:             email,
				Role:              role,
				Resources:         resources,
				RestrictResources: restrict,
			})
			if err != nil {
				return err
			}

			resourceList := cmdutil.FormatResourceNames(resources)
			status := "cleared"
			if restrict && len(resources) > 0 {
				status = fmt.Sprintf("set to %s", resourceList)
			} else if restrict {
				status = "restricted with no resources"
			}
			ch.PrintfSuccess("Updated resources for %q in project \"%s/%s\" (role %s, resources %s)\n", email, ch.Org, projectName, role, status)
			return nil
		},
	}

	cmd.Flags().StringVar(&ch.Org, "org", ch.Org, "Organization")
	cmd.Flags().StringVar(&projectName, "project", "", "Project (required)")
	cmd.Flags().StringVar(&email, "email", "", "Email of the user (required)")
	cmd.Flags().StringVar(&role, "role", "current", "Role of the user (defaults to current project role)")
	cmd.Flags().StringArrayVar(&explores, "explore", nil, "Explore Resource to set (repeat for multiple)")
	cmd.Flags().StringArrayVar(&canvases, "canvas", nil, "Canvas Resource to set (repeat for multiple)")
	cmd.Flags().BoolVar(&restrict, "restrict-resources", false, "Whether to restrict the user to the provided resources (defaults to true when resources are provided)")

	return cmd
}
