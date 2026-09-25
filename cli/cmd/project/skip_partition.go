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
	var pending, errored, local, force bool
	var partitionKey, partitionStart, partitionEnd string

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

			// Validate partition-range flags. All three must be set together. The range can be
			// narrowed by --pending or --errored (but not both), and cannot be combined with an
			// explicit --partition list.
			rangeMode := partitionKey != "" || partitionStart != "" || partitionEnd != ""
			if rangeMode {
				if partitionKey == "" || partitionStart == "" || partitionEnd == "" {
					return fmt.Errorf("--partition-key, --partition-start, and --partition-end must all be set together")
				}
				if len(partitions) > 0 {
					return fmt.Errorf("--partition-key cannot be combined with --partition")
				}
				if pending && errored {
					return fmt.Errorf("--pending and --errored cannot be combined")
				}
				if partitionStart > partitionEnd {
					return fmt.Errorf("--partition-start (%q) must be <= --partition-end (%q)", partitionStart, partitionEnd)
				}
			}

			rt, instanceID, err := ch.OpenRuntimeClient(cmd.Context(), ch.Org, project, branch, local)
			if err != nil {
				return err
			}

			// Resolve a partition range to concrete partition keys. When --pending or --errored is
			// also set, the range is narrowed to partitions in that state.
			if rangeMode {
				matched, err := resolvePartitionRange(cmd.Context(), rt, instanceID, model, partitionKey, partitionStart, partitionEnd, pending, errored, false)
				if err != nil {
					return err
				}
				if len(matched) == 0 {
					ch.Printf("No partitions match %s in [%s, %s] on model %q.\n", partitionKey, partitionStart, partitionEnd, model)
					return nil
				}

				ch.PrintModelPartitions(matched)

				if !force && ch.Interactive {
					if err := cmdutil.ConfirmPrompt(fmt.Sprintf("Skip %d partition(s)?", len(matched)), true); err != nil {
						return err
					}
				}

				partitions = make([]string, 0, len(matched))
				for _, p := range matched {
					partitions = append(partitions, p.Key)
				}
				// The resolved key list is authoritative; clear the state flags so only the matched
				// partitions are skipped rather than all pending/errored ones.
				pending = false
				errored = false
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
	skipCmd.Flags().StringVar(&partitionKey, "partition-key", "", "Name of the field in the partition data to range-filter on")
	skipCmd.Flags().StringVar(&partitionStart, "partition-start", "", "Inclusive lower bound for --partition-key (lexicographic string compare)")
	skipCmd.Flags().StringVar(&partitionEnd, "partition-end", "", "Inclusive upper bound for --partition-key (lexicographic string compare)")
	skipCmd.Flags().BoolVar(&force, "force", false, "Skip the partition-range confirmation prompt")
	skipCmd.MarkFlagsOneRequired("partition", "pending", "errored", "partition-key", "partition-start", "partition-end")
	skipCmd.Flags().BoolVar(&local, "local", false, "Target locally running Rill")

	return skipCmd
}
