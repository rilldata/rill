package worker

import (
	"context"
	"fmt"

	"github.com/rilldata/rill/admin/pkg/gitutil"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

const _unusedGithubRepoPageSize = 100

// deleteUnusedGithubRepo deletes unused Rill managed Github repositories from the database and Github.
// An unused repository is one that is not associated with any Rill project since more than 7 days.
func (w *Worker) deleteUnusedGithubRepo(ctx context.Context) error {
	for {
		// 1. Fetch repositories that are not associated with any Rill project
		repos, err := w.admin.DB.FindUnusedManagedGithubRepo(ctx, _unusedGithubRepoPageSize)
		if err != nil {
			return err
		}
		if len(repos) == 0 {
			return nil
		}

		// 2. Delete repos
		id, err := w.admin.Github.ManagedOrgInstallationID(ctx)
		if err != nil {
			return fmt.Errorf("failed to get managed org installation id: %w", err)
		}

		client, err := w.admin.Github.InstallationClient(id)
		if err != nil {
			return fmt.Errorf("failed to get github client: %w", err)
		}

		// Limit the number of concurrent deletes to 8
		group, cctx := errgroup.WithContext(ctx)
		group.SetLimit(8)
		var ids []string
		for _, repo := range repos {
			repo := repo
			ids = append(ids, repo.ID)
			group.Go(func() error {
				account, name, ok := gitutil.SplitGithubURL(repo.HTMLURL)
				if !ok {
					w.logger.Error("invalid github url", zap.String("url", repo.HTMLURL), zap.String("repo_id", repo.ID))
				}
				_, err := client.Repositories.Delete(cctx, account, name)
				if err != nil {
					return fmt.Errorf("failed to delete github repo %q: %w", repo.HTMLURL, err)
				}
				return nil
			})
		}
		err = group.Wait()
		if err != nil {
			return err
		}

		// 3. Delete the meta in the DB
		err = w.admin.DB.DeleteManagedGithubRepoMeta(ctx, ids)
		if err != nil {
			return err
		}

		if len(repos) < _unusedGithubRepoPageSize {
			// no more repos to delete
			return nil
		}
		// fetch again could be more repos
	}
}
