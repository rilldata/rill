package duckdb

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/storage"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func Test_connection_CreateTableAsSelect(t *testing.T) {
	temp := t.TempDir()

	handle, err := Driver{}.Open("default", map[string]any{}, storage.MustNew(temp, nil), activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)
	viewConnection := handle.(*connection)
	require.NoError(t, viewConnection.Migrate(context.Background()))
	viewConnection.AsOLAP("default")

	tests := []struct {
		testName    string
		name        string
		view        bool
		tableAsView bool
		c           *connection
	}{
		{
			testName: "select from view with external_table_storage",
			name:     "my-view",
			c:        viewConnection,
			view:     true,
		},
		{
			testName:    "select from table with external_table_storage",
			name:        "my-table",
			c:           viewConnection,
			tableAsView: true,
		},
	}
	ctx := context.Background()
	sql := "SELECT 1"
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			_, err := tt.c.CreateTableAsSelect(ctx, tt.name, sql, &drivers.CreateTableOptions{View: tt.view})
			require.NoError(t, err)
			res, err := tt.c.Execute(ctx, &drivers.Statement{Query: fmt.Sprintf("SELECT count(*) FROM %q", tt.name)})
			require.NoError(t, err)
			require.True(t, res.Next())
			var count int
			require.NoError(t, res.Scan(&count))
			require.Equal(t, 1, count)
			require.NoError(t, res.Close())

			if tt.tableAsView {
				res, err := tt.c.Execute(ctx, &drivers.Statement{Query: fmt.Sprintf("SELECT count(*) FROM information_schema.tables WHERE table_name='%s' AND table_type='VIEW'", tt.name)})
				require.NoError(t, err)
				require.True(t, res.Next())
				var count int
				require.NoError(t, res.Scan(&count))
				require.Equal(t, 1, count)
				require.NoError(t, res.Close())
			}
		})
	}
}

func Test_connection_CreateTableAsSelectMultipleTimes(t *testing.T) {
	temp := t.TempDir()

	handle, err := Driver{}.Open("default", map[string]any{}, storage.MustNew(temp, nil), activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)
	c := handle.(*connection)
	require.NoError(t, c.Migrate(context.Background()))
	c.AsOLAP("default")

	_, err = c.CreateTableAsSelect(context.Background(), "test-select-multiple", "select 1", &drivers.CreateTableOptions{})
	require.NoError(t, err)
	time.Sleep(2 * time.Millisecond)
	_, err = c.CreateTableAsSelect(context.Background(), "test-select-multiple", "select 'hello'", &drivers.CreateTableOptions{})
	require.NoError(t, err)

	_, err = c.CreateTableAsSelect(context.Background(), "test-select-multiple", "select fail query", &drivers.CreateTableOptions{})
	require.Error(t, err)

	res, err := c.Execute(context.Background(), &drivers.Statement{Query: fmt.Sprintf("SELECT * FROM %q", "test-select-multiple")})
	require.NoError(t, err)
	require.True(t, res.Next())
	var name string
	require.NoError(t, res.Scan(&name))
	require.Equal(t, "hello", name)
	require.False(t, res.Next())
	require.NoError(t, res.Close())
}

func Test_connection_DropTable(t *testing.T) {
	temp := t.TempDir()

	handle, err := Driver{}.Open("default", map[string]any{}, storage.MustNew(temp, nil), activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)
	c := handle.(*connection)
	require.NoError(t, c.Migrate(context.Background()))
	c.AsOLAP("default")

	_, err = c.CreateTableAsSelect(context.Background(), "test-drop", "select 1", &drivers.CreateTableOptions{})
	require.NoError(t, err)

	err = c.DropTable(context.Background(), "test-drop")
	require.NoError(t, err)

	res, err := c.Execute(context.Background(), &drivers.Statement{Query: "SELECT count(*) FROM information_schema.tables WHERE table_name='test-drop' AND table_type='VIEW'"})
	require.NoError(t, err)
	require.True(t, res.Next())
	var count int
	require.NoError(t, res.Scan(&count))
	require.Equal(t, 0, count)
	require.NoError(t, res.Close())
}

func Test_connection_InsertTableAsSelect_WithAppendStrategy(t *testing.T) {
	temp := t.TempDir()

	handle, err := Driver{}.Open("default", map[string]any{}, storage.MustNew(temp, nil), activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)
	c := handle.(*connection)
	require.NoError(t, c.Migrate(context.Background()))
	c.AsOLAP("default")

	_, err = c.CreateTableAsSelect(context.Background(), "test-insert", "select 1", &drivers.CreateTableOptions{})
	require.NoError(t, err)

	opts := &drivers.InsertTableOptions{
		ByName:   false,
		InPlace:  true,
		Strategy: drivers.IncrementalStrategyAppend,
	}
	_, err = c.InsertTableAsSelect(context.Background(), "test-insert", "select 2", opts)
	require.NoError(t, err)

	opts = &drivers.InsertTableOptions{
		ByName:   true,
		InPlace:  true,
		Strategy: drivers.IncrementalStrategyAppend,
	}
	_, err = c.InsertTableAsSelect(context.Background(), "test-insert", "select 3", opts)
	require.Error(t, err)

	res, err := c.Execute(context.Background(), &drivers.Statement{Query: "SELECT count(*) FROM 'test-insert'"})
	require.NoError(t, err)
	require.True(t, res.Next())
	var count int
	require.NoError(t, res.Scan(&count))
	require.Equal(t, 2, count)
	require.NoError(t, res.Close())
}

func Test_connection_InsertTableAsSelect_WithMergeStrategy(t *testing.T) {
	temp := t.TempDir()

	dbPath := filepath.Join(temp, "view.db")
	handle, err := Driver{}.Open("default", map[string]any{"path": dbPath, "external_table_storage": true}, storage.MustNew(temp, nil), activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)
	c := handle.(*connection)
	require.NoError(t, c.Migrate(context.Background()))
	c.AsOLAP("default")

	_, err = c.CreateTableAsSelect(context.Background(), "test-merge", "SELECT range, 'insert' AS strategy FROM range(0, 4)", &drivers.CreateTableOptions{})
	require.NoError(t, err)

	opts := &drivers.InsertTableOptions{
		ByName:    false,
		InPlace:   true,
		Strategy:  drivers.IncrementalStrategyMerge,
		UniqueKey: []string{"range"},
	}
	_, err = c.InsertTableAsSelect(context.Background(), "test-merge", "SELECT range, 'merge' AS strategy FROM range(2, 4)", opts)
	require.NoError(t, err)

	res, err := c.Execute(context.Background(), &drivers.Statement{Query: "SELECT range, strategy FROM 'test-merge' ORDER BY range"})
	require.NoError(t, err)

	var results []struct {
		Range    int
		Strategy string
	}
	for res.Next() {
		var r struct {
			Range    int
			Strategy string
		}
		require.NoError(t, res.Scan(&r.Range, &r.Strategy))
		results = append(results, r)
	}
	require.NoError(t, res.Close())

	exptected := []struct {
		Range    int
		Strategy string
	}{
		{0, "insert"},
		{1, "insert"},
		{2, "merge"},
		{3, "merge"},
	}
	require.Equal(t, exptected, results)
}

func Test_connection_RenameTable(t *testing.T) {
	temp := t.TempDir()

	handle, err := Driver{}.Open("default", map[string]any{}, storage.MustNew(temp, nil), activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)
	c := handle.(*connection)
	require.NoError(t, c.Migrate(context.Background()))
	c.AsOLAP("default")

	_, err = c.CreateTableAsSelect(context.Background(), "test-rename", "select 1", &drivers.CreateTableOptions{})
	require.NoError(t, err)

	err = c.RenameTable(context.Background(), "test-rename", "rename-test")
	require.NoError(t, err)

	res, err := c.Execute(context.Background(), &drivers.Statement{Query: "SELECT count(*) FROM 'rename-test'"})
	require.NoError(t, err)
	require.True(t, res.Next())
	var count int
	require.NoError(t, res.Scan(&count))
	require.Equal(t, 1, count)
	require.NoError(t, res.Close())
}

func Test_connection_RenameToExistingTable(t *testing.T) {
	temp := t.TempDir()
	handle, err := Driver{}.Open("default", map[string]any{}, storage.MustNew(temp, nil), activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)
	c := handle.(*connection)
	require.NoError(t, c.Migrate(context.Background()))
	c.AsOLAP("default")

	_, err = c.CreateTableAsSelect(context.Background(), "source", "SELECT 1 AS data", &drivers.CreateTableOptions{})
	require.NoError(t, err)

	_, err = c.CreateTableAsSelect(context.Background(), "_tmp_source", "SELECT 2 AS DATA", &drivers.CreateTableOptions{})
	require.NoError(t, err)

	err = c.RenameTable(context.Background(), "_tmp_source", "source")
	require.NoError(t, err)

	res, err := c.Execute(context.Background(), &drivers.Statement{Query: "SELECT * FROM 'source'"})
	require.NoError(t, err)
	require.True(t, res.Next())
	var num int
	require.NoError(t, res.Scan(&num))
	require.Equal(t, 2, num)
	require.NoError(t, res.Close())
}

func Test_connection_AddTableColumn(t *testing.T) {
	temp := t.TempDir()
	os.Mkdir(temp, fs.ModePerm)

	handle, err := Driver{}.Open("default", map[string]any{}, storage.MustNew(temp, nil), activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)
	c := handle.(*connection)
	require.NoError(t, c.Migrate(context.Background()))
	c.AsOLAP("default")

	_, err = c.CreateTableAsSelect(context.Background(), "test alter column", "select 1 as data", &drivers.CreateTableOptions{})
	require.NoError(t, err)

	res, err := c.Execute(context.Background(), &drivers.Statement{Query: "SELECT data_type FROM information_schema.columns WHERE table_name='test alter column'"})
	require.NoError(t, err)
	require.True(t, res.Next())
	var typ string
	require.NoError(t, res.Scan(&typ))
	require.Equal(t, "INTEGER", typ)
	require.NoError(t, res.Close())

	err = c.AlterTableColumn(context.Background(), "test alter column", "data", "VARCHAR")
	require.NoError(t, err)

	res, err = c.Execute(context.Background(), &drivers.Statement{Query: "SELECT data_type FROM information_schema.columns WHERE table_name='test alter column' AND table_schema=current_schema()"})
	require.NoError(t, err)
	require.True(t, res.Next())
	require.NoError(t, res.Scan(&typ))
	require.Equal(t, "VARCHAR", typ)
	require.NoError(t, res.Close())
}

func Test_connection_RenameToExistingTableOld(t *testing.T) {
	handle, err := Driver{}.Open("default", map[string]any{}, storage.MustNew(t.TempDir(), nil), activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)
	c := handle.(*connection)
	require.NoError(t, c.Migrate(context.Background()))
	c.AsOLAP("default")

	_, err = c.CreateTableAsSelect(context.Background(), "source", "SELECT 1 AS data", &drivers.CreateTableOptions{})
	require.NoError(t, err)

	_, err = c.CreateTableAsSelect(context.Background(), "_tmp_source", "SELECT 2 AS DATA", &drivers.CreateTableOptions{})
	require.NoError(t, err)

	err = c.RenameTable(context.Background(), "_tmp_source", "source")
	require.NoError(t, err)

	res, err := c.Execute(context.Background(), &drivers.Statement{Query: "SELECT * FROM 'source'"})
	require.NoError(t, err)
	require.True(t, res.Next())
	var num int
	require.NoError(t, res.Scan(&num))
	require.Equal(t, 2, num)
	require.NoError(t, res.Close())
}

func Test_connection_CreateTableAsSelectWithComments(t *testing.T) {
	temp := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(temp, "default"), fs.ModePerm))
	handle, err := Driver{}.Open("default", map[string]any{}, storage.MustNew(temp, nil), activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)
	normalConn := handle.(*connection)
	normalConn.AsOLAP("default")
	require.NoError(t, normalConn.Migrate(context.Background()))

	ctx := context.Background()
	sql := `
		-- lets write a stupid query
		with r as (select 1 as id ) 	
		select * from r
		-- that was a stupid query
		-- I hope to write not so stupid query
	`
	_, err = normalConn.CreateTableAsSelect(ctx, "test", sql, &drivers.CreateTableOptions{})
	require.NoError(t, err)

	_, err = normalConn.CreateTableAsSelect(ctx, "test_view", sql, &drivers.CreateTableOptions{View: true})
	require.NoError(t, err)

	sql = `
		with r as (select 1 as id ) 	
		select * from r
	`
	_, err = normalConn.CreateTableAsSelect(ctx, "test", sql, &drivers.CreateTableOptions{})
	require.NoError(t, err)

	_, err = normalConn.CreateTableAsSelect(ctx, "test_view", sql, &drivers.CreateTableOptions{View: true})
	require.NoError(t, err)
}

func verifyCount(t *testing.T, c *connection, table string, expected int) {
	res, err := c.Execute(context.Background(), &drivers.Statement{Query: fmt.Sprintf("SELECT count(*) from %s", table)})
	require.NoError(t, err)
	require.True(t, res.Next())
	var count int
	require.NoError(t, res.Scan(&count))
	require.Equal(t, expected, count)
	require.NoError(t, res.Close())
}
