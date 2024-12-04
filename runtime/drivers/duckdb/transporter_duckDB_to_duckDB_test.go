package duckdb

// func TestDuckDBToDuckDBTransfer(t *testing.T) {
// 	tempDir := t.TempDir()
// 	conn, err := Driver{}.Open("default", map[string]any{"path": fmt.Sprintf("%s.db", filepath.Join(tempDir, "tranfser")), "external_table_storage": false}, storage.MustNew(tempDir, nil), activity.NewNoopClient(), zap.NewNop())
// 	require.NoError(t, err)

// 	olap, ok := conn.AsOLAP("")
// 	require.True(t, ok)

// 	err = olap.Exec(context.Background(), &drivers.Statement{
// 		Query: "CREATE TABLE foo(bar VARCHAR, baz INTEGER)",
// 	})
// 	require.NoError(t, err)

// 	err = olap.Exec(context.Background(), &drivers.Statement{
// 		Query: "INSERT INTO foo VALUES ('a', 1), ('a', 2), ('b', 3), ('c', 4)",
// 	})
// 	require.NoError(t, err)
// 	require.NoError(t, conn.Close())

// to, err := Driver{}.Open("default", map[string]any{"path": filepath.Join(tempDir, "main.db"), "external_table_storage": false}, storage.MustNew(tempDir, nil), activity.NewNoopClient(), zap.NewNop())
// require.NoError(t, err)

// 	tr := newDuckDBToDuckDB(to.(*connection), zap.NewNop())

// 	// transfer once
// 	err = tr.Transfer(context.Background(), map[string]any{"sql": "SELECT * FROM foo", "db": filepath.Join(tempDir, "tranfser.db")}, map[string]any{"table": "test"}, &drivers.TransferOptions{})
// 	require.NoError(t, err)

// 	olap, ok = to.AsOLAP("")
// 	require.True(t, ok)

// 	rows, err := to.(*connection).Execute(context.Background(), &drivers.Statement{Query: "SELECT COUNT(*) FROM test"})
// 	require.NoError(t, err)

// 	var count int
// 	rows.Next()
// 	require.NoError(t, rows.Scan(&count))
// 	require.Equal(t, 4, count)
// 	require.NoError(t, rows.Close())

// 	// transfer again
// 	err = tr.Transfer(context.Background(), map[string]any{"sql": "SELECT * FROM foo", "db": filepath.Join(tempDir, "tranfser.db")}, map[string]any{"table": "test"}, &drivers.TransferOptions{})
// 	require.NoError(t, err)

// 	rows, err = olap.Execute(context.Background(), &drivers.Statement{Query: "SELECT COUNT(*) FROM test"})
// 	require.NoError(t, err)

// 	rows.Next()
// 	require.NoError(t, rows.Scan(&count))
// 	require.Equal(t, 4, count)
// 	require.NoError(t, rows.Close())
// }
