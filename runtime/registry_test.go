package runtime

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
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
				Variables:    map[string]string{"host": "localhost", "allow_host_credentials": "false"},
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
				Variables:    map[string]string{"host": "localhost", "allow_host_credentials": "false"},
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
				Variables:    map[string]string{"host": "localhost", "allow_host_credentials": "false"},
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
				Variables:    map[string]string{"host": "localhost", "allow_host_credentials": "false"},
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
			require.Equal(t, !tt.clearCache, rt.connCache.cache.Contains(inst.ID+inst.OLAPDriver+inst.OLAPDSN))
			require.Equal(t, !tt.clearCache, rt.connCache.cache.Contains(inst.ID+inst.RepoDriver+inst.RepoDSN))
			_, ok := rt.migrationMetaCache.cache[inst.ID]
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
			require.NoError(t, svc.Catalog.CreateEntry(ctx, "default", &drivers.CatalogEntry{
				Name: "data",
				Type: drivers.ObjectTypeTable,
				Object: &runtimev1.Table{
					Name:    "data",
					Managed: true,
				},
			}))

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
			require.False(t, rt.connCache.cache.Contains(inst.ID+inst.OLAPDriver+inst.OLAPDSN))
			require.False(t, rt.connCache.cache.Contains(inst.ID+inst.RepoDriver+inst.RepoDSN))
			_, ok := rt.migrationMetaCache.cache[inst.ID]
			require.False(t, ok)
			_, err = svc.Olap.Execute(context.Background(), &drivers.Statement{Query: "SELECT COUNT(*) FROM rill.migration_version"})
			require.True(t, err != nil)

			// verify db file is dropped if requested
			_, err = os.Stat(dbFile)
			require.Equal(t, tt.dropDB, os.IsNotExist(err))
		})
	}
}

// New returns a runtime configured for use in tests.
func NewTestRunTime(t *testing.T) *Runtime {
	opts := &Options{
		ConnectionCacheSize: 100,
		MetastoreDriver:     "sqlite",
		// Setting a test-specific name ensures a unique connection when "cache=shared" is enabled.
		// "cache=shared" is needed to prevent threading problems.
		MetastoreDSN:   fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name()),
		QueryCacheSize: 10000,
	}
	rt, err := New(opts, zap.NewNop())
	require.NoError(t, err)

	return rt
}
