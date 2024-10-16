package worker

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

func (w *Worker) repairOrgBilling(ctx context.Context) error {
	ids, err := w.admin.DB.FindOrganizationIDsWithoutBilling(ctx)
	if err != nil {
		return fmt.Errorf("failed to get organizations without billing id: %w", err)
	}

	for _, orgID := range ids {
		// TODO limit concurrency by a having a separate queue or submit limited jobs in each run
		_, err = w.admin.Jobs.RepairOrgBilling(ctx, orgID)
		if err != nil {
			w.logger.Named("billing").Error("failed to submit repair billing job", zap.String("org_id", orgID), zap.Error(err))
			continue
		}
	}
	return nil
}
