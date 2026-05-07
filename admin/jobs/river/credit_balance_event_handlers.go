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

// Work handles a credit_balance_dropped Orb webhook. We re-fetch the live balance to confirm it's actually below the low-credit threshold because the webhook is unordered and the customer may have topped up between trigger and delivery.
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
	if err := updateOnCreditTrialApproxBalance(ctx, w.admin, bi, md, balance); err != nil {
		return fmt.Errorf("failed to update on-credit-trial approximate credit balance for org %q: %w", org.Name, err)
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

	return nil
}

type CreditTrialLowBalanceRefreshArgs struct{}

func (CreditTrialLowBalanceRefreshArgs) Kind() string { return "credit_trial_low_balance_refresh" }

type CreditTrialLowBalanceRefreshWorker struct {
	river.WorkerDefaults[CreditTrialLowBalanceRefreshArgs]
	admin  *admin.Service
	logger *zap.Logger
}

// Work refreshes approximate credit-trial balances after the customer has crossed the single low-balance alert threshold.
func (w *CreditTrialLowBalanceRefreshWorker) Work(ctx context.Context, job *river.Job[CreditTrialLowBalanceRefreshArgs]) error {
	issues, err := w.admin.DB.FindBillingIssueByType(ctx, database.BillingIssueTypeOnCreditTrial)
	if err != nil {
		return fmt.Errorf("failed to find on-credit-trial billing issues: %w", err)
	}

	for _, issue := range issues {
		md, ok := issue.Metadata.(*database.BillingIssueMetadataOnCreditTrial)
		if !ok {
			return fmt.Errorf("unexpected metadata type for on-credit-trial issue %q", issue.ID)
		}
		if !md.LowCredit {
			continue
		}

		org, err := w.admin.DB.FindOrganization(ctx, issue.OrgID)
		if err != nil {
			if errors.Is(err, database.ErrNotFound) {
				w.logger.Named("billing").Warn("skipping credit-trial balance refresh: organization not found", zap.String("org_id", issue.OrgID), zap.String("billing_issue_id", issue.ID))
				continue
			}
			return fmt.Errorf("failed to find organization for on-credit-trial issue %q: %w", issue.ID, err)
		}
		if org.BillingCustomerID == "" {
			continue
		}

		balance, err := w.admin.Biller.GetCustomerCreditBalance(ctx, org.BillingCustomerID, billing.CreditsCurrency)
		if err != nil {
			return fmt.Errorf("failed to fetch credit balance for org %q: %w", org.Name, err)
		}
		if err := updateOnCreditTrialApproxBalance(ctx, w.admin, issue, md, balance); err != nil {
			return fmt.Errorf("failed to update on-credit-trial approximate credit balance for org %q: %w", org.Name, err)
		}
	}

	return nil
}

func updateOnCreditTrialApproxBalance(ctx context.Context, adm *admin.Service, issue *database.BillingIssue, md *database.BillingIssueMetadataOnCreditTrial, balance float64) error {
	md.ApproxLowCreditsBalance = balance
	if balance < billing.CreditTrialLowBalanceThreshold {
		md.LowCredit = true
	}
	_, err := adm.DB.UpsertBillingIssue(ctx, &database.UpsertBillingIssueOptions{
		OrgID:     issue.OrgID,
		Type:      database.BillingIssueTypeOnCreditTrial,
		Metadata:  md,
		EventTime: issue.EventTime,
	})
	return err
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
// we cancel the trial subscription, hibernate every project, raise BillingIssueTypeTrialCreditsDepleted (blocks API operations) and BillingIssueTypeSubscriptionCancelled.
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

	subEndDate, err := w.admin.Biller.CancelSubscriptionsForCustomer(ctx, org.BillingCustomerID, billing.SubscriptionCancellationOptionImmediate)
	if err != nil {
		return fmt.Errorf("failed to cancel trial subscription for org %q: %w", org.Name, err)
	}
	if subEndDate.IsZero() {
		subEndDate = time.Now().UTC()
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

	if _, err := w.admin.DB.UpsertBillingIssue(txCtx, &database.UpsertBillingIssueOptions{
		OrgID: org.ID,
		Type:  database.BillingIssueTypeSubscriptionCancelled,
		Metadata: &database.BillingIssueMetadataSubscriptionCancelled{
			EndDate: subEndDate,
		},
		EventTime: time.Now().UTC(),
	}); err != nil {
		return fmt.Errorf("failed to raise subscription-cancelled issue: %w", err)
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
		UpgradeURL:  w.admin.URLs.Billing(org.Name, false),
	})
	if err != nil {
		return fmt.Errorf("failed to send credit trial depleted email for org %q: %w", org.Name, err)
	}

	return nil
}
