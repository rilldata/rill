package devtool

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/rilldata/rill/admin/billing"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/cli/cmd/admin"
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func SubscriptionCmd(ch *cmdutil.Helper) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "subscription",
		Short: "Utilities for tweaking subscription",
	}

	cmd.AddCommand(AdvanceSubscriptionTimeCmd(ch))

	return cmd
}

func AdvanceSubscriptionTimeCmd(ch *cmdutil.Helper) *cobra.Command {
	var toDate string

	cmd := &cobra.Command{
		Use:   "advance-time",
		Short: "Tweak subscription by artificially advancing time",
		RunE: func(cmd *cobra.Command, args []string) error {
			adminClient, err := ch.Client()
			if err != nil {
				return err
			}

			ctx := cmd.Context()
			endDate := time.Now().AddDate(0, 0, -1)
			if toDate != "" {
				endDate, err = time.Parse(time.RFC3339, toDate)
				if err != nil {
					return err
				}
			}

			ch.Println("Using org", ch.Org)

			// Init config, used to start some clients
			// Load .env (note: fails silently if .env has errors)
			_ = godotenv.Load()
			var conf admin.Config
			err = envconfig.Process("rill_admin", &conf)
			if err != nil {
				return err
			}
			if conf.OrbAPIKey == "" {
				return errors.New("missing orb api key. make sure to run from rill git root to get keys from .env")
			}

			orgResp, err := adminClient.GetOrganization(ctx, &adminv1.GetOrganizationRequest{
				Org: ch.Org,
			})
			if err != nil {
				return err
			}

			subResp, err := adminClient.GetBillingSubscription(ctx, &adminv1.GetBillingSubscriptionRequest{
				Org: ch.Org,
			})
			if err != nil {
				return err
			}
			if subResp.Subscription == nil {
				return errors.New("org has no subscription")
			}

			resp, err := adminClient.ListOrganizationBillingIssues(ctx, &adminv1.ListOrganizationBillingIssuesRequest{
				Org: ch.Org,
			})
			if err != nil {
				return err
			}

			db, err := database.Open(conf.DatabaseDriver, conf.DatabaseURL, conf.DatabaseEncryptionKeyring)
			if err != nil {
				return err
			}

			var rerunJobs []string

			for _, issue := range resp.Issues {
				switch issue.Type {
				case adminv1.BillingIssueType_BILLING_ISSUE_TYPE_ON_TRIAL:
					_, err = db.UpsertBillingIssue(ctx, &database.UpsertBillingIssueOptions{
						OrgID: orgResp.Organization.Id,
						Type:  database.BillingIssueTypeOnTrial,
						Metadata: &database.BillingIssueMetadataOnTrial{
							SubID:              subResp.Subscription.Id,
							PlanID:             subResp.Subscription.Plan.Id,
							EndDate:            endDate,
							GracePeriodEndDate: issue.Metadata.GetOnTrial().GracePeriodEndDate.AsTime(),
						},
						EventTime: endDate.AddDate(0, 0, 1),
					})
					if err != nil {
						return err
					}
					rerunJobs = append(rerunJobs, "trial_ending_soon", "trial_end_check")
					ch.Println("Advanced trial issue to:", endDate.UTC().String())

				case adminv1.BillingIssueType_BILLING_ISSUE_TYPE_TRIAL_ENDED:
					_, err = db.UpsertBillingIssue(ctx, &database.UpsertBillingIssueOptions{
						OrgID: orgResp.Organization.Id,
						Type:  database.BillingIssueTypeTrialEnded,
						Metadata: &database.BillingIssueMetadataTrialEnded{
							SubID:              subResp.Subscription.Id,
							PlanID:             subResp.Subscription.Plan.Id,
							EndDate:            issue.Metadata.GetTrialEnded().EndDate.AsTime(),
							GracePeriodEndDate: endDate,
						},
						EventTime: endDate.AddDate(0, 0, 1),
					})
					if err != nil {
						return err
					}
					rerunJobs = append(rerunJobs, "trial_grace_period_check")
					ch.Println("Advanced trial ended issue to:", endDate.UTC().String())

				case adminv1.BillingIssueType_BILLING_ISSUE_TYPE_SUBSCRIPTION_CANCELLED:
					cfg := zap.NewProductionConfig()
					logger, err := cfg.Build()
					if err != nil {
						return err
					}
					biller := billing.NewOrb(logger, conf.OrbAPIKey, conf.OrbWebhookSecret, strings.ToLower(conf.OrbIntegratedTaxProvider))

					_, err = biller.UnscheduleCancellation(ctx, subResp.Subscription.Id)
					if err != nil {
						return err
					}

					_, err = biller.CancelSubscriptionsForCustomer(ctx, orgResp.Organization.BillingCustomerId, billing.SubscriptionCancellationOptionImmediate)
					if err != nil {
						return err
					}

					_, err = db.UpsertBillingIssue(ctx, &database.UpsertBillingIssueOptions{
						OrgID: orgResp.Organization.Id,
						Type:  database.BillingIssueTypeSubscriptionCancelled,
						Metadata: &database.BillingIssueMetadataSubscriptionCancelled{
							EndDate: endDate,
						},
						EventTime: endDate.AddDate(0, 0, 1),
					})
					if err != nil {
						return err
					}
					rerunJobs = append(rerunJobs, "subscription_cancellation_check")
					ch.Println("Advanced sub cancelled issue to:", endDate.UTC().String())

				case adminv1.BillingIssueType_BILLING_ISSUE_TYPE_PAYMENT_FAILED:
					invoices := make(map[string]*database.BillingIssueMetadataPaymentFailedMeta)
					m := issue.Metadata.GetPaymentFailed()
					for _, invoice := range m.Invoices {
						invoices[invoice.InvoiceId] = &database.BillingIssueMetadataPaymentFailedMeta{
							ID:                 invoice.InvoiceId,
							Number:             invoice.InvoiceNumber,
							URL:                invoice.InvoiceUrl,
							Amount:             invoice.AmountDue,
							Currency:           invoice.Currency,
							DueDate:            invoice.DueDate.AsTime(),
							FailedOn:           invoice.FailedOn.AsTime(),
							GracePeriodEndDate: endDate,
						}
					}
					_, err = db.UpsertBillingIssue(ctx, &database.UpsertBillingIssueOptions{
						OrgID: orgResp.Organization.Id,
						Type:  database.BillingIssueTypePaymentFailed,
						Metadata: &database.BillingIssueMetadataPaymentFailed{
							Invoices: invoices,
						},
						EventTime: endDate.AddDate(0, 0, 1),
					})
					if err != nil {
						return err
					}
					rerunJobs = append(rerunJobs, "payment_failed_grace_period_check")
					ch.Println("Advanced payment failed issue to:", endDate.UTC().String())
				}
			}

			if len(rerunJobs) > 0 {
				return rerunSelectRiverJobs(ctx, conf.RiverDatabaseURL, rerunJobs)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&ch.Org, "org", ch.Org, "Organization")
	cmd.Flags().StringVar(&toDate, "time", "", "Time to update the trial sub to")

	return cmd
}

// rerunSelectRiverJobs reruns jobs that are needed after advancing time
func rerunSelectRiverJobs(ctx context.Context, dsn string, kinds []string) error {
	riverClient, err := newRiverClient(ctx, dsn)
	if err != nil {
		return err
	}

	res, err := riverClient.JobList(ctx, river.NewJobListParams().Kinds(kinds...))
	if err != nil {
		return err
	}

	// there will be multiple instances of a job, one for every run.
	// so run just one of them by keeping track in kindsRemaining
	kindsRemaining := map[string]bool{}
	for _, kind := range kinds {
		kindsRemaining[kind] = true
	}

	for _, job := range res.Jobs {
		if !kindsRemaining[job.Kind] {
			continue
		}
		_, err = riverClient.JobRetry(ctx, job.ID)
		if err != nil {
			return err
		}
		delete(kindsRemaining, job.Kind)
	}

	return nil
}

// newRiverClient creates a barebones client to connect to river to rerun jobs
func newRiverClient(ctx context.Context, dsn string) (*river.Client[pgx.Tx], error) {
	dbPool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}

	tx, err := dbPool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	riverClient, err := river.NewClient(riverpgxv5.New(dbPool), &river.Config{
		JobTimeout:  time.Hour,
		MaxAttempts: 3, // default retry policy with backoff of attempt^4 seconds
	})
	if err != nil {
		return nil, err
	}

	return riverClient, nil
}
