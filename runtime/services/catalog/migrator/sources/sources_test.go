package sources_test

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	_ "github.com/rilldata/rill/runtime/drivers/duckdb"
	_ "github.com/rilldata/rill/runtime/drivers/file"
	_ "github.com/rilldata/rill/runtime/drivers/sqlite"
	"github.com/rilldata/rill/runtime/services/catalog"
	"github.com/rilldata/rill/runtime/services/catalog/migrator"
	_ "github.com/rilldata/rill/runtime/services/catalog/migrator/sources"
	"github.com/rilldata/rill/runtime/services/catalog/testutils"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
)

const TestDataPath = "../../../../../web-local/test/data"

var AdBidsCsvPath = filepath.Join(TestDataPath, "AdBids.csv")

const AdBidsRepoPath = "/sources/AdBids.yaml"

func TestSourceMigrator_Update(t *testing.T) {
	s, _ := testutils.GetService(t)
	testutils.CreateSource(t, s, "AdBids", AdBidsCsvPath, AdBidsRepoPath)
	result, err := s.Reconcile(context.Background(), catalog.ReconcileConfig{
		SafeSourceRefresh: true,
	})
	require.NoError(t, err)
	require.Len(t, result.Errors, 0)
	testutils.AssertTable(t, s, "AdBids", AdBidsRepoPath)

	// point to invalid file and reconcile
	testutils.CreateSource(t, s, "AdBids", "_"+AdBidsCsvPath, AdBidsRepoPath)
	result, err = s.Reconcile(context.Background(), catalog.ReconcileConfig{
		SafeSourceRefresh: true,
	})
	require.NoError(t, err)
	require.Len(t, result.Errors, 1)
	// table is persisted
	testutils.AssertTable(t, s, "AdBids", AdBidsRepoPath)
}

func TestConnectorWithSourceVariations(t *testing.T) {
	testdataPathRel := "../../../../../web-local/test/data"
	testdataPathAbs, err := filepath.Abs(testdataPathRel)
	require.NoError(t, err)

	sources := []struct {
		Connector       string
		Path            string
		AdditionalProps map[string]any
	}{
		{"local_file", filepath.Join(testdataPathAbs, "AdBids.csv"), nil},
		{"local_file", filepath.Join(testdataPathAbs, "AdBids.csv.gz"), nil},
		{"local_file", filepath.Join(testdataPathAbs, "AdBids.parquet"), map[string]any{
			"duckdb": map[string]any{
				"hive_partitioning": true,
			},
		}},
		{"local_file", filepath.Join(testdataPathAbs, "AdBids.parquet"), nil},
		{"local_file", filepath.Join(testdataPathAbs, "AdBids.txt"), nil},
		{"duckdb", "", map[string]any{
			"sql": fmt.Sprintf(`select * from read_csv_auto('%s')`, filepath.Join(testdataPathAbs, "AdBids.csv")),
		}},
		// something wrong with this particular file. duckdb fails to extract
		// TODO: move the generator to go and fix the parquet file
		//{"local_file", testdataPath + "AdBids.parquet.gz", nil},
		// only enable to do adhoc tests. needs credentials to work
		//{"s3", "s3://rill-developer.rilldata.io/AdBids.csv", nil},
		//{"s3", "s3://rill-developer.rilldata.io/AdBids.csv.gz", nil},
		//{"s3", "s3://rill-developer.rilldata.io/AdBids.parquet", nil},
		//{"s3", "s3://rill-developer.rilldata.io/AdBids.parquet.gz", nil},
		//{"gcs", "gs://scratch.rilldata.com/rill-developer/AdBids.csv", nil},
		//{"gcs", "gs://scratch.rilldata.com/rill-developer/AdBids.csv.gz", nil},
		//{"gcs", "gs://scratch.rilldata.com/rill-developer/AdBids.parquet", nil},
		//{"gcs", "gs://scratch.rilldata.com/rill-developer/AdBids.parquet.gz", nil},
		//{"duckdb", "", map[string]any{
		//	"sql": `select * from read_csv_auto('gs://scratch.rilldata.com/rill-developer/AdBids.csv.gz')`,
		//}},
	}

	ctx := context.Background()
	conn, err := drivers.Open("duckdb", map[string]any{"dsn": "?access_mode=read_write"}, zap.NewNop())
	require.NoError(t, err)
	olap, _ := conn.AsOLAP()

	fileStore, err := drivers.Open("file", map[string]any{"dsn": testdataPathRel}, zap.NewNop())
	require.NoError(t, err)
	repo, _ := fileStore.AsRepoStore()

	m := migrator.Migrators[drivers.ObjectTypeSource]
	opts := migrator.Options{InstanceEnv: map[string]string{"allow_host_access": "true"}, IngestStorageLimitInBytes: 1024 * 1024 * 1024}
	for _, tt := range sources {
		t.Run(fmt.Sprintf("%s - %s", tt.Connector, tt.Path), func(t *testing.T) {
			var props map[string]any
			if tt.AdditionalProps != nil {
				props = tt.AdditionalProps
			} else {
				props = make(map[string]any)
			}
			props["path"] = tt.Path

			p, err := structpb.NewStruct(props)
			require.NoError(t, err)
			source := &drivers.CatalogEntry{
				Type: drivers.ObjectType(runtimev1.ObjectType_OBJECT_TYPE_SOURCE),
				Object: &runtimev1.Source{
					Name:       "foo",
					Connector:  tt.Connector,
					Properties: p,
				},
			}
			err = m.Create(ctx, olap, repo, opts, source, zap.NewNop())
			require.NoError(t, err)

			var count int
			rows, err := olap.Execute(ctx, &drivers.Statement{Query: "SELECT count(timestamp) FROM foo"})
			require.NoError(t, err)
			require.True(t, rows.Next())
			require.NoError(t, rows.Scan(&count))
			require.GreaterOrEqual(t, count, 100)
			require.False(t, rows.Next())
			require.NoError(t, rows.Close())
		})
	}
}

func TestConnectorWithoutRootAccess(t *testing.T) {
	testdataPathRel := "../../../../../web-local/test/data"
	testdataPathAbs, err := filepath.Abs(testdataPathRel)
	require.NoError(t, err)

	sources := []struct {
		Connector string
		Path      string
		repoRoot  string
		isError   bool
	}{
		{"local_file", "AdBids.csv", testdataPathAbs, false},
		{"local_file", filepath.Join(testdataPathAbs, "AdBids.csv"), testdataPathAbs, false},
		{"local_file", "../../../../../runtime/testruntime/testdata/ad_bids/data/AdBids.csv.gz", testdataPathAbs, true},
	}

	ctx := context.Background()
	conn, err := drivers.Open("duckdb", map[string]any{"dsn": "?access_mode=read_write"}, zap.NewNop())
	require.NoError(t, err)
	olap, _ := conn.AsOLAP()

	fileStore, err := drivers.Open("file", map[string]any{"dsn": testdataPathRel}, zap.NewNop())
	require.NoError(t, err)
	repo, _ := fileStore.AsRepoStore()

	m := migrator.Migrators[drivers.ObjectTypeSource]
	opts := migrator.Options{InstanceEnv: map[string]string{"allow_host_access": "false"}, IngestStorageLimitInBytes: 1024 * 1024 * 1024}
	for _, tt := range sources {
		t.Run(fmt.Sprintf("%s - %s", tt.Connector, tt.Path), func(t *testing.T) {
			props := make(map[string]any, 0)
			props["path"] = tt.Path
			props["repo_root"] = tt.repoRoot

			p, err := structpb.NewStruct(props)
			require.NoError(t, err)
			source := &drivers.CatalogEntry{
				Type: drivers.ObjectType(runtimev1.ObjectType_OBJECT_TYPE_SOURCE),
				Object: &runtimev1.Source{
					Name:       "foo",
					Connector:  tt.Connector,
					Properties: p,
				},
			}
			err = m.Create(ctx, olap, repo, opts, source, zap.NewNop())
			if tt.isError {
				require.Error(t, err, "file connector cannot ingest source: path is outside repo root")
				return
			}
			require.NoError(t, err)

			var count int
			rows, err := olap.Execute(ctx, &drivers.Statement{Query: "SELECT count(timestamp) FROM foo"})
			require.NoError(t, err)
			require.True(t, rows.Next())
			require.NoError(t, rows.Scan(&count))
			require.GreaterOrEqual(t, count, 100)
			require.False(t, rows.Next())
			require.NoError(t, rows.Close())
		})
	}
}

func TestCSVDelimiter(t *testing.T) {
	testdataPathAbs, err := filepath.Abs("../../../../../web-local/test/data")
	require.NoError(t, err)
	testDelimiterCsvPath := filepath.Join(testdataPathAbs, "test-delimiter.csv")

	ctx := context.Background()
	conn, err := drivers.Open("duckdb", map[string]any{"dsn": "?access_mode=read_write"}, zap.NewNop())
	require.NoError(t, err)
	defer conn.Close()
	olap, _ := conn.AsOLAP()

	fileStore, err := drivers.Open("file", map[string]any{"dsn": testdataPathAbs}, zap.NewNop())
	require.NoError(t, err)
	defer fileStore.Close()
	repo, _ := fileStore.AsRepoStore()

	m := migrator.Migrators[drivers.ObjectTypeSource]
	opts := migrator.Options{InstanceEnv: map[string]string{"allow_host_access": "false"}, IngestStorageLimitInBytes: 1024 * 1024 * 1024}

	props := make(map[string]any, 0)
	props["path"] = testDelimiterCsvPath
	props["repo_root"] = testdataPathAbs
	props["duckdb"] = map[string]any{"delim": "'+'"}

	p, err := structpb.NewStruct(props)
	require.NoError(t, err)
	source := &drivers.CatalogEntry{
		Type: drivers.ObjectType(runtimev1.ObjectType_OBJECT_TYPE_SOURCE),
		Object: &runtimev1.Source{
			Name:       "foo",
			Connector:  "local_file",
			Properties: p,
		},
	}
	err = m.Create(ctx, olap, repo, opts, source, zap.NewNop())
	require.NoError(t, err)

	rows, err := olap.Execute(ctx, &drivers.Statement{Query: "SELECT * FROM foo"})
	require.NoError(t, err)
	cols, err := rows.Columns()
	require.NoError(t, err)
	// 3 columns because no delimiter is passed
	require.Len(t, cols, 2)
	require.NoError(t, rows.Close())
}

func TestFileFormatAndDelimiter(t *testing.T) {
	ctx := context.Background()
	conn, err := drivers.Open("duckdb", map[string]any{"dsn": "?access_mode=read_write"}, zap.NewNop())
	require.NoError(t, err)
	olap, _ := conn.AsOLAP()

	testdataPathAbs, err := filepath.Abs("../../../../../web-local/test/data")
	require.NoError(t, err)
	testDelimiterCsvPath := filepath.Join(testdataPathAbs, "test-format.log")

	repo := runRepoStore(t, testdataPathAbs)

	m := migrator.Migrators[drivers.ObjectTypeSource]
	opts := migrator.Options{InstanceEnv: map[string]string{"allow_host_access": "true"}, IngestStorageLimitInBytes: 1024 * 1024 * 1024}

	variations := []struct {
		title           string
		connector       string
		path            string
		additionalProps map[string]any
	}{
		{"direct file reference", "local_file", testDelimiterCsvPath, map[string]any{
			"duckdb": map[string]any{"delim": "' '"},
		}},
		{"sql with read_csv_auto", "duckdb", "", map[string]any{
			"sql": fmt.Sprintf(`from read_csv_auto('%s',delim=' ')`, testDelimiterCsvPath),
		}},
	}

	for _, tt := range variations {
		t.Run(tt.title, func(t *testing.T) {
			props := make(map[string]any, 0)
			if tt.additionalProps != nil {
				props = tt.additionalProps
			}
			props["path"] = tt.path
			props["repo_root"] = "."
			props["format"] = "csv"
			p, err := structpb.NewStruct(props)
			require.NoError(t, err)
			source := &drivers.CatalogEntry{
				Type: drivers.ObjectType(runtimev1.ObjectType_OBJECT_TYPE_SOURCE),
				Object: &runtimev1.Source{
					Name:       "foo",
					Connector:  tt.connector,
					Properties: p,
				},
			}
			err = m.Create(ctx, olap, repo, opts, source, zap.NewNop())
			require.NoError(t, err)

			rows, err := olap.Execute(ctx, &drivers.Statement{Query: "SELECT * FROM foo"})
			require.NoError(t, err)
			cols, err := rows.Columns()
			require.NoError(t, err)
			// 5 columns in file
			require.Len(t, cols, 5)
			require.NoError(t, rows.Close())

			var count int
			rows, err = olap.Execute(ctx, &drivers.Statement{Query: "SELECT count(timestamp) FROM foo"})
			require.NoError(t, err)
			require.True(t, rows.Next())
			require.NoError(t, rows.Scan(&count))
			require.Equal(t, count, 8)
			require.False(t, rows.Next())
			require.NoError(t, rows.Close())
		})
	}
}

func TestCSVIngestionWithColumns(t *testing.T) {
	olap := runOLAPStore(t)
	ctx := context.Background()
	filePath := createFilePath(t, "../../../../../web-local/test/data", "Users.csv")
	repo := runRepoStore(t, filepath.Dir(filePath))
	m := migrator.Migrators[drivers.ObjectTypeSource]
	opts := migrator.Options{InstanceEnv: map[string]string{"allow_host_access": "true"}, IngestStorageLimitInBytes: 1024 * 1024 * 1024}

	duckdbProps := map[string]any{
		"auto_detect":   false,
		"header":        true,
		"ignore_errors": true,
		"columns":       "{id:'INTEGER',name:'VARCHAR',country:'VARCHAR',city:'VARCHAR'}",
	}
	variations := []struct {
		title           string
		connector       string
		path            string
		additionalProps map[string]any
	}{
		{"direct file reference", "local_file", filePath, map[string]any{
			"duckdb": duckdbProps,
		}},
		{"sql with read_csv_auto", "duckdb", "", map[string]any{
			"sql": fmt.Sprintf(`
from read_csv_auto('%s',auto_detect=false,header=true,ignore_errors=true,
columns={id:'INTEGER',name:'VARCHAR',country:'VARCHAR',city:'VARCHAR'})`, filePath),
		}},
	}

	for _, tt := range variations {
		t.Run(tt.title, func(t *testing.T) {
			props := make(map[string]any, 0)
			if tt.additionalProps != nil {
				props = tt.additionalProps
			}
			props["path"] = tt.path
			props["repo_root"] = "."
			props["format"] = "csv"
			p, err := structpb.NewStruct(props)
			require.NoError(t, err)
			source := &drivers.CatalogEntry{
				Type: drivers.ObjectType(runtimev1.ObjectType_OBJECT_TYPE_SOURCE),
				Object: &runtimev1.Source{
					Name:       "csv_source",
					Connector:  tt.connector,
					Properties: p,
				},
			}
			err = m.Create(ctx, olap, repo, opts, source, zap.NewNop())
			require.NoError(t, err)

			rows, err := olap.Execute(ctx, &drivers.Statement{Query: "SELECT * FROM csv_source"})
			require.NoError(t, err)
			cols, err := rows.Columns()
			require.NoError(t, err)
			require.Len(t, cols, 4)
			require.ElementsMatch(t, cols, [4]string{"id", "name", "country", "city"})
			require.NoError(t, rows.Close())

			var count int
			rows, err = olap.Execute(ctx, &drivers.Statement{Query: "SELECT count(*) FROM csv_source"})
			require.NoError(t, err)
			require.True(t, rows.Next())
			require.NoError(t, rows.Scan(&count))
			require.Equal(t, count, 100)
			require.False(t, rows.Next())
			require.NoError(t, rows.Close())
		})
	}
}

func TestJsonIngestionDefault(t *testing.T) {
	olap := runOLAPStore(t)
	ctx := context.Background()
	filePath := createFilePath(t, "../../../../../web-local/test/data", "Users.json")
	repo := runRepoStore(t, filepath.Dir(filePath))
	m := migrator.Migrators[drivers.ObjectTypeSource]
	opts := migrator.Options{InstanceEnv: map[string]string{"allow_host_access": "true"}, IngestStorageLimitInBytes: 1024 * 1024 * 1024}

	props := make(map[string]any, 0)
	props["path"] = filePath
	props["repo_root"] = "."

	p, err := structpb.NewStruct(props)
	require.NoError(t, err)
	source := &drivers.CatalogEntry{
		Type: drivers.ObjectType(runtimev1.ObjectType_OBJECT_TYPE_SOURCE),
		Object: &runtimev1.Source{
			Name:       "json_source",
			Connector:  "local_file",
			Properties: p,
		},
	}
	err = m.Create(ctx, olap, repo, opts, source, zap.NewNop())
	require.NoError(t, err)

	rows, err := olap.Execute(ctx, &drivers.Statement{Query: "SELECT * FROM json_source"})
	require.NoError(t, err)
	cols, err := rows.Columns()
	require.NoError(t, err)
	require.Len(t, cols, 9)
	require.NoError(t, rows.Close())

	var count int
	rows, err = olap.Execute(ctx, &drivers.Statement{Query: "SELECT count(*) FROM json_source"})
	require.NoError(t, err)
	require.True(t, rows.Next())
	require.NoError(t, rows.Scan(&count))
	require.Equal(t, count, 10)
	require.False(t, rows.Next())
	require.NoError(t, rows.Close())
}

func TestJsonIngestionWithColumns(t *testing.T) {
	olap := runOLAPStore(t)
	ctx := context.Background()
	filePath := createFilePath(t, "../../../../../web-local/test/data", "Users.json")
	repo := runRepoStore(t, filepath.Dir(filePath))
	m := migrator.Migrators[drivers.ObjectTypeSource]
	opts := migrator.Options{InstanceEnv: map[string]string{"allow_host_access": "true"}, IngestStorageLimitInBytes: 1024 * 1024 * 1024}

	props := make(map[string]any, 0)
	props["path"] = filePath
	props["repo_root"] = "."
	props["duckdb"] = map[string]any{
		"columns": "{id:'INTEGER', name:'VARCHAR', isActive:'BOOLEAN', createdDate:'VARCHAR', address:'STRUCT(street VARCHAR, city VARCHAR, postalCode VARCHAR)', tags:'VARCHAR[]', projects:'STRUCT(projectId INTEGER, projectName VARCHAR, startDate VARCHAR, endDate VARCHAR)[]', scores:'INTEGER[]'}",
	}

	p, err := structpb.NewStruct(props)
	require.NoError(t, err)
	source := &drivers.CatalogEntry{
		Type: drivers.ObjectType(runtimev1.ObjectType_OBJECT_TYPE_SOURCE),
		Object: &runtimev1.Source{
			Name:       "json_source",
			Connector:  "local_file",
			Properties: p,
		},
	}
	err = m.Create(ctx, olap, repo, opts, source, zap.NewNop())
	require.NoError(t, err)

	rows, err := olap.Execute(ctx, &drivers.Statement{Query: "SELECT * FROM json_source"})
	require.NoError(t, err)
	cols, err := rows.Columns()
	require.NoError(t, err)
	require.Len(t, cols, 8)
	require.NoError(t, rows.Close())

	var count int
	rows, err = olap.Execute(ctx, &drivers.Statement{Query: "SELECT count(*) FROM json_source"})
	require.NoError(t, err)
	require.True(t, rows.Next())
	require.NoError(t, rows.Scan(&count))
	require.Equal(t, count, 10)
	require.False(t, rows.Next())
	require.NoError(t, rows.Close())
}

func TestJsonIngestionWithLessColumns(t *testing.T) {
	olap := runOLAPStore(t)
	ctx := context.Background()
	filePath := createFilePath(t, "../../../../../web-local/test/data", "Users.json")
	repo := runRepoStore(t, filepath.Dir(filePath))
	m := migrator.Migrators[drivers.ObjectTypeSource]
	opts := migrator.Options{InstanceEnv: map[string]string{"allow_host_access": "true"}, IngestStorageLimitInBytes: 1024 * 1024 * 1024}

	props := make(map[string]any, 0)
	props["path"] = filePath
	props["repo_root"] = "."
	props["duckdb"] = map[string]any{
		"columns": "{id:'INTEGER',name:'VARCHAR',isActive:'BOOLEAN',createdDate:'VARCHAR',}",
	}

	p, err := structpb.NewStruct(props)
	require.NoError(t, err)
	source := &drivers.CatalogEntry{
		Type: drivers.ObjectType(runtimev1.ObjectType_OBJECT_TYPE_SOURCE),
		Object: &runtimev1.Source{
			Name:       "json_source",
			Connector:  "local_file",
			Properties: p,
		},
	}
	err = m.Create(ctx, olap, repo, opts, source, zap.NewNop())
	require.NoError(t, err)

	require.NoError(t, err)
	rows, err := olap.Execute(ctx, &drivers.Statement{Query: "SELECT * FROM json_source"})
	require.NoError(t, err)
	cols, err := rows.Columns()
	require.NoError(t, err)
	require.Len(t, cols, 4)
	require.NoError(t, rows.Close())

	var count int
	rows, err = olap.Execute(ctx, &drivers.Statement{Query: "SELECT count(*) FROM json_source"})
	require.NoError(t, err)
	require.True(t, rows.Next())
	require.NoError(t, rows.Scan(&count))
	require.Equal(t, count, 10)
	require.False(t, rows.Next())
	require.NoError(t, rows.Close())
}

func TestJsonIngestionWithVariousParams(t *testing.T) {
	olap := runOLAPStore(t)
	ctx := context.Background()
	filePath := createFilePath(t, "../../../../../web-local/test/data", "Users.json")
	repo := runRepoStore(t, filepath.Dir(filePath))
	m := migrator.Migrators[drivers.ObjectTypeSource]
	opts := migrator.Options{InstanceEnv: map[string]string{"allow_host_access": "true"}, IngestStorageLimitInBytes: 1024 * 1024 * 1024}

	duckdbProps := map[string]any{
		"maximum_object_size": "9999999",
		"records":             true,
		"ignore_errors":       true,
		"columns":             "{id:'INTEGER',name:'VARCHAR',isActive:'BOOLEAN',createdDate:'VARCHAR'}",
		"auto_detect":         false,
		"sample_size":         -1,
		"dateformat":          "iso",
		"timestampformat":     "iso",
	}

	variations := []struct {
		title           string
		connector       string
		path            string
		additionalProps map[string]any
	}{
		{"direct file reference", "local_file", filePath, map[string]any{
			"duckdb": duckdbProps,
		}},
		{"sql with read_csv_auto", "duckdb", "", map[string]any{
			"sql": fmt.Sprintf(`
from read_json('%s',maximum_object_size=9999999,records=true,ignore_errors=true,
columns={id:'INTEGER',name:'VARCHAR',isActive:'BOOLEAN',createdDate:'VARCHAR'},
auto_detect=false,sample_size=-1,dateformat='iso',timestampformat='iso',format='auto')`, filePath),
		}},
	}

	props := make(map[string]any, 0)
	props["path"] = filePath
	props["repo_root"] = "."
	props["duckdb"] = map[string]any{
		"maximum_object_size": "9999999",
		"records":             true,
		"ignore_errors":       true,
		"columns":             "{id:'INTEGER',name:'VARCHAR',isActive:'BOOLEAN',createdDate:'VARCHAR',}",
		"auto_detect":         false,
		"sample_size":         -1,
		"dateformat":          "iso",
		"timestampformat":     "iso",
	}

	for _, tt := range variations {
		t.Run(tt.title, func(t *testing.T) {
			props := make(map[string]any, 0)
			if tt.additionalProps != nil {
				props = tt.additionalProps
			}
			props["path"] = tt.path
			props["repo_root"] = "."

			p, err := structpb.NewStruct(props)
			require.NoError(t, err)
			source := &drivers.CatalogEntry{
				Type: drivers.ObjectType(runtimev1.ObjectType_OBJECT_TYPE_SOURCE),
				Object: &runtimev1.Source{
					Name:       "json_source",
					Connector:  tt.connector,
					Properties: p,
				},
			}
			err = m.Create(ctx, olap, repo, opts, source, zap.NewNop())
			require.NoError(t, err)

			rows, err := olap.Execute(ctx, &drivers.Statement{Query: "SELECT * FROM json_source"})
			require.NoError(t, err)
			cols, err := rows.Columns()
			require.NoError(t, err)
			require.Len(t, cols, 4)
			require.NoError(t, rows.Close())

			var count int
			rows, err = olap.Execute(ctx, &drivers.Statement{Query: "SELECT count(*) FROM json_source"})
			require.NoError(t, err)
			require.True(t, rows.Next())
			require.NoError(t, rows.Scan(&count))
			require.Equal(t, count, 10)
			require.False(t, rows.Next())
			require.NoError(t, rows.Close())
		})
	}
}

func TestJsonIngestionWithInvalidParam(t *testing.T) {
	olap := runOLAPStore(t)
	ctx := context.Background()
	filePath := createFilePath(t, "../../../../../web-local/test/data", "Users.json")
	repo := runRepoStore(t, filepath.Dir(filePath))
	m := migrator.Migrators[drivers.ObjectTypeSource]
	opts := migrator.Options{InstanceEnv: map[string]string{"allow_host_access": "true"}, IngestStorageLimitInBytes: 1024 * 1024 * 1024}

	props := make(map[string]any, 0)
	props["path"] = filePath
	props["repo_root"] = "."
	props["duckdb"] = map[string]any{
		"json": map[string]any{
			"invalid_param": "auto",
		},
	}

	p, err := structpb.NewStruct(props)
	require.NoError(t, err)
	source := &drivers.CatalogEntry{
		Type: drivers.ObjectType(runtimev1.ObjectType_OBJECT_TYPE_SOURCE),
		Object: &runtimev1.Source{
			Name:       "json_source",
			Connector:  "local_file",
			Properties: p,
		},
	}
	err = m.Create(ctx, olap, repo, opts, source, zap.NewNop())
	require.Error(t, err, "Invalid named parameter \"invalid_param\" for function read_json")
}

func TestPropertiesEquals(t *testing.T) {
	s1 := &runtimev1.Source{
		Name:       "s1",
		Connector:  "local",
		Properties: newStruct(t, map[string]any{"a": 100, "b": "hello world"}),
	}

	s2 := &runtimev1.Source{
		Name:       "s2",
		Connector:  "local",
		Properties: newStruct(t, map[string]any{"a": 100, "b": "hello world"}),
	}

	s3 := &runtimev1.Source{
		Name:       "s3",
		Connector:  "local",
		Properties: newStruct(t, map[string]any{"a": 101, "b": "hello world"}),
	}

	s4 := &runtimev1.Source{
		Name:       "s4",
		Connector:  "local",
		Properties: newStruct(t, map[string]any{"a": 100, "c": "hello world"}),
	}

	s5 := &runtimev1.Source{
		Name:      "s5",
		Connector: "local",
		Properties: newStruct(t, map[string]any{
			"number": 0,
			"string": "hello world",
			"nestedMap": map[string]any{
				"nestedMap": map[string]any{
					"string": "value",
					"number": 2,
				},
				"string": "value",
				"number": 1,
			},
		}),
	}

	s6 := &runtimev1.Source{
		Name:      "s6",
		Connector: "local",
		Properties: newStruct(t, map[string]any{
			"number": 0,
			"string": "hello world",
			"nestedMap": map[string]any{
				"number": 1,
				"string": "value",
				"nestedMap": map[string]any{
					"number": 2,
					"string": "value",
				},
			},
		}),
	}

	s7 := &runtimev1.Source{
		Name:      "s7",
		Connector: "local",
		Properties: newStruct(t, map[string]any{
			"number": 0,
			"string": "hello world",
			"nestedMap": map[string]any{
				"number": 1,
				"string": "value",
			},
		}),
	}

	m := migrator.Migrators[drivers.ObjectTypeSource]
	// s1 and s2 should be equal
	ctx := context.Background()
	require.True(t, m.IsEqual(ctx, &drivers.CatalogEntry{Object: s1}, &drivers.CatalogEntry{Object: s2}))
	require.True(t, m.IsEqual(ctx, &drivers.CatalogEntry{Object: s2}, &drivers.CatalogEntry{Object: s1}))

	// s1 should not equal s3 or s4
	require.False(t, m.IsEqual(ctx, &drivers.CatalogEntry{Object: s1}, &drivers.CatalogEntry{Object: s3}))
	require.False(t, m.IsEqual(ctx, &drivers.CatalogEntry{Object: s3}, &drivers.CatalogEntry{Object: s1}))
	require.False(t, m.IsEqual(ctx, &drivers.CatalogEntry{Object: s1}, &drivers.CatalogEntry{Object: s4}))
	require.False(t, m.IsEqual(ctx, &drivers.CatalogEntry{Object: s4}, &drivers.CatalogEntry{Object: s1}))

	// s2 should not equal s3 or s4
	require.False(t, m.IsEqual(ctx, &drivers.CatalogEntry{Object: s2}, &drivers.CatalogEntry{Object: s3}))
	require.False(t, m.IsEqual(ctx, &drivers.CatalogEntry{Object: s3}, &drivers.CatalogEntry{Object: s2}))
	require.False(t, m.IsEqual(ctx, &drivers.CatalogEntry{Object: s2}, &drivers.CatalogEntry{Object: s4}))
	require.False(t, m.IsEqual(ctx, &drivers.CatalogEntry{Object: s4}, &drivers.CatalogEntry{Object: s2}))

	// s5 and s6 should be equal
	require.True(t, m.IsEqual(ctx, &drivers.CatalogEntry{Object: s5}, &drivers.CatalogEntry{Object: s6}))
	require.True(t, m.IsEqual(ctx, &drivers.CatalogEntry{Object: s6}, &drivers.CatalogEntry{Object: s5}))

	// s6 and s7 should not be equal
	require.False(t, m.IsEqual(ctx, &drivers.CatalogEntry{Object: s7}, &drivers.CatalogEntry{Object: s6}))
	require.False(t, m.IsEqual(ctx, &drivers.CatalogEntry{Object: s6}, &drivers.CatalogEntry{Object: s7}))
}

func TestSqlIngestionWithFiltersAndColumns(t *testing.T) {
	ctx := context.Background()
	conn, err := drivers.Open("duckdb", map[string]any{"dsn": "?access_mode=read_write"}, zap.NewNop())
	require.NoError(t, err)
	olap, _ := conn.AsOLAP()
	m := migrator.Migrators[drivers.ObjectTypeSource]
	opts := migrator.Options{InstanceEnv: map[string]string{"allow_host_access": "true"}, IngestStorageLimitInBytes: 1024 * 1024 * 1024}

	testdataPathAbs, err := filepath.Abs("../../../../../web-local/test/data")
	require.NoError(t, err)
	testCsvPath := filepath.Join(testdataPathAbs, "AdBids.csv")

	repo := runRepoStore(t, testdataPathAbs)

	props := make(map[string]any, 0)
	props["repo_root"] = "."
	props["sql"] = fmt.Sprintf(`
select * exclude(publisher),
(case when publisher = 'Yahoo' then 0 when publisher = 'Google' then 1 else 2 end) as pub,
from read_csv_auto('%s') where publisher in ('Yahoo', 'Google')`, testCsvPath)

	p, err := structpb.NewStruct(props)
	require.NoError(t, err)
	source := &drivers.CatalogEntry{
		Type: drivers.ObjectType(runtimev1.ObjectType_OBJECT_TYPE_SOURCE),
		Object: &runtimev1.Source{
			Name:       "csv_source",
			Connector:  "duckdb",
			Properties: p,
		},
	}
	err = m.Create(ctx, olap, repo, opts, source, zap.NewNop())
	require.NoError(t, err)

	rows, err := olap.Execute(ctx, &drivers.Statement{Query: "SELECT * FROM csv_source"})
	require.NoError(t, err)
	cols, err := rows.Columns()
	require.NoError(t, err)
	require.Len(t, cols, 5)
	require.ElementsMatch(t, cols, [5]string{"id", "timestamp", "domain", "bid_price", "pub"})
	require.NoError(t, rows.Close())

	var count int
	rows, err = olap.Execute(ctx, &drivers.Statement{Query: "SELECT count(*) FROM csv_source"})
	require.NoError(t, err)
	require.True(t, rows.Next())
	require.NoError(t, rows.Scan(&count))
	require.Equal(t, count, 37356)
	require.False(t, rows.Next())
	require.NoError(t, rows.Close())

	rows, err = olap.Execute(ctx, &drivers.Statement{Query: "SELECT count(pub) FROM csv_source group by pub"})
	require.NoError(t, err)
	counts := make([]int, 0)
	for rows.Next() {
		var count int
		require.NoError(t, rows.Scan(&count))
		counts = append(counts, count)
	}
	// Only 2 rows present since we filtered out other publishers
	require.ElementsMatch(t, counts, [2]int{18593, 18763})
}

func newStruct(t *testing.T, m map[string]any) *structpb.Struct {
	v, err := structpb.NewStruct(m)
	require.NoError(t, err)
	return v
}

func createFilePath(t *testing.T, dirPath string, fileName string) string {
	testdataPathAbs, err := filepath.Abs(dirPath)
	require.NoError(t, err)
	filePath := filepath.Join(testdataPathAbs, fileName)
	return filePath
}

func runOLAPStore(t *testing.T) drivers.OLAPStore {
	conn, err := drivers.Open("duckdb", map[string]any{"dsn": "?access_mode=read_write"}, zap.NewNop())
	require.NoError(t, err)
	olap, canServe := conn.AsOLAP()
	require.True(t, canServe)
	return olap
}

func runRepoStore(t *testing.T, testdataPathAbs string) drivers.RepoStore {
	fileStore, err := drivers.Open("file", map[string]any{"dsn": testdataPathAbs}, zap.NewNop())
	require.NoError(t, err)
	repo, _ := fileStore.AsRepoStore()
	return repo
}
