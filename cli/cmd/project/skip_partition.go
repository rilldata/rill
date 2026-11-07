package project

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/spf13/cobra"
)

func SkipPartitionCmd(ch *cmdutil.Helper) *cobra.Command {
	var project, path, branch, model string
	var partitions []string
	var pending, errored, local bool

	skipCmd := &cobra.Command{
		Use:   "skip-partition [<project>] <model>",
		Args:  cobra.RangeArgs(1, 2),
		Short: "Skip partitions for a model",
		Long: "Mark partitions as skipped so they are excluded from execution and from the model's error state. " +
			"Skipped partitions remain skipped until they are explicitly triggered (e.g. via 'rill project refresh --partition').",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 1 {
				model = args[0]
			} else if len(args) == 2 {
				project = args[0]
				model = args[1]
			}

			if !local && project == "" {
				var err error
				project, err = ch.InferProjectName(cmd.Context(), path, "use --project to specify the name or --local to target a local Rill process")
				if err != nil {
					return err
				}
			}

			rt, instanceID, err := ch.OpenRuntimeClient(cmd.Context(), ch.Org, project, branch, local)
			if err != nil {
				return err
			}

			_, err = rt.SkipModelPartitions(cmd.Context(), &runtimev1.SkipModelPartitionsRequest{
				InstanceId: instanceID,
				Model:      model,
				Partitions: partitions,
				Pending:    pending,
				Errored:    errored,
			})
			if err != nil {
				return fmt.Errorf("failed to skip model partitions: %w", err)
			}

			ch.PrintfSuccess("Skipped partitions for model %q.\n", model)

			return nil
		},
	}

	skipCmd.Flags().SortFlags = false
	skipCmd.Flags().StringVar(&project, "project", "", "Project Name")
	skipCmd.Flags().StringVar(&path, "path", ".", "Project directory")
	skipCmd.Flags().StringVar(&branch, "branch", "", "Target deployment by Git branch (default: primary deployment)")
	skipCmd.Flags().StringVar(&model, "model", "", "Model Name")
	skipCmd.Flags().StringSliceVar(&partitions, "partition", nil, "Skip specific partitions by key")
	skipCmd.Flags().BoolVar(&pending, "pending", false, "Skip all pending partitions")
	skipCmd.Flags().BoolVar(&errored, "errored", false, "Skip all errored partitions")
	skipCmd.MarkFlagsOneRequired("partition", "pending", "errored")
	skipCmd.Flags().BoolVar(&local, "local", false, "Target locally running Rill")

	return skipCmd
}
