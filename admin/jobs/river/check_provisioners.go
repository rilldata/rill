package river

import (
	"context"
	"fmt"

	"github.com/rilldata/rill/admin"
	"github.com/riverqueue/river"
	"go.uber.org/zap"
)

type CheckProvisionersArgs struct{}

func (CheckProvisionersArgs) Kind() string { return "check_provisioners" }

type CheckProvisionersWorker struct {
	river.WorkerDefaults[CheckProvisionersArgs]
	admin  *admin.Service
	logger *zap.Logger
}

func (w *CheckProvisionersWorker) Work(ctx context.Context, job *river.Job[CheckProvisionersArgs]) error {
	// Check every provisioner in the set individually
	for _, p := range w.admin.ProvisionerSet {
		err := p.Check(ctx)
		if err != nil {
			return fmt.Errorf("failed to check provisioner capacity: %w", err)
		}
	}

	return nil
}
