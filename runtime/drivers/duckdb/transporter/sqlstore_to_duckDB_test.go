package transporter_test

import (
	"context"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/drivers/duckdb/transporter"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// Mock SQLStore implementation for testing
type mockSQLStore struct{}

func (m *mockSQLStore) Query(ctx context.Context, props map[string]any, qry string) (drivers.RowIterator, error) {
	// Return a mock iterator
	return &mockRowIterator{}, nil
}

func (m *mockSQLStore) WithRaw(ctx context.Context, priority int, fn func(any) error) error {
	return nil
}

// Mock Iterator implementation for testing
type mockRowIterator struct {
	count int
}

func (m *mockRowIterator) Next(ctx context.Context) ([]any, error) {
	if m.count == 10 { // send ten rows and stop
		return nil, drivers.ErrIteratorDone
	}
	m.count++
	return []any{"value1", m.count}, nil
}

func (m *mockRowIterator) Close() error {
	return nil
}

func (m *mockRowIterator) Schema(ctx context.Context) (*runtimev1.StructType, error) {
	// Return a mock schema with two columns
	return &runtimev1.StructType{
		Fields: []*runtimev1.StructType_Field{
			{Name: "col1", Type: &runtimev1.Type{Code: runtimev1.Type_CODE_STRING}},
			{Name: "col2", Type: &runtimev1.Type{Code: runtimev1.Type_CODE_INT64}},
		},
	}, nil
}

func (m *mockRowIterator) Size(unit drivers.ProgressUnit) (uint64, bool) {
	return 10, true
}

// Mock Progress implementation for testing
type mockProgress struct {
	target int64
	count  int64
}

func (m *mockProgress) Target(total int64, unit drivers.ProgressUnit) {
	m.target = total
}

func (m *mockProgress) Observe(delta int64, unit drivers.ProgressUnit) {
	m.count += delta
}

func TestTransfer(t *testing.T) {
	logger := zap.NewNop()

	fromStore := &mockSQLStore{}
	olap := runOLAPStore(t)

	// Create a sqlStoreToDuckDB transporter for testing
	transporter := transporter.NewSQLStoreToDuckDB(fromStore, olap, logger)

	// Create mock Source and Sink for testing
	source := &drivers.DatabaseSource{}
	sink := &drivers.DatabaseSink{Table: "test_table"}

	ctx := context.Background()
	p := &mockProgress{}
	// Run the Transfer function and check for errors
	err := transporter.Transfer(ctx, source, sink, &drivers.TransferOpts{}, p)
	if err != nil {
		t.Errorf("Transfer function returned an error: %v", err)
	}
	require.Equal(t, int64(10), p.target)
	require.Equal(t, int64(10), p.count)

	rows, err := olap.Execute(ctx, &drivers.Statement{Query: "SELECT count(*) FROM test_table where col1 = 'value1'"})
	require.NoError(t, err)
	require.True(t, rows.Next())
	var count int
	require.NoError(t, rows.Scan(&count))
	require.Equal(t, 10, count)
	require.NoError(t, rows.Close())

	rows, err = olap.Execute(ctx, &drivers.Statement{Query: "SELECT count(*) FROM test_table where col1 = 'value1' and col2 = 5"})
	require.NoError(t, err)
	require.True(t, rows.Next())
	require.NoError(t, rows.Scan(&count))
	require.Equal(t, 1, count)
	require.NoError(t, rows.Close())
}
