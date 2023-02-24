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
	repoDsn := t.TempDir()
	tests := []struct {
		name            string
		inst            *drivers.Instance
		wantErr         bool
		savedInst       *drivers.Instance
		olapConnChanged bool
		repoConnChanged bool
	}{
		{
			name: "edit env",
			inst: &drivers.Instance{
				ID:           "default",
				OLAPDriver:   "duckdb",
				OLAPDSN:      "",
				RepoDriver:   "file",
				RepoDSN:      repoDsn,
				EmbedCatalog: true,
				Env:          map[string]string{"host": "localhost"},
			},
			savedInst: &drivers.Instance{
				ID:           "default",
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
				ID:           "default",
				OLAPDriver:   "duckdb",
				OLAPDSN:      "",
				RepoDriver:   "file",
				RepoDSN:      repoDsn,
				EmbedCatalog: false,
				Env:          map[string]string{"host": "localhost"},
			},
			savedInst: &drivers.Instance{
				ID:           "default",
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
				ID:           "default",
				OLAPDriver:   "duckdb",
				OLAPDSN:      "?access_mode=read_write",
				RepoDriver:   "file",
				RepoDSN:      repoDsn,
				EmbedCatalog: false,
				Env:          map[string]string{"host": "localhost"},
			},
			savedInst: &drivers.Instance{
				ID:           "default",
				OLAPDriver:   "duckdb",
				OLAPDSN:      "?access_mode=read_write",
				RepoDriver:   "file",
				EmbedCatalog: false,
				Env:          map[string]string{"host": "localhost", "allow_host_credentials": "false"},
			},
			olapConnChanged: true,
		},
		{
			name: "edit repo dsn",
			inst: &drivers.Instance{
				ID:           "default",
				OLAPDriver:   "duckdb",
				OLAPDSN:      "",
				RepoDriver:   "file",
				RepoDSN:      t.TempDir(),
				EmbedCatalog: false,
				Env:          map[string]string{"host": "localhost"},
			},
			savedInst: &drivers.Instance{
				ID:           "default",
				OLAPDriver:   "duckdb",
				OLAPDSN:      "",
				RepoDriver:   "file",
				EmbedCatalog: false,
				Env:          map[string]string{"host": "localhost", "allow_host_credentials": "false"},
			},
			repoConnChanged: true,
		},
		{
			name: "invalid id",
			inst: &drivers.Instance{
				ID:           "default1",
				OLAPDriver:   "duckdb",
				OLAPDSN:      "",
				RepoDriver:   "file",
				RepoDSN:      t.TempDir(),
				EmbedCatalog: false,
				Env:          map[string]string{"host": "localhost"},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			// create instance
			rt, inst := NewInstance(t, repoDsn)
			// get olap connection
			olapConn, err := rt.connCache.get(ctx, inst.ID, inst.OLAPDriver, inst.OLAPDSN)
			require.NoError(t, err)
			olap, ok := olapConn.OLAPStore()
			require.True(t, ok)

			// get repo connection
			_, err = rt.connCache.get(ctx, inst.ID, inst.RepoDriver, inst.RepoDSN)
			require.NoError(t, err)

			// edit instance
			err = rt.EditInstance(ctx, tt.inst)
			if (err != nil) != tt.wantErr {
				t.Errorf("Runtime.EditInstance() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}

			// verify db instances are correctly updated
			dbInst, err := rt.FindInstance(ctx, inst.ID)
			require.NoError(t, err)
			require.Equal(t, tt.savedInst.ID, dbInst.ID)
			require.Equal(t, tt.savedInst.OLAPDriver, dbInst.OLAPDriver)
			require.Equal(t, tt.savedInst.OLAPDSN, dbInst.OLAPDSN)
			require.Equal(t, tt.savedInst.RepoDriver, dbInst.RepoDriver)
			require.Equal(t, tt.inst.RepoDSN, dbInst.RepoDSN)
			require.Equal(t, tt.savedInst.EmbedCatalog, dbInst.EmbedCatalog)
			require.Greater(t, time.Since(dbInst.CreatedOn), time.Since(dbInst.UpdatedOn))
			require.Equal(t, tt.savedInst.Env, dbInst.Env)

			// verify older olap connection is closed and cache updated if olap changed
			_, err = olap.Execute(context.Background(), &drivers.Statement{Query: "SELECT COUNT(*) FROM rill.migration_version"})
			_, ok = rt.connCache.cache.Get(inst.ID + inst.OLAPDriver + inst.OLAPDSN)
			if tt.olapConnChanged {
				require.Error(t, err)
				require.False(t, ok)
			} else {
				require.NoError(t, err)
				require.True(t, ok)
			}

			// verify cache updated if repo changed
			_, ok = rt.connCache.cache.Get(inst.ID + inst.RepoDriver + inst.RepoDSN)
			require.Equal(t, !tt.repoConnChanged, ok)
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

			// get olap connection
			olapConn, err := rt.connCache.get(ctx, inst.ID, inst.OLAPDriver, inst.OLAPDSN)
			require.NoError(t, err)
			olap, ok := olapConn.OLAPStore()
			require.True(t, ok)
			// ingest some data
			require.NoError(t, olap.Exec(ctx, &drivers.Statement{Query: "CREATE TABLE data(id INTEGER, name VARCHAR)"}))
			require.NoError(t, olap.Exec(ctx, &drivers.Statement{Query: "INSERT INTO data VALUES (1, 'Mark'), (2, 'Hannes')"}))
			cat, _ := olapConn.CatalogStore()
			require.NoError(t, cat.CreateEntry(ctx, "default", &drivers.CatalogEntry{
				Name: "data",
				Type: drivers.ObjectTypeTable,
				Object: &runtimev1.Table{
					Name:    "data",
					Managed: true,
				},
			}))

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
			_, err = olap.Execute(context.Background(), &drivers.Statement{Query: "SELECT COUNT(*) FROM rill.migration_version"})
			require.Error(t, err)
			require.False(t, rt.connCache.cache.Contains(inst.ID+inst.OLAPDriver+inst.OLAPDSN))

			// verify repo cache updated
			require.False(t, rt.connCache.cache.Contains(inst.ID+inst.RepoDriver+inst.RepoDSN))

			// verify catalog cache updated
			_, ok = rt.catalogCache.cache[inst.ID]
			require.False(t, ok)

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

// NewInstance creates a runtime and an instance for use in tests.
// The instance's repo is a temp directory that will be cleared when the tests finish.
func NewInstance(t *testing.T, repodsn string) (*Runtime, *drivers.Instance) {
	rt := NewTestRunTime(t)

	inst := &drivers.Instance{
		ID:           "default",
		OLAPDriver:   "duckdb",
		OLAPDSN:      "",
		RepoDriver:   "file",
		RepoDSN:      repodsn,
		EmbedCatalog: true,
	}

	err := rt.CreateInstance(context.Background(), inst)
	require.NoError(t, err)
	require.NotEmpty(t, inst.ID)

	return rt, inst
}
