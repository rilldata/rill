package duckdb_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
)

func TestLocalFileToDuckDBModel(t *testing.T) {
	tempDir := t.TempDir()
	defer os.RemoveAll(tempDir)
	fw1, err := os.Create(filepath.Join(tempDir, "data1.csv"))
	require.NoError(t, err)
	_, err = fw1.WriteString("id,country\n1,US\n2,CA\n")
	require.NoError(t, err)
	require.NoError(t, fw1.Close())

	fw2, err := os.Create(filepath.Join(tempDir, "data2.csv"))
	require.NoError(t, err)
	_, err = fw2.WriteString("id,country\n3,IN\n4,UK\n")
	require.NoError(t, err)
	require.NoError(t, fw2.Close())

	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"rill.yaml": "",
			"models/nochange.yaml": `
type: model
connector: local_file
path: ` + filepath.Join(tempDir, "data*.csv") + `
output:
  connector: duckdb
`,
			"models/change.yaml": `
type: model
connector: local_file
path: ` + filepath.Join(tempDir, "data*.csv") + `
invalidate_on_change: true
output:
  connector: duckdb
`}})

	olap, release, err := rt.OLAP(context.Background(), instanceID, "")
	require.NoError(t, err)
	defer release()

	require.Equal(t, 4, rows(t, olap, "nochange"))
	require.Equal(t, 4, rows(t, olap, "change"))

	// add new data to both files
	fw1, err = os.OpenFile(filepath.Join(tempDir, "data1.csv"), os.O_APPEND|os.O_WRONLY, 0o644)
	require.NoError(t, err)
	_, err = fw1.WriteString("5,JP\n")
	require.NoError(t, err)
	require.NoError(t, fw1.Close())

	fw2, err = os.OpenFile(filepath.Join(tempDir, "data2.csv"), os.O_APPEND|os.O_WRONLY, 0o644)
	require.NoError(t, err)
	_, err = fw2.WriteString("6,DE\n")
	require.NoError(t, err)
	require.NoError(t, fw2.Close())

	testruntime.ReconcileAndWait(t, rt, instanceID, &runtimev1.ResourceName{
		Kind: runtime.ResourceKindModel,
		Name: "nochange",
	})
	testruntime.ReconcileAndWait(t, rt, instanceID, &runtimev1.ResourceName{
		Kind: runtime.ResourceKindModel,
		Name: "change",
	})

	require.Equal(t, 4, rows(t, olap, "nochange"))
	require.Equal(t, 6, rows(t, olap, "change"))
}

func rows(t *testing.T, olap drivers.OLAPStore, tbl string) int {
	rows, err := olap.Execute(context.Background(), &drivers.Statement{Query: "SELECT count(*) FROM " + tbl})
	require.NoError(t, err)
	defer rows.Close()

	var count int
	require.True(t, rows.Next())
	require.NoError(t, rows.Scan(&count))
	require.NoError(t, rows.Err())
	return count
}
