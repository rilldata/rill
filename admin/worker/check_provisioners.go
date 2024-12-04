package worker

import (
	"context"
	"fmt"
)

func (w *Worker) checkProvisioners(ctx context.Context) error {
	// Check every provisioner in the set individually
	for _, p := range w.admin.ProvisionerSet {
		err := p.Check(ctx)
		if err != nil {
			return fmt.Errorf("failed to check provisioner capacity: %w", err)
		}
	}

	return nil
}
