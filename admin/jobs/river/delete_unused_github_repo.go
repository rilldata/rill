package river

import (
	"context"
	"fmt"
	"net/http"

	"github.com/rilldata/rill/admin"
	"github.com/rilldata/rill/admin/pkg/gitutil"
	"github.com/riverqueue/river"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

const _unusedGithubRepoPageSize = 100

type deleteUnusedGithubReposArgs struct{}

func (deleteUnusedGithubReposArgs) Kind() string { return "delete_unused_github_repos" }

type deleteUnusedGithubReposWorker struct {
	river.WorkerDefaults[deleteUnusedGithubReposArgs]
	admin  *admin.Service
	logger *zap.Logger
}

// deleteUnusedGithubRepos deletes unused Rill managed Github repositories from the database and Github.
// An unused repository is one that is not associated with any Rill project since more than 7 days.
func (w *deleteUnusedGithubReposWorker) Work(ctx context.Context, job *river.Job[deleteUnusedGithubReposArgs]) error {
	for {
		// 1. Fetch repositories that are not associated with any Rill project
		repos, err := w.admin.DB.FindUnusedManagedGitRepos(ctx, _unusedGithubRepoPageSize)
		if err != nil {
			return err
		}
		if len(repos) == 0 {
			return nil
		}

		// 2. Delete repos
		id, err := w.admin.Github.ManagedOrgInstallationID()
		if err != nil {
			return fmt.Errorf("failed to get managed org installation id: %w", err)
		}

		client := w.admin.Github.InstallationClient(id, nil)

		// Limit the number of concurrent deletes to 8
		group, cctx := errgroup.WithContext(ctx)
		group.SetLimit(8)
		var ids []string
		for _, repo := range repos {
			repo := repo
			ids = append(ids, repo.ID)
			group.Go(func() error {
				account, name, ok := gitutil.SplitGithubRemote(repo.Remote)
				if !ok {
					w.logger.Error("invalid github url", zap.String("remote", repo.Remote), zap.String("repo_id", repo.ID))
				}
				resp, err := client.Repositories.Delete(cctx, account, name)
				if err != nil {
					if resp != nil && resp.StatusCode == http.StatusNotFound {
						// already deleted, ignore
						return nil
					}
					return fmt.Errorf("failed to delete github repo %q: %w", repo.Remote, err)
				}
				return nil
			})
		}
		err = group.Wait()
		if err != nil {
			return err
		}

		// 3. Delete the meta in the DB
		err = w.admin.DB.DeleteManagedGitRepos(ctx, ids)
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
