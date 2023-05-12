package testruntime

import (
	"context"
	"fmt"
	"path/filepath"
	goruntime "runtime"
	"strings"

	"github.com/c2h5oh/datasize"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/stretchr/testify/require"

	// Load database drivers for testing.
	_ "github.com/rilldata/rill/runtime/drivers/duckdb"
	_ "github.com/rilldata/rill/runtime/drivers/file"
	_ "github.com/rilldata/rill/runtime/drivers/sqlite"
)

// TestingT satisfies both *testing.T and *testing.B.
type TestingT interface {
	Name() string
	TempDir() string
	FailNow()
	Errorf(format string, args ...interface{})
}

// New returns a runtime configured for use in tests.
func New(t TestingT) *runtime.Runtime {
	opts := &runtime.Options{
		ConnectionCacheSize: 100,
		MetastoreDriver:     "sqlite",
		// Setting a test-specific name ensures a unique connection when "cache=shared" is enabled.
		// "cache=shared" is needed to prevent threading problems.
		MetastoreDSN:          fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name()),
		QueryCacheSizeInBytes: int64(datasize.MB * 10),
		AllowHostAccess:       true,
	}
	rt, err := runtime.New(opts, nil)
	require.NoError(t, err)

	return rt
}

// NewInstance creates a runtime and an instance for use in tests.
// The instance's repo is a temp directory that will be cleared when the tests finish.
func NewInstance(t TestingT) (*runtime.Runtime, string) {
	rt := New(t)

	inst := &drivers.Instance{
		OLAPDriver:   "duckdb",
		OLAPDSN:      "",
		RepoDriver:   "file",
		RepoDSN:      t.TempDir(),
		EmbedCatalog: true,
	}

	err := rt.CreateInstance(context.Background(), inst)
	require.NoError(t, err)
	require.NotEmpty(t, inst.ID)

	return rt, inst.ID
}

// NewInstanceWithModel creates a runtime and an instance for use in tests.
// The passed model name and SQL SELECT statement will be loaded into the instance.
func NewInstanceWithModel(t TestingT, name, sql string) (*runtime.Runtime, string) {
	rt, instanceID := NewInstance(t)

	path := filepath.Join("models", name+".sql")
	err := rt.PutFile(context.Background(), instanceID, path, strings.NewReader(sql), true, false)
	require.NoError(t, err)

	res, err := rt.Reconcile(context.Background(), instanceID, nil, nil, false, false)
	require.NoError(t, err)
	require.Empty(t, res.Errors)

	return rt, instanceID
}

// NewInstanceForProject creates a runtime and an instance for use in tests.
// The passed name should match a test project in the testdata folder.
// You should not do mutable repo operations on the returned instance.
func NewInstanceForProject(t TestingT, name string) (*runtime.Runtime, string) {
	rt := New(t)

	_, currentFile, _, _ := goruntime.Caller(0)

	inst := &drivers.Instance{
		OLAPDriver:   "duckdb",
		OLAPDSN:      "?access_mode=read_write",
		RepoDriver:   "file",
		RepoDSN:      filepath.Join(currentFile, "..", "testdata", name),
		EmbedCatalog: true,
	}

	err := rt.CreateInstance(context.Background(), inst)
	require.NoError(t, err)
	require.NotEmpty(t, inst.ID)

	res, err := rt.Reconcile(context.Background(), inst.ID, nil, nil, false, false)
	require.NoError(t, err)
	require.Empty(t, res.Errors)

	return rt, inst.ID
}
