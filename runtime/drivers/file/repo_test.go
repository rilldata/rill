package file

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestWatch(t *testing.T) {
	dir := t.TempDir()
	// create /tmp directory and ensure watcher does not watch it
	tmpDir := filepath.Join(dir, "tmp")
	err := os.Mkdir(tmpDir, 0777)
	require.NoError(t, err)
	createFile(t, filepath.Join(tmpDir, "file3"))
	c := connection{root: dir, logger: zap.NewNop()}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch := make(chan drivers.WatchEvent, 10)
	go func() {
		err := c.Watch(ctx, func(es []drivers.WatchEvent) {
			for _, e := range es {
				ch <- e
			}
		})
		require.ErrorIs(t, err, ctx.Err())
	}()

	time.Sleep(time.Second) // Ensure Watcher had time to be initialized

	fullname1 := filepath.Join(dir, "file1")
	createFile(t, fullname1)

	subDirName := filepath.Join(dir, "subdir")
	err = os.Mkdir(subDirName, 0777)
	require.NoError(t, err)

	fullname2 := filepath.Join(subDirName, "file2")
	createFile(t, fullname2)

	time.Sleep(2 * time.Second) // Ensure Watcher picks up events

	err = os.Remove(fullname1)
	require.NoError(t, err)
	err = os.Remove(fullname2)
	require.NoError(t, err)

	time.Sleep(2 * time.Second) // Ensure Watcher picks up events

	// Consume events
	var res []drivers.WatchEvent
	done := false
	for !done {
		select {
		case e := <-ch:
			res = append(res, e)
		default:
			done = true
		}
	}
	require.Len(t, res, 5)

	// Split events into two batches and sort (since order of events happening close to each other is not guaranteed)
	var batch1, batch2 []drivers.WatchEvent
	for _, e := range res[0:3] {
		batch1 = append(batch1, e)
	}
	for _, e := range res[3:] {
		batch2 = append(batch2, e)
	}
	less := func(a, b drivers.WatchEvent) int {
		return strings.Compare(a.Path, b.Path)
	}

	slices.SortFunc(batch1, less)
	slices.SortFunc(batch2, less)

	require.Equal(t, runtimev1.FileEvent_FILE_EVENT_WRITE, batch1[0].Type)
	require.Equal(t, "/file1", batch1[0].Path)

	require.Equal(t, runtimev1.FileEvent_FILE_EVENT_WRITE, batch1[1].Type)
	require.Equal(t, "/subdir", batch1[1].Path)
	require.Equal(t, true, batch1[1].Dir)

	require.Equal(t, runtimev1.FileEvent_FILE_EVENT_WRITE, batch1[2].Type)
	require.Equal(t, "/subdir/file2", batch1[2].Path)

	require.Equal(t, runtimev1.FileEvent_FILE_EVENT_DELETE, batch2[0].Type)
	require.Equal(t, "/file1", batch2[0].Path)

	require.Equal(t, runtimev1.FileEvent_FILE_EVENT_DELETE, batch2[1].Type)
	require.Equal(t, "/subdir/file2", batch2[1].Path)

	files := c.watcher.watcher.WatchList()
	require.NotContains(t, files, tmpDir)
}

func createFile(t *testing.T, fullname string) {
	f, err := os.Create(fullname)
	require.NoError(t, err)
	err = f.Close()
	require.NoError(t, err)
}
