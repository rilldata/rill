package river

import (
	"context"
	"fmt"
	"time"

	"github.com/rilldata/rill/admin"
	"github.com/rilldata/rill/admin/database"
	"github.com/riverqueue/river"
	"go.uber.org/zap"
)

type CreditCheckArgs struct{}

func (CreditCheckArgs) Kind() string { return "credit_check" }

type CreditCheckWorker struct {
	river.WorkerDefaults[CreditCheckArgs]
	admin  *admin.Service
	logger *zap.Logger
}

func (w *CreditCheckWorker) Work(ctx context.Context, job *river.Job[CreditCheckArgs]) error {
	limit := 100
	afterName := ""

	for {
		orgs, err := w.admin.DB.FindOrganizationsByBillingPlanName(ctx, "free-plan", afterName, limit)
		if err != nil {
			return fmt.Errorf("failed to find free-plan orgs: %w", err)
		}

		for _, org := range orgs {
			if err := w.checkOrg(ctx, org); err != nil {
				w.logger.Warn("credit check failed for org", zap.String("org_id", org.ID), zap.String("org_name", org.Name), zap.Error(err))
			}
			afterName = org.Name
		}

		if len(orgs) < limit {
			break
		}
	}

	return nil
}

func (w *CreditCheckWorker) checkOrg(ctx context.Context, org *database.Organization) error {
	if org.BillingCustomerID == "" {
		return nil
	}

	bal, err := w.admin.Biller.GetCreditBalance(ctx, org.BillingCustomerID)
	if err != nil {
		return fmt.Errorf("failed to get credit balance: %w", err)
	}
	if bal == nil || bal.TotalCredit <= 0 {
		return nil
	}

	usedFraction := bal.UsedCredit / bal.TotalCredit
	now := time.Now().UTC()

	switch {
	case bal.RemainingCredit <= 0:
		// 100%: exhausted — upsert exhausted issue and hibernate all projects
		_, err = w.admin.DB.UpsertBillingIssue(ctx, &database.UpsertBillingIssueOptions{
			OrgID: org.ID,
			Type:  database.BillingIssueTypeCreditExhausted,
			Metadata: &database.BillingIssueMetadataCreditExhausted{
				CreditTotal:  bal.TotalCredit,
				CreditExpiry: bal.ExpiryDate,
				ExhaustedOn:  now,
			},
			EventTime: now,
		})
		if err != nil {
			return fmt.Errorf("failed to upsert credit exhausted issue: %w", err)
		}

		// clean up lower-severity issues since exhausted supersedes them
		for _, t := range []database.BillingIssueType{database.BillingIssueTypeCreditLow, database.BillingIssueTypeCreditCritical} {
			if delErr := w.admin.DB.DeleteBillingIssueByTypeForOrg(ctx, org.ID, t); delErr != nil {
				w.logger.Warn("failed to delete lower-severity credit issue", zap.String("org_id", org.ID), zap.Error(delErr))
			}
		}

		// hibernate all active projects
		projLimit := 10
		afterProjectName := ""
		for {
			projs, projErr := w.admin.DB.FindProjectsForOrganization(ctx, org.ID, afterProjectName, projLimit)
			if projErr != nil {
				return fmt.Errorf("failed to find projects for org: %w", projErr)
			}
			for _, proj := range projs {
				if _, hibErr := w.admin.HibernateProject(ctx, proj); hibErr != nil {
					w.logger.Warn("failed to hibernate project on credit exhaustion", zap.String("project_id", proj.ID), zap.Error(hibErr))
				}
				afterProjectName = proj.Name
			}
			if len(projs) < projLimit {
				break
			}
		}

		w.logger.Warn("credit exhausted: hibernated all projects",
			zap.String("org_id", org.ID),
			zap.String("org_name", org.Name),
			zap.Float64("total_credit", bal.TotalCredit),
			zap.Float64("used_credit", bal.UsedCredit),
		)

	case usedFraction >= 0.95:
		// 95%+: critical
		_, err = w.admin.DB.UpsertBillingIssue(ctx, &database.UpsertBillingIssueOptions{
			OrgID: org.ID,
			Type:  database.BillingIssueTypeCreditCritical,
			Metadata: &database.BillingIssueMetadataCreditCritical{
				CreditRemaining: bal.RemainingCredit,
				CreditTotal:     bal.TotalCredit,
				CreditExpiry:    bal.ExpiryDate,
			},
			EventTime: now,
		})
		if err != nil {
			return fmt.Errorf("failed to upsert credit critical issue: %w", err)
		}
		// low is superseded by critical
		if delErr := w.admin.DB.DeleteBillingIssueByTypeForOrg(ctx, org.ID, database.BillingIssueTypeCreditLow); delErr != nil {
			w.logger.Warn("failed to delete credit low issue", zap.String("org_id", org.ID), zap.Error(delErr))
		}

	case usedFraction >= 0.80:
		// 80%+: low
		_, err = w.admin.DB.UpsertBillingIssue(ctx, &database.UpsertBillingIssueOptions{
			OrgID: org.ID,
			Type:  database.BillingIssueTypeCreditLow,
			Metadata: &database.BillingIssueMetadataCreditLow{
				CreditRemaining: bal.RemainingCredit,
				CreditTotal:     bal.TotalCredit,
				CreditExpiry:    bal.ExpiryDate,
			},
			EventTime: now,
		})
		if err != nil {
			return fmt.Errorf("failed to upsert credit low issue: %w", err)
		}
	}

	return nil
}
