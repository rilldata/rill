package worker

import (
	"context"

	"go.uber.org/zap"
)

func (w *Worker) checkSlots(ctx context.Context) error {
	slotsUsedByRuntime, err := w.admin.DB.ResolveRuntimeSlotsUsed(ctx)
	if err != nil {
		return err
	}

	var slotsTotal, slotsUsed int
	minPctUsed := 1.0

	for _, spec := range w.admin.Provisioner.Spec.Runtimes {
		slotsTotal += spec.Slots
		for _, status := range slotsUsedByRuntime {
			if spec.Host == status.RuntimeHost {
				slotsUsed += status.SlotsUsed
				pctUsed := float64(status.SlotsUsed) / float64(spec.Slots)
				if pctUsed < minPctUsed {
					minPctUsed = pctUsed
				}
			}
		}
	}

	// Log info status
	w.logger.Info(`slots check: status`, zap.Int("runtimes", len(w.admin.Provisioner.Spec.Runtimes)), zap.Int("slots_total", slotsTotal), zap.Int("slots_used", slotsUsed), zap.Float64("min_pct_used", minPctUsed))

	// Check there's at least 20% free slots
	if float64(slotsUsed)/float64(slotsTotal) >= 0.8 {
		w.logger.Warn(`slots check: +80% of all slots used`, zap.Int("slots_total", slotsTotal), zap.Int("slots_used", slotsUsed), zap.Float64("min_pct_used", minPctUsed))
	}

	// Check there's at least one runtime with at least 30% free slots
	if slotsUsed != 0 && minPctUsed >= 0.7 {
		w.logger.Warn(`slots check: +70% of slots used on every runtime`, zap.Int("slots_total", slotsTotal), zap.Int("slots_used", slotsUsed), zap.Float64("min_pct_used", minPctUsed))
	}

	return nil
}
