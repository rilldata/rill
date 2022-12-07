package models

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

type parseTest struct {
	query  string
	tables []string
}

func Test_extractCTEs(t *testing.T) {
	cteTests := []parseTest{
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
`, []string{"cte1", "tbl1", "tbl2"}},
		// duckdb 0.6 syntax
		{`
WITH cte1 AS (
    from tbl1 LIMIT 100
),
cte2 AS (
    from cte1
),
cte3 AS (
    select created_date, count(*) from tbl2 GROUP BY created_date
)   
        SELECt    
    date_trunc('day', created_date) AS whatever,
    another_column,
    a_third as the_third_column
from cte1;
`, []string{"cte1", "tbl1", "tbl2"}},
		// this query is somewhat malformed after the CTEs,
		// but the CTEs can still be extracted.
		{`
with x AS (select * from whatever),
y AS (select dt from another_table),
whatever is next is what is next.
`, []string{"whatever", "another_table"}},
		// works with doubly-nested CTEs in that it ignores the nested CTEs.
		// one shouldn't even do this in practice but we'll still support it.
		{`
WITH x AS (WITH y as (select * from test) select * from y) select * from x)
SELECt * from x;
`, []string{"test", "y", "x"}},
	}

	for i, tt := range cteTests {
		t.Run(fmt.Sprintf("CTE_%d", i), func(t *testing.T) {
			require.ElementsMatch(t, ExtractTableNames(tt.query), tt.tables)
		})
	}
}

func Test_extractFromStatements(t *testing.T) {
	fromTests := []parseTest{
		{
			"SELECt * from table1",
			[]string{"table1"},
		},
		{
			"SELECt * from table2",
			[]string{"table2"},
		},
		{
			"from table1",
			[]string{"table1"},
		},
		{
			"fRoM table1",
			[]string{"table1"},
		},
		{
			"from ",
			nil,
		},
		{
			"select * from ",
			nil,
		},
		{`          select * 
        
        
        
        
        from table3       
        
        
        `,
			[]string{"table3"}},
		{`
        
        
        
        
        from table3       
        
        
        `,
			[]string{"table3"}},
		{`with 
        x as (select * from whatever),
        abcd_wxyz as (select * from x)
           SELECT * from       abcd_wxyz   ;
        `,
			[]string{"whatever", "x", "abcd_wxyz"}},
		// handles nested from statements
		{
			"select * from (select * from abc_xyz)",
			[]string{"abc_xyz"},
		},
		{
			"   select something from (select * from abc_xyz)    ",
			[]string{"abc_xyz"},
		},
		{
			"   select something from            (       select * from abc_xyz         )    ",
			[]string{"abc_xyz"},
		},
		{
			"   select something from            (       from      abc_xyz         )    ",
			[]string{"abc_xyz"},
		},
		// add where clause
		{
			"   select something from table WHERE id IS NOT NULL;",
			[]string{"table"},
		},
		// add GROUP BY clause
		{
			"   select something, count(*) from       table        GROUP BY count(*);",
			[]string{"table"},
		},
		// check wraps for ?
		{
			"\nselect something, count(*) from       table        \n    LEFT JOIN cruds ON cruds.id = table.id;",
			[]string{"table", "cruds"},
		},
		{
			"\n        select something, count(*) from       table    abc    \n            LEFT JOIN cruds ON cruds.id = table.id;",
			[]string{"table", "cruds"},
		},
		{
			`FROM "s3://path/to/bucket.parquet"`,
			[]string{`"s3://path/to/bucket.parquet"`},
		},
		{
			`FROM "s3://path/to/bucket.parquet" as tbl`,
			[]string{`"s3://path/to/bucket.parquet"`},
		},
		{
			` FROM tbl JOIN "s3://path/to/bucket.parquet" as tbl2 ON tbl2.id = tbl.id`,
			[]string{"tbl", `"s3://path/to/bucket.parquet"`},
		},
	}

	for i, tt := range fromTests {
		t.Run(fmt.Sprintf("FromStatement_%d", i), func(t *testing.T) {
			require.ElementsMatch(t, ExtractTableNames(tt.query), tt.tables)
		})
	}
}

func Test_extractJoins(t *testing.T) {
	joinTests := []parseTest{
		{
			"SELECt * from whatever inner join another ON another.id = whatever.another_id",
			[]string{"whatever", "another"},
		},
		{
			"fRom whatever inner join another ON another.id = whatever.another_id",
			[]string{"whatever", "another"},
		},
		{
			`select * from tbl JOIN 
  
  x ON tbl.id = x.id`,
			[]string{"tbl", "x"},
		},
		{
			`from tbl JOIN 
  
  x ON tbl.id = x.id`,
			[]string{"tbl", "x"},
		},
		{`with 
        x as (select * from whatever),
        abcd_wxyz as (select * from x)
           SELECT * from       abcd_wxyz    join    y        ON        y.id = abcd_wxyz.whatever   ;
        `,
			[]string{"whatever", "x", "abcd_wxyz", "y"}},
		{`with 
        x as (select * from whatever),
        abcd_wxyz as (select * from x)
           SELECT * from       abcd_wxyz    join    (select * from y)        ON        y.id = abcd_wxyz.whatever   ;
        `, []string{"whatever", "x", "abcd_wxyz", "y"}},
	}

	for i, tt := range joinTests {
		t.Run(fmt.Sprintf("JoinQuery_%d", i), func(t *testing.T) {
			require.ElementsMatch(t, ExtractTableNames(tt.query), tt.tables)
		})
	}
}
