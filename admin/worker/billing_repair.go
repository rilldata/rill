package worker

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

func (w *Worker) repairOrgBilling(ctx context.Context) error {
	orgs, err := w.admin.DB.FindOrganizationsWithoutPaymentCustomerID(ctx)
	if err != nil {
		return fmt.Errorf("failed to get organizations without payment id: %w", err)
	}

	for _, org := range orgs {
		_, _, err = w.admin.RepairOrgBilling(ctx, org)
		if err != nil {
			w.logger.Error("failed to repair billing for organization", zap.String("org_name", org.Name), zap.String("org_id", org.ID), zap.Error(err))
			continue
		}
		w.logger.Info("repaired billing for organization", zap.String("org_name", org.Name), zap.String("org_id", org.ID))
	}
	return nil
}
