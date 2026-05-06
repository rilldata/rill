package devtool

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/rilldata/rill/admin/billing"
	"github.com/rilldata/rill/cli/cmd/admin"
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// ReportUsageCmd posts a single usage event directly to Orb for testing purposes.
func ReportUsageCmd(ch *cmdutil.Helper) *cobra.Command {
	var orgID, eventName, endTimeStr string
	var amount float64

	cmd := &cobra.Command{
		Use:   "report-usage",
		Short: "Report a single usage event to Orb for testing credit-trial flows",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			if orgID == "" {
				return errors.New("--org-id is required")
			}
			if eventName == "" {
				return errors.New("--event is required")
			}
			if amount <= 0 {
				return errors.New("--amount must be > 0")
			}

			endTime := time.Now().UTC()
			if endTimeStr != "" {
				t, err := time.Parse(time.RFC3339, endTimeStr)
				if err != nil {
					return fmt.Errorf("invalid --end-time (expected RFC3339): %w", err)
				}
				endTime = t.UTC()
			}

			// Load .env (silently ignores missing files) and read the admin Orb config.
			_ = godotenv.Load()
			var conf admin.Config
			if err := envconfig.Process("rill_admin", &conf); err != nil {
				return err
			}
			if conf.OrbAPIKey == "" {
				return errors.New("missing orb api key; make sure RILL_ADMIN_ORB_API_KEY is set (run from the repo root so .env is picked up)")
			}

			cfg := zap.NewProductionConfig()
			logger, err := cfg.Build()
			if err != nil {
				return err
			}
			biller := billing.NewOrb(logger, conf.OrbAPIKey, conf.OrbWebhookSecret, strings.ToLower(conf.OrbIntegratedTaxProvider))

			usage := &billing.Usage{
				CustomerID:     orgID,
				MetricName:     eventName,
				Value:          amount,
				ReportingGrain: billing.UsageReportingGranularityHour,
				StartTime:      endTime.Add(-time.Hour),
				EndTime:        endTime,
				Metadata: map[string]interface{}{
					"org_id":       orgID,
					"project_id":   "devtool-project-id",
					"project_name": "devtool-project",
				},
			}

			if err := biller.ReportUsage(ctx, []*billing.Usage{usage}); err != nil {
				return fmt.Errorf("failed to report usage: %w", err)
			}

			ch.PrintfSuccess("Reported usage event %q for org %q (amount=%g, period=%s..%s)\n",
				eventName, orgID, amount,
				usage.StartTime.Format(time.RFC3339), usage.EndTime.Format(time.RFC3339))
			return nil
		},
	}

	cmd.Flags().SortFlags = false
	cmd.Flags().StringVar(&orgID, "org-id", "", "Org ID (used as Orb external_customer_id)")
	cmd.Flags().StringVar(&eventName, "event", "", "Event/metric name (e.g. slot_seconds_spend, duckdb_estimated_size_bytes)")
	cmd.Flags().Float64Var(&amount, "amount", 0, "Numeric amount to report")
	cmd.Flags().StringVar(&endTimeStr, "end-time", "", "End time of the reporting window in RFC3339 (defaults to current time)")
	return cmd
}
