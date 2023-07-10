package file

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/stretchr/testify/require"
)

func TestWatch(t *testing.T) {
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
