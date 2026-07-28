package admin

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	virtualRetryN    = 3
	virtualRetryWait = 2 * time.Second
	virtualPageSize  = 100
	virtualMaxPages  = 500
)

// virtualRepo represents a repository of virtual files loaded from the Rill Admin service.
// Virtual files act as a staging buffer for files that are promoted into the project's Git repo.
// They are presented at their real path (e.g. user_files/<seg>/alerts/x.yaml) so that a staged file
// overlays — and a staged deletion (tombstone) hides — the committed copy of the same path in gitRepo.
// It is unsafe for concurrent reads and writes.
type virtualRepo struct {
	h             *Handle
	tmpDir        string
	nextPageToken string
	// tombstones tracks paths that were deleted in the staging buffer. A tombstoned path must be hidden
	// across all underlying repos, otherwise a still-committed Git copy would resurface a deleted resource.
	tombstones map[string]bool
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
			if err := r.applyFile(vf); err != nil {
				return err
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

// applyFile applies one file from the PullVirtualRepo feed to the staging overlay.
func (r *virtualRepo) applyFile(vf *adminv1.VirtualFile) error {
	if r.tombstones == nil {
		r.tombstones = make(map[string]bool)
	}

	rel := filepath.FromSlash(vf.Path)
	path := filepath.Join(r.tmpDir, rel)

	// A synced file is confirmed in Git (or, for a synced deletion, confirmed gone from Git),
	// so the Git copy is authoritative: drop the staged copy so it stops shadowing the Git copy.
	if vf.Deleted || vf.Synced {
		if vf.Deleted && !vf.Synced {
			// A staged (not yet synced) deletion also hides the committed Git copy, if any.
			r.tombstones[filepath.ToSlash(rel)] = true
		} else {
			delete(r.tombstones, filepath.ToSlash(rel))
		}
		err := os.Remove(path)
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove virtual file %q: %w", path, err)
		}
		return nil
	}

	// A re-created file clears any prior tombstone for its path.
	delete(r.tombstones, filepath.ToSlash(rel))

	err := os.MkdirAll(filepath.Dir(path), os.ModePerm)
	if err != nil {
		return fmt.Errorf("could not create directory for virtual file %q: %w", path, err)
	}

	err = os.WriteFile(path, vf.Data, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to write virtual file %q: %w", path, err)
	}
	return nil
}

func (r *virtualRepo) root() string {
	return r.tmpDir
}

// isTombstoned reports whether path was deleted in the staging buffer. path is a forward-slash
// repo-relative path, optionally with a leading slash.
func (r *virtualRepo) isTombstoned(path string) bool {
	if r == nil || len(r.tombstones) == 0 {
		return false
	}
	return r.tombstones[strings.TrimPrefix(path, "/")]
}
