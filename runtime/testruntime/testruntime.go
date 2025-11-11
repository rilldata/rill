package testruntime

import (
	"context"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	goruntime "runtime"
	"slices"
	"strconv"
	"testing"

	"github.com/c2h5oh/datasize"
	"github.com/joho/godotenv"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/drivers/clickhouse/testclickhouse"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/pkg/email"
	"github.com/rilldata/rill/runtime/storage"
	"github.com/rilldata/rill/runtime/testruntime/testmode"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"

	// Load database drivers for testing.
	_ "github.com/rilldata/rill/runtime/drivers/admin"
	_ "github.com/rilldata/rill/runtime/drivers/athena"
	_ "github.com/rilldata/rill/runtime/drivers/bigquery"
	_ "github.com/rilldata/rill/runtime/drivers/clickhouse"
	_ "github.com/rilldata/rill/runtime/drivers/druid"
	_ "github.com/rilldata/rill/runtime/drivers/duckdb"
	_ "github.com/rilldata/rill/runtime/drivers/file"
	_ "github.com/rilldata/rill/runtime/drivers/gcs"
	_ "github.com/rilldata/rill/runtime/drivers/https"
	_ "github.com/rilldata/rill/runtime/drivers/mock/ai"
	_ "github.com/rilldata/rill/runtime/drivers/openai"
	_ "github.com/rilldata/rill/runtime/drivers/postgres"
	_ "github.com/rilldata/rill/runtime/drivers/redshift"
	_ "github.com/rilldata/rill/runtime/drivers/s3"
	_ "github.com/rilldata/rill/runtime/drivers/snowflake"
	_ "github.com/rilldata/rill/runtime/drivers/sqlite"
	_ "github.com/rilldata/rill/runtime/reconcilers"
)

// TestingT satisfies both *testing.T and *testing.B.
type TestingT interface {
	Name() string
	TempDir() string
	FailNow()
	SkipNow()
	Errorf(format string, args ...interface{})
	Cleanup(f func())
	Context() context.Context
}

// New returns a runtime configured for use in tests.
func New(t TestingT, allowHostAccess bool) *runtime.Runtime {
	ctx := t.Context()
	opts := &runtime.Options{
		MetastoreConnector: "metastore",
		SystemConnectors: []*runtimev1.Connector{
			{
				Type: "sqlite",
				Name: "metastore",
				// Setting a test-specific name ensures a unique connection when "cache=shared" is enabled.
				// "cache=shared" is needed to prevent threading problems.
				Config: Must(structpb.NewStruct(map[string]any{"dsn": fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())})),
			},
		},
		ConnectionCacheSize:          100,
		QueryCacheSizeBytes:          int64(datasize.MB * 100),
		SecurityEngineCacheSize:      100,
		ControllerLogBufferCapacity:  10000,
		ControllerLogBufferSizeBytes: int64(datasize.MB * 16),
		AllowHostAccess:              allowHostAccess,
	}

	logger := zap.NewNop()
	var err error
	if os.Getenv("DEBUG") == "1" {
		logger, err = zap.NewDevelopment()
		require.NoError(t, err)
	}

	rt, err := runtime.New(ctx, opts, logger, storage.MustNew(t.TempDir(), nil), activity.NewNoopClient(), email.New(email.NewTestSender()))
	require.NoError(t, err)
	t.Cleanup(func() { rt.Close() })

	return rt
}

// InstanceOptions enables configuration of the instance options that are configurable in tests.
type InstanceOptions struct {
	Files             map[string]string
	Variables         map[string]string
	WatchRepo         bool
	StageChanges      bool
	DisableHostAccess bool
	EnableLLM         bool
	TestConnectors    []string
	FrontendURL       string
}

// NewInstanceWithOptions creates a runtime and an instance for use in tests.
// The instance's repo is a temp directory that will be cleared when the tests finish.
func NewInstanceWithOptions(t TestingT, opts InstanceOptions) (*runtime.Runtime, string) {
	rt := New(t, !opts.DisableHostAccess)
	ctx := t.Context()

	olapDriver := os.Getenv("RILL_RUNTIME_TEST_OLAP_DRIVER")
	if olapDriver == "" {
		olapDriver = "duckdb"
	}
	olapDSN := os.Getenv("RILL_RUNTIME_TEST_OLAP_DSN")
	if olapDSN == "" {
		olapDSN = ":memory:"
	}

	vars := make(map[string]string)
	maps.Copy(vars, opts.Variables)
	if vars["rill.stage_changes"] == "" {
		vars["rill.stage_changes"] = strconv.FormatBool(opts.StageChanges)
	}
	if vars["rill.watch_repo"] == "" {
		vars["rill.watch_repo"] = strconv.FormatBool(opts.WatchRepo)
	}

	// Making LLM completions in tests is disabled by default.
	// If enabled, we skip the test in CI (short mode) to prevent running up costs.
	var aiConnector string
	if opts.EnableLLM {
		// Mark AI tests as expensive
		testmode.Expensive(t)

		// Add "openai" to the test connectors if not already present.
		if !slices.Contains(opts.TestConnectors, "openai") {
			opts.TestConnectors = append(opts.TestConnectors, "openai")
		}

		// Set the "openai" test connector as the instance's default AI connector.
		// This enables LLM completions.
		aiConnector = "openai"
	}

	for _, conn := range opts.TestConnectors {
		acquire, ok := Connectors[conn]
		require.True(t, ok, "unknown test connector %q", conn)
		cfg := acquire(t)
		for k, v := range cfg {
			k = fmt.Sprintf("connector.%s.%s", conn, k)
			vars[k] = v
		}
	}

	tmpDir := t.TempDir()
	inst := &drivers.Instance{
		Environment:      "test",
		OLAPConnector:    olapDriver,
		RepoConnector:    "repo",
		AIConnector:      aiConnector,
		CatalogConnector: "catalog",
		Connectors: []*runtimev1.Connector{
			{
				Type:   "file",
				Name:   "repo",
				Config: Must(structpb.NewStruct(map[string]any{"dsn": tmpDir})),
			},
			{
				Type:   olapDriver,
				Name:   olapDriver,
				Config: Must(structpb.NewStruct(map[string]any{"dsn": olapDSN})),
			},
			{
				Type: "sqlite",
				Name: "catalog",
				// Setting a test-specific name ensures a unique connection when "cache=shared" is enabled.
				// "cache=shared" is needed to prevent threading problems.
				Config: Must(structpb.NewStruct(map[string]any{"dsn": fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())})),
			},
		},
		Variables:   vars,
		FrontendURL: opts.FrontendURL,
	}

	if opts.Files == nil {
		opts.Files = make(map[string]string)
	}
	if _, ok := opts.Files["rill.yaml"]; !ok {
		opts.Files["rill.yaml"] = ""
	}

	for path, data := range opts.Files {
		abs := filepath.Join(tmpDir, path)
		require.NoError(t, os.MkdirAll(filepath.Dir(abs), os.ModePerm))
		require.NoError(t, os.WriteFile(abs, []byte(data), 0o644))
	}

	err := rt.CreateInstance(ctx, inst)
	require.NoError(t, err)
	require.NotEmpty(t, inst.ID)

	ctrl, err := rt.Controller(ctx, inst.ID)
	require.NoError(t, err)

	_, err = ctrl.Get(ctx, runtime.GlobalProjectParserName, false)
	require.NoError(t, err)

	err = ctrl.WaitUntilIdle(ctx, opts.WatchRepo)
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
	rt := New(t, true)
	ctx := t.Context()

	_, currentFile, _, _ := goruntime.Caller(0)
	projectPath := filepath.Join(currentFile, "..", "testdata", name)

	olapDriver := os.Getenv("RILL_RUNTIME_TEST_OLAP_DRIVER") // todo: refactor a couple of tests that use envs
	if olapDriver == "" {
		olapDriver = "duckdb"
	}
	olapDSN := os.Getenv("RILL_RUNTIME_TEST_OLAP_DSN")
	if olapDSN == "" {
		olapDSN = ":memory:"
	}

	inst := &drivers.Instance{
		Environment:      "test",
		OLAPConnector:    olapDriver,
		RepoConnector:    "repo",
		CatalogConnector: "catalog",
		Connectors: []*runtimev1.Connector{
			{
				Type:   "file",
				Name:   "repo",
				Config: Must(structpb.NewStruct(map[string]any{"dsn": projectPath})),
			},
			{
				Type:   olapDriver,
				Name:   olapDriver,
				Config: Must(structpb.NewStruct(map[string]any{"dsn": olapDSN})),
			},
			{
				Type: "sqlite",
				Name: "catalog",
				// Setting a test-specific name ensures a unique connection when "cache=shared" is enabled.
				// "cache=shared" is needed to prevent threading problems.
				Config: Must(structpb.NewStruct(map[string]any{"dsn": fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())})),
			},
		},
	}

	err := rt.CreateInstance(ctx, inst)
	require.NoError(t, err)
	require.NotEmpty(t, inst.ID)

	ctrl, err := rt.Controller(ctx, inst.ID)
	require.NoError(t, err)

	_, err = ctrl.Get(ctx, runtime.GlobalProjectParserName, false)
	require.NoError(t, err)

	err = ctrl.WaitUntilIdle(ctx, false)
	require.NoError(t, err)

	return rt, inst.ID
}

func NewInstanceForDruidProject(t *testing.T) (*runtime.Runtime, string, error) {
	_, currentFile, _, _ := goruntime.Caller(0)
	envPath := filepath.Join(currentFile, "..", "..", "..", ".env")
	_, err := os.Stat(envPath)
	if err == nil { // avoid .env in CI environment
		require.NoError(t, godotenv.Load(envPath))
	}
	if os.Getenv("RILL_RUNTIME_DRUID_TEST_DSN") == "" {
		t.Skip("skipping the test without the test instance")
	}

	rt := New(t, true)
	ctx := t.Context()

	_, currentFile, _, _ = goruntime.Caller(0)
	projectPath := filepath.Join(currentFile, "..", "testdata", "ad_bids_druid")
	dsn := os.Getenv("RILL_RUNTIME_DRUID_TEST_DSN")

	inst := &drivers.Instance{
		Environment:      "test",
		OLAPConnector:    "druid",
		RepoConnector:    "repo",
		CatalogConnector: "catalog",
		Connectors: []*runtimev1.Connector{
			{
				Type:   "file",
				Name:   "repo",
				Config: Must(structpb.NewStruct(map[string]any{"dsn": projectPath})),
			},
			{
				Type:   "druid",
				Name:   "druid",
				Config: Must(structpb.NewStruct(map[string]any{"dsn": dsn})),
			},
			{
				Type: "sqlite",
				Name: "catalog",
				// Setting a test-specific name ensures a unique connection when "cache=shared" is enabled.
				// "cache=shared" is needed to prevent threading problems.
				Config: Must(structpb.NewStruct(map[string]any{"dsn": fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())})),
			},
		},
	}

	err = rt.CreateInstance(ctx, inst)
	require.NoError(t, err)
	require.NotEmpty(t, inst.ID)

	ctrl, err := rt.Controller(ctx, inst.ID)
	require.NoError(t, err)

	_, err = ctrl.Get(ctx, runtime.GlobalProjectParserName, false)
	require.NoError(t, err)

	err = ctrl.WaitUntilIdle(ctx, false)
	require.NoError(t, err)

	return rt, inst.ID, nil
}

func NewInstanceWithClickhouseProject(t TestingT, withCluster bool) (*runtime.Runtime, string) {
	dsn, cluster := testclickhouse.StartCluster(t)

	rt := New(t, true)
	ctx := t.Context()

	_, currentFile, _, _ := goruntime.Caller(0)
	projectPath := filepath.Join(currentFile, "..", "testdata", "ad_bids_clickhouse")

	olapConfig := map[string]any{"dsn": dsn, "mode": "readwrite"}
	if withCluster {
		olapConfig["cluster"] = cluster
		olapConfig["log_queries"] = "true"
	}
	inst := &drivers.Instance{
		Environment:      "test",
		OLAPConnector:    "duckdb",
		RepoConnector:    "repo",
		CatalogConnector: "catalog",
		Connectors: []*runtimev1.Connector{
			{
				Type:   "file",
				Name:   "repo",
				Config: Must(structpb.NewStruct(map[string]any{"dsn": projectPath})),
			},
			{
				Type:   "clickhouse",
				Name:   "clickhouse",
				Config: Must(structpb.NewStruct(olapConfig)),
			},
			{
				Type: "sqlite",
				Name: "catalog",
				// Setting a test-specific name ensures a unique connection when "cache=shared" is enabled.
				// "cache=shared" is needed to prevent threading problems.
				Config: Must(structpb.NewStruct(map[string]any{"dsn": fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())})),
			},
		},
		Variables: map[string]string{"rill.stage_changes": "false"},
	}

	err := rt.CreateInstance(ctx, inst)
	require.NoError(t, err)
	require.NotEmpty(t, inst.ID)

	ctrl, err := rt.Controller(ctx, inst.ID)
	require.NoError(t, err)

	_, err = ctrl.Get(ctx, runtime.GlobalProjectParserName, false)
	require.NoError(t, err)

	err = ctrl.WaitUntilIdle(ctx, false)
	require.NoError(t, err)

	return rt, inst.ID
}
