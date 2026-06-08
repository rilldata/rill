package project

import (
	"context"
	"fmt"
	"strconv"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"google.golang.org/protobuf/types/known/structpb"
)

func RefreshCmd(ch *cmdutil.Helper) *cobra.Command {
	var project, path, branch string
	var local bool
	var models, modelPartitions, sources, metricViews, alerts, reports, connectors []string
	var all, full, erroredPartitions, parser, yes bool
	var partitionKey, partitionStart, partitionEnd string

	refreshCmd := &cobra.Command{
		Use:               "refresh [<project-name>]",
		Args:              cobra.MaximumNArgs(1),
		Short:             "Refresh one or more resources",
		PersistentPreRunE: cmdutil.CheckChain(cmdutil.CheckAuth(ch), cmdutil.CheckOrganization(ch)),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Determine project name
			if len(args) > 0 {
				project = args[0]
			}
			if !local && project == "" {
				var err error
				project, err = ch.InferProjectName(cmd.Context(), path, "use --project to specify the name or --local to target a local Rill process")
				if err != nil {
					return err
				}
			}

			// Connect to the runtime
			rt, instanceID, err := ch.OpenRuntimeClient(cmd.Context(), ch.Org, project, branch, local)
			if err != nil {
				return err
			}

			// If only meta flags are set, default to an incremental refresh of all sources and models.
			var numMetaFlags int
			cmd.Flags().Visit(func(f *pflag.Flag) {
				// Count all inherited flags as meta flags.
				if cmd.InheritedFlags().Lookup(f.Name) != nil {
					numMetaFlags++
				}
			})
			if cmd.Flags().Changed("project") {
				numMetaFlags++
			}
			if cmd.Flags().Changed("path") {
				numMetaFlags++
			}
			if cmd.Flags().Changed("local") {
				numMetaFlags++
			}
			if numMetaFlags == cmd.Flags().NFlag() {
				all = true
			}

			// Build non-model resources
			var resources []*runtimev1.ResourceName
			for _, v := range metricViews {
				resources = append(resources, &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: v})
			}
			for _, a := range alerts {
				resources = append(resources, &runtimev1.ResourceName{Kind: runtime.ResourceKindAlert, Name: a})
			}
			for _, r := range reports {
				resources = append(resources, &runtimev1.ResourceName{Kind: runtime.ResourceKindReport, Name: r})
			}
			for _, c := range connectors {
				resources = append(resources, &runtimev1.ResourceName{Kind: runtime.ResourceKindConnector, Name: c})
			}

			// Merge sources into models since sources have been deprecated and are no longer created on the backend.
			models = append(models, sources...)

			// Validate partition-range flags. All three must be set together, and the range mode
			// is mutually exclusive with --partition and --errored-partitions.
			rangeMode := partitionKey != "" || partitionStart != "" || partitionEnd != ""
			if rangeMode {
				if partitionKey == "" || partitionStart == "" || partitionEnd == "" {
					return fmt.Errorf("--partition-key, --partition-start, and --partition-end must all be set together")
				}
				if len(modelPartitions) > 0 || erroredPartitions {
					return fmt.Errorf("--partition-key cannot be combined with --partition or --errored-partitions")
				}
				if partitionStart > partitionEnd {
					return fmt.Errorf("--partition-start (%q) must be <= --partition-end (%q)", partitionStart, partitionEnd)
				}
			}

			// Build model triggers
			if len(modelPartitions) > 0 || erroredPartitions || rangeMode {
				// If partitions are specified, ensure exactly one model is specified.
				if len(models) != 1 {
					return fmt.Errorf("must specify exactly one --model when using --partition, --errored-partitions, or --partition-key")
				}

				// Since it's a common error, do an early check to ensure the model is incremental.
				// (This error will also be logged by the reconciler, but surfacing it here is more user-friendly.)
				mn := models[0]
				resp, err := rt.GetResource(cmd.Context(), &runtimev1.GetResourceRequest{
					InstanceId: instanceID,
					Name:       &runtimev1.ResourceName{Kind: runtime.ResourceKindModel, Name: mn},
				})
				if err != nil {
					return fmt.Errorf("failed to get model %q: %w", mn, err)
				}
				m := resp.Resource.GetModel()
				if !m.Spec.Incremental {
					return fmt.Errorf("can't refresh partitions on model %q because it is not incremental", mn)
				}
			}

			// Resolve partition range to concrete partition keys.
			if rangeMode {
				matched, err := resolvePartitionRange(cmd.Context(), rt, instanceID, models[0], partitionKey, partitionStart, partitionEnd)
				if err != nil {
					return err
				}
				if len(matched) == 0 {
					ch.Printf("No partitions match %s in [%s, %s] on model %q.\n", partitionKey, partitionStart, partitionEnd, models[0])
					return nil
				}

				ch.PrintModelPartitions(matched)

				if !yes && ch.Interactive {
					if err := cmdutil.ConfirmPrompt(fmt.Sprintf("Refresh %d partition(s)?", len(matched)), true); err != nil {
						return err
					}
				}

				modelPartitions = make([]string, 0, len(matched))
				for _, p := range matched {
					modelPartitions = append(modelPartitions, p.Key)
				}
			}
			var modelTriggers []*runtimev1.RefreshModelTrigger
			for _, m := range models {
				modelTriggers = append(modelTriggers, &runtimev1.RefreshModelTrigger{
					Model:                m,
					Full:                 full,
					AllErroredPartitions: erroredPartitions,
					Partitions:           modelPartitions,
				})
			}

			// Return an error for ineffective use of --full
			if full && !all && len(models) == 0 {
				return fmt.Errorf("the --full flag can only be used with --all or --model")
			}

			// Send request
			_, err = rt.CreateTrigger(cmd.Context(), &runtimev1.CreateTriggerRequest{
				InstanceId: instanceID,
				Resources:  resources,
				Models:     modelTriggers,
				Parser:     parser,
				All:        all && !full,
				AllFull:    all && full,
			})
			if err != nil {
				return fmt.Errorf("failed to create trigger: %w", err)
			}

			// Print status
			if local {
				ch.Printf("Refresh initiated. Check the project logs for status updates.\n")
			} else {
				ch.Printf("Refresh initiated. To check the status, run `rill project status` or `rill project logs`.\n")
			}

			return nil
		},
	}

	refreshCmd.Flags().SortFlags = false
	refreshCmd.Flags().StringVar(&project, "project", "", "Project name")
	refreshCmd.Flags().StringVar(&path, "path", ".", "Project directory")
	refreshCmd.Flags().StringVar(&branch, "branch", "", "Target deployment by Git branch (default: primary deployment)")
	refreshCmd.Flags().BoolVar(&local, "local", false, "Target locally running Rill")
	refreshCmd.Flags().BoolVar(&all, "all", false, "Refresh all resources except alerts and reports (default)")
	refreshCmd.Flags().BoolVar(&full, "full", false, "Fully reload the targeted models (use with --all or --model)")
	refreshCmd.Flags().StringSliceVar(&models, "model", nil, "Refresh a model")
	refreshCmd.Flags().StringSliceVar(&modelPartitions, "partition", nil, "Refresh a model partition (must set --model)")
	refreshCmd.Flags().BoolVar(&erroredPartitions, "errored-partitions", false, "Refresh all model partitions with errors (must set --model)")
	refreshCmd.Flags().StringVar(&partitionKey, "partition-key", "", "Name of the field in the partition data to range-filter on (must set --model)")
	refreshCmd.Flags().StringVar(&partitionStart, "partition-start", "", "Inclusive lower bound for --partition-key (lexicographic string compare)")
	refreshCmd.Flags().StringVar(&partitionEnd, "partition-end", "", "Inclusive upper bound for --partition-key (lexicographic string compare)")
	refreshCmd.Flags().BoolVar(&yes, "yes", false, "Skip the partition-range refresh confirmation prompt")
	refreshCmd.Flags().StringSliceVar(&sources, "source", nil, "Refresh a source")
	refreshCmd.Flags().StringSliceVar(&metricViews, "metrics-view", nil, "Refresh a metrics view")
	refreshCmd.Flags().StringSliceVar(&alerts, "alert", nil, "Refresh an alert")
	refreshCmd.Flags().StringSliceVar(&reports, "report", nil, "Refresh a report")
	refreshCmd.Flags().StringSliceVar(&connectors, "connector", nil, "Re-validate a connector")
	refreshCmd.Flags().BoolVar(&parser, "parser", false, "Refresh the parser (forces a pull from Github)")

	return refreshCmd
}

// resolvePartitionRange lists all partitions of the given model and returns those whose
// data field `key` falls within [start, end] (inclusive, lexicographic string compare).
func resolvePartitionRange(ctx context.Context, rt runtimev1.RuntimeServiceClient, instanceID, model, key, start, end string) ([]*runtimev1.ModelPartition, error) {
	var matched []*runtimev1.ModelPartition
	var pageToken string
	var sawAnyPartition bool
	for {
		res, err := rt.GetModelPartitions(ctx, &runtimev1.GetModelPartitionsRequest{
			InstanceId: instanceID,
			Model:      model,
			PageSize:   100,
			PageToken:  pageToken,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list partitions for model %q: %w", model, err)
		}

		for _, p := range res.Partitions {
			sawAnyPartition = true
			if p.Data == nil {
				continue
			}
			fields := p.Data.GetFields()
			v, ok := fields[key]
			if !ok {
				available := make([]string, 0, len(fields))
				for f := range fields {
					available = append(available, f)
				}
				return nil, fmt.Errorf("partition field %q not found on partition %q; available fields: %v", key, p.Key, available)
			}

			var s string
			switch k := v.Kind.(type) {
			case *structpb.Value_StringValue:
				s = k.StringValue
			case *structpb.Value_NumberValue:
				s = strconv.FormatFloat(k.NumberValue, 'f', -1, 64)
			case *structpb.Value_BoolValue:
				s = strconv.FormatBool(k.BoolValue)
			default:
				return nil, fmt.Errorf("partition %q: unsupported type %T for field %q", p.Key, v.Kind, key)
			}
			if s >= start && s <= end {
				matched = append(matched, p)
			}
		}

		if res.NextPageToken == "" {
			break
		}
		pageToken = res.NextPageToken
	}

	if !sawAnyPartition {
		return nil, fmt.Errorf("model %q has no partitions to filter", model)
	}
	return matched, nil
}
