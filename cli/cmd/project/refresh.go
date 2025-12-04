package project

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func RefreshCmd(ch *cmdutil.Helper) *cobra.Command {
	var project, path string
	var local bool
	var models, modelPartitions, sources, metricViews, alerts, reports, connectors []string
	var all, full, erroredPartitions, parser bool

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
				if !ch.Interactive {
					return fmt.Errorf("project not specified and could not be inferred from context")
				}
				var err error
				project, err = ch.InferProjectName(cmd.Context(), ch.Org, path)
				if err != nil {
					return fmt.Errorf("unable to infer project name (use `--project` to explicitly specify the name): %w", err)
				}
			}

			// Connect to the runtime
			rt, instanceID, err := ch.OpenRuntimeClient(cmd.Context(), ch.Org, project, local)
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

			// Build model triggers
			if len(modelPartitions) > 0 || erroredPartitions {
				// If partitions are specified, ensure exactly one model is specified.
				if len(models) != 1 {
					return fmt.Errorf("must specify exactly one --model when using --partition or --errored-partitions")
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
	refreshCmd.Flags().BoolVar(&local, "local", false, "Target locally running Rill")
	refreshCmd.Flags().BoolVar(&all, "all", false, "Refresh all resources except alerts and reports (default)")
	refreshCmd.Flags().BoolVar(&full, "full", false, "Fully reload the targeted models (use with --all or --model)")
	refreshCmd.Flags().StringSliceVar(&models, "model", nil, "Refresh a model")
	refreshCmd.Flags().StringSliceVar(&modelPartitions, "partition", nil, "Refresh a model partition (must set --model)")
	refreshCmd.Flags().BoolVar(&erroredPartitions, "errored-partitions", false, "Refresh all model partitions with errors (must set --model)")
	refreshCmd.Flags().StringSliceVar(&sources, "source", nil, "Refresh a source")
	refreshCmd.Flags().StringSliceVar(&metricViews, "metrics-view", nil, "Refresh a metrics view")
	refreshCmd.Flags().StringSliceVar(&alerts, "alert", nil, "Refresh an alert")
	refreshCmd.Flags().StringSliceVar(&reports, "report", nil, "Refresh a report")
	refreshCmd.Flags().StringSliceVar(&connectors, "connector", nil, "Re-validate a connector")
	refreshCmd.Flags().BoolVar(&parser, "parser", false, "Refresh the parser (forces a pull from Github)")

	return refreshCmd
}
