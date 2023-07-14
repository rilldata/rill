package transporter_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/rilldata/rill/runtime/drivers"
	_ "github.com/rilldata/rill/runtime/drivers/duckdb"
	"github.com/rilldata/rill/runtime/drivers/duckdb/transporter"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type mockObjectStore struct {
	mockIterator drivers.FileIterator
}

func (m *mockObjectStore) DownloadFiles(ctx context.Context, src *drivers.BucketSource) (drivers.FileIterator, error) {
	return m.mockIterator, nil
}

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

func (m *mockIterator) Size(unit drivers.ProgressUnit) (int64, bool) {
	return 0, false
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

	mockConnector := &mockObjectStore{}
	for _, test := range tests {
		mockConnector.mockIterator = &test.mockIterator
		olap := runOLAPStore(t)
		ctx := context.Background()
		tr := transporter.NewObjectStoreToDuckDB(mockConnector, olap, zap.NewNop())

		err = tr.Transfer(ctx, &drivers.BucketSource{}, &drivers.DatabaseSink{Table: test.name}, drivers.NewTransferOpts(),
			drivers.NoOpProgress{})
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
			hasError: true,
		},
		{
			mockIterator: mockIterator{batches: [][]string{
				{file1, file2, file3, file4},
			}},
			name:     "columns_jumbled",
			hasError: true,
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

	mockConnector := &mockObjectStore{}
	for _, test := range tests {

		mockConnector.mockIterator = &test.mockIterator
		olap := runOLAPStore(t)
		ctx := context.Background()
		tr := transporter.NewObjectStoreToDuckDB(mockConnector, olap, zap.NewNop())

		err = tr.Transfer(ctx, &drivers.BucketSource{Properties: map[string]any{"allow_schema_relaxation": false}},
			&drivers.DatabaseSink{Table: test.name}, drivers.NewTransferOpts(),
			drivers.NoOpProgress{})
		if test.hasError {
			require.Error(t, err, fmt.Errorf("error expected for %s got nil", test.name))
		} else {
			require.NoError(t, err, fmt.Errorf("no error expected for %s got %s", test.name, err))
		}
	}

}

func TestIterativeParquetIngestionWithVariableSchema(t *testing.T) {
	file1 := filepath.Join("../../../testruntime/testdata/variable-schema", "data.parquet")
	file2 := filepath.Join("../../../testruntime/testdata/variable-schema", "data1.parquet")
	file3 := filepath.Join("../../../testruntime/testdata/variable-schema", "data2.parquet")
	file4 := filepath.Join("../../../testruntime/testdata/variable-schema", "data3.parquet")

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

	mockConnector := &mockObjectStore{}
	for _, test := range tests {
		mockConnector.mockIterator = &test.mockIterator
		olap := runOLAPStore(t)
		ctx := context.Background()
		tr := transporter.NewObjectStoreToDuckDB(mockConnector, olap, zap.NewNop())

		err := tr.Transfer(ctx, &drivers.BucketSource{}, &drivers.DatabaseSink{Table: test.name},
			drivers.NewTransferOpts(), drivers.NoOpProgress{})
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

	mockConnector := &mockObjectStore{}
	for _, test := range tests {
		mockConnector.mockIterator = &test.mockIterator
		olap := runOLAPStore(t)
		ctx := context.Background()
		tr := transporter.NewObjectStoreToDuckDB(mockConnector, olap, zap.NewNop())

		err := tr.Transfer(ctx, &drivers.BucketSource{}, &drivers.DatabaseSink{Table: test.name},
			drivers.NewTransferOpts(), drivers.NoOpProgress{})
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

func runOLAPStore(t *testing.T) drivers.OLAPStore {
	conn, err := drivers.Open("duckdb", map[string]any{"dsn": "?access_mode=read_write"}, zap.NewNop())
	require.NoError(t, err)
	olap, canServe := conn.AsOLAP()
	require.True(t, canServe)
	return olap
}
