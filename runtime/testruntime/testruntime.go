package testruntime

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	goruntime "runtime"

	"github.com/c2h5oh/datasize"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/pkg/email"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	// Load database drivers for testing.
	_ "github.com/rilldata/rill/runtime/drivers/admin"
	_ "github.com/rilldata/rill/runtime/drivers/bigquery"
	_ "github.com/rilldata/rill/runtime/drivers/clickhouse"
	_ "github.com/rilldata/rill/runtime/drivers/druid"
	_ "github.com/rilldata/rill/runtime/drivers/duckdb"
	_ "github.com/rilldata/rill/runtime/drivers/file"
	_ "github.com/rilldata/rill/runtime/drivers/gcs"
	_ "github.com/rilldata/rill/runtime/drivers/github"
	_ "github.com/rilldata/rill/runtime/drivers/https"
	_ "github.com/rilldata/rill/runtime/drivers/postgres"
	_ "github.com/rilldata/rill/runtime/drivers/s3"
	_ "github.com/rilldata/rill/runtime/drivers/sqlite"
	_ "github.com/rilldata/rill/runtime/reconcilers"
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
	opts := &runtime.Options{
		MetastoreConnector: "metastore",
		SystemConnectors: []*runtimev1.Connector{
			{
				Type: "sqlite",
				Name: "metastore",
				// Setting a test-specific name ensures a unique connection when "cache=shared" is enabled.
				// "cache=shared" is needed to prevent threading problems.
				Config: map[string]string{"dsn": fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())},
			},
		},
		ConnectionCacheSize:          100,
		QueryCacheSizeBytes:          int64(datasize.MB * 100),
		SecurityEngineCacheSize:      100,
		ControllerLogBufferCapacity:  10000,
		ControllerLogBufferSizeBytes: int64(datasize.MB * 16),
		AllowHostAccess:              true,
	}

	logger := zap.NewNop()
	// nolint
	// logger, err := zap.NewDevelopment()
	// require.NoError(t, err)

	rt, err := runtime.New(context.Background(), opts, logger, activity.NewNoopClient(), email.New(email.NewNoopSender()))
	require.NoError(t, err)
	t.Cleanup(func() { rt.Close() })

	return rt
}

// InstanceOptions enables configuration of the instance options that are configurable in tests.
type InstanceOptions struct {
	Files                        map[string]string
	Variables                    map[string]string
	WatchRepo                    bool
	StageChanges                 bool
	ModelDefaultMaterialize      bool
	ModelMaterializeDelaySeconds uint32
}

// NewInstanceWithOptions creates a runtime and an instance for use in tests.
// The instance's repo is a temp directory that will be cleared when the tests finish.
func NewInstanceWithOptions(t TestingT, opts InstanceOptions) (*runtime.Runtime, string) {
	rt := New(t)

	olapDriver := os.Getenv("OLAP_DRIVER")
	if olapDriver == "" {
		olapDriver = "duckdb"
	}
	olapDSN := os.Getenv("OLAP_DSN")

	tmpDir := t.TempDir()
	inst := &drivers.Instance{
		OLAPConnector:    olapDriver,
		RepoConnector:    "repo",
		CatalogConnector: "catalog",
		Connectors: []*runtimev1.Connector{
			{
				Type:   "file",
				Name:   "repo",
				Config: map[string]string{"dsn": tmpDir},
			},
			{
				Type:   olapDriver,
				Name:   olapDriver,
				Config: map[string]string{"dsn": olapDSN},
			},
			{
				Type: "sqlite",
				Name: "catalog",
				// Setting a test-specific name ensures a unique connection when "cache=shared" is enabled.
				// "cache=shared" is needed to prevent threading problems.
				Config: map[string]string{"dsn": fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())},
			},
		},
		Variables:                    opts.Variables,
		WatchRepo:                    opts.WatchRepo,
		StageChanges:                 opts.StageChanges,
		ModelDefaultMaterialize:      opts.ModelDefaultMaterialize,
		ModelMaterializeDelaySeconds: opts.ModelMaterializeDelaySeconds,
	}

	for path, data := range opts.Files {
		abs := filepath.Join(tmpDir, path)
		require.NoError(t, os.MkdirAll(filepath.Dir(abs), os.ModePerm))
		require.NoError(t, os.WriteFile(abs, []byte(data), 0o644))
	}

	err := rt.CreateInstance(context.Background(), inst)
	require.NoError(t, err)
	require.NotEmpty(t, inst.ID)

	ctrl, err := rt.Controller(context.Background(), inst.ID)
	require.NoError(t, err)

	_, err = ctrl.Get(context.Background(), runtime.GlobalProjectParserName, false)
	require.NoError(t, err)

	err = ctrl.WaitUntilIdle(context.Background(), opts.WatchRepo)
	require.NoError(t, err)

	return rt, inst.ID
}

// NewInstance is a convenience wrapper around NewInstanceWithOptions, using defaults sensible for most tests.
func NewInstance(t TestingT) (*runtime.Runtime, string) {
	return NewInstanceWithOptions(t, InstanceOptions{
		Files: map[string]string{"rill.yaml": ""},
	})
}

// NewInstanceWithModel creates a runtime and an instance for use in tests.
// The passed model name and SQL SELECT statement will be loaded into the instance.
func NewInstanceWithModel(t TestingT, name, sql string) (*runtime.Runtime, string) {
	path := filepath.Join("models", name+".sql")
	return NewInstanceWithOptions(t, InstanceOptions{
		Files: map[string]string{
			"rill.yaml": "",
			path:        sql,
		},
	})
}

// NewInstanceForProject creates a runtime and an instance for use in tests.
// The passed name should match a test project in the testdata folder.
// You should not do mutable repo operations on the returned instance.
func NewInstanceForProject(t TestingT, name string) (*runtime.Runtime, string) {
	rt := New(t)

	_, currentFile, _, _ := goruntime.Caller(0)
	projectPath := filepath.Join(currentFile, "..", "testdata", name)

	inst := &drivers.Instance{
		OLAPConnector:    "duckdb",
		RepoConnector:    "repo",
		CatalogConnector: "catalog",
		Connectors: []*runtimev1.Connector{
			{
				Type:   "file",
				Name:   "repo",
				Config: map[string]string{"dsn": projectPath},
			},
			{
				Type:   "duckdb",
				Name:   "duckdb",
				Config: map[string]string{"dsn": ""},
			},
			{
				Type: "sqlite",
				Name: "catalog",
				// Setting a test-specific name ensures a unique connection when "cache=shared" is enabled.
				// "cache=shared" is needed to prevent threading problems.
				Config: map[string]string{"dsn": fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())},
			},
		},
		EmbedCatalog: true,
	}

	err := rt.CreateInstance(context.Background(), inst)
	require.NoError(t, err)
	require.NotEmpty(t, inst.ID)

	ctrl, err := rt.Controller(context.Background(), inst.ID)
	require.NoError(t, err)

	_, err = ctrl.Get(context.Background(), runtime.GlobalProjectParserName, false)
	require.NoError(t, err)

	err = ctrl.WaitUntilIdle(context.Background(), false)
	require.NoError(t, err)

	return rt, inst.ID
}
