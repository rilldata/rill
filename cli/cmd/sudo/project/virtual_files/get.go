package virtual_files

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
		Use:   "get <org> <project> <path>",
		Args:  cobra.ExactArgs(3),
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

			org := args[0]
			project := args[1]
			path := args[2]

			environment, err := adminenv.Infer(ch.AdminURL())
			if err != nil {
				return err
			}

			const pageSize = 100

			allFiles, err := pullVirtualFiles(ctx, client, project, environment, uint32(pageSize))
			if err != nil {
				return err
			}

			// Find the file with matching path
			var file *adminv1.VirtualFile
			for _, f := range allFiles {
				if f.Path == path {
					file = f
					break
				}
			}

			if file == nil {
				return fmt.Errorf("no file found at path %q", path)
			}

			if file.Deleted {
				ch.PrintfWarn("File at path %q is marked as deleted\n", path)
				return nil
			}

			fileType := GetFileType(path)
			ch.PrintfSuccess("Content of virtual file %q in project %q (org %q):\n", path, project, org)
			if fileType != FileTypeUnknown {
				ch.PrintfSuccess("Type: %s\n", fileType)
			}

			data := string(file.Data)
			var obj interface{}
			if err := yaml.Unmarshal(file.Data, &obj); err != nil {
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

	return getCmd
}
