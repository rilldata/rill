package admin

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	virtualDir       = "__virtual__"
	virtualRetryN    = 3
	virtualRetryWait = 2 * time.Second
	virtualPageSize  = 100
	virtualMaxPages  = 500
)

// virtualRepo represents a repository of virtual files loaded from the Rill Admin service.
// It presents all virtual files as residing under the /__virtual__ path, in order to avoid conflicts with files in archiveRepo or gitRepo.
// It is unsafe for concurrent reads and writes.
type virtualRepo struct {
	h             *Handle
	tmpDir        string
	nextPageToken string
}

func (r *virtualRepo) sync(ctx context.Context) error {
	// Call syncInner with retries
	var err error
	for i := 0; i < virtualRetryN; i++ {
		err = r.syncInner(ctx)
		if err == nil {
			break
		}
		code := status.Code(err)
		if code != codes.Unavailable && code != codes.Internal {
			break
		}
		select {
		case <-time.After(virtualRetryWait):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return err
}

func (r *virtualRepo) syncInner(ctx context.Context) error {
	i := 0
	for i = 0; i < virtualMaxPages; i++ { // Just a failsafe to avoid infinite loops
		res, err := r.h.admin.PullVirtualRepo(ctx, &adminv1.PullVirtualRepoRequest{
			ProjectId: r.h.config.ProjectID,
			PageSize:  virtualPageSize,
			PageToken: r.nextPageToken,
		})
		if err != nil {
			return fmt.Errorf("failed to sync virtual repo: %w", err)
		}

		for _, vf := range res.Files {
			path := filepath.Join(r.tmpDir, virtualDir, vf.Path)

			if vf.Deleted {
				err = os.Remove(path)
				if err != nil && !os.IsNotExist(err) {
					return fmt.Errorf("failed to remove virtual file %q: %w", path, err)
				}
				continue
			}

			err = os.MkdirAll(filepath.Dir(path), os.ModePerm)
			if err != nil {
				return fmt.Errorf("could not create directory for virtual file %q: %w", path, err)
			}

			err = os.WriteFile(path, vf.Data, os.ModePerm)
			if err != nil {
				return fmt.Errorf("failed to write virtual file %q: %w", path, err)
			}
		}

		r.nextPageToken = res.NextPageToken

		// If there are no more files, we're done for now.
		// We can't just check NextPageToken because it will still be set, enabling us to pull new changes next time this function is called.
		if len(res.Files) == 0 {
			break
		}
	}

	if i == virtualMaxPages {
		return fmt.Errorf("internal: virtualRepo: syncInner ran for over %d iterations", virtualMaxPages)
	}

	return nil
}

func (r *virtualRepo) root() string {
	return r.tmpDir
}
