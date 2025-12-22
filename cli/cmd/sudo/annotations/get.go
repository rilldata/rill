package annotations

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func GetCmd(ch *cmdutil.Helper) *cobra.Command {
	getCmd := &cobra.Command{
		Use:   "get <org> <project>",
		Args:  cobra.ExactArgs(2),
		Short: "Get annotations for project in an organization",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			client, err := ch.Client()
			if err != nil {
				return err
			}
			res, err := client.GetProject(ctx, &adminv1.GetProjectRequest{
				Org:                  args[0],
				Project:              args[1],
				SuperuserForceAccess: true,
			})
			if err != nil {
				return err
			}

			if len(res.Project.Annotations) == 0 {
				ch.PrintfWarn("No annotations found\n")
				return nil
			}

			for k, v := range res.Project.Annotations {
				fmt.Printf("%s=%s\n", k, v)
			}

			return nil
		},
	}

	return getCmd
}
