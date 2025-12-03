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
	var explores []string
	var canvases []string
	var restrict bool

	cmd := &cobra.Command{
		Use:   "set-resources",
		Short: "Set a user's project resources and restriction flag (overwrites existing list)",
		RunE: func(cmd *cobra.Command, args []string) error {
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

			if len(resources) == 0 && !cmd.Flags().Changed("restrict-resources") {
				// error out if no resources and restrict not explicitly set
				return fmt.Errorf("either resources must be provided or --restrict-resources must be set to true or false to enforce or clear restrictions")
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
				Resources:         resources,
				RestrictResources: &restrict,
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
			ch.PrintfSuccess("Updated resources for %q in project \"%s/%s\" (resources %s)\n", email, ch.Org, projectName, status)
			return nil
		},
	}

	cmd.Flags().StringVar(&ch.Org, "org", ch.Org, "Organization")
	cmd.Flags().StringVar(&projectName, "project", "", "Project (required)")
	cmd.Flags().StringVar(&email, "email", "", "Email of the user (required)")
	cmd.Flags().StringArrayVar(&explores, "explore", nil, "Explore Resource to set (repeat for multiple)")
	cmd.Flags().StringArrayVar(&canvases, "canvas", nil, "Canvas Resource to set (repeat for multiple)")
	cmd.Flags().BoolVar(&restrict, "restrict-resources", false, "Whether to restrict the user to the provided resources (defaults to true when resources are provided)")

	return cmd
}
