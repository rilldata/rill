package project

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/spf13/cobra"
)

func SplitsCmd(ch *cmdutil.Helper) *cobra.Command {
	var project, path, model string
	var pending, errored, local bool
	var pageSize uint32
	var pageToken string

	splitsCmd := &cobra.Command{
		Use:   "splits [<project>] <model>",
		Args:  cobra.RangeArgs(1, 2),
		Short: "List splits for a model",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 1 {
				model = args[0]
			} else if len(args) == 2 {
				project = args[0]
				model = args[1]
			}

			if !local && !cmd.Flags().Changed("project") && len(args) <= 1 && ch.Interactive {
				var err error
				project, err = ch.InferProjectName(cmd.Context(), ch.Org, path)
				if err != nil {
					return err
				}
			}

			rt, instanceID, err := ch.OpenRuntimeClient(cmd.Context(), ch.Org, project, local)
			if err != nil {
				return err
			}

			res, err := rt.GetModelSplits(cmd.Context(), &runtimev1.GetModelSplitsRequest{
				InstanceId: instanceID,
				Model:      model,
				Pending:    pending,
				Errored:    errored,
				PageSize:   pageSize,
				PageToken:  pageToken,
			})
			if err != nil {
				return fmt.Errorf("failed to get model splits: %w", err)
			}

			ch.PrintModelSplits(res.Splits)

			if res.NextPageToken != "" {
				cmd.Println()
				cmd.Printf("Next page token: %s\n", res.NextPageToken)
			}

			return nil
		},
	}

	splitsCmd.Flags().SortFlags = false
	splitsCmd.Flags().StringVar(&project, "project", "", "Project Name")
	splitsCmd.Flags().StringVar(&path, "path", ".", "Project directory")
	splitsCmd.Flags().StringVar(&model, "model", "", "Model Name")
	splitsCmd.Flags().BoolVar(&pending, "pending", false, "Only fetch pending splits")
	splitsCmd.Flags().BoolVar(&errored, "errored", false, "Only fetch errored splits")
	splitsCmd.MarkFlagsMutuallyExclusive("pending", "errored")
	splitsCmd.Flags().BoolVar(&local, "local", false, "Target locally running Rill")
	splitsCmd.Flags().Uint32Var(&pageSize, "page-size", 50, "Number of splits to return per page")
	splitsCmd.Flags().StringVar(&pageToken, "page-token", "", "Pagination token")

	return splitsCmd
}
