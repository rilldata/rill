package pinot

import (
	"testing"
	"time"

	sqlDriver "database/sql/driver"
	"github.com/stretchr/testify/require"
)

func Test_main(t *testing.T) {
	query := "SELECT * FROM users WHERE id = ? AND name = ?"
	args := []sqlDriver.NamedValue{
		{Ordinal: 1, Value: 123},
		{Ordinal: 2, Value: "John"},
	}
	q, err := completeQuery(query, args)
	require.NoError(t, err)
	require.Equal(t, "SELECT * FROM users WHERE id = 123 AND name = 'John'", q)

	args = []sqlDriver.NamedValue{
		{Ordinal: 1, Value: 123.5},
		{Ordinal: 2, Value: ""},
	}
	q, err = completeQuery(query, args)
	require.NoError(t, err)
	require.Equal(t, "SELECT * FROM users WHERE id = 123.5 AND name = ''", q)

	now := time.Now()
	args = []sqlDriver.NamedValue{
		{Ordinal: 1, Value: now},
		{Ordinal: 2, Value: true},
	}
	q, err = completeQuery(query, args)
	require.NoError(t, err)
	require.Equal(t, "SELECT * FROM users WHERE id = '"+now.Format("2006-01-02 15:04:05.000Z")+"' AND name = true", q)
}
