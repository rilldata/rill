package runtime

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/c2h5oh/datasize"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/compilers/rillv1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestRuntime_EditInstance(t *testing.T) {
	repodsn := t.TempDir()
	rt := NewTestRunTime(t)
	tests := []struct {
		name       string
		inst       *drivers.Instance
		wantErr    bool
		savedInst  *drivers.Instance
		clearCache bool
	}{
		{
			name: "edit env",
			inst: &drivers.Instance{
				OLAPDriver:   "duckdb",
				OLAPDSN:      "",
				RepoDriver:   "file",
				RepoDSN:      repodsn,
				EmbedCatalog: true,
				Variables:    map[string]string{"host": "localhost"},
			},
			savedInst: &drivers.Instance{
				OLAPDriver:   "duckdb",
				OLAPDSN:      "",
				RepoDriver:   "file",
				EmbedCatalog: true,
				Variables:    map[string]string{"host": "localhost", "allow_host_access": "true"},
			},
		},
		{
			name: "edit env and embed catalog",
			inst: &drivers.Instance{
				OLAPDriver:   "duckdb",
				OLAPDSN:      "",
				RepoDriver:   "file",
				RepoDSN:      repodsn,
				EmbedCatalog: false,
				Variables:    map[string]string{"host": "localhost"},
			},
			savedInst: &drivers.Instance{
				OLAPDriver:   "duckdb",
				OLAPDSN:      "",
				RepoDriver:   "file",
				EmbedCatalog: false,
				Variables:    map[string]string{"host": "localhost", "allow_host_access": "true"},
			},
		},
		{
			name: "edit olap dsn",
			inst: &drivers.Instance{
				OLAPDriver:   "duckdb",
				OLAPDSN:      "?access_mode=read_write",
				RepoDriver:   "file",
				RepoDSN:      repodsn,
				EmbedCatalog: false,
				Variables:    map[string]string{"host": "localhost"},
			},
			savedInst: &drivers.Instance{
				OLAPDriver:   "duckdb",
				OLAPDSN:      "?access_mode=read_write",
				RepoDriver:   "file",
				EmbedCatalog: false,
				Variables:    map[string]string{"host": "localhost", "allow_host_access": "true"},
			},
			clearCache: true,
		},
		{
			name: "edit repo dsn",
			inst: &drivers.Instance{
				OLAPDriver:   "duckdb",
				OLAPDSN:      "",
				RepoDriver:   "file",
				RepoDSN:      t.TempDir(),
				EmbedCatalog: false,
				Variables:    map[string]string{"host": "localhost"},
			},
			savedInst: &drivers.Instance{
				OLAPDriver:   "duckdb",
				OLAPDSN:      "",
				RepoDriver:   "file",
				EmbedCatalog: false,
				Variables:    map[string]string{"host": "localhost", "allow_host_access": "true"},
			},
			clearCache: true,
		},
		{
			name: "invalid olap driver",
			inst: &drivers.Instance{
				OLAPDriver:   "file",
				OLAPDSN:      "",
				RepoDriver:   "file",
				RepoDSN:      t.TempDir(),
				EmbedCatalog: false,
				Variables:    map[string]string{"host": "localhost"},
			},
			wantErr: true,
		},
		{
			name: "invalid repo driver",
			inst: &drivers.Instance{
				OLAPDriver:   "duckdb",
				OLAPDSN:      "",
				RepoDriver:   "duckdb",
				RepoDSN:      t.TempDir(),
				EmbedCatalog: false,
				Variables:    map[string]string{"host": "localhost"},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			//create instance
			inst := &drivers.Instance{
				OLAPDriver:   "duckdb",
				OLAPDSN:      "",
				RepoDriver:   "file",
				RepoDSN:      repodsn,
				EmbedCatalog: true,
			}
			require.NoError(t, rt.CreateInstance(context.Background(), inst))
			// load all caches
			svc, err := rt.NewCatalogService(ctx, inst.ID)
			require.NoError(t, err)

			// edit instance
			tt.inst.ID = inst.ID
			err = rt.EditInstance(ctx, tt.inst)
			if (err != nil) != tt.wantErr {
				t.Errorf("Runtime.EditInstance() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}

			// verify db instances are correctly updated
			newInst, err := rt.FindInstance(ctx, inst.ID)
			require.NoError(t, err)
			require.Equal(t, inst.ID, newInst.ID)
			require.Equal(t, tt.savedInst.OLAPDriver, newInst.OLAPDriver)
			require.Equal(t, tt.savedInst.OLAPDSN, newInst.OLAPDSN)
			require.Equal(t, tt.savedInst.RepoDriver, newInst.RepoDriver)
			require.Equal(t, tt.inst.RepoDSN, newInst.RepoDSN)
			require.Equal(t, tt.savedInst.EmbedCatalog, newInst.EmbedCatalog)
			require.Greater(t, time.Since(newInst.CreatedOn), time.Since(newInst.UpdatedOn))
			require.Equal(t, tt.savedInst.Variables, newInst.Variables)

			// verify older olap connection is closed and cache updated if olap changed
			_, ok := rt.connCache.cache[inst.ID+inst.OLAPDriver+generateKey(variables("olap", map[string]string{"dsn": ""}, inst.ResolveVariables()))]
			require.Equal(t, !tt.clearCache, ok)
			_, ok = rt.connCache.cache[inst.ID+inst.RepoDriver+generateKey(variables("repo", map[string]string{"dsn": repodsn}, inst.ResolveVariables()))]
			require.Equal(t, !tt.clearCache, ok)
			_, ok = rt.migrationMetaCache.cache.Get(inst.ID)
			require.Equal(t, !tt.clearCache, ok)
			_, err = svc.Olap.Execute(context.Background(), &drivers.Statement{Query: "SELECT COUNT(*) FROM rill.migration_version"})
			require.Equal(t, tt.clearCache, err != nil)
		})
	}
}

func TestRuntime_DeleteInstance(t *testing.T) {
	repodsn := t.TempDir()
	rt := NewTestRunTime(t)
	tests := []struct {
		name       string
		instanceID string
		dropDB     bool
		wantErr    bool
	}{
		{"delete valid no drop", "default", false, false},
		{"delete valid drop", "default", true, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// create test data
			ctx := context.Background()
			dbFile := filepath.Join(t.TempDir(), "test.db")
			inst := &drivers.Instance{
				ID:           "default",
				OLAPDriver:   "duckdb",
				OLAPDSN:      dbFile,
				RepoDriver:   "file",
				RepoDSN:      repodsn,
				EmbedCatalog: true,
			}
			require.NoError(t, rt.CreateInstance(context.Background(), inst))
			// load all caches
			svc, err := rt.NewCatalogService(ctx, inst.ID)
			require.NoError(t, err)

			// ingest some data
			require.NoError(t, svc.Olap.Exec(ctx, &drivers.Statement{Query: "CREATE TABLE data(id INTEGER, name VARCHAR)"}))
			require.NoError(t, svc.Olap.Exec(ctx, &drivers.Statement{Query: "INSERT INTO data VALUES (1, 'Mark'), (2, 'Hannes')"}))
			require.NoError(t, svc.Catalog.CreateEntry(ctx, "", &drivers.CatalogEntry{
				Name: "data",
				Type: drivers.ObjectTypeTable,
				Object: &runtimev1.Table{
					Name:    "data",
					Managed: true,
				},
			}))
			require.ErrorContains(t, svc.Catalog.CreateEntry(ctx, "", &drivers.CatalogEntry{
				Name: "data",
				Type: drivers.ObjectTypeModel,
				Object: &runtimev1.Model{
					Name: "data",
				},
			}), "catalog entry with name \"data\" already exists")

			// delete instance
			err = rt.DeleteInstance(ctx, tt.instanceID, tt.dropDB)
			if (err != nil) != tt.wantErr {
				t.Errorf("Runtime.DeleteInstance() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}

			// verify db is correctly cleared
			_, err = rt.FindInstance(ctx, inst.ID)
			require.Error(t, err)

			// verify older olap connection is closed and cache updated
			require.False(t, rt.connCache.lruCache.Contains(inst.ID+inst.OLAPDriver+fmt.Sprintf("dsn:%s ", inst.OLAPDSN)))
			require.False(t, rt.connCache.lruCache.Contains(inst.ID+inst.RepoDriver+fmt.Sprintf("dsn:%s ", inst.RepoDSN)))
			_, ok := rt.migrationMetaCache.cache.Get(inst.ID)
			require.False(t, ok)
			_, err = svc.Olap.Execute(context.Background(), &drivers.Statement{Query: "SELECT COUNT(*) FROM rill.migration_version"})
			require.True(t, err != nil)

			// verify db file is dropped if requested
			_, err = os.Stat(dbFile)
			require.Equal(t, tt.dropDB, os.IsNotExist(err))
		})
	}
}

func TestRuntime_DeleteInstance_DropCorrupted(t *testing.T) {
	// We require the ability to delete instances and drop database files created with old versions of DuckDB, which can no longer be opened.

	// Prepare
	ctx := context.Background()
	rt := NewTestRunTime(t)
	dbpath := filepath.Join(t.TempDir(), "test.db")

	// Create instance
	inst := &drivers.Instance{
		ID:           "default",
		OLAPDriver:   "duckdb",
		OLAPDSN:      dbpath,
		RepoDriver:   "file",
		RepoDSN:      t.TempDir(),
		EmbedCatalog: true,
	}
	err := rt.CreateInstance(context.Background(), inst)
	require.NoError(t, err)

	// Put some data into it to create a .db file on disk
	olap, release, err := rt.OLAP(ctx, inst.ID)
	require.NoError(t, err)
	defer release()
	err = olap.Exec(ctx, &drivers.Statement{Query: "CREATE TABLE data(id INTEGER, name VARCHAR)"})
	require.NoError(t, err)

	// Close OLAP connection
	c, _, _ := rt.opts.OLAPDef(inst.OLAPDSN)
	evicted := rt.connCache.evict(ctx, inst.ID, inst.OLAPDriver, variables("olap", c.Defaults, inst.ResolveVariables()))
	require.True(t, evicted)

	// Corrupt database file
	err = os.WriteFile(dbpath, []byte("corrupted"), 0644)
	require.NoError(t, err)

	// Check we can't open it anymore
	_, _, err = rt.OLAP(ctx, inst.ID)
	require.Error(t, err)
	require.FileExists(t, dbpath)

	// Delete instance and check it still drops the .db file
	err = rt.DeleteInstance(ctx, inst.ID, true)
	require.NoError(t, err)
	require.NoFileExists(t, dbpath)
}

// New returns a runtime configured for use in tests.
func NewTestRunTime(t *testing.T) *Runtime {
	globalConnectors := []*rillv1.ConnectorDef{
		{
			Type: "sqlite",
			Name: "metastore",
			// Setting a test-specific name ensures a unique connection when "cache=shared" is enabled.
			// "cache=shared" is needed to prevent threading problems.
			Defaults: map[string]string{"dsn": "file:rill?mode=memory&cache=shared"},
		},
	}
	privateConnectors := []*rillv1.ConnectorDef{
		{
			Type: "file",
			Name: "repo",
		},
		{
			Name:     "olap",
			Type:     "duckdb",
			Defaults: map[string]string{"dsn": ""},
		},
	}

	opts := &Options{
		ConnectionCacheSize: 100,
		QueryCacheSizeBytes: int64(datasize.MB) * 100,
		AllowHostAccess:     true,
		GlobalDrivers:       globalConnectors,
		PrivateDrivers:      privateConnectors,
	}
	rt, err := New(opts, zap.NewNop())
	t.Cleanup(func() {
		rt.Close()
	})
	require.NoError(t, err)

	return rt
}
