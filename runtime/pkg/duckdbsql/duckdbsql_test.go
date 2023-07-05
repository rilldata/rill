package duckdbsql

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

//func TestFormat(t *testing.T) {
//	res, err := Format("select    10+20 from  read_csv( 'data.csv')")
//	require.NoError(t, err)
//	require.Equal(t, "SELECT (10 + 20) FROM read_csv('data.csv')", res)
//}

// Comments are not parsed
func TestExtractTableRefs(t *testing.T) {
	sqlVariations := []struct {
		title string
		sql   string
	}{
		{
			"comments",
			`-- Some comment
-- @materialise
select 1`,
		},
		{
			"read_csv",
			`select * from read_csv( 'data.csv', delim='|', columns={'A':'Date'})`,
		},
		{
			"join",
			`select * from AdBid a join AdImp i on a.id=i.id where a='1' group by b limit 2`,
		},
	}

	for _, tt := range sqlVariations {
		t.Run(tt.title, func(t *testing.T) {
			ast, err := Parse(tt.sql)
			require.NoError(t, err)

			fmt.Println("\n" + tt.sql)
			for _, node := range ast.fromNodes {
				fmt.Println(node.ref)
			}
		})
	}
}
