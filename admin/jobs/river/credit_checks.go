package river

import (
	"context"

	"github.com/rilldata/rill/admin"
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

// Work is a no-op; credit checks are now handled by Orb.
func (w *CreditCheckWorker) Work(ctx context.Context, job *river.Job[CreditCheckArgs]) error {
	w.logger.Debug("credit check: no-op, handled by Orb")
	return nil
}
