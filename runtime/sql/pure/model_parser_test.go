package pure

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_extractCTEs(t *testing.T) {
	cteTests := []struct {
		query string
		ctes  []*table
	}{
		// this query has multiple CTEs.
		{`
WITH cte1 AS (
    SELECt * from tbl1 LIMIT 100
),
cte2 AS (
    SELECT * from cte1
),
cte3 AS (
    select created_date, count(*) from tbl2 GROUP BY created_date
)   
        SELECt    
    date_trunc('day', created_date) AS whatever,
    another_column,
    a_third as the_third_column
from cte1;
`, []*table{
			{20, 49, "cte1", "SELECt * from tbl1 LIMIT 100"},
			{66, 85, "cte2", "SELECT * from cte1"},
			{102, 164, "cte3", "select created_date, count(*) from tbl2 GROUP BY created_date"},
		}},
		// this query doesn't have a cte.
		{`
SELECt * from whatever;
`, nil},
		// this query doesn't even technically work.
		{"this is just a random string", nil},
		// this query is somewhat malformed after the CTEs,
		// but the CTEs can still be extracted.
		{`
with x AS (select * from whatever),
y AS (select dt from another_table),
whatever is next is what is next.
`, []*table{
			{12, 34, "x", "select * from whatever"},
			{43, 71, "y", "select dt from another_table"},
		}},
		// works with doubly-nested CTEs in that it ignores the nested CTEs.
		// one shouldn't even do this in practice but we'll still support it.
		{`
WITH x AS (WITH y as (select * from test) select * from y) select * from x)
SELECt * from x;
`, []*table{
			{12, 58, "x", "WITH y as (select * from test) select * from y"},
		}},
	}

	for i, tt := range cteTests {
		t.Run(fmt.Sprintf("CTE_%d", i), func(t *testing.T) {
			require.Equal(t, tt.ctes, extractCTEs(tt.query))
		})
	}
}

func Test_extractFromStatements(t *testing.T) {
	fromTests := []struct {
		query  string
		tables []*table
	}{
		{
			"SELECt * from table1",
			[]*table{{14, 20, "table1", ""}},
		},
		{
			"SELECt * from table2",
			[]*table{{14, 20, "table2", ""}},
		},
		{`          select * 
        
        
        
        
        from table3       
        
        
        `,
			[]*table{{69, 75, "table3", ""}}},
		{`with 
        x as (select * from whatever),
        abcd_wxyz as (select * from x)
           SELECT * from       abcd_wxyz   ;
        `,
			[]*table{
				{34, 42, "whatever", ""},
				{81, 82, "x", ""},
				{115, 124, "abcd_wxyz", ""},
			}},
		// handles nested from statements
		{
			"   select something from (select * from abc_xyz)    ",
			[]*table{{40, 47, "abc_xyz", ""}},
		},
		{
			"   select something from            (       select * from abc_xyz         )    ",
			[]*table{{58, 65, "abc_xyz", ""}},
		},
		// add where clause
		{
			"   select something from table WHERE id IS NOT NULL;",
			[]*table{{25, 30, "table", ""}},
		},
		// add GROUP BY clause
		{
			"   select something, count(*) from       table        GROUP BY count(*);",
			[]*table{{41, 46, "table", ""}},
		},
		// check wraps for ?
		{
			"\nselect something, count(*) from       table        \n    LEFT JOIN cruds ON cruds.id = table.id;",
			[]*table{{39, 44, "table", ""}},
		},
		{
			"\n        select something, count(*) from       table    abc    \n            LEFT JOIN cruds ON cruds.id = table.id;",
			[]*table{{47, 52, "table", ""}},
		},
	}

	for i, tt := range fromTests {
		t.Run(fmt.Sprintf("FromStatement_%d", i), func(t *testing.T) {
			require.Equal(t, tt.tables, extractFromStatements(tt.query))
		})
	}
}

func Test_extractJoins(t *testing.T) {
	joinTests := []struct {
		query  string
		tables []*table
	}{
		{
			"SELECt * from whatever inner join another ON another.id = whatever.another_id",
			[]*table{{34, 41, "another", ""}},
		},
		{`with 
        x as (select * from whatever),
        abcd_wxyz as (select * from x)
           SELECT * from       abcd_wxyz    join    y        ON        y.id = abcd_wxyz.whatever   ;
        `,
			[]*table{{136, 137, "y", ""}}},
		{`with 
        x as (select * from whatever),
        abcd_wxyz as (select * from x)
           SELECT * from       abcd_wxyz    join    (select * from y)        ON        y.id = abcd_wxyz.whatever   ;
        `, nil},
	}

	for i, tt := range joinTests {
		t.Run(fmt.Sprintf("JoinQuery_%d", i), func(t *testing.T) {
			require.Equal(t, tt.tables, extractJoins(tt.query))
		})
	}
}

func TestExtractTableNames(t *testing.T) {
	extractTests := []struct {
		query string
		names []string
	}{
		{`
WITH cte1 AS (
    SELECt * from tbl1 LIMIT 100
),
cte2 AS (
    SELECT * from cte1
),
cte3 AS (
    select created_date, count(*) from tbl2 GROUP BY created_date
)   
        SELECt    
    date_trunc('day', created_date) AS whatever,
    another_column,
    a_third as the_third_column
from cte1;
`,
			[]string{"tbl1", "tbl2"}},
		{`
with x AS (select * from whatever),
y AS (select dt from another_table),
whatever is next is what is next.
`,
			[]string{"whatever", "another_table"}},
		{`with 
        x as (select * from whatever),
        abcd_wxyz as (select * from x)
           SELECT * from       abcd_wxyz   ;
        `,
			[]string{"whatever"}},
		{`
WITH x AS (WITH y as (select * from test) select * from y) select * from x)
SELECt * from x;
`,
			// TODO: y is identified as a table
			[]string{"test", "y"}},
		{`with 
        x as (select * from whatever),
        abcd_wxyz as (select * from x)
           SELECT * from       abcd_wxyz    join    (select * from y)        ON        y.id = abcd_wxyz.whatever   ;
        `, []string{"whatever", "y"}},
		{`with 
        x as (select * from whatever),
        abcd_wxyz as (select * from x)
           SELECT * from       abcd_wxyz    join    y        ON        y.id = abcd_wxyz.whatever   ;
        `, []string{"whatever", "y"}},
	}

	for i, tt := range extractTests {
		t.Run(fmt.Sprintf("Extract_%d", i), func(t *testing.T) {
			require.Equal(t, tt.names, ExtractTableNames(tt.query))
		})
	}
}
