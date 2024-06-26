package worker

import (
	"context"
	"net/url"
	"strings"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

func (w *Worker) deleteUnusedAssets(ctx context.Context) error {
	assets, err := w.admin.DB.DeleteUnusedAssets(ctx)
	if err != nil {
		return err
	}
	if len(assets) == 0 {
		return nil
	}

	// Limit the number of concurrent deletes to 8
	// TODO: Use batch API once google-cloud-go supports it
	group, ctx := errgroup.WithContext(ctx)
	group.SetLimit(8)
	for _, asset := range assets {
		asset := asset
		group.Go(func() error {
			parsed, err := url.Parse(asset.Path)
			if err != nil {
				w.logger.Warn("failed to parse asset path", zap.String("path", asset.Path), zap.Error(err))
				return nil
			}
			err = w.admin.AssetsBucket.Object(strings.TrimPrefix(parsed.Path, "/")).Delete(ctx)
			if err != nil {
				w.logger.Warn("failed to delete asset", zap.String("path", asset.Path), zap.Error(err))
			}
			return nil
		})
	}
	_ = group.Wait()
	return nil
}
