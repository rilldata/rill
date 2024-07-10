package worker

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

func (w *Worker) repairOrgBilling(ctx context.Context) error {
	startTime := time.Now().UTC()
	t, err := w.admin.DB.FindBillingRepairedOn(ctx)
	if err != nil {
		return fmt.Errorf("failed to get last billing repair time: %w", err)
	}

	ids, err := w.admin.DB.FindOrganizationIDsWithoutPaymentCreatedOnOrAfter(ctx, t)
	if err != nil {
		return fmt.Errorf("failed to get organizations without payment created after %s: %w", t, err)
	}

	for _, id := range ids {
		org, err := w.admin.DB.FindOrganization(ctx, id)
		if err != nil {
			w.logger.Error("failed to get organization", zap.String("org_id", id), zap.Error(err))
			continue
		}
		_, _, err = w.admin.RepairOrgBilling(ctx, org)
		if err != nil {
			w.logger.Error("failed to repair billing for organization", zap.String("org_name", org.Name), zap.String("org_id", id), zap.Error(err))
			continue
		}
		w.logger.Info("repaired billing for organization", zap.String("org_name", org.Name), zap.String("org_id", id))
	}

	err = w.admin.DB.UpdateBillingRepairedOn(ctx, startTime)
	if err != nil {
		return fmt.Errorf("failed to update last billing repair time: %w", err)
	}
	return nil
}
