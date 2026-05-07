package billing

import (
	"fmt"
	"time"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func MockUsageCmd(ch *cmdutil.Helper) *cobra.Command {
	var eventName, eventTimeStr, projectName string
	var value float64

	cmd := &cobra.Command{
		Use:     "mock-usage <org-name>",
		Short:   "Report a mock usage event for an organization",
		Example: "rill sudo billing mock-usage my-org --event slot_seconds_spend --value 3600",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			org := args[0]

			if eventName == "" {
				return fmt.Errorf("please set --event")
			}
			if value <= 0 {
				return fmt.Errorf("please set --value to a positive value")
			}

			var eventTime *timestamppb.Timestamp
			if eventTimeStr != "" {
				t, err := time.Parse(time.RFC3339, eventTimeStr)
				if err != nil {
					return fmt.Errorf("invalid --event-time (expected RFC3339): %w", err)
				}
				eventTime = timestamppb.New(t)
			}

			client, err := ch.Client()
			if err != nil {
				return err
			}

			res, err := client.SudoReportUsage(ctx, &adminv1.SudoReportUsageRequest{
				Org:         org,
				EventName:   eventName,
				Value:       value,
				EndTime:     eventTime,
				ProjectName: projectName,
			})
			if err != nil {
				return err
			}

			ch.PrintfSuccess("Reported usage event %q for organization %q (value=%g, period=%s..%s)\n",
				res.EventName,
				org,
				res.Value,
				res.StartTime.AsTime().Format(time.RFC3339),
				res.EndTime.AsTime().Format(time.RFC3339),
			)
			return nil
		},
	}

	cmd.Flags().SortFlags = false
	cmd.Flags().StringVar(&eventName, "event", "", "Event/metric name (for example, slot_seconds_spend or duckdb_estimated_size_bytes)")
	cmd.Flags().Float64Var(&value, "value", 0, "Numeric value to report")
	cmd.Flags().StringVar(&eventTimeStr, "event-time", "", "Event time of the reporting window in RFC3339 (defaults to current server time)")
	cmd.Flags().StringVar(&projectName, "project-name", "", "Optional project name to attribute the mock event to (defaults to a placeholder)")
	return cmd
}
