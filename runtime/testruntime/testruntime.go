package testruntime

import (
	"context"
	"fmt"
	"path/filepath"
	goruntime "runtime"
	"strings"

	"github.com/c2h5oh/datasize"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

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
	Cleanup(f func())
}

// New returns a runtime configured for use in tests.
func New(t TestingT) *runtime.Runtime {
	systemConnectors := []*runtimev1.Connector{
		{
			Type: "sqlite",
			Name: "metastore",
			// Setting a test-specific name ensures a unique connection when "cache=shared" is enabled.
			// "cache=shared" is needed to prevent threading problems.
			Config: map[string]string{"dsn": fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())},
		},
	}
	opts := &runtime.Options{
		ConnectionCacheSize:     100,
		MetastoreConnector:      "metastore",
		QueryCacheSizeBytes:     int64(datasize.MB * 100),
		AllowHostAccess:         true,
		SystemConnectors:        systemConnectors,
		SecurityEngineCacheSize: 100,
	}
	rt, err := runtime.New(opts, zap.NewNop(), nil)
	require.NoError(t, err)
	t.Cleanup(func() {
		rt.Close()
	})
	return rt
}

// NewInstance creates a runtime and an instance for use in tests.
// The instance's repo is a temp directory that will be cleared when the tests finish.
func NewInstance(t TestingT) (*runtime.Runtime, string) {
	rt := New(t)

	inst := &drivers.Instance{
		OLAPConnector: "duckdb",
		RepoConnector: "repo",
		EmbedCatalog:  true,
		Connectors: []*runtimev1.Connector{
			{
				Type:   "file",
				Name:   "repo",
				Config: map[string]string{"dsn": t.TempDir()},
			},
			{
				Type:   "duckdb",
				Name:   "duckdb",
				Config: map[string]string{"dsn": ""},
			},
		},
	}

	err := rt.CreateInstance(context.Background(), inst)
	require.NoError(t, err)
	require.NotEmpty(t, inst.ID)

	err = rt.PutFile(context.Background(), inst.ID, "rill.yaml", strings.NewReader(""), true, false)
	require.NoError(t, err)

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
		OLAPConnector: "duckdb",
		RepoConnector: "repo",
		EmbedCatalog:  true,
		Connectors: []*runtimev1.Connector{
			{
				Type:   "file",
				Name:   "repo",
				Config: map[string]string{"dsn": filepath.Join(currentFile, "..", "testdata", name)},
			},
			{
				Type:   "duckdb",
				Name:   "duckdb",
				Config: map[string]string{"dsn": "?access_mode=read_write"},
			},
		},
	}

	err := rt.CreateInstance(context.Background(), inst)
	require.NoError(t, err)
	require.NotEmpty(t, inst.ID)

	res, err := rt.Reconcile(context.Background(), inst.ID, nil, nil, false, false)
	require.NoError(t, err)
	require.Empty(t, res.Errors)

	return rt, inst.ID
}
