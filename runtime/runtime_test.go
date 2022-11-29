package runtime

// func NewTestRuntime(t *testing.T) *Runtime {
// 	opts := &Options{
// 		ConnectionCacheSize: 100,
// 		MetastoreDriver:     "sqlite",
// 		// Setting a test-specific name ensures a unique connection when "cache=shared" is enabled.
// 		// "cache=shared" is needed to prevent threading problems.
// 		MetastoreDSN: fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name()),
// 	}
// 	rt, err := New(opts, nil)
// 	require.NoError(t, err)

// 	return rt
// }

// func newTestInstance(t *testing.T) (*Runtime, string) {
// 	rt := newTestRuntime(t)

// 	inst := &drivers.Instance{
// 		OLAPDriver:   "duckdb",
// 		OLAPDSN:      "",
// 		RepoDriver:   "file",
// 		RepoDSN:      t.TempDir(),
// 		EmbedCatalog: true,
// 	}
// 	err = rt.CreateInstance(context.Background(), inst)
// 	require.NoError(t, err)
// 	require.NotEmpty(t, inst.ID)

// 	opts := &Options{
// 		ConnectionCacheSize: 100,
// 		MetastoreDriver:     "sqlite",
// 		// Setting a test-specific name ensures a unique connection when "cache=shared" is enabled.
// 		// "cache=shared" is needed to prevent threading problems.
// 		MetastoreDSN: fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name()),
// 	}
// 	rt, err := New(opts, nil)
// 	require.NoError(t, err)

// 	return rt
// }
