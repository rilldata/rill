package duckdbsql

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFormat(t *testing.T) {
	res, err := Format("select    10+20 from  read_csv( 'data.csv')")
	require.NoError(t, err)
	require.Equal(t, "SELECT (10 + 20) FROM read_csv('data.csv')", res)
}
