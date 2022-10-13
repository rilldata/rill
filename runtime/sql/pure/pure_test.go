package pure

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreateSource(t *testing.T) {
	sql := ` CREATE SOURCE
	foobar 
		WITH ( connector =
		's3'
	
	'hello.world'= 200, ) `

	stmt, err := Parse(sql)
	require.NoError(t, err)
	require.Equal(t, "foobar", stmt.CreateSource.Name)
	require.Equal(t, 2, len(stmt.CreateSource.With.Properties))
	require.Equal(t, "connector", stmt.CreateSource.With.Properties[0].Key)
	require.Equal(t, "s3", *stmt.CreateSource.With.Properties[0].Value.String)
	require.Equal(t, "hello.world", stmt.CreateSource.With.Properties[1].Key)
	require.Equal(t, float64(200), *stmt.CreateSource.With.Properties[1].Value.Number)
}
