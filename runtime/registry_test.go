package runtime

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/stretchr/testify/require"
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
				Env:          map[string]string{"host": "localhost"},
			},
			savedInst: &drivers.Instance{
				OLAPDriver:   "duckdb",
				OLAPDSN:      "",
				RepoDriver:   "file",
				EmbedCatalog: true,
				Env:          map[string]string{"host": "localhost", "allow_host_credentials": "false"},
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
				Env:          map[string]string{"host": "localhost"},
			},
			savedInst: &drivers.Instance{
				OLAPDriver:   "duckdb",
				OLAPDSN:      "",
				RepoDriver:   "file",
				EmbedCatalog: false,
				Env:          map[string]string{"host": "localhost", "allow_host_credentials": "false"},
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
				Env:          map[string]string{"host": "localhost"},
			},
			savedInst: &drivers.Instance{
				OLAPDriver:   "duckdb",
				OLAPDSN:      "?access_mode=read_write",
				RepoDriver:   "file",
				EmbedCatalog: false,
				Env:          map[string]string{"host": "localhost", "allow_host_credentials": "false"},
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
				Env:          map[string]string{"host": "localhost"},
			},
			savedInst: &drivers.Instance{
				OLAPDriver:   "duckdb",
				OLAPDSN:      "",
				RepoDriver:   "file",
				EmbedCatalog: false,
				Env:          map[string]string{"host": "localhost", "allow_host_credentials": "false"},
			},
			clearCache: true,
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
			svc, err := rt.catalogCache.get(ctx, rt, inst.ID)
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
			require.Equal(t, tt.savedInst.Env, newInst.Env)

			// verify older olap connection is closed and cache updated if olap changed
			require.Equal(t, !tt.clearCache, rt.connCache.cache.Contains(inst.ID+inst.OLAPDriver+inst.OLAPDSN))
			require.Equal(t, !tt.clearCache, rt.connCache.cache.Contains(inst.ID+inst.RepoDriver+inst.RepoDSN))
			_, ok := rt.catalogCache.cache[inst.ID]
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
		{"delete invalid", "default1", true, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// create test data
			ctx := context.Background()
			inst := &drivers.Instance{
				ID:           "default",
				OLAPDriver:   "duckdb",
				OLAPDSN:      filepath.Join(t.TempDir(), "test.db"),
				RepoDriver:   "file",
				RepoDSN:      repodsn,
				EmbedCatalog: true,
			}
			require.NoError(t, rt.CreateInstance(context.Background(), inst))
			// load all caches
			svc, err := rt.catalogCache.get(ctx, rt, inst.ID)
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
			_, ok := rt.catalogCache.cache[inst.ID]
			require.False(t, ok)
			_, err = svc.Olap.Execute(context.Background(), &drivers.Statement{Query: "SELECT COUNT(*) FROM rill.migration_version"})
			require.True(t, err != nil)

			if tt.dropDB {
				// get a new connection to verify db schema and ingested data is dropped
				olap, err := drivers.Open(inst.OLAPDriver, inst.OLAPDSN, rt.logger)
				require.NoError(t, err)
				olapStore, _ := olap.OLAPStore()

				// verify rillschema is dropped
				rows, err := olapStore.Execute(ctx, &drivers.Statement{Query: "SELECT count(*) FROM information_schema.schemata WHERE schema_name = 'rill'"})
				require.NoError(t, err)
				var count int
				require.True(t, rows.Next())
				require.NoError(t, rows.Scan(&count))
				require.Equal(t, 0, count)
				require.NoError(t, rows.Close())

				// verify ingested data is dropped
				rows, err = olapStore.Execute(ctx, &drivers.Statement{Query: "SELECT count(*) FROM information_schema.tables WHERE table_name = 'data'"})
				require.NoError(t, err)
				require.True(t, rows.Next())
				require.NoError(t, rows.Scan(&count))
				require.Equal(t, 0, count)
				require.NoError(t, rows.Close())

			}
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
	rt, err := New(opts, nil)
	require.NoError(t, err)

	return rt
}
