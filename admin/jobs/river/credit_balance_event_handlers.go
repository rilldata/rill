package river

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/rilldata/rill/admin"
	"github.com/rilldata/rill/admin/billing"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/runtime/pkg/email"
	"github.com/riverqueue/river"
	"go.uber.org/zap"
)

type CreditBalanceDroppedArgs struct {
	BillingCustomerID string
}

func (CreditBalanceDroppedArgs) Kind() string { return "credit_balance_dropped" }

type CreditBalanceDroppedWorker struct {
	river.WorkerDefaults[CreditBalanceDroppedArgs]
	admin  *admin.Service
	logger *zap.Logger
}

// Work handles a credit_balance_dropped Orb webhook. We re-fetch the live balance to confirm it's actually below the low-credit threshold (the webhook is unordered and the customer may have topped up between trigger and delivery), send the warning email once per trial, and persist the LowCredit flag on the OnCreditTrial issue so subsequent dropped events for the same trial are no-ops.
func (w *CreditBalanceDroppedWorker) Work(ctx context.Context, job *river.Job[CreditBalanceDroppedArgs]) error {
	org, err := w.admin.DB.FindOrganizationForBillingCustomerID(ctx, job.Args.BillingCustomerID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil
		}
		return fmt.Errorf("failed to find organization for billing customer %q: %w", job.Args.BillingCustomerID, err)
	}

	bi, err := w.admin.DB.FindBillingIssueByTypeForOrg(ctx, org.ID, database.BillingIssueTypeOnCreditTrial)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			w.logger.Named("billing").Debug("credit_balance_dropped received but org is not on credit trial; ignoring", zap.String("org_id", org.ID))
			return nil
		}
		return fmt.Errorf("failed to find on-credit-trial issue for org %q: %w", org.Name, err)
	}
	md, ok := bi.Metadata.(*database.BillingIssueMetadataOnCreditTrial)
	if !ok {
		return fmt.Errorf("unexpected metadata type for on-credit-trial issue for org %q", org.Name)
	}

	balance, err := w.admin.Biller.GetCustomerCreditBalance(ctx, org.BillingCustomerID, billing.CreditsCurrency)
	if err != nil {
		return fmt.Errorf("failed to fetch credit balance for org %q: %w", org.Name, err)
	}
	if balance >= billing.CreditTrialLowBalanceThreshold {
		w.logger.Named("billing").Info("credit_balance_dropped webhook ignored: balance is no longer below the low-credit threshold", zap.String("org_id", org.ID), zap.Float64("balance", balance), zap.Float64("threshold", billing.CreditTrialLowBalanceThreshold))
		return nil
	}

	err = w.admin.Email.SendCreditTrialLow(&email.CreditTrialLow{
		ToEmail:          org.BillingEmail,
		ToName:           org.Name,
		OrgName:          org.Name,
		FrontendURL:      w.admin.URLs.Frontend(),
		UpgradeURL:       w.admin.URLs.Billing(org.Name, true),
		RemainingBalance: balance,
	})
	if err != nil {
		return fmt.Errorf("failed to send credit trial low email for org %q: %w", org.Name, err)
	}

	md.LowCredit = true
	if _, err := w.admin.DB.UpsertBillingIssue(ctx, &database.UpsertBillingIssueOptions{
		OrgID:     org.ID,
		Type:      database.BillingIssueTypeOnCreditTrial,
		Metadata:  md,
		EventTime: bi.EventTime,
	}); err != nil {
		return fmt.Errorf("failed to mark on-credit-trial low_credit flag for org %q: %w", org.Name, err)
	}

	return nil
}

type CreditBalanceDepletedArgs struct {
	BillingCustomerID string
}

func (CreditBalanceDepletedArgs) Kind() string { return "credit_balance_depleted" }

type CreditBalanceDepletedWorker struct {
	river.WorkerDefaults[CreditBalanceDepletedArgs]
	admin  *admin.Service
	logger *zap.Logger
}

// Work handles a credit_balance_depleted Orb webhook. We re-fetch the balance, on a confirmed zero balance for an org that is on the credit trial,
// we cancel the trial subscription, hibernate every project, raise BillingIssueTypeTrialCreditsDepleted (blocks API operations).
func (w *CreditBalanceDepletedWorker) Work(ctx context.Context, job *river.Job[CreditBalanceDepletedArgs]) error {
	org, err := w.admin.DB.FindOrganizationForBillingCustomerID(ctx, job.Args.BillingCustomerID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil
		}
		return fmt.Errorf("failed to find organization for billing customer %q: %w", job.Args.BillingCustomerID, err)
	}

	onTrial, err := w.admin.DB.FindBillingIssueByTypeForOrg(ctx, org.ID, database.BillingIssueTypeOnCreditTrial)
	if err != nil && !errors.Is(err, database.ErrNotFound) {
		return fmt.Errorf("failed to find on-credit-trial issue: %w", err)
	}
	if onTrial == nil {
		w.logger.Named("billing").Warn("credit_balance_depleted webhook ignored: org is not on the credit trial", zap.String("org_id", org.ID), zap.String("org_name", org.Name))
		return nil
	}

	balance, err := w.admin.Biller.GetCustomerCreditBalance(ctx, org.BillingCustomerID, billing.CreditsCurrency)
	if err != nil {
		return fmt.Errorf("failed to fetch credit balance for org %q: %w", org.Name, err)
	}
	if balance > 0 {
		w.logger.Named("billing").Info("credit_balance_depleted webhook ignored: balance recovered before processing", zap.String("org_id", org.ID), zap.Float64("balance", balance))
		return nil
	}

	var subID, planID string
	if m, ok := onTrial.Metadata.(*database.BillingIssueMetadataOnCreditTrial); ok {
		subID = m.SubID
		planID = m.PlanID
	}

	_, err = w.admin.Biller.CancelSubscriptionsForCustomer(ctx, org.BillingCustomerID, billing.SubscriptionCancellationOptionImmediate)
	if err != nil {
		return fmt.Errorf("failed to cancel trial subscription for org %q: %w", org.Name, err)
	}

	txCtx, tx, err := w.admin.DB.NewTx(ctx, false)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if err := w.admin.DB.DeleteBillingIssue(txCtx, onTrial.ID); err != nil {
		return fmt.Errorf("failed to delete on-credit-trial issue: %w", err)
	}

	if _, err := w.admin.DB.UpsertBillingIssue(txCtx, &database.UpsertBillingIssueOptions{
		OrgID: org.ID,
		Type:  database.BillingIssueTypeTrialCreditsDepleted,
		Metadata: &database.BillingIssueMetadataTrialCreditsDepleted{
			SubID:      subID,
			PlanID:     planID,
			DepletedOn: time.Now().UTC(),
		},
		EventTime: time.Now().UTC(),
	}); err != nil {
		return fmt.Errorf("failed to raise trial-credits-depleted issue: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit billing issue updates: %w", err)
	}

	limit := 10
	afterProjectName := ""
	for {
		projs, err := w.admin.DB.FindProjectsForOrganization(ctx, org.ID, afterProjectName, limit)
		if err != nil {
			return err
		}
		for _, proj := range projs {
			if _, err := w.admin.HibernateProject(ctx, proj); err != nil {
				return fmt.Errorf("failed to hibernate project %q: %w", proj.Name, err)
			}
			afterProjectName = proj.Name
		}
		if len(projs) < limit {
			break
		}
	}

	w.logger.Named("billing").Warn("trial subscription cancelled and projects hibernated due to depleted trial credits", zap.String("org_id", org.ID), zap.String("org_name", org.Name))

	err = w.admin.Email.SendCreditTrialDepleted(&email.CreditTrialDepleted{
		ToEmail:     org.BillingEmail,
		ToName:      org.Name,
		OrgName:     org.Name,
		FrontendURL: w.admin.URLs.Frontend(),
		UpgradeURL:  w.admin.URLs.Billing(org.Name, true),
	})
	if err != nil {
		return fmt.Errorf("failed to send credit trial depleted email for org %q: %w", org.Name, err)
	}

	return nil
}
