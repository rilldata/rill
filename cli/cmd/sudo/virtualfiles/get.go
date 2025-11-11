package virtualfiles

import (
	"context"
	"fmt"
	"time"

	"github.com/rilldata/rill/cli/pkg/adminenv"
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func GetCmd(ch *cmdutil.Helper) *cobra.Command {
	var timeout time.Duration

	getCmd := &cobra.Command{
		Use:   "get <project> <path>",
		Args:  cobra.ExactArgs(2),
		Short: "Get the content of a specific virtual file",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			// Apply timeout if specified
			if timeout > 0 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, timeout)
				defer cancel()
			}

			client, err := ch.Client()
			if err != nil {
				return err
			}

			project := args[0]
			path := args[1]

			environment, err := adminenv.Infer(ch.AdminURL())
			if err != nil {
				return err
			}

			resp, err := client.GetVirtualFile(ctx, &adminv1.GetVirtualFileRequest{
				ProjectId:   project,
				Environment: environment,
				Path:        path,
			})
			if err != nil {
				return err
			}

			if resp.File.Deleted {
				ch.PrintfWarn("File at path %q is marked as deleted\n", path)
				return nil
			}

			fileType := GetFileType(path)
			ch.PrintfSuccess("Content of virtual file %q in project %q:\n", path, project)
			if fileType != FileTypeUnknown {
				ch.PrintfSuccess("Type: %s\n", fileType)
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

	getCmd.Flags().DurationVar(&timeout, "timeout", 30*time.Second, "Request timeout")
	getCmd.PersistentFlags().StringVar(&ch.Org, "org", ch.Org, "Organization Name")

	return getCmd
}
