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
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestRuntime_EditInstance(t *testing.T) {
	repodsn := t.TempDir()
	newRepodsn := t.TempDir()
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
				OLAPDriver:   "olap",
				RepoDriver:   "repo",
				EmbedCatalog: true,
				Variables:    map[string]string{"connectors.s3.region": "us-east-1"},
				Connectors: []*runtimev1.Connector{
					{
						Type:   "file",
						Name:   "repo",
						Config: map[string]string{"dsn": repodsn},
					},
					{
						Type:   "duckdb",
						Name:   "olap",
						Config: map[string]string{"dsn": ""},
					},
				},
			},
			savedInst: &drivers.Instance{
				OLAPDriver:   "olap",
				RepoDriver:   "repo",
				EmbedCatalog: true,
				Variables:    map[string]string{"connectors.s3.region": "us-east-1", "allow_host_access": "true"},
				Connectors: []*runtimev1.Connector{
					{
						Type:   "file",
						Name:   "repo",
						Config: map[string]string{"dsn": repodsn},
					},
					{
						Type:   "duckdb",
						Name:   "olap",
						Config: map[string]string{"dsn": ""},
					},
				},
			},
		},
		{
			name: "edit drivers",
			inst: &drivers.Instance{
				OLAPDriver:   "olap1",
				RepoDriver:   "repo1",
				EmbedCatalog: true,
				Variables:    map[string]string{"host": "localhost"},
				Connectors: []*runtimev1.Connector{
					{
						Type:   "file",
						Name:   "repo1",
						Config: map[string]string{"dsn": repodsn},
					},
					{
						Type:   "duckdb",
						Name:   "olap1",
						Config: map[string]string{"dsn": ""},
					},
				},
			},
			savedInst: &drivers.Instance{
				OLAPDriver:   "olap1",
				RepoDriver:   "repo1",
				EmbedCatalog: true,
				Variables:    map[string]string{"host": "localhost", "allow_host_access": "true"},
				Connectors: []*runtimev1.Connector{
					{
						Type:   "file",
						Name:   "repo1",
						Config: map[string]string{"dsn": repodsn},
					},
					{
						Type:   "duckdb",
						Name:   "olap1",
						Config: map[string]string{"dsn": ""},
					},
				},
			},
			clearCache: true,
		},
		{
			name: "edit env and embed catalog",
			inst: &drivers.Instance{
				OLAPDriver:   "olap",
				RepoDriver:   "repo",
				EmbedCatalog: false,
				Variables:    map[string]string{"connectors.s3.region": "us-east-1"},
				Connectors: []*runtimev1.Connector{
					{
						Type:   "file",
						Name:   "repo",
						Config: map[string]string{"dsn": repodsn},
					},
					{
						Type:   "duckdb",
						Name:   "olap",
						Config: map[string]string{"dsn": ""},
					},
				},
			},
			savedInst: &drivers.Instance{
				OLAPDriver:   "olap",
				RepoDriver:   "repo",
				EmbedCatalog: false,
				Variables:    map[string]string{"connectors.s3.region": "us-east-1", "allow_host_access": "true"},
				Connectors: []*runtimev1.Connector{
					{
						Type:   "file",
						Name:   "repo",
						Config: map[string]string{"dsn": repodsn},
					},
					{
						Type:   "duckdb",
						Name:   "olap",
						Config: map[string]string{"dsn": ""},
					},
				},
			},
		},
		{
			name: "edit olap dsn",
			inst: &drivers.Instance{
				OLAPDriver:   "olap",
				RepoDriver:   "repo",
				EmbedCatalog: false,
				Variables:    map[string]string{"host": "localhost"},
				Connectors: []*runtimev1.Connector{
					{
						Type:   "file",
						Name:   "repo",
						Config: map[string]string{"dsn": repodsn},
					},
					{
						Type:   "duckdb",
						Name:   "olap",
						Config: map[string]string{"dsn": "?access_mode=read_write"},
					},
				},
			},
			savedInst: &drivers.Instance{
				OLAPDriver:   "olap",
				RepoDriver:   "repo",
				EmbedCatalog: false,
				Variables:    map[string]string{"host": "localhost", "allow_host_access": "true"},
				Connectors: []*runtimev1.Connector{
					{
						Type:   "file",
						Name:   "repo",
						Config: map[string]string{"dsn": repodsn},
					},
					{
						Type:   "duckdb",
						Name:   "olap",
						Config: map[string]string{"dsn": "?access_mode=read_write"},
					},
				},
			},
			clearCache: true,
		},
		{
			name: "edit repo dsn",
			inst: &drivers.Instance{
				OLAPDriver:   "olap",
				RepoDriver:   "repo",
				EmbedCatalog: false,
				Variables:    map[string]string{"host": "localhost"},
				Connectors: []*runtimev1.Connector{
					{
						Type:   "file",
						Name:   "repo",
						Config: map[string]string{"dsn": newRepodsn},
					},
					{
						Type:   "duckdb",
						Name:   "olap",
						Config: map[string]string{"dsn": ""},
					},
				},
			},
			savedInst: &drivers.Instance{
				OLAPDriver:   "olap",
				RepoDriver:   "repo",
				EmbedCatalog: false,
				Variables:    map[string]string{"host": "localhost", "allow_host_access": "true"},
				Connectors: []*runtimev1.Connector{
					{
						Type:   "file",
						Name:   "repo",
						Config: map[string]string{"dsn": newRepodsn},
					},
					{
						Type:   "duckdb",
						Name:   "olap",
						Config: map[string]string{"dsn": ""},
					},
				},
			},
			clearCache: true,
		},
		{
			name: "edit annotations",
			inst: &drivers.Instance{
				OLAPDriver:   "olap",
				RepoDriver:   "repo",
				EmbedCatalog: false,
				Variables:    map[string]string{"host": "localhost"},
				Connectors: []*runtimev1.Connector{
					{
						Type:   "file",
						Name:   "repo",
						Config: map[string]string{"dsn": newRepodsn},
					},
					{
						Type:   "duckdb",
						Name:   "olap",
						Config: map[string]string{"dsn": ""},
					},
				},
				Annotations: map[string]string{
					"organization_name": "org_name",
				},
			},
			savedInst: &drivers.Instance{
				OLAPDriver:   "olap",
				RepoDriver:   "repo",
				EmbedCatalog: false,
				Variables:    map[string]string{"host": "localhost", "allow_host_access": "true"},
				Connectors: []*runtimev1.Connector{
					{
						Type:   "file",
						Name:   "repo",
						Config: map[string]string{"dsn": newRepodsn},
					},
					{
						Type:   "duckdb",
						Name:   "olap",
						Config: map[string]string{"dsn": ""},
					},
				},
				Annotations: map[string]string{
					"organization_name": "org_name",
				},
			},
			clearCache: true,
		},
		{
			name: "invalid olap driver",
			inst: &drivers.Instance{
				OLAPDriver:   "olap1",
				RepoDriver:   "repo",
				EmbedCatalog: false,
				Variables:    map[string]string{"host": "localhost"},
				Connectors: []*runtimev1.Connector{
					{
						Type:   "file",
						Name:   "repo",
						Config: map[string]string{"dsn": repodsn},
					},
					{
						Type:   "duckdb",
						Name:   "olap",
						Config: map[string]string{"dsn": ""},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid repo driver",
			inst: &drivers.Instance{
				OLAPDriver:   "olap",
				RepoDriver:   "repo1",
				EmbedCatalog: false,
				Variables:    map[string]string{"host": "localhost"},
				Connectors: []*runtimev1.Connector{
					{
						Type:   "file",
						Name:   "repo",
						Config: map[string]string{"dsn": repodsn},
					},
					{
						Type:   "duckdb",
						Name:   "olap",
						Config: map[string]string{"dsn": ""},
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			//create instance
			inst := &drivers.Instance{
				OLAPDriver:   "olap",
				RepoDriver:   "repo",
				EmbedCatalog: true,
				Connectors: []*runtimev1.Connector{
					{
						Type:   "file",
						Name:   "repo",
						Config: map[string]string{"dsn": repodsn},
					},
					{
						Type:   "duckdb",
						Name:   "olap",
						Config: map[string]string{"dsn": ""},
					},
				},
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
			require.True(t, equal(tt.savedInst.Connectors[0], newInst.Connectors[0]))
			require.True(t, equal(tt.savedInst.Connectors[1], newInst.Connectors[1]))
			require.Equal(t, tt.savedInst.RepoDriver, newInst.RepoDriver)
			require.Equal(t, tt.savedInst.EmbedCatalog, newInst.EmbedCatalog)
			require.Greater(t, time.Since(newInst.CreatedOn), time.Since(newInst.UpdatedOn))
			require.Equal(t, tt.savedInst.Variables, newInst.Variables)

			// verify older olap connection is closed and cache updated if olap changed
			c, _ := rt.connectorDef(inst, inst.OLAPDriver)
			_, ok := rt.connCache.cache[inst.ID+c.Type+generateKey(rt.connectorConfig(inst.OLAPDriver, c.Config, inst.ResolveVariables()))]
			require.Equal(t, !tt.clearCache, ok)
			c, _ = rt.connectorDef(inst, inst.RepoDriver)
			_, ok = rt.connCache.cache[inst.ID+c.Type+generateKey(rt.connectorConfig(inst.RepoDriver, c.Config, inst.ResolveVariables()))]
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
				OLAPDriver:   "olap",
				RepoDriver:   "repo",
				EmbedCatalog: true,
				Connectors: []*runtimev1.Connector{
					{
						Type:   "file",
						Name:   "repo",
						Config: map[string]string{"dsn": repodsn},
					},
					{
						Type:   "duckdb",
						Name:   "olap",
						Config: map[string]string{"dsn": dbFile},
					},
				},
			}
			require.NoError(t, rt.CreateInstance(context.Background(), inst))
			// load all caches
			svc, err := rt.NewCatalogService(ctx, inst.ID)
			require.NoError(t, err)

			// ingest some data
			require.NoError(t, svc.Olap.Exec(ctx, &drivers.Statement{Query: "CREATE TABLE data(id INTEGER, name VARCHAR)"}))
			require.NoError(t, svc.Olap.Exec(ctx, &drivers.Statement{Query: "INSERT INTO data VALUES (1, 'Mark'), (2, 'Hannes')"}))
			require.NoError(t, svc.Catalog.CreateEntry(ctx, &drivers.CatalogEntry{
				Name: "data",
				Type: drivers.ObjectTypeTable,
				Object: &runtimev1.Table{
					Name:    "data",
					Managed: true,
				},
			}))
			require.ErrorContains(t, svc.Catalog.CreateEntry(ctx, &drivers.CatalogEntry{
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
			require.False(t, rt.connCache.lruCache.Contains(inst.ID+"duckdb"+fmt.Sprintf("dsn:%s ", dbFile)))
			require.False(t, rt.connCache.lruCache.Contains(inst.ID+"file"+fmt.Sprintf("dsn:%s ", repodsn)))
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
		OLAPDriver:   "olap",
		RepoDriver:   "repo",
		EmbedCatalog: true,
		Connectors: []*runtimev1.Connector{
			{
				Type:   "file",
				Name:   "repo",
				Config: map[string]string{"dsn": t.TempDir()},
			},
			{
				Type:   "duckdb",
				Name:   "olap",
				Config: map[string]string{"dsn": dbpath},
			},
		},
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
	c, _ := rt.connectorDef(inst, inst.OLAPDriver)
	evicted := rt.connCache.evict(ctx, inst.ID, c.Type, rt.connectorConfig("olap", c.Config, inst.ResolveVariables()))
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
	globalConnectors := []*runtimev1.Connector{
		{
			Type: "sqlite",
			Name: "metastore",
			// Setting a test-specific name ensures a unique connection when "cache=shared" is enabled.
			// "cache=shared" is needed to prevent threading problems.
			Config: map[string]string{"dsn": "file:rill?mode=memory&cache=shared"},
		},
	}

	opts := &Options{
		ConnectionCacheSize:   100,
		MetastoreConnector:    "metastore",
		QueryCacheSizeBytes:   int64(datasize.MB) * 100,
		PolicyEngineCacheSize: 100,
		AllowHostAccess:       true,
		SystemConnectors:      globalConnectors,
	}
	rt, err := New(opts, zap.NewNop(), nil)
	t.Cleanup(func() {
		rt.Close()
	})
	require.NoError(t, err)

	return rt
}
