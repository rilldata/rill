package runtime

import (
	"context"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/c2h5oh/datasize"
	"github.com/duckdb/duckdb-go/v2"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/pkg/email"
	"github.com/rilldata/rill/runtime/storage"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestRuntime_EditInstance(t *testing.T) {
	repodsn := t.TempDir()
	newRepodsn := t.TempDir()
	tests := []struct {
		name      string
		inst      *drivers.Instance
		wantErr   bool
		savedInst *drivers.Instance
	}{
		{
			name: "edit env",
			inst: &drivers.Instance{
				Environment:      "test",
				OLAPConnector:    "duckdb",
				RepoConnector:    "repo",
				CatalogConnector: "catalog",
				Variables:        map[string]string{"connector.s3.region": "us-east-1"},
				Connectors: []*runtimev1.Connector{
					{
						Type:   "file",
						Name:   "repo",
						Config: must(structpb.NewStruct(map[string]any{"dsn": repodsn})),
					},
					{
						Type:   "duckdb",
						Name:   "duckdb",
						Config: must(structpb.NewStruct(map[string]any{"dsn": ":memory:"})),
					},
					{
						Type:   "sqlite",
						Name:   "catalog",
						Config: must(structpb.NewStruct(map[string]any{"dsn": ":memory:"})),
					},
				},
			},
			savedInst: &drivers.Instance{
				Environment:      "test",
				OLAPConnector:    "duckdb",
				RepoConnector:    "repo",
				CatalogConnector: "catalog",
				Variables:        map[string]string{"connector.s3.region": "us-east-1"},
				Connectors: []*runtimev1.Connector{
					{
						Type:   "file",
						Name:   "repo",
						Config: must(structpb.NewStruct(map[string]any{"dsn": repodsn})),
					},
					{
						Type:   "duckdb",
						Name:   "duckdb",
						Config: must(structpb.NewStruct(map[string]any{"dsn": ":memory:"})),
					},
					{
						Type:   "sqlite",
						Name:   "catalog",
						Config: must(structpb.NewStruct(map[string]any{"dsn": ":memory:"})),
					},
				},
			},
		},
		{
			name: "edit drivers",
			inst: &drivers.Instance{
				Environment:      "test",
				OLAPConnector:    "olap1",
				RepoConnector:    "repo1",
				CatalogConnector: "catalog1",
				Variables:        map[string]string{"host": "localhost"},
				Connectors: []*runtimev1.Connector{
					{
						Type:   "file",
						Name:   "repo1",
						Config: must(structpb.NewStruct(map[string]any{"dsn": repodsn})),
					},
					{
						Type:   "duckdb",
						Name:   "olap1",
						Config: must(structpb.NewStruct(map[string]any{"dsn": ":memory:"})),
					},
					{
						Type:   "sqlite",
						Name:   "catalog1",
						Config: must(structpb.NewStruct(map[string]any{"dsn": ":memory:"})),
					},
				},
			},
			savedInst: &drivers.Instance{
				Environment:      "test",
				OLAPConnector:    "olap1",
				RepoConnector:    "repo1",
				CatalogConnector: "catalog1",
				Variables:        map[string]string{"host": "localhost"},
				Connectors: []*runtimev1.Connector{
					{
						Type:   "file",
						Name:   "repo1",
						Config: must(structpb.NewStruct(map[string]any{"dsn": repodsn})),
					},
					{
						Type:   "duckdb",
						Name:   "olap1",
						Config: must(structpb.NewStruct(map[string]any{"dsn": ":memory:"})),
					},
					{
						Type:   "sqlite",
						Name:   "catalog1",
						Config: must(structpb.NewStruct(map[string]any{"dsn": ":memory:"})),
					},
				},
			},
		},
		{
			name: "edit olap dsn",
			inst: &drivers.Instance{
				Environment:      "test",
				OLAPConnector:    "duckdb",
				RepoConnector:    "repo",
				CatalogConnector: "catalog",
				Variables:        map[string]string{"host": "localhost"},
				Connectors: []*runtimev1.Connector{
					{
						Type:   "file",
						Name:   "repo",
						Config: must(structpb.NewStruct(map[string]any{"dsn": repodsn})),
					},
					{
						Type:   "duckdb",
						Name:   "duckdb",
						Config: must(structpb.NewStruct(map[string]any{"dsn": ":memory:?access_mode=read_write"})),
					},
					{
						Type:   "sqlite",
						Name:   "catalog",
						Config: must(structpb.NewStruct(map[string]any{"dsn": ":memory:"})),
					},
				},
			},
			savedInst: &drivers.Instance{
				Environment:      "test",
				OLAPConnector:    "duckdb",
				RepoConnector:    "repo",
				CatalogConnector: "catalog",
				Variables:        map[string]string{"host": "localhost"},
				Connectors: []*runtimev1.Connector{
					{
						Type:   "file",
						Name:   "repo",
						Config: must(structpb.NewStruct(map[string]any{"dsn": repodsn})),
					},
					{
						Type:   "duckdb",
						Name:   "duckdb",
						Config: must(structpb.NewStruct(map[string]any{"dsn": ":memory:?access_mode=read_write"})),
					},
					{
						Type:   "sqlite",
						Name:   "catalog",
						Config: must(structpb.NewStruct(map[string]any{"dsn": ":memory:"})),
					},
				},
			},
		},
		{
			name: "edit repo dsn",
			inst: &drivers.Instance{
				Environment:      "test",
				OLAPConnector:    "duckdb",
				RepoConnector:    "repo",
				CatalogConnector: "catalog",
				Variables:        map[string]string{"host": "localhost"},
				Connectors: []*runtimev1.Connector{
					{
						Type:   "file",
						Name:   "repo",
						Config: must(structpb.NewStruct(map[string]any{"dsn": newRepodsn})),
					},
					{
						Type:   "duckdb",
						Name:   "duckdb",
						Config: must(structpb.NewStruct(map[string]any{"dsn": ":memory:"})),
					},
					{
						Type:   "sqlite",
						Name:   "catalog",
						Config: must(structpb.NewStruct(map[string]any{"dsn": ":memory:"})),
					},
				},
			},
			savedInst: &drivers.Instance{
				Environment:      "test",
				OLAPConnector:    "duckdb",
				RepoConnector:    "repo",
				CatalogConnector: "catalog",
				Variables:        map[string]string{"host": "localhost"},
				Connectors: []*runtimev1.Connector{
					{
						Type:   "file",
						Name:   "repo",
						Config: must(structpb.NewStruct(map[string]any{"dsn": newRepodsn})),
					},
					{
						Type:   "duckdb",
						Name:   "duckdb",
						Config: must(structpb.NewStruct(map[string]any{"dsn": ":memory:"})),
					},
					{
						Type:   "sqlite",
						Name:   "catalog",
						Config: must(structpb.NewStruct(map[string]any{"dsn": ":memory:"})),
					},
				},
			},
		},
		{
			name: "edit annotations",
			inst: &drivers.Instance{
				Environment:      "test",
				OLAPConnector:    "duckdb",
				RepoConnector:    "repo",
				CatalogConnector: "catalog",
				Variables:        map[string]string{"host": "localhost"},
				Connectors: []*runtimev1.Connector{
					{
						Type:   "file",
						Name:   "repo",
						Config: must(structpb.NewStruct(map[string]any{"dsn": newRepodsn})),
					},
					{
						Type:   "duckdb",
						Name:   "duckdb",
						Config: must(structpb.NewStruct(map[string]any{"dsn": ":memory:"})),
					},
					{
						Type:   "sqlite",
						Name:   "catalog",
						Config: must(structpb.NewStruct(map[string]any{"dsn": ":memory:"})),
					},
				},
				Annotations: map[string]string{
					"organization_name": "org_name",
				},
			},
			savedInst: &drivers.Instance{
				Environment:      "test",
				OLAPConnector:    "duckdb",
				RepoConnector:    "repo",
				CatalogConnector: "catalog",
				Variables:        map[string]string{"host": "localhost"},
				Connectors: []*runtimev1.Connector{
					{
						Type:   "file",
						Name:   "repo",
						Config: must(structpb.NewStruct(map[string]any{"dsn": newRepodsn})),
					},
					{
						Type:   "duckdb",
						Name:   "duckdb",
						Config: must(structpb.NewStruct(map[string]any{"dsn": ":memory:"})),
					},
					{
						Type:   "sqlite",
						Name:   "catalog",
						Config: must(structpb.NewStruct(map[string]any{"dsn": ":memory:"})),
					},
				},
				Annotations: map[string]string{
					"organization_name": "org_name",
				},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			rt := newTestRuntime(t)
			defer rt.Close()
			ctx := context.Background()

			// Create instance
			inst := &drivers.Instance{
				Environment:   "test",
				OLAPConnector: "duckdb",
				RepoConnector: "repo",
				Connectors: []*runtimev1.Connector{
					{
						Type:   "file",
						Name:   "repo",
						Config: must(structpb.NewStruct(map[string]any{"dsn": repodsn})),
					},
					{
						Type:   "duckdb",
						Name:   "duckdb",
						Config: must(structpb.NewStruct(map[string]any{"dsn": ":memory:"})),
					},
				},
			}
			require.NoError(t, rt.CreateInstance(context.Background(), inst))
			_, err := rt.Controller(ctx, inst.ID)
			require.NoError(t, err)

			// Acquire OLAP (to make sure it's opened)
			firstOlap, release, err := rt.OLAP(ctx, inst.ID, "")
			require.NoError(t, err)
			release()

			// Edit instance
			tt.inst.ID = inst.ID
			err = rt.EditInstance(ctx, tt.inst, true)
			if (err != nil) != tt.wantErr {
				t.Errorf("Runtime.EditInstance() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}

			// Wait for controller restart
			time.Sleep(2 * time.Second)
			_, err = rt.Controller(ctx, inst.ID)
			require.NoError(t, err)

			// Verify db instances are correctly updated
			newInst, err := rt.registryCache.store.FindInstance(ctx, inst.ID)
			require.NoError(t, err)
			require.Equal(t, inst.ID, newInst.ID)
			require.Equal(t, tt.savedInst.OLAPConnector, newInst.OLAPConnector)
			require.True(t, connectorsEqual(tt.savedInst.Connectors[0], newInst.Connectors[0]))
			require.True(t, connectorsEqual(tt.savedInst.Connectors[1], newInst.Connectors[1]))
			require.Equal(t, tt.savedInst.RepoConnector, newInst.RepoConnector)
			require.Equal(t, tt.savedInst.CatalogConnector, newInst.CatalogConnector)
			require.Greater(t, time.Since(newInst.CreatedOn), time.Since(newInst.UpdatedOn))
			require.True(t, time.Since(newInst.UpdatedOn) < 10*time.Second)
			require.Equal(t, tt.savedInst.Variables, newInst.Variables)

			// Verify new olap connection is opened
			olap, release, err := rt.OLAP(ctx, inst.ID, "")
			require.NoError(t, err)
			defer release()

			// Verify new olap is not the old one
			require.NotEqual(t, firstOlap, olap)
		})
	}
}

func TestRuntime_DeleteInstance(t *testing.T) {
	repodsn := t.TempDir()
	rt := newTestRuntime(t)
	tests := []struct {
		name    string
		wantErr bool
	}{
		{"delete valid drop", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// create test data
			ctx := context.Background()
			inst := &drivers.Instance{
				Environment:   "test",
				OLAPConnector: "duckdb",
				RepoConnector: "repo",
				Connectors: []*runtimev1.Connector{
					{
						Type:   "file",
						Name:   "repo",
						Config: must(structpb.NewStruct(map[string]any{"dsn": repodsn})),
					},
					{
						Type:   "duckdb",
						Name:   "duckdb",
						Config: must(structpb.NewStruct(map[string]any{})),
					},
				},
			}
			require.NoError(t, rt.CreateInstance(context.Background(), inst))
			_, err := rt.Controller(ctx, inst.ID)
			require.NoError(t, err)
			dbFile := filepath.Join(t.TempDir(), inst.ID, "duckdb", "main.db")

			// Acquire OLAP
			olap, release, err := rt.OLAP(ctx, inst.ID, "")
			require.NoError(t, err)
			defer release()

			// ingest some data
			require.NoError(t, olap.Exec(ctx, &drivers.Statement{Query: "CREATE TABLE data(id INTEGER, name VARCHAR)"}))
			require.NoError(t, olap.Exec(ctx, &drivers.Statement{Query: "INSERT INTO data VALUES (1, 'Mark'), (2, 'Hannes')"}))

			// delete instance
			err = rt.DeleteInstance(ctx, inst.ID)
			if (err != nil) != tt.wantErr {
				t.Errorf("Runtime.DeleteInstance() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}

			// verify db is correctly cleared
			_, err = rt.Instance(ctx, inst.ID)
			require.Error(t, err)

			// verify older olap connection is closed and cache updated
			// require.False(t, rt.connCache.lru.Contains(inst.ID+"duckdb"+fmt.Sprintf("dsn:%s ", dbFile)))
			// require.False(t, rt.connCache.lru.Contains(inst.ID+"file"+fmt.Sprintf("dsn:%s ", repodsn)))
			time.Sleep(2 * time.Second)
			err = olap.Exec(context.Background(), &drivers.Statement{Query: "SELECT COUNT(*) FROM rill.migration_version"})
			require.True(t, err != nil)

			// verify db file is dropped
			_, err = os.Stat(dbFile)
			require.True(t, os.IsNotExist(err))
		})
	}
}

func TestRuntime_DeleteInstance_DropCorrupted(t *testing.T) {
	// We require the ability to delete instances and drop database files created with old versions of DuckDB, which can no longer be opened.

	// Prepare
	ctx := context.Background()
	rt := newTestRuntime(t)
	// Create instance
	inst := &drivers.Instance{
		Environment:   "test",
		OLAPConnector: "duckdb",
		RepoConnector: "repo",
		Connectors: []*runtimev1.Connector{
			{
				Type:   "file",
				Name:   "repo",
				Config: must(structpb.NewStruct(map[string]any{"dsn": t.TempDir()})),
			},
			{
				Type:   "duckdb",
				Name:   "duckdb",
				Config: must(structpb.NewStruct(map[string]any{})),
			},
		},
	}
	err := rt.CreateInstance(context.Background(), inst)
	require.NoError(t, err)

	dataDir, err := rt.storage.DataDir(inst.ID, "duckdb")
	require.NoError(t, err)
	dbpath := filepath.Join(dataDir, "main.db")

	// Put some data into it to create a .db file on disk
	olap, release, err := rt.OLAP(ctx, inst.ID, "")
	require.NoError(t, err)
	defer release()
	err = olap.Exec(ctx, &drivers.Statement{Query: "CREATE TABLE data(id INTEGER, name VARCHAR)"})
	require.NoError(t, err)

	// Close OLAP connection
	rt.evictInstanceConnections(inst.ID)

	// Corrupt database file
	err = os.WriteFile(dbpath, []byte("corrupted"), 0644)
	require.NoError(t, err)

	// Check we can't open it anymore
	conn, err := duckdb.NewConnector(dbpath, nil)
	require.Error(t, err)
	require.Nil(t, conn)

	// Delete instance and check it still drops the .db file for DuckDB
	err = rt.DeleteInstance(ctx, inst.ID)
	require.NoError(t, err)
	require.NoFileExists(t, dbpath)
}

// New returns a runtime configured for use in tests.
func newTestRuntime(t *testing.T) *Runtime {
	globalConnectors := []*runtimev1.Connector{
		{
			Type: "sqlite",
			Name: "metastore",
			// Setting a test-specific name ensures a unique connection when "cache=shared" is enabled.
			// "cache=shared" is needed to prevent threading problems.
			Config: must(structpb.NewStruct(map[string]any{"dsn": fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())})),
		},
	}

	opts := &Options{
		ConnectionCacheSize:          100,
		MetastoreConnector:           "metastore",
		QueryCacheSizeBytes:          int64(datasize.MB) * 100,
		SecurityEngineCacheSize:      100,
		AllowHostAccess:              true,
		SystemConnectors:             globalConnectors,
		ControllerLogBufferCapacity:  10000,
		ControllerLogBufferSizeBytes: int64(datasize.MB * 16),
	}

	rt, err := New(context.Background(), opts, zap.NewNop(), storage.MustNew(t.TempDir(), nil), activity.NewNoopClient(), email.New(email.NewNoopSender()))
	t.Cleanup(func() {
		rt.Close()
	})
	require.NoError(t, err)

	return rt
}

func connectorsEqual(a, b *runtimev1.Connector) bool {
	if (a != nil) != (b != nil) {
		return false
	}
	return a.Name == b.Name && a.Type == b.Type && maps.Equal(a.Config.AsMap(), b.Config.AsMap())
}

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}
