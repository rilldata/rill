package file

import (
	"context"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/bmatcuk/doublestar/v4"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/stretchr/testify/require"
)

func TestWatch_replay(t *testing.T) {
	dir, err := ioutil.TempDir("", "testwatch")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	_, err = os.Create(filepath.Join(dir, "file1"))
	require.NoError(t, err)

	subDirName := filepath.Join(dir, "subdir")
	err = os.Mkdir(subDirName, 0777)
	require.NoError(t, err)

	_, err = os.Create(filepath.Join(subDirName, "file2"))
	require.NoError(t, err)

	c := connection{
		root: dir,
	}

	events := make([]drivers.WatchEvent, 0, 2)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		c.Watch(ctx, true, func(event drivers.WatchEvent) error {
			events = append(events, event)
			return nil
		})
	}()
	time.Sleep(3 * time.Second)

	require.Equal(t, runtimev1.FileEvent_FILE_EVENT_WRITE, events[0].Type)
	require.Equal(t, "/file1", events[0].Path)

	require.Equal(t, runtimev1.FileEvent_FILE_EVENT_WRITE, events[1].Type)
	require.Equal(t, "/subdir/file2", events[1].Path)
}

func TestWatch_create_remove(t *testing.T) {
	dir, err := ioutil.TempDir("", "testwatch")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	events := make([]drivers.WatchEvent, 0, 4)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c := connection{
		root: dir,
	}

	eventsChan := make(chan drivers.WatchEvent)
	allReceivedChannel := make(chan bool)
	go func() {
		err := c.Watch(ctx, true, func(event drivers.WatchEvent) error {
			fmt.Printf("type %s path %s\n", event.Type.String(), event.Path)
			eventsChan <- event
			return nil
		})
		require.NoError(t, err)
	}()

	time.Sleep(3 * time.Second) // to be sure Watcher had time to be initialized by another thread

	fullname1 := filepath.Join(dir, "file1")
	f1, err := os.Create(fullname1)
	require.NoError(t, err)
	f1.Close()

	subDirName := filepath.Join(dir, "subdir")
	err = os.Mkdir(subDirName, 0777)
	require.NoError(t, err)

	received2 := make(chan bool)
	go func() {
		for {
			select {
			case a := <-eventsChan:
				events = append(events, a)
			}
			if len(events) == 2 {
				received2 <- true
			} else if len(events) == 5 {
				allReceivedChannel <- true
				return
			}
			time.Sleep(1 * time.Second)
		}
	}()

	waitWithSecondsTimeout(t, received2, 30)

	fullname2 := filepath.Join(subDirName, "file2")
	f2, err := os.Create(fullname2)
	require.NoError(t, err)
	f2.Close()

	fsRoot := os.DirFS(c.root)
	doublestar.GlobWalk(fsRoot, "**", func(f string, de fs.DirEntry) error {
		if !de.IsDir() {
			fmt.Printf("dir %s\n", f)
		} else {
			fmt.Printf("file %s\n", f)
		}
		return nil
	})

	time.Sleep(3 * time.Second) // wating for created files to be noticed

	err = os.Remove(fullname1)
	require.NoError(t, err)
	err = os.Remove(fullname2)
	require.NoError(t, err)

	waitWithSecondsTimeout(t, allReceivedChannel, 30)

	require.Equal(t, 5, len(events), "%v", events)
	require.Equal(t, runtimev1.FileEvent_FILE_EVENT_WRITE, events[0].Type)
	require.Equal(t, "/file1", events[0].Path)

	require.Equal(t, runtimev1.FileEvent_FILE_EVENT_WRITE, events[1].Type)
	require.Equal(t, "/subdir", events[1].Path)

	require.Equal(t, runtimev1.FileEvent_FILE_EVENT_WRITE, events[2].Type)
	require.Equal(t, "/subdir/file2", events[2].Path)

	require.Equal(t, runtimev1.FileEvent_FILE_EVENT_DELETE, events[3].Type)
	require.Equal(t, "/file1", events[3].Path)

	require.Equal(t, runtimev1.FileEvent_FILE_EVENT_DELETE, events[4].Type)
	require.Equal(t, "/subdir/file2", events[4].Path)
}

func waitWithSecondsTimeout(t *testing.T, c <-chan bool, seconds time.Duration) {
	select {
	case <-c:
	case <-time.After(seconds * time.Second):
		t.FailNow()
	}
}
