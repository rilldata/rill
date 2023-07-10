package file

import (
	"context"
	"io/ioutil"
	"os"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/stretchr/testify/require"
)

func TestWatch(t *testing.T) {
	dir, err := ioutil.TempDir("", "prefix")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	f, err := ioutil.TempFile(dir, "test.*")
	require.NoError(t, err)

	c := connection{
		root: dir,
	}

	c.Watch(context.Background(), true, func(event drivers.WatchEvent) error {
		require.Equal(t, runtimev1.FileEvent_FILE_EVENT_WRITE, event.Type)
		require.Equal(t, f.Name(), event.Path)
		return nil
	})
}
