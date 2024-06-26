package worker

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

const _defaultPageSize = 20

func (w *Worker) deleteUnusedAssets(ctx context.Context) error {
	// We skip unused assets created in last 15 minutes to prevent race condition
	// where somebody just created an asset but is yet to use it
	createdBefore := time.Now().Add(-15 * time.Minute)
	for {
		// 1. Fetch unused assets
		assets, err := w.admin.DB.FindUnusedAssets(ctx, createdBefore, _defaultPageSize)
		if err != nil {
			return err
		}
		if len(assets) == 0 {
			return nil
		}
		createdBefore = assets[len(assets)-1].CreatedOn
		ids := make([]string, len(assets))

		// 2. Delete objects from cloud storage
		// Limit the number of concurrent deletes to 8
		// TODO: Use batch API once google-cloud-go supports it
		group, cctx := errgroup.WithContext(ctx)
		group.SetLimit(8)
		for j := 0; j < len(assets); j++ {
			i := j
			group.Go(func() error {
				parsed, err := url.Parse(assets[i].Path)
				if err != nil {
					w.logger.Warn("failed to parse asset path", zap.String("path", assets[i].Path), zap.Error(err))
					return nil
				}
				err = w.admin.AssetsBucket.Object(strings.TrimPrefix(parsed.Path, "/")).Delete(cctx)
				if err != nil && !errors.Is(err, storage.ErrObjectNotExist) {
					w.logger.Warn("failed to delete asset", zap.String("path", assets[i].Path), zap.Error(err))
					return nil
				}
				// collect ids for which delete was successful or object was not found
				ids[i] = assets[i].ID
				return nil
			})
		}
		_ = group.Wait()

		// 3. Delete the assets in the DB
		var finalIDs []string
		for _, id := range ids {
			if id != "" {
				finalIDs = append(finalIDs, id)
			}
		}
		if len(finalIDs) == 0 {
			// No assets were safely deleted so could be an issue with google cloud storage,network etc
			// we return and execute again in the next run of this job
			return nil
		}
		err = w.admin.DB.DeleteAssets(ctx, finalIDs)
		if err != nil {
			return err
		}

		if len(assets) < _defaultPageSize {
			// no more assets to delete
			return nil
		}
	}
}
