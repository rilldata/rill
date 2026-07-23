package river

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/rilldata/rill/admin"
	"github.com/rilldata/rill/admin/billing"
	"github.com/rilldata/rill/admin/database"
	_ "github.com/rilldata/rill/admin/database/postgres"
	"github.com/rilldata/rill/admin/pkg/pgtestcontainer"
	"github.com/rilldata/rill/runtime/pkg/email"
	"github.com/riverqueue/river"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestBillingLifecycleExactTimeBoundaries(t *testing.T) {
	// Lifecycle checks use inclusive deadlines. One nanosecond before a deadline
	// must be a no-op, while the exact instant must durably advance state.
	boundary := time.Date(2026, time.July, 23, 12, 0, 0, 0, time.UTC)

	t.Run("trial ending soon", func(t *testing.T) {
		// The seven-day warning becomes eligible exactly at end-date minus seven days.
		adm, db, _, sender := newBillingLifecycleFixture(t)
		org := db.addOrganization("ending")
		issue := db.addIssue(org.ID, database.BillingIssueTypeOnTrial, false, &database.BillingIssueMetadataOnTrial{EndDate: boundary.AddDate(0, 0, 7)}, boundary)
		now := boundary.Add(-time.Nanosecond)
		worker := &TrialEndingSoonWorker{admin: adm, logger: zap.NewNop(), now: func() time.Time { return now }}

		require.NoError(t, worker.Work(t.Context(), nil))
		require.False(t, db.isProcessed(issue.ID))
		require.Equal(t, 0, sender.attemptCount())

		now = boundary
		require.NoError(t, worker.Work(t.Context(), nil))
		require.True(t, db.isProcessed(issue.ID))
		require.Equal(t, 1, sender.attemptCount())
	})

	t.Run("trial ended", func(t *testing.T) {
		// The trial-to-grace transaction runs at the exact subscription trial end.
		adm, db, biller, sender := newBillingLifecycleFixture(t)
		org := db.addOrganization("ended")
		db.addIssue(org.ID, database.BillingIssueTypeOnTrial, true, &database.BillingIssueMetadataOnTrial{SubID: "trial-sub", PlanID: "trial-plan", EndDate: boundary, GracePeriodEndDate: boundary.AddDate(0, 0, 7)}, boundary)
		biller.subscriptions[org.BillingCustomerID] = &billing.Subscription{ID: "trial-sub", Plan: &billing.Plan{ID: "trial-plan"}}
		now := boundary.Add(-time.Nanosecond)
		worker := &TrialEndCheckWorker{admin: adm, logger: zap.NewNop(), now: func() time.Time { return now }}

		require.NoError(t, worker.Work(t.Context(), nil))
		require.NotNil(t, db.issueForOrg(org.ID, database.BillingIssueTypeOnTrial))
		require.Nil(t, db.issueForOrg(org.ID, database.BillingIssueTypeTrialEnded))

		now = boundary
		require.NoError(t, worker.Work(t.Context(), nil))
		require.Nil(t, db.issueForOrg(org.ID, database.BillingIssueTypeOnTrial))
		require.NotNil(t, db.issueForOrg(org.ID, database.BillingIssueTypeTrialEnded))
		require.Equal(t, 1, sender.attemptCount())
	})

	t.Run("trial grace period", func(t *testing.T) {
		// Trial projects are hibernated at the exact grace-period end, not before it.
		adm, db, biller, sender := newBillingLifecycleFixture(t)
		org := db.addOrganization("grace")
		issue := db.addIssue(org.ID, database.BillingIssueTypeTrialEnded, false, &database.BillingIssueMetadataTrialEnded{SubID: "trial-sub", PlanID: "trial-plan", GracePeriodEndDate: boundary}, boundary)
		biller.subscriptions[org.BillingCustomerID] = &billing.Subscription{ID: "trial-sub", Plan: &billing.Plan{ID: "trial-plan"}}
		now := boundary.Add(-time.Nanosecond)
		worker := &TrialGracePeriodCheckWorker{admin: adm, logger: zap.NewNop(), now: func() time.Time { return now }}

		require.NoError(t, worker.Work(t.Context(), nil))
		require.False(t, db.isProcessed(issue.ID))
		require.Empty(t, biller.cancellations)

		now = boundary
		require.NoError(t, worker.Work(t.Context(), nil))
		require.True(t, db.isProcessed(issue.ID))
		require.Equal(t, []string{org.BillingCustomerID}, biller.cancellations)
		requireOrganizationQuotasZero(t, db.organization(org.ID))
		require.Equal(t, 1, sender.attemptCount())
	})

	t.Run("subscription cancellation", func(t *testing.T) {
		// Cancellation intentionally waits through the day after the recorded end date.
		adm, db, _, sender := newBillingLifecycleFixture(t)
		org := db.addOrganization("cancelled")
		issue := db.addIssue(org.ID, database.BillingIssueTypeSubscriptionCancelled, false, &database.BillingIssueMetadataSubscriptionCancelled{EndDate: boundary.AddDate(0, 0, -1)}, boundary)
		now := boundary.Add(-time.Nanosecond)
		worker := &SubscriptionCancellationCheckWorker{admin: adm, logger: zap.NewNop(), now: func() time.Time { return now }}

		require.NoError(t, worker.Work(t.Context(), nil))
		require.False(t, db.isProcessed(issue.ID))

		now = boundary
		require.NoError(t, worker.Work(t.Context(), nil))
		require.True(t, db.isProcessed(issue.ID))
		requireOrganizationQuotasZero(t, db.organization(org.ID))
		require.Equal(t, 1, sender.attemptCount())
	})

	t.Run("failed invoice grace period", func(t *testing.T) {
		// An unpaid invoice becomes hibernation-eligible exactly one day after grace.
		adm, db, biller, sender := newBillingLifecycleFixture(t)
		org := db.addOrganization("overdue")
		issue := db.addIssue(org.ID, database.BillingIssueTypePaymentFailed, false, &database.BillingIssueMetadataPaymentFailed{Invoices: map[string]*database.BillingIssueMetadataPaymentFailedMeta{
			"invoice": {ID: "invoice", GracePeriodEndDate: boundary.AddDate(0, 0, -1)},
		}}, boundary)
		biller.invoices["invoice"] = &billing.Invoice{ID: "invoice", Status: "open"}
		now := boundary.Add(-time.Nanosecond)
		worker := &PaymentFailedGracePeriodCheckWorker{admin: adm, logger: zap.NewNop(), now: func() time.Time { return now }}

		require.NoError(t, worker.Work(t.Context(), nil))
		require.False(t, db.isProcessed(issue.ID))

		now = boundary
		require.NoError(t, worker.Work(t.Context(), nil))
		require.True(t, db.isProcessed(issue.ID))
		require.Equal(t, 1, sender.attemptCount())
	})
}

func TestBillingLifecycleNotificationAttemptIsDurablyAtMostOnce(t *testing.T) {
	// The existing schema can provide at-most-once SMTP attempts by committing a
	// terminal marker first; these cases pin both the guarantee and its loss tradeoff.
	now := time.Date(2026, time.July, 23, 12, 0, 0, 0, time.UTC)

	t.Run("processed marker failure happens before email", func(t *testing.T) {
		// A marker outage leaves the issue retryable and sends nothing until marking succeeds.
		adm, db, _, sender := newBillingLifecycleFixture(t)
		org := db.addOrganization("marker")
		issue := db.addIssue(org.ID, database.BillingIssueTypeSubscriptionCancelled, false, &database.BillingIssueMetadataSubscriptionCancelled{EndDate: now.AddDate(0, 0, -1)}, now)
		db.processedFailures[issue.ID] = 1
		worker := &SubscriptionCancellationCheckWorker{admin: adm, logger: zap.NewNop(), now: func() time.Time { return now }}

		require.ErrorContains(t, worker.Work(t.Context(), nil), "processed")
		require.False(t, db.isProcessed(issue.ID))
		require.Equal(t, 0, sender.attemptCount())

		require.NoError(t, worker.Work(t.Context(), nil))
		require.NoError(t, worker.Work(t.Context(), nil))
		require.True(t, db.isProcessed(issue.ID))
		require.Equal(t, 1, sender.attemptCount())
	})

	t.Run("transaction commit failure happens before email", func(t *testing.T) {
		// A failed trial transition commit retains the old issue and cannot leak an email.
		adm, db, biller, sender := newBillingLifecycleFixture(t)
		org := db.addOrganization("commit")
		db.addIssue(org.ID, database.BillingIssueTypeOnTrial, true, &database.BillingIssueMetadataOnTrial{SubID: "trial-sub", PlanID: "trial-plan", EndDate: now, GracePeriodEndDate: now.AddDate(0, 0, 7)}, now)
		biller.subscriptions[org.BillingCustomerID] = &billing.Subscription{ID: "trial-sub", Plan: &billing.Plan{ID: "trial-plan"}}
		db.commitFailures = 1
		worker := &TrialEndCheckWorker{admin: adm, logger: zap.NewNop(), now: func() time.Time { return now }}

		require.ErrorContains(t, worker.Work(t.Context(), nil), "commit")
		require.NotNil(t, db.issueForOrg(org.ID, database.BillingIssueTypeOnTrial))
		require.Nil(t, db.issueForOrg(org.ID, database.BillingIssueTypeTrialEnded))
		require.Equal(t, 0, sender.attemptCount())

		require.NoError(t, worker.Work(t.Context(), nil))
		require.NoError(t, worker.Work(t.Context(), nil))
		require.Nil(t, db.issueForOrg(org.ID, database.BillingIssueTypeOnTrial))
		require.NotNil(t, db.issueForOrg(org.ID, database.BillingIssueTypeTrialEnded))
		require.Equal(t, 1, sender.attemptCount())
	})

	t.Run("email failure is not retried after durable completion", func(t *testing.T) {
		// SMTP failure after the marker demonstrates the explicit at-most-once/loss tradeoff.
		adm, db, biller, sender := newBillingLifecycleFixture(t)
		org := db.addOrganization("smtp")
		issue := db.addIssue(org.ID, database.BillingIssueTypeTrialEnded, false, &database.BillingIssueMetadataTrialEnded{SubID: "trial-sub", PlanID: "trial-plan", GracePeriodEndDate: now}, now)
		biller.subscriptions[org.BillingCustomerID] = &billing.Subscription{ID: "trial-sub", Plan: &billing.Plan{ID: "trial-plan"}}
		sender.failures = 1
		worker := &TrialGracePeriodCheckWorker{admin: adm, logger: zap.NewNop(), now: func() time.Time { return now }}

		require.NoError(t, worker.Work(t.Context(), nil))
		require.True(t, db.isProcessed(issue.ID))
		require.Equal(t, 1, sender.attemptCount())
		require.Equal(t, 0, sender.successCount())

		require.NoError(t, worker.Work(t.Context(), nil))
		require.Equal(t, 1, sender.attemptCount())
	})

	t.Run("failed-payment duplicate does not repeat a failed email", func(t *testing.T) {
		// The persisted invoice entry suppresses retry sends even when SMTP returned an error.
		adm, db, biller, sender := newBillingLifecycleFixture(t)
		org := db.addOrganization("failed-email")
		biller.invoices["invoice"] = &billing.Invoice{ID: "invoice", Status: "open"}
		sender.failures = 1
		worker := &PaymentFailedWorker{admin: adm, logger: zap.NewNop()}
		job := paymentFailedJob(org.BillingCustomerID, "invoice", now)

		require.ErrorContains(t, worker.Work(t.Context(), job), "email")
		require.NotNil(t, db.issueForOrg(org.ID, database.BillingIssueTypePaymentFailed))
		require.Equal(t, 1, sender.attemptCount())

		require.NoError(t, worker.Work(t.Context(), job))
		require.Equal(t, 1, sender.attemptCount())
	})

	t.Run("successful-payment duplicate does not repeat a failed email", func(t *testing.T) {
		// Deleting the invoice issue before SMTP makes the success notification one attempt.
		adm, db, biller, sender := newBillingLifecycleFixture(t)
		org := db.addOrganization("success-email")
		db.addIssue(org.ID, database.BillingIssueTypePaymentFailed, false, paymentFailureMetadata("invoice", now), now)
		biller.invoices["invoice"] = &billing.Invoice{ID: "invoice", Status: "paid"}
		sender.failures = 1
		worker := &PaymentSuccessWorker{admin: adm, logger: zap.NewNop(), now: func() time.Time { return now }}
		job := &river.Job[PaymentSuccessArgs]{Args: PaymentSuccessArgs{BillingCustomerID: org.BillingCustomerID, InvoiceID: "invoice"}}

		require.NoError(t, worker.Work(t.Context(), job))
		require.Nil(t, db.issueForOrg(org.ID, database.BillingIssueTypePaymentFailed))
		require.Equal(t, 1, sender.attemptCount())
		require.Equal(t, 0, sender.successCount())

		require.NoError(t, worker.Work(t.Context(), job))
		require.Equal(t, 1, sender.attemptCount())
	})
}

func TestInvoiceLifecycleDuplicateAndOutOfOrderDeliveryIsMonotonic(t *testing.T) {
	// Provider truth and per-invoice timestamps prevent duplicate or stale events
	// from regressing the durable failed-invoice set.
	adm, db, biller, sender := newBillingLifecycleFixture(t)
	org := db.addOrganization("invoice-order")
	failureWorker := &PaymentFailedWorker{admin: adm, logger: zap.NewNop()}
	successWorker := &PaymentSuccessWorker{admin: adm, logger: zap.NewNop(), now: func() time.Time {
		return time.Date(2026, time.July, 24, 0, 0, 0, 0, time.UTC)
	}}
	newer := time.Date(2026, time.July, 23, 10, 0, 0, 0, time.UTC)
	older := newer.Add(-time.Hour)
	biller.invoices["invoice"] = &billing.Invoice{ID: "invoice", Status: "open"}

	require.NoError(t, failureWorker.Work(t.Context(), paymentFailedJob(org.BillingCustomerID, "invoice", newer)))
	require.NoError(t, failureWorker.Work(t.Context(), paymentFailedJob(org.BillingCustomerID, "invoice", newer)))
	require.NoError(t, failureWorker.Work(t.Context(), paymentFailedJob(org.BillingCustomerID, "invoice", older)))
	issue := db.issueForOrg(org.ID, database.BillingIssueTypePaymentFailed)
	require.NotNil(t, issue)
	require.Equal(t, newer, issue.EventTime)
	require.Equal(t, newer, issue.Metadata.(*database.BillingIssueMetadataPaymentFailed).Invoices["invoice"].FailedOn)
	require.Equal(t, 1, sender.attemptCount())

	// A success event is stale while the provider still reports the invoice unpaid.
	successJob := &river.Job[PaymentSuccessArgs]{Args: PaymentSuccessArgs{BillingCustomerID: org.BillingCustomerID, InvoiceID: "invoice"}}
	require.NoError(t, successWorker.Work(t.Context(), successJob))
	require.NotNil(t, db.issueForOrg(org.ID, database.BillingIssueTypePaymentFailed))
	require.Equal(t, 1, sender.attemptCount())

	// Once paid, success clears exactly once; a delayed failure cannot resurrect it.
	biller.invoices["invoice"].Status = "paid"
	require.NoError(t, successWorker.Work(t.Context(), successJob))
	require.NoError(t, successWorker.Work(t.Context(), successJob))
	require.Nil(t, db.issueForOrg(org.ID, database.BillingIssueTypePaymentFailed))
	require.Equal(t, 2, sender.attemptCount())
	require.NoError(t, failureWorker.Work(t.Context(), paymentFailedJob(org.BillingCustomerID, "invoice", older)))
	require.Nil(t, db.issueForOrg(org.ID, database.BillingIssueTypePaymentFailed))
	require.Equal(t, 2, sender.attemptCount())

	// An older failure for another invoice is retained without moving aggregate time backward.
	biller.invoices["newer-invoice"] = &billing.Invoice{ID: "newer-invoice", Status: "open"}
	biller.invoices["older-invoice"] = &billing.Invoice{ID: "older-invoice", Status: "open"}
	require.NoError(t, failureWorker.Work(t.Context(), paymentFailedJob(org.BillingCustomerID, "newer-invoice", newer)))
	require.NoError(t, failureWorker.Work(t.Context(), paymentFailedJob(org.BillingCustomerID, "older-invoice", older)))
	issue = db.issueForOrg(org.ID, database.BillingIssueTypePaymentFailed)
	require.Equal(t, newer, issue.EventTime)
	require.Len(t, issue.Metadata.(*database.BillingIssueMetadataPaymentFailed).Invoices, 2)
}

func TestSubscriptionCancellationResumesPartialHibernationAndContinuesOrganizations(t *testing.T) {
	// A replacement subscription and a one-time second-page failure must not block
	// later organizations; a retry must converge all already-durable project state.
	now := time.Date(2026, time.July, 23, 12, 0, 0, 0, time.UTC)
	adm, db, biller, sender := newBillingLifecycleFixture(t)
	activeOrg := db.addOrganization("a-active")
	brokenOrg := db.addOrganization("b-broken")
	laterOrg := db.addOrganization("c-later")
	activeIssue := db.addIssue(activeOrg.ID, database.BillingIssueTypeSubscriptionCancelled, false, &database.BillingIssueMetadataSubscriptionCancelled{EndDate: now.AddDate(0, 0, -1)}, now)
	brokenIssue := db.addIssue(brokenOrg.ID, database.BillingIssueTypeSubscriptionCancelled, false, &database.BillingIssueMetadataSubscriptionCancelled{EndDate: now.AddDate(0, 0, -1)}, now)
	laterIssue := db.addIssue(laterOrg.ID, database.BillingIssueTypeSubscriptionCancelled, false, &database.BillingIssueMetadataSubscriptionCancelled{EndDate: now.AddDate(0, 0, -1)}, now)
	db.addProjects(activeOrg.ID, 1)
	db.addProjects(brokenOrg.ID, 11)
	db.addProjects(laterOrg.ID, 1)
	db.pageFailures[projectPageKey(brokenOrg.ID, "project-09")] = 1
	biller.subscriptions[activeOrg.BillingCustomerID] = &billing.Subscription{ID: "replacement", Plan: &billing.Plan{ID: "paid"}}
	worker := &SubscriptionCancellationCheckWorker{admin: adm, logger: zap.NewNop(), now: func() time.Time { return now }}

	require.ErrorContains(t, worker.Work(t.Context(), nil), "second project page unavailable")
	require.True(t, db.isProcessed(activeIssue.ID))
	require.False(t, db.projectHibernated(activeOrg.ID, "project-00"))
	require.Equal(t, 7, db.organization(activeOrg.ID).QuotaProjects)
	require.False(t, db.isProcessed(brokenIssue.ID))
	for i := range 10 {
		require.True(t, db.projectHibernated(brokenOrg.ID, fmt.Sprintf("project-%02d", i)))
	}
	require.False(t, db.projectHibernated(brokenOrg.ID, "project-10"))
	require.True(t, db.isProcessed(laterIssue.ID))
	require.True(t, db.projectHibernated(laterOrg.ID, "project-00"))
	require.Equal(t, 1, sender.attemptCount())

	require.NoError(t, worker.Work(t.Context(), nil))
	require.True(t, db.isProcessed(brokenIssue.ID))
	for i := range 11 {
		require.True(t, db.projectHibernated(brokenOrg.ID, fmt.Sprintf("project-%02d", i)))
	}
	require.Equal(t, 2, db.hibernateCalls[projectKey(brokenOrg.ID, "project-00")])
	require.Equal(t, 2, sender.attemptCount())

	require.NoError(t, worker.Work(t.Context(), nil))
	require.Equal(t, 2, sender.attemptCount())
}

func TestTrialEndReplacementSubscriptionDoesNotBlockLaterOrganization(t *testing.T) {
	// A stale trial issue protected by a replacement subscription is retired, and
	// the next organization still transitions atomically into its grace period.
	now := time.Date(2026, time.July, 23, 12, 0, 0, 0, time.UTC)
	adm, db, biller, sender := newBillingLifecycleFixture(t)
	staleOrg := db.addOrganization("a-stale-trial")
	validOrg := db.addOrganization("b-valid-trial")
	db.addIssue(staleOrg.ID, database.BillingIssueTypeOnTrial, true, &database.BillingIssueMetadataOnTrial{SubID: "old", PlanID: "trial", EndDate: now, GracePeriodEndDate: now.AddDate(0, 0, 7)}, now)
	db.addIssue(validOrg.ID, database.BillingIssueTypeOnTrial, true, &database.BillingIssueMetadataOnTrial{SubID: "expected", PlanID: "trial", EndDate: now, GracePeriodEndDate: now.AddDate(0, 0, 7)}, now)
	biller.subscriptions[staleOrg.BillingCustomerID] = &billing.Subscription{ID: "replacement", Plan: &billing.Plan{ID: "paid"}}
	biller.subscriptions[validOrg.BillingCustomerID] = &billing.Subscription{ID: "expected", Plan: &billing.Plan{ID: "trial"}}
	worker := &TrialEndCheckWorker{admin: adm, logger: zap.NewNop(), now: func() time.Time { return now }}

	require.NoError(t, worker.Work(t.Context(), nil))
	require.Nil(t, db.issueForOrg(staleOrg.ID, database.BillingIssueTypeOnTrial))
	require.Nil(t, db.issueForOrg(staleOrg.ID, database.BillingIssueTypeTrialEnded))
	require.Nil(t, db.issueForOrg(validOrg.ID, database.BillingIssueTypeOnTrial))
	require.NotNil(t, db.issueForOrg(validOrg.ID, database.BillingIssueTypeTrialEnded))
	require.Equal(t, 1, sender.attemptCount())
}

func TestBillingLifecyclePostgresClaimsAndTransitions(t *testing.T) {
	// PostgreSQL is used only for the atomic claim and unique transactional
	// transition that an in-memory fixture cannot faithfully prove.
	if testing.Short() {
		t.Skip("requires a PostgreSQL test container")
	}
	pg := pgtestcontainer.New(t)
	t.Cleanup(func() { pg.Terminate(t) })
	db, err := database.Open("postgres", pg.DatabaseURL, "")
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })
	require.NoError(t, db.Migrate(t.Context()))
	urls, err := admin.NewURLs("https://admin.example.com", "https://app.example.com")
	require.NoError(t, err)
	now := time.Date(2026, time.July, 23, 12, 0, 0, 0, time.UTC)

	t.Run("processed marker is a single atomic claim", func(t *testing.T) {
		// Two workers holding the same stale issue may claim it once and emit one email.
		org := insertPostgresLifecycleOrg(t, db, "atomic-claim")
		metadata := &database.BillingIssueMetadataOnTrial{EndDate: now.AddDate(0, 0, 7)}
		issue, err := db.UpsertBillingIssue(t.Context(), &database.UpsertBillingIssueOptions{OrgID: org.ID, Type: database.BillingIssueTypeOnTrial, Metadata: metadata, EventTime: now})
		require.NoError(t, err)
		sender := &billingLifecycleEmailSender{}
		adm := &admin.Service{DB: db, Email: email.New(sender), URLs: urls, Logger: zap.NewNop()}
		worker := &TrialEndingSoonWorker{admin: adm, logger: zap.NewNop()}

		errs := runConcurrently(2, func() error {
			return worker.processIssue(t.Context(), issue, metadata, now)
		})
		require.Equal(t, 1, countNilErrors(errs))
		require.Equal(t, 1, sender.attemptCount())
	})

	t.Run("trial transition remains unique under duplicate work", func(t *testing.T) {
		// Competing transactions may create one grace issue and only the winner sends.
		org := insertPostgresLifecycleOrg(t, db, "atomic-transition")
		metadata := &database.BillingIssueMetadataOnTrial{SubID: "trial-sub", PlanID: "trial-plan", EndDate: now, GracePeriodEndDate: now.AddDate(0, 0, 7)}
		issue, err := db.UpsertBillingIssue(t.Context(), &database.UpsertBillingIssueOptions{OrgID: org.ID, Type: database.BillingIssueTypeOnTrial, Metadata: metadata, EventTime: now})
		require.NoError(t, err)
		require.NoError(t, db.UpdateBillingIssueOverdueAsProcessed(t.Context(), issue.ID))
		biller := &billingLifecycleBiller{subscriptions: map[string]*billing.Subscription{
			org.BillingCustomerID: {ID: "trial-sub", Plan: &billing.Plan{ID: "trial-plan"}},
		}, invoices: make(map[string]*billing.Invoice)}
		sender := &billingLifecycleEmailSender{}
		adm := &admin.Service{DB: db, Biller: biller, Email: email.New(sender), URLs: urls, Logger: zap.NewNop()}
		worker := &TrialEndCheckWorker{admin: adm, logger: zap.NewNop()}

		errs := runConcurrently(2, func() error {
			return worker.processIssue(t.Context(), issue, metadata)
		})
		require.Equal(t, 1, countNilErrors(errs))
		require.Equal(t, 1, sender.attemptCount())
		_, err = db.FindBillingIssueByTypeForOrg(t.Context(), org.ID, database.BillingIssueTypeOnTrial)
		require.ErrorIs(t, err, database.ErrNotFound)
		ended, err := db.FindBillingIssueByTypeForOrg(t.Context(), org.ID, database.BillingIssueTypeTrialEnded)
		require.NoError(t, err)
		require.True(t, ended.EventTime.Equal(now))
	})
}

type billingLifecycleIssue struct {
	issue     *database.BillingIssue
	processed bool
}

type billingLifecycleDB struct {
	database.DB
	organizations     map[string]*database.Organization
	issues            map[string]*billingLifecycleIssue
	projects          map[string][]*database.Project
	processedFailures map[string]int
	pageFailures      map[string]int
	hibernateCalls    map[string]int
	commitFailures    int
	nextIssueID       int
}

type billingLifecycleTxContextKey struct{}

type billingLifecycleTxState struct {
	db     *billingLifecycleDB
	issues map[string]*billingLifecycleIssue
	closed bool
}

func (d *billingLifecycleDB) NewTx(ctx context.Context, _ bool) (context.Context, database.Tx, error) {
	tx := &billingLifecycleTxState{db: d, issues: cloneLifecycleIssueStore(d.issues)}
	return context.WithValue(ctx, billingLifecycleTxContextKey{}, tx), tx, nil
}

func (tx *billingLifecycleTxState) Commit() error {
	if tx.closed {
		return errors.New("transaction already closed")
	}
	tx.closed = true
	if tx.db.commitFailures > 0 {
		tx.db.commitFailures--
		return errors.New("injected commit failure")
	}
	tx.db.issues = cloneLifecycleIssueStore(tx.issues)
	return nil
}

func (tx *billingLifecycleTxState) Rollback() error {
	tx.closed = true
	return nil
}

func (d *billingLifecycleDB) FindBillingIssueByTypeAndOverdueProcessed(ctx context.Context, issueType database.BillingIssueType, processed bool) ([]*database.BillingIssue, error) {
	var issues []*database.BillingIssue
	for _, record := range d.issueStore(ctx) {
		if record.issue.Type == issueType && record.processed == processed {
			issues = append(issues, cloneBillingIssue(record.issue))
		}
	}
	sort.Slice(issues, func(i, j int) bool {
		if issues[i].OrgID == issues[j].OrgID {
			return issues[i].ID < issues[j].ID
		}
		return issues[i].OrgID < issues[j].OrgID
	})
	return issues, nil
}

func (d *billingLifecycleDB) FindBillingIssueByTypeForOrg(ctx context.Context, orgID string, issueType database.BillingIssueType) (*database.BillingIssue, error) {
	for _, record := range d.issueStore(ctx) {
		if record.issue.OrgID == orgID && record.issue.Type == issueType {
			return cloneBillingIssue(record.issue), nil
		}
	}
	return nil, database.ErrNotFound
}

func (d *billingLifecycleDB) UpsertBillingIssue(ctx context.Context, opts *database.UpsertBillingIssueOptions) (*database.BillingIssue, error) {
	store := d.issueStore(ctx)
	for _, record := range store {
		if record.issue.OrgID == opts.OrgID && record.issue.Type == opts.Type {
			record.issue.Metadata = cloneBillingMetadata(opts.Metadata)
			record.issue.EventTime = opts.EventTime
			return cloneBillingIssue(record.issue), nil
		}
	}
	d.nextIssueID++
	issue := &database.BillingIssue{
		ID:        fmt.Sprintf("generated-issue-%d", d.nextIssueID),
		OrgID:     opts.OrgID,
		Type:      opts.Type,
		Metadata:  cloneBillingMetadata(opts.Metadata),
		EventTime: opts.EventTime,
	}
	store[issue.ID] = &billingLifecycleIssue{issue: issue}
	return cloneBillingIssue(issue), nil
}

func (d *billingLifecycleDB) UpdateBillingIssueOverdueAsProcessed(ctx context.Context, id string) error {
	if d.processedFailures[id] > 0 {
		d.processedFailures[id]--
		return errors.New("injected processed marker failure")
	}
	record := d.issueStore(ctx)[id]
	if record == nil || record.processed {
		return database.ErrNotFound
	}
	record.processed = true
	return nil
}

func (d *billingLifecycleDB) DeleteBillingIssue(ctx context.Context, id string) error {
	store := d.issueStore(ctx)
	if store[id] == nil {
		return database.ErrNotFound
	}
	delete(store, id)
	return nil
}

func (d *billingLifecycleDB) FindOrganization(_ context.Context, id string) (*database.Organization, error) {
	org := d.organizations[id]
	if org == nil {
		return nil, database.ErrNotFound
	}
	copy := *org
	return &copy, nil
}

func (d *billingLifecycleDB) FindOrganizationForBillingCustomerID(_ context.Context, customerID string) (*database.Organization, error) {
	for _, org := range d.organizations {
		if org.BillingCustomerID == customerID {
			copy := *org
			return &copy, nil
		}
	}
	return nil, database.ErrNotFound
}

func (d *billingLifecycleDB) CountProjectsForOrganization(_ context.Context, orgID string) (int, error) {
	return len(d.projects[orgID]), nil
}

func (d *billingLifecycleDB) UpdateOrganization(_ context.Context, id string, opts *database.UpdateOrganizationOptions) (*database.Organization, error) {
	org := d.organizations[id]
	if org == nil {
		return nil, database.ErrNotFound
	}
	org.Name = opts.Name
	org.DisplayName = opts.DisplayName
	org.Description = opts.Description
	org.QuotaProjects = opts.QuotaProjects
	org.QuotaDeployments = opts.QuotaDeployments
	org.QuotaSlotsTotal = opts.QuotaSlotsTotal
	org.QuotaSlotsPerDeployment = opts.QuotaSlotsPerDeployment
	org.QuotaOutstandingInvites = opts.QuotaOutstandingInvites
	org.QuotaStorageLimitBytesPerDeployment = opts.QuotaStorageLimitBytesPerDeployment
	org.QuotaSeats = opts.QuotaSeats
	org.BillingCustomerID = opts.BillingCustomerID
	org.PaymentCustomerID = opts.PaymentCustomerID
	org.BillingEmail = opts.BillingEmail
	org.BillingPlanName = opts.BillingPlanName
	org.BillingPlanDisplayName = opts.BillingPlanDisplayName
	copy := *org
	return &copy, nil
}

func (d *billingLifecycleDB) FindProjectsForOrganization(_ context.Context, orgID, afterProjectName string, limit int) ([]*database.Project, error) {
	key := projectPageKey(orgID, afterProjectName)
	if d.pageFailures[key] > 0 {
		d.pageFailures[key]--
		return nil, errors.New("second project page unavailable")
	}
	projects := d.projects[orgID]
	start := 0
	for start < len(projects) && projects[start].Name <= afterProjectName {
		start++
	}
	end := min(start+limit, len(projects))
	result := make([]*database.Project, 0, end-start)
	for _, project := range projects[start:end] {
		copy := *project
		result = append(result, &copy)
	}
	return result, nil
}

func (d *billingLifecycleDB) FindDeploymentsForProject(context.Context, string, string, string) ([]*database.Deployment, error) {
	return nil, nil
}

func (d *billingLifecycleDB) UpdateProject(_ context.Context, id string, opts *database.UpdateProjectOptions) (*database.Project, error) {
	for orgID, projects := range d.projects {
		for _, project := range projects {
			if project.ID != id {
				continue
			}
			project.PrimaryDeploymentID = opts.PrimaryDeploymentID
			d.hibernateCalls[projectKey(orgID, project.Name)]++
			copy := *project
			return &copy, nil
		}
	}
	return nil, database.ErrNotFound
}

func (d *billingLifecycleDB) issueStore(ctx context.Context) map[string]*billingLifecycleIssue {
	if tx, ok := ctx.Value(billingLifecycleTxContextKey{}).(*billingLifecycleTxState); ok {
		return tx.issues
	}
	return d.issues
}

func (d *billingLifecycleDB) addOrganization(name string) *database.Organization {
	org := &database.Organization{
		ID:                                  name,
		Name:                                name,
		DisplayName:                         strings.ToUpper(name),
		QuotaProjects:                       7,
		QuotaDeployments:                    7,
		QuotaSlotsTotal:                     7,
		QuotaSlotsPerDeployment:             7,
		QuotaOutstandingInvites:             7,
		QuotaStorageLimitBytesPerDeployment: 7,
		QuotaSeats:                          7,
		BillingCustomerID:                   "customer-" + name,
		PaymentCustomerID:                   "payment-" + name,
		BillingEmail:                        name + "@example.com",
	}
	d.organizations[org.ID] = org
	return org
}

func (d *billingLifecycleDB) addIssue(orgID string, issueType database.BillingIssueType, processed bool, metadata database.BillingIssueMetadata, eventTime time.Time) *database.BillingIssue {
	d.nextIssueID++
	issue := &database.BillingIssue{ID: fmt.Sprintf("issue-%d", d.nextIssueID), OrgID: orgID, Type: issueType, Metadata: cloneBillingMetadata(metadata), EventTime: eventTime}
	d.issues[issue.ID] = &billingLifecycleIssue{issue: issue, processed: processed}
	return cloneBillingIssue(issue)
}

func (d *billingLifecycleDB) addProjects(orgID string, count int) {
	for i := range count {
		deploymentID := fmt.Sprintf("deployment-%s-%02d", orgID, i)
		d.projects[orgID] = append(d.projects[orgID], &database.Project{
			ID:                  fmt.Sprintf("project-id-%s-%02d", orgID, i),
			OrganizationID:      orgID,
			Name:                fmt.Sprintf("project-%02d", i),
			PrimaryDeploymentID: &deploymentID,
		})
	}
}

func (d *billingLifecycleDB) isProcessed(id string) bool {
	return d.issues[id] != nil && d.issues[id].processed
}

func (d *billingLifecycleDB) issueForOrg(orgID string, issueType database.BillingIssueType) *database.BillingIssue {
	issue, err := d.FindBillingIssueByTypeForOrg(context.Background(), orgID, issueType)
	if err != nil {
		return nil
	}
	return issue
}

func (d *billingLifecycleDB) organization(id string) *database.Organization {
	org := d.organizations[id]
	copy := *org
	return &copy
}

func (d *billingLifecycleDB) projectHibernated(orgID, name string) bool {
	for _, project := range d.projects[orgID] {
		if project.Name == name {
			return project.PrimaryDeploymentID == nil
		}
	}
	return false
}

type billingLifecycleBiller struct {
	billing.Biller
	subscriptions map[string]*billing.Subscription
	invoices      map[string]*billing.Invoice
	cancellations []string
}

func (b *billingLifecycleBiller) GetActiveSubscription(_ context.Context, customerID string) (*billing.Subscription, error) {
	subscription := b.subscriptions[customerID]
	if subscription == nil {
		return nil, billing.ErrNotFound
	}
	copy := *subscription
	if subscription.Plan != nil {
		plan := *subscription.Plan
		copy.Plan = &plan
	}
	return &copy, nil
}

func (b *billingLifecycleBiller) CancelSubscriptionsForCustomer(_ context.Context, customerID string, _ billing.SubscriptionCancellationOption) (time.Time, error) {
	b.cancellations = append(b.cancellations, customerID)
	delete(b.subscriptions, customerID)
	return time.Time{}, nil
}

func (b *billingLifecycleBiller) GetInvoice(_ context.Context, invoiceID string) (*billing.Invoice, error) {
	invoice := b.invoices[invoiceID]
	if invoice == nil {
		return nil, billing.ErrNotFound
	}
	copy := *invoice
	return &copy, nil
}

func (b *billingLifecycleBiller) IsInvoiceValid(_ context.Context, invoice *billing.Invoice) bool {
	return invoice != nil && !strings.EqualFold(invoice.Status, "void")
}

func (b *billingLifecycleBiller) IsInvoicePaid(_ context.Context, invoice *billing.Invoice) bool {
	return invoice != nil && strings.EqualFold(invoice.Status, "paid")
}

type billingLifecycleEmailAttempt struct {
	to      string
	subject string
	failed  bool
}

type billingLifecycleEmailSender struct {
	mu       sync.Mutex
	attempts []billingLifecycleEmailAttempt
	failures int
}

func (s *billingLifecycleEmailSender) Send(toEmail, _ string, subject, _ string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	attempt := billingLifecycleEmailAttempt{to: toEmail, subject: subject}
	if s.failures > 0 {
		s.failures--
		attempt.failed = true
		s.attempts = append(s.attempts, attempt)
		return errors.New("injected email failure")
	}
	s.attempts = append(s.attempts, attempt)
	return nil
}

func (s *billingLifecycleEmailSender) attemptCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.attempts)
}

func (s *billingLifecycleEmailSender) successCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	count := 0
	for _, attempt := range s.attempts {
		if !attempt.failed {
			count++
		}
	}
	return count
}

func newBillingLifecycleFixture(t *testing.T) (*admin.Service, *billingLifecycleDB, *billingLifecycleBiller, *billingLifecycleEmailSender) {
	t.Helper()
	db := &billingLifecycleDB{
		organizations:     make(map[string]*database.Organization),
		issues:            make(map[string]*billingLifecycleIssue),
		projects:          make(map[string][]*database.Project),
		processedFailures: make(map[string]int),
		pageFailures:      make(map[string]int),
		hibernateCalls:    make(map[string]int),
	}
	biller := &billingLifecycleBiller{subscriptions: make(map[string]*billing.Subscription), invoices: make(map[string]*billing.Invoice)}
	sender := &billingLifecycleEmailSender{}
	urls, err := admin.NewURLs("https://admin.example.com", "https://app.example.com")
	require.NoError(t, err)
	adm := &admin.Service{DB: db, Biller: biller, Email: email.New(sender), URLs: urls, Logger: zap.NewNop()}
	return adm, db, biller, sender
}

func insertPostgresLifecycleOrg(t *testing.T, db database.DB, name string) *database.Organization {
	t.Helper()
	org, err := db.InsertOrganization(t.Context(), &database.InsertOrganizationOptions{
		Name:              name,
		DisplayName:       name,
		QuotaProjects:     7,
		QuotaDeployments:  7,
		QuotaSlotsTotal:   7,
		QuotaSeats:        7,
		BillingCustomerID: "customer-" + name,
		BillingEmail:      name + "@example.com",
	})
	require.NoError(t, err)
	return org
}

func runConcurrently(count int, fn func() error) []error {
	start := make(chan struct{})
	results := make(chan error, count)
	for range count {
		go func() {
			<-start
			results <- fn()
		}()
	}
	close(start)
	errs := make([]error, 0, count)
	for range count {
		errs = append(errs, <-results)
	}
	return errs
}

func countNilErrors(errs []error) int {
	count := 0
	for _, err := range errs {
		if err == nil {
			count++
		}
	}
	return count
}

func paymentFailedJob(customerID, invoiceID string, failedAt time.Time) *river.Job[PaymentFailedArgs] {
	return &river.Job[PaymentFailedArgs]{Args: PaymentFailedArgs{
		BillingCustomerID: customerID,
		InvoiceID:         invoiceID,
		InvoiceNumber:     "number-" + invoiceID,
		InvoiceURL:        "https://billing.example.com/" + invoiceID,
		Amount:            "100",
		Currency:          "USD",
		DueDate:           failedAt.AddDate(0, 0, 1),
		FailedAt:          failedAt,
	}}
}

func paymentFailureMetadata(invoiceID string, failedAt time.Time) *database.BillingIssueMetadataPaymentFailed {
	return &database.BillingIssueMetadataPaymentFailed{Invoices: map[string]*database.BillingIssueMetadataPaymentFailedMeta{
		invoiceID: {
			ID:                 invoiceID,
			FailedOn:           failedAt,
			GracePeriodEndDate: failedAt.AddDate(0, 0, database.BillingGracePeriodDays),
		},
	}}
}

func requireOrganizationQuotasZero(t *testing.T, org *database.Organization) {
	t.Helper()
	require.Zero(t, org.QuotaProjects)
	require.Zero(t, org.QuotaDeployments)
	require.Zero(t, org.QuotaSlotsTotal)
	require.Zero(t, org.QuotaSlotsPerDeployment)
	require.Zero(t, org.QuotaOutstandingInvites)
	require.Zero(t, org.QuotaStorageLimitBytesPerDeployment)
	require.Zero(t, org.QuotaSeats)
}

func projectPageKey(orgID, after string) string {
	return orgID + "\x00" + after
}

func projectKey(orgID, name string) string {
	return orgID + "\x00" + name
}

func cloneLifecycleIssueStore(source map[string]*billingLifecycleIssue) map[string]*billingLifecycleIssue {
	result := make(map[string]*billingLifecycleIssue, len(source))
	for id, record := range source {
		result[id] = &billingLifecycleIssue{issue: cloneBillingIssue(record.issue), processed: record.processed}
	}
	return result
}

func cloneBillingIssue(issue *database.BillingIssue) *database.BillingIssue {
	if issue == nil {
		return nil
	}
	copy := *issue
	copy.Metadata = cloneBillingMetadata(issue.Metadata)
	return &copy
}

func cloneBillingMetadata(metadata database.BillingIssueMetadata) database.BillingIssueMetadata {
	switch value := metadata.(type) {
	case *database.BillingIssueMetadataOnTrial:
		copy := *value
		return &copy
	case *database.BillingIssueMetadataTrialEnded:
		copy := *value
		return &copy
	case *database.BillingIssueMetadataSubscriptionCancelled:
		copy := *value
		return &copy
	case *database.BillingIssueMetadataPaymentFailed:
		copy := &database.BillingIssueMetadataPaymentFailed{Invoices: make(map[string]*database.BillingIssueMetadataPaymentFailedMeta, len(value.Invoices))}
		for id, invoice := range value.Invoices {
			invoiceCopy := *invoice
			copy.Invoices[id] = &invoiceCopy
		}
		return copy
	default:
		return metadata
	}
}
