package river

import (
	"context"
	"errors"

	"github.com/rilldata/rill/admin"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/runtime/pkg/email"
	"github.com/riverqueue/river"
	"go.uber.org/zap"
)

type InitOrgBillingArgs struct {
	OrgID string
}

func (InitOrgBillingArgs) Kind() string { return "init_org_billing" }

type InitOrgBillingWorker struct {
	river.WorkerDefaults[InitOrgBillingArgs]
	admin  *admin.Service
	logger *zap.Logger
}

// Work This worker initializes the billing for an organization
func (w *InitOrgBillingWorker) Work(ctx context.Context, job *river.Job[InitOrgBillingArgs]) error {
	org, err := w.admin.DB.FindOrganization(ctx, job.Args.OrgID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			// org got deleted, ignore
			return nil
		}
		w.logger.Error("failed to find organization", zap.String("org_id", job.Args.OrgID), zap.Error(err))
		return err
	}

	if job.Attempt > 1 {
		// rare case but if its retried, we should repair the billing as it might be in some inconsistent state
		_, _, err = w.admin.RepairOrganizationBilling(ctx, org, false)
		if err != nil {
			w.logger.Error("failed to init billing for organization", zap.String("org_name", org.Name), zap.String("org_id", org.ID), zap.Error(err))
			return err
		}
		return nil
	}

	_, err = w.admin.InitOrganizationBilling(ctx, org)
	if err != nil {
		w.logger.Error("failed to init billing for organization", zap.String("org_name", org.Name), zap.String("org_id", org.ID), zap.Error(err))
		return err
	}
	return nil
}

type RepairOrgBillingArgs struct {
	OrgID string
}

func (RepairOrgBillingArgs) Kind() string { return "repair_org_billing" }

type RepairOrgBillingWorker struct {
	river.WorkerDefaults[RepairOrgBillingArgs]
	admin  *admin.Service
	logger *zap.Logger
}

// Work This worker repairs the billing for an organization
func (w *RepairOrgBillingWorker) Work(ctx context.Context, job *river.Job[RepairOrgBillingArgs]) error {
	org, err := w.admin.DB.FindOrganization(ctx, job.Args.OrgID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			// org got deleted, ignore
			return nil
		}
		w.logger.Error("failed to find organization", zap.String("org_id", job.Args.OrgID), zap.Error(err))
		return err
	}

	_, _, err = w.admin.RepairOrganizationBilling(ctx, org, true)
	if err != nil {
		w.logger.Error("failed to repair billing for organization", zap.String("org_name", org.Name), zap.String("org_id", org.ID), zap.Error(err))
		return err
	}
	return nil
}

type StartTrialArgs struct {
	OrgID string
}

func (StartTrialArgs) Kind() string { return "start_trial" }

type StartTrialWorker struct {
	river.WorkerDefaults[StartTrialArgs]
	admin  *admin.Service
	logger *zap.Logger
}

// Work This worker starts the trial for an organization
func (w *StartTrialWorker) Work(ctx context.Context, job *river.Job[StartTrialArgs]) error {
	org, err := w.admin.DB.FindOrganization(ctx, job.Args.OrgID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			// org got deleted, ignore
			return nil
		}
		w.logger.Error("failed to find organization", zap.String("org_id", job.Args.OrgID), zap.Error(err))
		return err
	}

	org, sub, err := w.admin.StartTrial(ctx, org)
	if err != nil {
		w.logger.Error("failed to start trial for organization", zap.String("org_id", job.Args.OrgID), zap.Error(err))
		return err
	}

	// send trial started email
	err = w.admin.Email.SendTrialStarted(&email.TrialStarted{
		ToEmail:      org.BillingEmail,
		ToName:       org.Name,
		OrgName:      org.Name,
		FrontendURL:  w.admin.URLs.Frontend(),
		TrialEndDate: sub.TrialEndDate,
	})
	if err != nil {
		w.logger.Error("failed to send trial started email", zap.String("org_name", org.Name), zap.String("org_id", org.ID), zap.String("billing_email", org.BillingEmail), zap.Error(err))
	}

	return nil
}

type PurgeOrgArgs struct {
	OrgID string
}

func (PurgeOrgArgs) Kind() string { return "purge_org" }

type PurgeOrgWorker struct {
	river.WorkerDefaults[PurgeOrgArgs]
	admin  *admin.Service
	logger *zap.Logger
}

// Work This worker handles the deletion of an organization and all its associated data
func (w *PurgeOrgWorker) Work(ctx context.Context, job *river.Job[PurgeOrgArgs]) error {
	org, err := w.admin.DB.FindOrganization(ctx, job.Args.OrgID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			// org got deleted, ignore
			return nil
		}
		return err
	}

	if org.BillingCustomerID != "" {
		err = w.admin.Biller.DeleteCustomer(ctx, org.BillingCustomerID)
		if err != nil {
			w.logger.Error("failed to delete billing customer", zap.String("org_id", org.ID), zap.String("org_name", org.Name), zap.Error(err))
		}
	}

	if org.PaymentCustomerID != "" {
		err = w.admin.PaymentProvider.DeleteCustomer(ctx, org.PaymentCustomerID)
		if err != nil {
			w.logger.Error("failed to delete payment customer", zap.String("org_id", org.ID), zap.String("org_name", org.Name), zap.Error(err))
		}
	}

	// delete org, billing issues will be cascade deleted
	err = w.admin.DB.DeleteOrganization(ctx, org.Name)
	if err != nil {
		return err
	}

	w.logger.Warn("organization purged", zap.String("org_id", org.ID), zap.String("org_name", org.Name))

	return nil
}
