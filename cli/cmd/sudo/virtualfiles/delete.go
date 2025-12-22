package virtualfiles

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func DeleteCmd(ch *cmdutil.Helper) *cobra.Command {
	var force bool

	deleteCmd := &cobra.Command{
		Use:   "delete <org> <project> <path>",
		Args:  cobra.ExactArgs(3),
		Short: "Delete a specific virtual file",
		Long: `Delete a specific virtual file by marking it as deleted.

Virtual files are stored in the virtual repository and represent runtime resources
like alerts and reports. Deleting a virtual file will mark it as deleted
in the database, which will cause the runtime to remove the corresponding resource.

This command can delete virtual files even if they have parse errors.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			client, err := ch.Client()
			if err != nil {
				return err
			}

			org := args[0]
			project := args[1]
			path := args[2]

			if org == "" || project == "" || path == "" {
				return fmt.Errorf("org, project, and path cannot be empty")
			}

			if !force {
				ok, err := cmdutil.ConfirmPrompt(fmt.Sprintf("Delete virtual file %q in project %q (org %q)?", path, project, org), "", false)
				if err != nil {
					return err
				}
				if !ok {
					return nil
				}
			}

			// Get the project ID
			projResp, err := client.GetProject(ctx, &adminv1.GetProjectRequest{
				Org:                  org,
				Project:              project,
				SuperuserForceAccess: true,
			})
			if err != nil {
				return fmt.Errorf("failed to get project: %w", err)
			}
			projectID := projResp.Project.Id

			// Delete the virtual file directly
			_, err = client.DeleteVirtualFile(ctx, &adminv1.DeleteVirtualFileRequest{
				ProjectId:            projectID,
				Environment:          "prod",
				Path:                 path,
				SuperuserForceAccess: true,
			})
			if err != nil {
				return fmt.Errorf("failed to delete virtual file: %w", err)
			}

			ch.PrintfSuccess("Successfully deleted virtual file %q\n", path)
			return nil
		},
	}

	deleteCmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation prompt")

	return deleteCmd
}
