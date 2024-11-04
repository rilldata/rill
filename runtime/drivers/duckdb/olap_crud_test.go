package duckdb

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func Test_connection_CreateTableAsSelect(t *testing.T) {
	temp := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(temp, "default"), fs.ModePerm))
	dbPath := filepath.Join(temp, "default", "normal.db")
	handle, err := Driver{}.Open("default", map[string]any{"path": dbPath, "external_table_storage": false}, activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)
	normalConn := handle.(*connection)
	normalConn.AsOLAP("default")
	require.NoError(t, normalConn.Migrate(context.Background()))

	dbPath = filepath.Join(temp, "default")
	handle, err = Driver{}.Open("default", map[string]any{"data_dir": dbPath, "external_table_storage": true}, activity.NewNoopClient(), zap.NewNop())
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
			testName: "select from view",
			name:     "my-view",
			view:     true,
			c:        normalConn,
		},
		{
			testName: "select from table",
			name:     "my-table",
			c:        normalConn,
		},
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
			err := tt.c.CreateTableAsSelect(ctx, tt.name, tt.view, sql, nil)
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
				contents, err := os.ReadFile(filepath.Join(temp, "default", "read", tt.name, "version.txt"))
				require.NoError(t, err)
				version, err := strconv.ParseInt(string(contents), 10, 64)
				require.NoError(t, err)
				require.True(t, time.Since(time.UnixMilli(version)).Seconds() < 1)
			}
		})
	}
}

func Test_connection_CreateTableAsSelectMultipleTimes(t *testing.T) {
	temp := t.TempDir()

	handle, err := Driver{}.Open("default", map[string]any{"data_dir": temp, "external_table_storage": true}, activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)
	c := handle.(*connection)
	require.NoError(t, c.Migrate(context.Background()))
	c.AsOLAP("default")

	err = c.CreateTableAsSelect(context.Background(), "test-select-multiple", false, "select 1", nil)
	require.NoError(t, err)
	time.Sleep(2 * time.Millisecond)
	err = c.CreateTableAsSelect(context.Background(), "test-select-multiple", false, "select 'hello'", nil)
	require.NoError(t, err)

	dirs, err := os.ReadDir(filepath.Join(temp, "read", "test-select-multiple"))
	require.NoError(t, err)
	names := make([]string, 0)
	for _, dir := range dirs {
		names = append(names, dir.Name())
	}

	err = c.CreateTableAsSelect(context.Background(), "test-select-multiple", false, "select fail query", nil)
	require.Error(t, err)

	dirs, err = os.ReadDir(filepath.Join(temp, "read", "test-select-multiple"))
	require.NoError(t, err)
	newNames := make([]string, 0)
	for _, dir := range dirs {
		newNames = append(newNames, dir.Name())
	}

	require.Equal(t, names, newNames)

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

	handle, err := Driver{}.Open("default", map[string]any{"data_dir": temp, "external_table_storage": true}, activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)
	c := handle.(*connection)
	require.NoError(t, c.Migrate(context.Background()))
	c.AsOLAP("default")

	err = c.CreateTableAsSelect(context.Background(), "test-drop", false, "select 1", nil)
	require.NoError(t, err)

	// Note: true since at lot of places we look at information_schema lookup on main db to determine whether tbl is a view or table
	err = c.DropTable(context.Background(), "test-drop", true)
	require.NoError(t, err)

	_, err = os.ReadDir(filepath.Join(temp, "read", "test-drop"))
	require.True(t, os.IsNotExist(err))

	res, err := c.Execute(context.Background(), &drivers.Statement{Query: "SELECT count(*) FROM information_schema.tables WHERE table_name='test-drop' AND table_type='VIEW'"})
	require.NoError(t, err)
	require.True(t, res.Next())
	var count int
	require.NoError(t, res.Scan(&count))
	require.Equal(t, 0, count)
	require.NoError(t, res.Close())
}

func Test_connection_InsertTableAsSelect(t *testing.T) {
	temp := t.TempDir()

	handle, err := Driver{}.Open("default", map[string]any{"data_dir": temp, "external_table_storage": true}, activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)
	c := handle.(*connection)
	require.NoError(t, c.Migrate(context.Background()))
	c.AsOLAP("default")

	err = c.CreateTableAsSelect(context.Background(), "test-insert", false, "select 1", nil)
	require.NoError(t, err)

	err = c.InsertTableAsSelect(context.Background(), "test-insert", "select 2", false, true, drivers.IncrementalStrategyAppend, nil)
	require.NoError(t, err)

	err = c.InsertTableAsSelect(context.Background(), "test-insert", "select 3", true, true, drivers.IncrementalStrategyAppend, nil)
	require.Error(t, err)

	res, err := c.Execute(context.Background(), &drivers.Statement{Query: "SELECT count(*) FROM 'test-insert'"})
	require.NoError(t, err)
	require.True(t, res.Next())
	var count int
	require.NoError(t, res.Scan(&count))
	require.Equal(t, 2, count)
	require.NoError(t, res.Close())
}

func Test_connection_RenameTable(t *testing.T) {
	temp := t.TempDir()

	handle, err := Driver{}.Open("default", map[string]any{"data_dir": temp, "external_table_storage": true}, activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)
	c := handle.(*connection)
	require.NoError(t, c.Migrate(context.Background()))
	c.AsOLAP("default")

	err = c.CreateTableAsSelect(context.Background(), "test-rename", false, "select 1", nil)
	require.NoError(t, err)

	err = c.RenameTable(context.Background(), "test-rename", "rename-test", false)
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

	handle, err := Driver{}.Open("default", map[string]any{"data_dir": temp, "external_table_storage": true}, activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)
	c := handle.(*connection)
	require.NoError(t, c.Migrate(context.Background()))
	c.AsOLAP("default")

	err = c.CreateTableAsSelect(context.Background(), "source", false, "SELECT 1 AS data", nil)
	require.NoError(t, err)

	err = c.CreateTableAsSelect(context.Background(), "_tmp_source", false, "SELECT 2 AS DATA", nil)
	require.NoError(t, err)

	err = c.RenameTable(context.Background(), "_tmp_source", "source", false)
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
	handle, err := Driver{}.Open("default", map[string]any{"data_dir": temp, "external_table_storage": true}, activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)
	c := handle.(*connection)
	require.NoError(t, c.Migrate(context.Background()))
	c.AsOLAP("default")

	err = c.CreateTableAsSelect(context.Background(), "test alter column", false, "select 1 as data", nil)
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

	res, err = c.Execute(context.Background(), &drivers.Statement{Query: "SELECT data_type FROM information_schema.columns WHERE table_name='test alter column'"})
	require.NoError(t, err)
	require.True(t, res.Next())
	require.NoError(t, res.Scan(&typ))
	require.Equal(t, "VARCHAR", typ)
	require.NoError(t, res.Close())
}

func Test_connection_RenameToExistingTableOld(t *testing.T) {
	handle, err := Driver{}.Open("default", map[string]any{"dsn": ":memory:", "external_table_storage": false}, activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)
	c := handle.(*connection)
	require.NoError(t, c.Migrate(context.Background()))
	c.AsOLAP("default")

	err = c.CreateTableAsSelect(context.Background(), "source", false, "SELECT 1 AS data", nil)
	require.NoError(t, err)

	err = c.CreateTableAsSelect(context.Background(), "_tmp_source", false, "SELECT 2 AS DATA", nil)
	require.NoError(t, err)

	err = c.RenameTable(context.Background(), "_tmp_source", "source", false)
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
	dbPath := filepath.Join(temp, "default", "normal.db")
	handle, err := Driver{}.Open("default", map[string]any{"path": dbPath, "external_table_storage": false}, activity.NewNoopClient(), zap.NewNop())
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
	err = normalConn.CreateTableAsSelect(ctx, "test", false, sql, nil)
	require.NoError(t, err)

	err = normalConn.CreateTableAsSelect(ctx, "test_view", true, sql, nil)
	require.NoError(t, err)

	sql = `
		with r as (select 1 as id ) 	
		select * from r
	`
	err = normalConn.CreateTableAsSelect(ctx, "test", false, sql, nil)
	require.NoError(t, err)

	err = normalConn.CreateTableAsSelect(ctx, "test_view", true, sql, nil)
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
