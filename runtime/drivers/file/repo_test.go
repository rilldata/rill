package file

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/stretchr/testify/require"
)

func TestWatch(t *testing.T) {
	dir := t.TempDir()
	c := connection{root: dir}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch := make(chan drivers.WatchEvent, 10)
	go func() {
		err := c.Watch(ctx, "", func(es []drivers.WatchEvent) {
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
	err := os.Mkdir(subDirName, 0777)
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
	var events []drivers.WatchEvent
	done := false
	for !done {
		select {
		case e := <-ch:
			events = append(events, e)
		default:
			done = true
		}
	}
	require.Len(t, events, 5)

	require.Equal(t, 5, len(events), "%v", events)
	require.Equal(t, runtimev1.FileEvent_FILE_EVENT_WRITE, events[0].Type)
	require.Equal(t, "/file1", events[0].Path)

	require.Equal(t, runtimev1.FileEvent_FILE_EVENT_WRITE, events[1].Type)
	require.Equal(t, "/subdir", events[1].Path)
	require.Equal(t, true, events[1].Dir)

	require.Equal(t, runtimev1.FileEvent_FILE_EVENT_WRITE, events[2].Type)
	require.Equal(t, "/subdir/file2", events[2].Path)

	require.Equal(t, runtimev1.FileEvent_FILE_EVENT_DELETE, events[3].Type)
	require.Equal(t, "/file1", events[3].Path)

	require.Equal(t, runtimev1.FileEvent_FILE_EVENT_DELETE, events[4].Type)
	require.Equal(t, "/subdir/file2", events[4].Path)
}

func createFile(t *testing.T, fullname string) {
	f, err := os.Create(fullname)
	require.NoError(t, err)
	err = f.Close()
	require.NoError(t, err)
}
