package virtualfiles

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func GetCmd(ch *cmdutil.Helper) *cobra.Command {
	getCmd := &cobra.Command{
		Use:   "get <org> <project> <path>",
		Args:  cobra.ExactArgs(3),
		Short: "Get the content of a specific virtual file",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			client, err := ch.Client()
			if err != nil {
				return err
			}

			org := args[0]
			project := args[1]
			path := args[2]

			if org == "" {
				return fmt.Errorf("org cannot be empty")
			}

			projResp, err := client.GetProject(ctx, &adminv1.GetProjectRequest{
				Org:                  org,
				Project:              project,
				SuperuserForceAccess: true,
			})
			if err != nil {
				return fmt.Errorf("failed to get project: %w", err)
			}
			projectID := projResp.Project.Id

			resp, err := client.GetVirtualFile(ctx, &adminv1.GetVirtualFileRequest{
				ProjectId:            projectID,
				Environment:          "prod",
				Path:                 path,
				SuperuserForceAccess: true,
			})
			if err != nil {
				return err
			}

			if resp.File.Deleted {
				ch.PrintfWarn("File at path %q is marked as deleted\n", path)
				return nil
			}

			fmt.Print(string(resp.File.Data))

			return nil
		},
	}

	return getCmd
}
