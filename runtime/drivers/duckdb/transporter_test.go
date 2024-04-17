package duckdb_test

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/drivers/duckdb"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type mockObjectStore struct {
	mockIterator drivers.FileIterator
}

func (m *mockObjectStore) DownloadFiles(ctx context.Context, srcProps map[string]any) (drivers.FileIterator, error) {
	return m.mockIterator, nil
}

type mockIterator struct {
	batches [][]string
	index   int
}

func (m *mockIterator) Close() error {
	return nil
}

func (m *mockIterator) Next() ([]string, error) {
	if m.index == len(m.batches) {
		return nil, io.EOF
	}
	m.index += 1
	return m.batches[m.index-1], nil
}

func (m *mockIterator) Size(unit drivers.ProgressUnit) (int64, bool) {
	return 0, false
}

func (m *mockIterator) Format() string {
	return ""
}

var _ drivers.FileIterator = &mockIterator{}

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
		files       [][]string
		query       bool
		name        string
		count       int
		filterCount int
		colCount    int
	}

	tests := []test{
		{
			files: [][]string{
				{file1, file1},
				{file1, file1},
			},
			name:        "same_schema",
			count:       8,
			filterCount: 4,
			colCount:    2,
		},
		{
			files: [][]string{
				{file1, file2, file3, file4},
			},
			name:        "variable_schema_ingested_at_once",
			count:       8,
			filterCount: 4,
			colCount:    3,
		},
		{
			files: [][]string{
				{file1},
				{file2},
				{file3},
				{file4},
			},
			name:        "changing_schema",
			count:       8,
			filterCount: 4,
			colCount:    3,
		},
		{
			files: [][]string{
				{file1},
				{file3, file1},
				{file2, file1},
				{file1, file4},
			},
			name:        "schema_combination",
			count:       14,
			filterCount: 7,
			colCount:    3,
		},
	}
	queryTests := make([]test, len(tests))
	for i, t := range tests {
		queryTests[i] = test{
			files:       t.files,
			query:       true,
			name:        t.name,
			count:       t.count,
			filterCount: t.filterCount,
			colCount:    t.colCount,
		}
	}
	tests = append(tests, queryTests...)

	mockConnector := &mockObjectStore{}
	for _, test := range tests {
		t.Run(fmt.Sprintf("%s - query=%v", test.name, test.query), func(t *testing.T) {
			mockConnector.mockIterator = &mockIterator{batches: test.files}
			olap := runOLAPStore(t)
			ctx := context.Background()
			tr := duckdb.NewObjectStoreToDuckDB(mockConnector, olap, zap.NewNop())

			var src map[string]any
			if test.query {
				src = map[string]any{"sql": "select * from read_csv_auto('path',union_by_name=true,sample_size=200000)", "allow_schema_relaxation": true}
			} else {
				src = map[string]any{"allow_schema_relaxation": true}
			}

			err = tr.Transfer(ctx, src, map[string]any{"table": test.name}, mockTransferOptions())
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
			require.NoError(t, rows.Close())
			require.Equal(t, test.colCount, colCount)
		})
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
		files    [][]string
		query    bool
		name     string
		hasError bool
	}

	tests := []test{
		{
			files: [][]string{
				{file1, file1},
				{file1, file1},
			},
			name:     "same_schema",
			hasError: false,
		},
		{
			files: [][]string{
				{file1, file2, file3, file4},
			},
			name:     "variable_schema_ingested_at_once",
			hasError: true,
		},
		{
			files: [][]string{
				{file1, file2, file3, file4},
			},
			name:     "columns_jumbled",
			hasError: true,
		},
		{
			files: [][]string{
				{file1},
				{file2},
			},
			name:     "new_columns",
			hasError: true,
		},
		{
			files: [][]string{
				{file2},
				{file1},
			},
			name:     "less_columns",
			hasError: true,
		},
		{
			files: [][]string{
				{file1},
				{file4},
			},
			name:     "datatype_change",
			hasError: true,
		},
	}
	queryTests := make([]test, len(tests))
	for i, t := range tests {
		queryTests[i] = test{
			files:    t.files,
			query:    true,
			name:     t.name,
			hasError: t.hasError,
		}
	}
	tests = append(tests, queryTests...)

	mockConnector := &mockObjectStore{}
	for i, test := range tests {
		t.Run(fmt.Sprintf("%d - query=%v", i, test.query), func(t *testing.T) {
			mockConnector.mockIterator = &mockIterator{batches: test.files}
			olap := runOLAPStore(t)
			ctx := context.Background()
			tr := duckdb.NewObjectStoreToDuckDB(mockConnector, olap, zap.NewNop())

			var src map[string]any
			if test.query {
				src = map[string]any{"sql": "select * from read_csv_auto('path')"}
			} else {
				src = map[string]any{}
			}

			err = tr.Transfer(ctx, src, map[string]any{"table": test.name}, mockTransferOptions())
			if test.hasError {
				require.Error(t, err, fmt.Errorf("error expected for %s got nil", test.name))
			} else {
				require.NoError(t, err, fmt.Errorf("no error expected for %s got %s", test.name, err))
			}
		})
	}
}

func TestIterativeParquetIngestionWithVariableSchema(t *testing.T) {
	file1 := filepath.Join("../../testruntime/testdata/variable-schema", "data.parquet")
	file2 := filepath.Join("../../testruntime/testdata/variable-schema", "data1.parquet")
	file3 := filepath.Join("../../testruntime/testdata/variable-schema", "data2.parquet")
	file4 := filepath.Join("../../testruntime/testdata/variable-schema", "data3.parquet")

	type test struct {
		files       [][]string
		query       bool
		name        string
		count       int
		filterCount int
		colCount    int
	}

	tests := []test{
		{
			files: [][]string{
				{file1, file1},
				{file1, file1},
			},
			name:        "same_schema",
			count:       8,
			filterCount: 4,
			colCount:    2,
		},
		{
			files: [][]string{
				{file1, file2, file3, file4},
			},
			name:        "variable_schema_ingested_at_once",
			count:       8,
			filterCount: 4,
			colCount:    3,
		},
		{
			files: [][]string{
				{file1},
				{file2},
				{file3},
				{file4},
			},
			name:        "changing_schema",
			count:       8,
			filterCount: 4,
			colCount:    3,
		},
		{
			files: [][]string{
				{file1},
				{file3, file1},
				{file2, file1},
				{file1, file4},
			},
			name:        "schema_combination",
			count:       14,
			filterCount: 7,
			colCount:    3,
		},
	}
	queryTests := make([]test, len(tests))
	for i, t := range tests {
		queryTests[i] = test{
			files:       t.files,
			query:       true,
			name:        t.name,
			count:       t.count,
			filterCount: t.filterCount,
			colCount:    t.colCount,
		}
	}
	tests = append(tests, queryTests...)

	mockConnector := &mockObjectStore{}
	for i, test := range tests {
		t.Run(fmt.Sprintf("%d - query=%v", i, test.query), func(t *testing.T) {
			mockConnector.mockIterator = &mockIterator{batches: test.files}
			olap := runOLAPStore(t)
			ctx := context.Background()
			tr := duckdb.NewObjectStoreToDuckDB(mockConnector, olap, zap.NewNop())

			var src map[string]any
			if test.query {
				src = map[string]any{"sql": "select * from read_parquet('path',union_by_name=true,hive_partitioning=true)", "allow_schema_relaxation": true}
			} else {
				src = map[string]any{"allow_schema_relaxation": true}
			}

			err := tr.Transfer(ctx, src, map[string]any{"table": test.name}, mockTransferOptions())
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
			require.NoError(t, rows.Close())
			require.Equal(t, test.colCount, colCount)
		})
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
		files       [][]string
		query       bool
		name        string
		count       int
		filterCount int
		colCount    int
	}

	tests := []test{
		{
			files: [][]string{
				{file1, file1},
				{file1, file1},
			},
			name:        "same_schema",
			count:       8,
			filterCount: 4,
			colCount:    2,
		},
		{
			files: [][]string{
				{file1, file2, file3, file4},
			},
			name:        "variable_schema_ingested_at_once",
			count:       8,
			filterCount: 4,
			colCount:    3,
		},
		{
			files: [][]string{
				{file1},
				{file2},
				{file3},
				{file4},
			},
			name:        "changing_schema",
			count:       8,
			filterCount: 4,
			colCount:    3,
		},
		{
			files: [][]string{
				{file1},
				{file3, file1},
				{file2, file1},
				{file1, file4},
			},
			name:        "schema_combination",
			count:       14,
			filterCount: 7,
			colCount:    3,
		},
	}
	queryTests := make([]test, len(tests))
	for i, t := range tests {
		queryTests[i] = test{
			files:       t.files,
			query:       true,
			name:        t.name,
			count:       t.count,
			filterCount: t.filterCount,
			colCount:    t.colCount,
		}
	}
	tests = append(tests, queryTests...)

	mockConnector := &mockObjectStore{}
	for _, test := range tests {
		t.Run(fmt.Sprintf("%s - query=%v", test.name, test.query), func(t *testing.T) {
			m := &mockIterator{batches: test.files}
			mockConnector.mockIterator = m
			olap := runOLAPStore(t)
			ctx := context.Background()
			tr := duckdb.NewObjectStoreToDuckDB(mockConnector, olap, zap.NewNop())

			var src map[string]any
			if test.query {
				files := make([]string, 0)
				for _, f := range test.files {
					files = append(files, f...)
				}
				m.batches = [][]string{files}
				src = map[string]any{"sql": "select * from read_json('path',format='auto',union_by_name=true,auto_detect=true,sample_size=200000)", "batch_size": "-1"}
			} else {
				src = map[string]any{"allow_schema_relaxation": true}
			}

			err := tr.Transfer(ctx, src, map[string]any{"table": test.name}, mockTransferOptions())
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
			require.NoError(t, rows.Close())
			require.Equal(t, test.colCount, colCount)
		})
	}
}

func runOLAPStore(t *testing.T) drivers.OLAPStore {
	conn, err := drivers.Open("duckdb", "default", map[string]any{"dsn": ":memory:?access_mode=read_write"}, activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)
	olap, canServe := conn.AsOLAP("")
	require.True(t, canServe)
	return olap
}

func mockTransferOptions() *drivers.TransferOptions {
	return &drivers.TransferOptions{
		AllowHostAccess: true,
		Progress:        drivers.NoOpProgress{},
		AcquireConnector: func(name string) (drivers.Handle, func(), error) {
			return nil, nil, fmt.Errorf("not found")
		},
	}
}
