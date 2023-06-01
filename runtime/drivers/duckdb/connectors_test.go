package duckdb_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/rilldata/rill/runtime/connectors"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/drivers/duckdb"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	_ "github.com/rilldata/rill/runtime/connectors/gcs"
	_ "github.com/rilldata/rill/runtime/connectors/s3"
)

type mockConnector struct {
	anonymous    bool
	mockIterator connectors.FileIterator
}

func (m *mockConnector) Spec() connectors.Spec {
	return connectors.Spec{}
}

func (m *mockConnector) ConsumeAsIterator(ctx context.Context, env *connectors.Env, source *connectors.Source, logger *zap.Logger) (connectors.FileIterator, error) {
	return m.mockIterator, nil
}

func (m *mockConnector) HasAnonymousAccess(ctx context.Context, env *connectors.Env, source *connectors.Source) (bool, error) {
	return m.anonymous, nil
}

var _ connectors.Connector = &mockConnector{}

type mockIterator struct {
	batches [][]string
	index   int
}

func (m *mockIterator) Close() error {
	return nil
}

func (m *mockIterator) NextBatch(limit int) ([]string, error) {
	m.index += 1
	return m.batches[m.index-1], nil
}

func (m *mockIterator) HasNext() bool {
	return m.index < len(m.batches)
}

var _ connectors.FileIterator = &mockIterator{}

func TestConnectorWithSourceVariations(t *testing.T) {
	testdataPathRel := "../../../web-local/test/data"
	testdataPathAbs, err := filepath.Abs(testdataPathRel)
	require.NoError(t, err)

	sources := []struct {
		Connector       string
		Path            string
		AdditionalProps map[string]any
	}{
		{"local_file", filepath.Join(testdataPathRel, "AdBids.csv"), nil},
		{"local_file", filepath.Join(testdataPathRel, "AdBids.csv"), map[string]any{"csv.delimiter": ","}},
		{"local_file", filepath.Join(testdataPathRel, "AdBids.csv.gz"), nil},
		{"local_file", filepath.Join(testdataPathRel, "AdBids.parquet"), map[string]any{"hive_partitioning": true}},
		{"local_file", filepath.Join(testdataPathAbs, "AdBids.parquet"), nil},
		{"local_file", filepath.Join(testdataPathAbs, "AdBids.txt"), nil},
		{"local_file", "../../../runtime/testruntime/testdata/ad_bids/data/AdBids.csv.gz", nil},
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
	}

	ctx := context.Background()
	conn, err := duckdb.Driver{}.Open("?access_mode=read_write", zap.NewNop())
	require.NoError(t, err)
	olap, _ := conn.OLAPStore()

	for _, tt := range sources {
		t.Run(fmt.Sprintf("%s - %s", tt.Connector, tt.Path), func(t *testing.T) {
			var props map[string]any
			if tt.AdditionalProps != nil {
				props = tt.AdditionalProps
			} else {
				props = make(map[string]any)
			}
			props["path"] = tt.Path

			e := &connectors.Env{
				RepoDriver:      "file",
				RepoRoot:        ".",
				AllowHostAccess: true,
			}
			s := &connectors.Source{
				Name:       "foo",
				Connector:  tt.Connector,
				Properties: props,
			}
			_, err = olap.Ingest(ctx, e, s)
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

func TestConnectorWithGithubRepoDriver(t *testing.T) {
	testdataPathRel := "../../../web-local/test/data"
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
		{"local_file", "../../../runtime/testruntime/testdata/ad_bids/data/AdBids.csv.gz", testdataPathAbs, true},
	}

	ctx := context.Background()
	conn, err := duckdb.Driver{}.Open("?access_mode=read_write", zap.NewNop())
	require.NoError(t, err)
	olap, _ := conn.OLAPStore()

	for _, tt := range sources {
		t.Run(fmt.Sprintf("%s - %s", tt.Connector, tt.Path), func(t *testing.T) {
			props := make(map[string]any)
			props["path"] = tt.Path

			e := &connectors.Env{
				RepoDriver:      "github",
				RepoRoot:        tt.repoRoot,
				AllowHostAccess: false,
			}
			s := &connectors.Source{
				Name:       "foo",
				Connector:  tt.Connector,
				Properties: props,
			}
			_, err = olap.Ingest(ctx, e, s)
			if tt.isError {
				require.Error(t, err)
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
	ctx := context.Background()
	conn, err := duckdb.Driver{}.Open("?access_mode=read_write", zap.NewNop())
	require.NoError(t, err)
	olap, _ := conn.OLAPStore()

	testdataPathAbs, err := filepath.Abs("../../../web-local/test/data")
	require.NoError(t, err)
	testDelimiterCsvPath := filepath.Join(testdataPathAbs, "test-delimiter.csv")

	_, err = olap.Ingest(ctx, &connectors.Env{
		RepoDriver:      "file",
		RepoRoot:        ".",
		AllowHostAccess: true,
	}, &connectors.Source{
		Name:      "foo",
		Connector: "local_file",
		Properties: map[string]any{
			"path": testDelimiterCsvPath,
		},
	})
	require.NoError(t, err)
	rows, err := olap.Execute(ctx, &drivers.Statement{Query: "SELECT * FROM foo"})
	require.NoError(t, err)
	cols, err := rows.Columns()
	require.NoError(t, err)
	// 3 columns because no delimiter is passed
	require.Len(t, cols, 3)
	require.NoError(t, rows.Close())

	_, err = olap.Ingest(ctx, &connectors.Env{
		RepoDriver:      "file",
		RepoRoot:        ".",
		AllowHostAccess: true,
	}, &connectors.Source{
		Name:      "foo",
		Connector: "local_file",
		Properties: map[string]any{
			"path":   testDelimiterCsvPath,
			"duckdb": map[string]any{"delim": "'+'"},
		},
	})
	require.NoError(t, err)
	rows, err = olap.Execute(ctx, &drivers.Statement{Query: "SELECT * FROM foo"})
	require.NoError(t, err)
	cols, err = rows.Columns()
	require.NoError(t, err)
	// 3 columns because no delimiter is passed
	require.Len(t, cols, 2)
	require.NoError(t, rows.Close())
}

func TestFileFormatAndDelimiter(t *testing.T) {
	ctx := context.Background()
	conn, err := duckdb.Driver{}.Open("?access_mode=read_write", zap.NewNop())
	require.NoError(t, err)
	olap, _ := conn.OLAPStore()

	testdataPathAbs, err := filepath.Abs("../../../web-local/test/data")
	require.NoError(t, err)
	testDelimiterCsvPath := filepath.Join(testdataPathAbs, "test-format.log")

	_, err = olap.Ingest(ctx, &connectors.Env{
		RepoDriver:      "file",
		RepoRoot:        ".",
		AllowHostAccess: true,
	}, &connectors.Source{
		Name:      "foo",
		Connector: "local_file",
		Properties: map[string]any{
			"path":   testDelimiterCsvPath,
			"format": "csv",
			"duckdb": map[string]any{"delim": "' '"},
		},
	})
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
}

func TestCSVIngestionWithColumns(t *testing.T) {
	olap := runOLAPStore(t)
	ctx := context.Background()
	filePath := createFilePath(t, "../../../web-local/test/data", "Users.csv")

	_, err := olap.Ingest(ctx, &connectors.Env{
		RepoDriver:      "file",
		RepoRoot:        ".",
		AllowHostAccess: true,
	}, &connectors.Source{
		Name:      "csv_source",
		Connector: "local_file",
		Properties: map[string]any{
			"path": filePath,
			"duckdb": map[string]any{
				"auto_detect":   false,
				"header":        true,
				"ignore_errors": true,
				"columns":       "{id:'INTEGER',name:'VARCHAR',country:'VARCHAR',city:'VARCHAR'}",
			},
		},
	})
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
}

func TestJsonIngestionDefault(t *testing.T) {
	olap := runOLAPStore(t)
	ctx := context.Background()
	filePath := createFilePath(t, "../../../web-local/test/data", "Users.json")

	_, err := olap.Ingest(ctx, &connectors.Env{
		RepoDriver:      "file",
		RepoRoot:        ".",
		AllowHostAccess: true,
	}, &connectors.Source{
		Name:      "json_source",
		Connector: "local_file",
		Properties: map[string]any{
			"path": filePath,
		},
	})
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
	filePath := createFilePath(t, "../../../web-local/test/data", "Users.json")

	_, err := olap.Ingest(ctx, &connectors.Env{
		RepoDriver:      "file",
		RepoRoot:        ".",
		AllowHostAccess: true,
	}, &connectors.Source{
		Name:      "json_source",
		Connector: "local_file",
		Properties: map[string]any{
			"path": filePath,
			"duckdb": map[string]any{
				"columns": "{id:'INTEGER', name:'VARCHAR', isActive:'BOOLEAN', createdDate:'VARCHAR', address:'STRUCT(street VARCHAR, city VARCHAR, postalCode VARCHAR)', tags:'VARCHAR[]', projects:'STRUCT(projectId INTEGER, projectName VARCHAR, startDate VARCHAR, endDate VARCHAR)[]', scores:'INTEGER[]'}",
			},
		},
	})
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
	filePath := createFilePath(t, "../../../web-local/test/data", "Users.json")

	_, err := olap.Ingest(ctx, &connectors.Env{
		RepoDriver:      "file",
		RepoRoot:        ".",
		AllowHostAccess: true,
	}, &connectors.Source{
		Name:      "json_source",
		Connector: "local_file",
		Properties: map[string]any{
			"path": filePath,
			"duckdb": map[string]any{
				"columns": "{id:'INTEGER',name:'VARCHAR',isActive:'BOOLEAN',createdDate:'VARCHAR',}",
			},
		},
	})
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
	filePath := createFilePath(t, "../../../web-local/test/data", "Users.json")

	_, err := olap.Ingest(ctx, &connectors.Env{
		RepoDriver:      "file",
		RepoRoot:        ".",
		AllowHostAccess: true,
	}, &connectors.Source{
		Name:      "json_source",
		Connector: "local_file",
		Properties: map[string]any{
			"path": filePath,
			"duckdb": map[string]any{
				"maximum_object_size": "9999999",
				"records":             true,
				"ignore_errors":       true,
				"columns":             "{id:'INTEGER',name:'VARCHAR',isActive:'BOOLEAN',createdDate:'VARCHAR',}",
				"auto_detect":         false,
				"sample_size":         -1,
				"dateformat":          "iso",
				"timestampformat":     "iso",
			},
		},
	})
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

func TestJsonIngestionWithInvalidParam(t *testing.T) {
	olap := runOLAPStore(t)
	ctx := context.Background()
	filePath := createFilePath(t, "../../../web-local/test/data", "Users.json")

	_, err := olap.Ingest(ctx, &connectors.Env{
		RepoDriver:      "file",
		RepoRoot:        ".",
		AllowHostAccess: true,
	}, &connectors.Source{
		Name:      "json_source",
		Connector: "local_file",
		Properties: map[string]any{
			"path": filePath,
			"duckdb": map[string]any{
				"json": map[string]any{
					"invalid_param": "auto",
				},
			},
		},
	})
	require.Error(t, err, "Invalid named parameter \"invalid_param\" for function read_json")
}

func TestIterativeCSVIngestionWithVariableSchema(t *testing.T) {
	tempDir := t.TempDir()
	file1 := filepath.Join(tempDir, "data1.csv")
	temp, err := os.Create(file1)
	require.NoError(t, err)
	_, err = temp.WriteString(`id,city
1,bglr
2,mum`)
	require.NoError(t, err)

	file2 := filepath.Join(tempDir, "data2.csv")
	temp, err = os.Create(file2)
	require.NoError(t, err)
	_, err = temp.WriteString(`id,city,country
3,bglr,IND
4,mum,IND`)
	require.NoError(t, err)

	file3 := filepath.Join(tempDir, "data3.csv")
	temp, err = os.Create(file3)
	require.NoError(t, err)
	_, err = temp.WriteString(`city,id
bglr,5
mum,6`)
	require.NoError(t, err)

	file4 := filepath.Join(tempDir, "data4.csv")
	temp, err = os.Create(file4)
	require.NoError(t, err)
	_, err = temp.WriteString(`city,id
bglr,7.1
mum,8.2`)
	require.NoError(t, err)

	type test struct {
		mockIterator mockIterator
		name         string
		count        int
		filterCount  int
		colCount     int
	}

	tests := []test{
		{
			mockIterator: mockIterator{batches: [][]string{
				{file1, file1},
				{file1, file1},
			}},
			name:        "same_schema",
			count:       8,
			filterCount: 4,
			colCount:    2,
		},
		{
			mockIterator: mockIterator{batches: [][]string{
				{file1, file2, file3, file4},
			}},
			name:        "variable_schema_ingested_at_once",
			count:       8,
			filterCount: 4,
			colCount:    3,
		},
		{
			mockIterator: mockIterator{batches: [][]string{
				{file1},
				{file2},
				{file3},
				{file4},
			}},
			name:        "changing_schema",
			count:       8,
			filterCount: 4,
			colCount:    3,
		},
		{
			mockIterator: mockIterator{batches: [][]string{
				{file1},
				{file3, file1},
				{file2, file1},
				{file1, file4},
			}},
			name:        "schema_combination",
			count:       14,
			filterCount: 7,
			colCount:    3,
		},
	}

	mockConnector := &mockConnector{}
	connectors.Register("mock-csv", mockConnector)
	for _, test := range tests {
		mockConnector.mockIterator = &test.mockIterator
		olap := runOLAPStore(t)
		ctx := context.Background()

		_, err = olap.Ingest(ctx, &connectors.Env{
			RepoDriver:      "file",
			RepoRoot:        ".",
			AllowHostAccess: true,
		}, &connectors.Source{
			Name:      test.name,
			Connector: "mock-csv",
			Properties: map[string]any{
				"path": filepath.Join(tempDir, "*.csv"),
			},
		})
		require.NoError(t, err, "no err expected test %s", test.name)

		var count int
		rows, err := olap.Execute(ctx, &drivers.Statement{Query: fmt.Sprintf("SELECT count(*) FROM %s", test.name)})
		require.NoError(t, err)
		require.True(t, rows.Next())
		require.NoError(t, rows.Scan(&count))
		require.Equal(t, test.count, count)
		require.NoError(t, rows.Close())

		rows, err = olap.Execute(ctx, &drivers.Statement{Query: fmt.Sprintf("SELECT count(*) FROM %s where city='bglr'", test.name)})
		require.NoError(t, err)
		require.True(t, rows.Next())
		require.NoError(t, rows.Scan(&count))
		require.Equal(t, test.filterCount, count)
		require.NoError(t, rows.Close())

		rows, err = olap.Execute(ctx, &drivers.Statement{Query: fmt.Sprintf("DESCRIBE %s", test.name)})
		require.NoError(t, err)
		colCount := 0
		for rows.Next() {
			colCount++
		}
		require.Equal(t, test.colCount, colCount)
	}

}

func TestIterativeCSVIngestionWithVariableSchemaError(t *testing.T) {
	tempDir := t.TempDir()
	file1 := filepath.Join(tempDir, "data1.csv")
	temp, err := os.Create(file1)
	require.NoError(t, err)
	_, err = temp.WriteString(`id,city
1,bglr
2,mum`)
	require.NoError(t, err)

	file2 := filepath.Join(tempDir, "data2.csv")
	temp, err = os.Create(file2)
	require.NoError(t, err)
	_, err = temp.WriteString(`id,city,country
3,bglr,IND
4,mum,IND`)
	require.NoError(t, err)

	file3 := filepath.Join(tempDir, "data3.csv")
	temp, err = os.Create(file3)
	require.NoError(t, err)
	_, err = temp.WriteString(`city,id
bglr,5
mum,6`)
	require.NoError(t, err)

	file4 := filepath.Join(tempDir, "data4.csv")
	temp, err = os.Create(file4)
	require.NoError(t, err)
	_, err = temp.WriteString(`city,id
bglr,7.1
mum,8.2`)
	require.NoError(t, err)

	type test struct {
		mockIterator mockIterator
		name         string
		hasError     bool
	}

	tests := []test{
		{
			mockIterator: mockIterator{batches: [][]string{
				{file1, file1},
				{file1, file1},
			}},
			name:     "same_schema",
			hasError: false,
		},
		{
			mockIterator: mockIterator{batches: [][]string{
				{file1, file2, file3, file4},
			}},
			name:     "variable_schema_ingested_at_once",
			hasError: false,
		},
		{
			mockIterator: mockIterator{batches: [][]string{
				{file1, file2, file3, file4},
			}},
			name:     "columns_jumbled",
			hasError: false,
		},
		{
			mockIterator: mockIterator{batches: [][]string{
				{file1},
				{file2},
			}},
			name:     "new_columns",
			hasError: true,
		},
		{
			mockIterator: mockIterator{batches: [][]string{
				{file2},
				{file1},
			}},
			name:     "less_columns",
			hasError: true,
		},
		{
			mockIterator: mockIterator{batches: [][]string{
				{file1},
				{file4},
			}},
			name:     "datatype_change",
			hasError: true,
		},
	}

	mockConnector := &mockConnector{}
	connectors.Register("mock-csv-error", mockConnector)
	for _, test := range tests {
		mockConnector.mockIterator = &test.mockIterator
		olap := runOLAPStore(t)
		ctx := context.Background()

		_, err = olap.Ingest(ctx, &connectors.Env{
			RepoDriver:      "file",
			RepoRoot:        ".",
			AllowHostAccess: true,
		}, &connectors.Source{
			Name:      test.name,
			Connector: "mock-csv-error",
			Properties: map[string]any{
				"path":                   filepath.Join(tempDir, "*.csv"),
				"allow_field_addition":   false,
				"allow_field_relaxation": false,
			},
		})
		if test.hasError {
			require.Error(t, err, fmt.Errorf("error expected for %s got nil", test.name))
		} else {
			require.NoError(t, err, fmt.Errorf("no error expected for %s got %s", test.name, err))
		}
	}

}

func TestIterativeParquetIngestionWithVariableSchema(t *testing.T) {
	file1 := filepath.Join("../../testruntime/testdata/variable-schema", "data.parquet")
	file2 := filepath.Join("../../testruntime/testdata/variable-schema", "data1.parquet")
	file3 := filepath.Join("../../testruntime/testdata/variable-schema", "data2.parquet")
	file4 := filepath.Join("../../testruntime/testdata/variable-schema", "data3.parquet")

	type test struct {
		mockIterator mockIterator
		name         string
		count        int
		filterCount  int
		colCount     int
	}

	tests := []test{
		{
			mockIterator: mockIterator{batches: [][]string{
				{file1, file1},
				{file1, file1},
			}},
			name:        "same_schema",
			count:       8,
			filterCount: 4,
			colCount:    2,
		},
		{
			mockIterator: mockIterator{batches: [][]string{
				{file1, file2, file3, file4},
			}},
			name:        "variable_schema_ingested_at_once",
			count:       8,
			filterCount: 4,
			colCount:    3,
		},
		{
			mockIterator: mockIterator{batches: [][]string{
				{file1},
				{file2},
				{file3},
				{file4},
			}},
			name:        "changing_schema",
			count:       8,
			filterCount: 4,
			colCount:    3,
		},
		{
			mockIterator: mockIterator{batches: [][]string{
				{file1},
				{file3, file1},
				{file2, file1},
				{file1, file4},
			}},
			name:        "schema_combination",
			count:       14,
			filterCount: 7,
			colCount:    3,
		},
	}

	mockConnector := &mockConnector{}
	connectors.Register("mock-parquet", mockConnector)
	for _, test := range tests {
		mockConnector.mockIterator = &test.mockIterator
		olap := runOLAPStore(t)
		ctx := context.Background()

		_, err := olap.Ingest(ctx, &connectors.Env{
			RepoDriver:      "file",
			RepoRoot:        ".",
			AllowHostAccess: true,
		}, &connectors.Source{
			Name:      test.name,
			Connector: "mock-parquet",
		})
		require.NoError(t, err)

		var count int
		rows, err := olap.Execute(ctx, &drivers.Statement{Query: fmt.Sprintf("SELECT count(*) FROM %s", test.name)})
		require.NoError(t, err)
		require.True(t, rows.Next())
		require.NoError(t, rows.Scan(&count))
		require.Equal(t, test.count, count)
		require.NoError(t, rows.Close())

		rows, err = olap.Execute(ctx, &drivers.Statement{Query: fmt.Sprintf("SELECT count(*) FROM %s where city='bglr'", test.name)})
		require.NoError(t, err)
		require.True(t, rows.Next())
		require.NoError(t, rows.Scan(&count))
		require.Equal(t, test.filterCount, count)
		require.NoError(t, rows.Close())

		rows, err = olap.Execute(ctx, &drivers.Statement{Query: fmt.Sprintf("DESCRIBE %s", test.name)})
		require.NoError(t, err)
		colCount := 0
		for rows.Next() {
			colCount++
		}
		require.Equal(t, test.colCount, colCount)
	}

}

func TestIterativeJSONIngestionWithVariableSchema(t *testing.T) {
	tempDir := t.TempDir()
	file1 := filepath.Join(tempDir, "data1.ndjson")
	temp, err := os.Create(file1)
	require.NoError(t, err)
	_, err = temp.WriteString(`{"id":1, "city":"bglr"}
{"id":2, "city":"mum"}`)
	require.NoError(t, err)

	file2 := filepath.Join(tempDir, "data2.ndjson")
	temp, err = os.Create(file2)
	require.NoError(t, err)
	_, err = temp.WriteString(`{"id":3, "city":"bglr", "country":"IND"}
{"id":4, "city":"mum","country":"IND"}`)
	require.NoError(t, err)

	file3 := filepath.Join(tempDir, "data3.ndjson")
	temp, err = os.Create(file3)
	require.NoError(t, err)
	_, err = temp.WriteString(`{"city":"bglr", "id":3}
{"city":"mum","id":4}`)
	require.NoError(t, err)

	file4 := filepath.Join(tempDir, "data4.ndjson")
	temp, err = os.Create(file4)
	require.NoError(t, err)
	_, err = temp.WriteString(`{"city":"bglr", "id":3.2}
{"city":"mum","id":4.5}`)
	require.NoError(t, err)

	type test struct {
		mockIterator mockIterator
		name         string
		count        int
		filterCount  int
		colCount     int
	}

	tests := []test{
		{
			mockIterator: mockIterator{batches: [][]string{
				{file1, file1},
				{file1, file1},
			}},
			name:        "same_schema",
			count:       8,
			filterCount: 4,
			colCount:    2,
		},
		{
			mockIterator: mockIterator{batches: [][]string{
				{file1, file2, file3, file4},
			}},
			name:        "variable_schema_ingested_at_once",
			count:       8,
			filterCount: 4,
			colCount:    3,
		},
		{
			mockIterator: mockIterator{batches: [][]string{
				{file1},
				{file2},
				{file3},
				{file4},
			}},
			name:        "changing_schema",
			count:       8,
			filterCount: 4,
			colCount:    3,
		},
		{
			mockIterator: mockIterator{batches: [][]string{
				{file1},
				{file3, file1},
				{file2, file1},
				{file1, file4},
			}},
			name:        "schema_combination",
			count:       14,
			filterCount: 7,
			colCount:    3,
		},
	}

	mockConnector := &mockConnector{}
	connectors.Register("mock-json", mockConnector)
	for _, test := range tests {
		mockConnector.mockIterator = &test.mockIterator
		olap := runOLAPStore(t)
		ctx := context.Background()

		_, err = olap.Ingest(ctx, &connectors.Env{
			RepoDriver:      "file",
			RepoRoot:        ".",
			AllowHostAccess: true,
		}, &connectors.Source{
			Name:      test.name,
			Connector: "mock-json",
			Properties: map[string]any{
				"path": filepath.Join(tempDir, "*.csv"),
			},
		})
		require.NoError(t, err, "no err expected test %s", test.name)

		var count int
		rows, err := olap.Execute(ctx, &drivers.Statement{Query: fmt.Sprintf("SELECT count(*) FROM %s", test.name)})
		require.NoError(t, err)
		require.True(t, rows.Next())
		require.NoError(t, rows.Scan(&count))
		require.Equal(t, test.count, count)
		require.NoError(t, rows.Close())

		rows, err = olap.Execute(ctx, &drivers.Statement{Query: fmt.Sprintf("SELECT count(*) FROM %s where city='bglr'", test.name)})
		require.NoError(t, err)
		require.True(t, rows.Next())
		require.NoError(t, rows.Scan(&count))
		require.Equal(t, test.filterCount, count)
		require.NoError(t, rows.Close())

		rows, err = olap.Execute(ctx, &drivers.Statement{Query: fmt.Sprintf("DESCRIBE %s", test.name)})
		require.NoError(t, err)
		colCount := 0
		for rows.Next() {
			colCount++
		}
		require.Equal(t, test.colCount, colCount)
	}
}

func createFilePath(t *testing.T, dirPath string, fileName string) string {
	testdataPathAbs, err := filepath.Abs(dirPath)
	require.NoError(t, err)
	filePath := filepath.Join(testdataPathAbs, fileName)
	return filePath
}

func runOLAPStore(t *testing.T) drivers.OLAPStore {
	conn, err := duckdb.Driver{}.Open("?access_mode=read_write", zap.NewNop())
	require.NoError(t, err)
	olap, canServe := conn.OLAPStore()
	require.True(t, canServe)
	return olap
}
