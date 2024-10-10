package worker

import (
	"context"

	"github.com/rilldata/rill/admin/database"
)

func (w *Worker) deleteUnusedAssets(ctx context.Context) error {
	return w.admin.BatchDeleteAssets(ctx, func(ctx context.Context, pageSize int) ([]*database.Asset, error) {
		return w.admin.DB.FindUnusedAssets(ctx, pageSize)
	})
}
