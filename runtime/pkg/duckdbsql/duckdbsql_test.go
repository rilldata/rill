package duckdbsql

import (
	"testing"

	"github.com/rilldata/rill/runtime/pkg/fileutil"
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
		title     string
		sql       string
		tableRefs []*TableRef
	}{
		{
			"comments",
			`-- Some comment
-- @materialise
select 1`,
			[]*TableRef{},
		},
		{
			"read_csv",
			`select * from read_csv( 'data.csv', delim='|', columns={'A':'Date'})`,
			[]*TableRef{
				{
					Function: "read_csv",
					Path:     "data.csv",
					Properties: map[string]any{
						"delim": "|",
						"columns": map[string]any{
							"A": "Date",
						},
					},
				},
			},
		},
		{
			"join",
			`select * from AdBid a join AdImp i on a.id=i.id where a='1' group by b limit 2`,
			[]*TableRef{
				{
					Name: "AdBid",
				},
				{
					Name: "AdImp",
				},
			},
		},
	}

	for _, tt := range sqlVariations {
		t.Run(tt.title, func(t *testing.T) {
			ast, err := Parse(tt.sql)
			require.NoError(t, err)

			actualTableRefs := make([]*TableRef, 0)
			for _, node := range ast.fromNodes {
				actualTableRefs = append(actualTableRefs, node.ref)
			}
			require.EqualValues(t, tt.tableRefs, actualTableRefs)
		})
	}
}

func TestReplaceTableRefs(t *testing.T) {
	sqlVariations := []struct {
		title       string
		sql         string
		replacedSql string
	}{
		{
			"no replace",
			`select * from AdBid a join AdImp i on a.id=i.id where a='1' group by b limit 2`,
			`SELECT * FROM AdBid AS a INNER JOIN AdImp AS i ON ((a.id = i.id)) WHERE (a = '1') GROUP BY b LIMIT 2`,
		},
		{
			"simple replace",
			`select * from read_csv( 'AdBids.csv', delim='|', columns={'timestamp':'TIMESTAMP'})`,
			`SELECT * FROM AdBids`,
		},
		{
			"replace with join and alias",
			`
select * from
	read_csv( 'AdBids.csv', delim='|', columns={'timestamp':'TIMESTAMP'}) as b join
	AdImpressions i on b.id=i.id
`,
			`SELECT * FROM AdBids AS b INNER JOIN AdImpressions AS i ON ((b.id = i.id))`,
		},
	}

	for _, tt := range sqlVariations {
		t.Run(tt.title, func(t *testing.T) {
			ast, err := Parse(tt.sql)
			require.NoError(t, err)

			err = ast.RewriteTableRefs(func(table *TableRef) (*TableRef, bool) {
				if table.Path == "" {
					return nil, false
				}

				return &TableRef{
					Name: fileutil.Stem(table.Path),
				}, true
			})

			actualSql, err := ast.Format()
			require.NoError(t, err)
			require.EqualValues(t, tt.replacedSql, actualSql)
		})
	}
}
