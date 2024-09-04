package worker

import (
	"context"

	"go.uber.org/zap"
)

func (w *Worker) trialEndCheck(ctx context.Context) error {
	subs, err := w.admin.Biller.FindSubscriptionsPastTrialPeriod(ctx)
	if err != nil {
		return err
	}

	for _, sub := range subs {
		w.logger.Warn("subscription past trial period", zap.String("org_id", sub.Customer.ID), zap.String("org_name", sub.Customer.Name), zap.String("end_date", sub.TrialEndDate.Format("2006-01-02")))
	}
	return nil
}
