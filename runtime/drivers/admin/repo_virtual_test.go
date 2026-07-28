package admin

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/rilldata/rill/runtime/pkg/ctxsync"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/singleflight"
)

// newOverlayTestRepo builds a repo whose git root holds the given files and whose virtual root is empty,
// bypassing the admin handshake. Feed events are applied with virtualRepo.applyFile.
func newOverlayTestRepo(t *testing.T, gitFiles map[string]string) *repo {
	gitDir := t.TempDir()
	for p, content := range gitFiles {
		full := filepath.Join(gitDir, filepath.FromSlash(p))
		require.NoError(t, os.MkdirAll(filepath.Dir(full), os.ModePerm))
		require.NoError(t, os.WriteFile(full, []byte(content), 0o644))
	}
	return &repo{
		mu:           ctxsync.NewRWMutex(),
		singleflight: &singleflight.Group{},
		ready:        true,
		git:          &gitRepo{repoDir: gitDir},
		virtual:      &virtualRepo{tmpDir: t.TempDir()},
	}
}

func TestVirtualRepoOverlay(t *testing.T) {
	ctx := context.Background()
	alertPath := "user_files/jane-u1/alerts/my-alert.yaml"
	r := newOverlayTestRepo(t, map[string]string{
		alertPath:      "git copy",
		"models/m.sql": "select 1",
	})

	// A staged (dirty) file shadows the committed Git copy at the same path.
	require.NoError(t, r.virtual.applyFile(&adminv1.VirtualFile{Path: alertPath, Data: []byte("staged copy")}))
	data, err := r.Get(ctx, "/"+alertPath)
	require.NoError(t, err)
	require.Equal(t, "staged copy", data)

	// ListGlob returns the shadowed path only once.
	entries, err := r.ListGlob(ctx, "**", true)
	require.NoError(t, err)
	var paths []string
	for _, e := range entries {
		paths = append(paths, e.Path)
	}
	require.ElementsMatch(t, []string{"/" + alertPath, "/models/m.sql"}, paths)

	// Once synced, the staged copy is dropped and the Git copy is served (no permanent shadowing).
	require.NoError(t, r.virtual.applyFile(&adminv1.VirtualFile{Path: alertPath, Data: []byte("staged copy"), Synced: true}))
	data, err = r.Get(ctx, "/"+alertPath)
	require.NoError(t, err)
	require.Equal(t, "git copy", data)
}

func TestVirtualRepoTombstones(t *testing.T) {
	ctx := context.Background()
	alertPath := "user_files/jane-u1/alerts/my-alert.yaml"
	r := newOverlayTestRepo(t, map[string]string{alertPath: "git copy"})

	// A staged deletion hides the committed Git copy across Get, Stat and ListGlob.
	require.NoError(t, r.virtual.applyFile(&adminv1.VirtualFile{Path: alertPath, Deleted: true}))
	_, err := r.Get(ctx, "/"+alertPath)
	require.ErrorIs(t, err, os.ErrNotExist)
	_, err = r.Stat(ctx, "/"+alertPath)
	require.ErrorIs(t, err, os.ErrNotExist)
	entries, err := r.ListGlob(ctx, "**", true)
	require.NoError(t, err)
	require.Empty(t, entries)

	// Once the deletion is confirmed in Git (synced), the tombstone is cleared: if a user later re-adds
	// a file at the path directly in Git, it must not stay hidden.
	require.NoError(t, r.virtual.applyFile(&adminv1.VirtualFile{Path: alertPath, Deleted: true, Synced: true}))
	require.False(t, r.virtual.isTombstoned("/"+alertPath))
	data, err := r.Get(ctx, "/"+alertPath) // the git copy still exists in this test setup
	require.NoError(t, err)
	require.Equal(t, "git copy", data)

	// A re-created staged file clears a tombstone too.
	require.NoError(t, r.virtual.applyFile(&adminv1.VirtualFile{Path: alertPath, Deleted: true}))
	require.True(t, r.virtual.isTombstoned("/"+alertPath))
	require.NoError(t, r.virtual.applyFile(&adminv1.VirtualFile{Path: alertPath, Data: []byte("recreated")}))
	require.False(t, r.virtual.isTombstoned("/"+alertPath))
	data, err = r.Get(ctx, "/"+alertPath)
	require.NoError(t, err)
	require.Equal(t, "recreated", data)
}
