package virtualfiles

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func GetCmd(ch *cmdutil.Helper) *cobra.Command {
	getCmd := &cobra.Command{
		Use:   "get <project> <path>",
		Args:  cobra.ExactArgs(2),
		Short: "Get the content of a specific virtual file",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			client, err := ch.Client()
			if err != nil {
				return err
			}

			project := args[0]
			path := args[1]

			org := ch.Org
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

			data := string(resp.File.Data)
			var obj interface{}
			if err := yaml.Unmarshal(resp.File.Data, &obj); err != nil {
				// fallback to plain text
				fmt.Println(data)
				return nil
			}

			yamlData, err := yaml.Marshal(obj)
			if err != nil {
				fmt.Println(data)
			} else {
				fmt.Println(string(yamlData))
			}

			return nil
		},
	}

	getCmd.PersistentFlags().StringVar(&ch.Org, "org", ch.Org, "Organization Name")

	return getCmd
}
