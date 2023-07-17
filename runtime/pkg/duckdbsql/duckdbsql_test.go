package duckdbsql

import (
	"testing"

	"github.com/rilldata/rill/runtime/pkg/fileutil"
	"github.com/stretchr/testify/require"
)

// Comments are not parsed
func TestParse_TableRefs(t *testing.T) {
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
				{Name: "AdBid"},
				{Name: "AdImp"},
			},
		},
		{
			"join with sub query",
			`select * from AdBid a join (select * from AdImp where city='Bengaluru') i on a.id=i.id where a='1' group by b limit 2`,
			[]*TableRef{
				{Name: "AdBid"},
				{Name: "AdImp"},
			},
		},
		{
			"simple CTEs",
			`with tbl2 as (select col1 from tbl1), tbl3 as (select col1 from tbl1) select col1 from tbl2 join tbl3 on tbl2.id = tbl3.id`,
			[]*TableRef{
				{Name: "tbl1"},
				{Name: "tbl2", LocalAlias: true},
				{Name: "tbl3", LocalAlias: true},
			},
		},
		{
			"CTEs with union",
			`with tbl2 as (select col1 from tbl1), tbl3 as (select col1 from tbl1) select col1 from tbl2 union all select col1 from tbl3`,
			[]*TableRef{
				{Name: "tbl1"},
				{Name: "tbl2", LocalAlias: true},
				{Name: "tbl3", LocalAlias: true},
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

func TestParse_ColumnRefs(t *testing.T) {
	sqlVariations := []struct {
		title      string
		sql        string
		columnRefs []*ColumnRef
	}{
		{
			"select with exclude column",
			`select *, exclude(id), count(*), avg(bid_price) as bid_price from AdBids`,
			[]*ColumnRef{
				{IsStar: true},
				{Name: "id", IsExclude: true},
				{Expr: "count_star()"},
				{Name: "bid_price", Expr: "avg(bid_price)"},
			},
		},
		{
			"joins with exclude column",
			`
select b.*, i.city, i.country as i_cnt, exclude(i.id), count(*), avg(b.bid_price) as bid_price from
  AdBids b join (select * from AdImpressions i1 where i1.city='Bengaluru') i on b.id = i.id
`,
			[]*ColumnRef{
				{IsStar: true, RelationName: "b"},
				{Name: "i.city"},
				{Name: "i_cnt"},
				{Name: "i.id", IsExclude: true},
				{Expr: "count_star()"},
				{Name: "bid_price", Expr: "avg(b.bid_price)"},
			},
		},
		{
			"CTEs with join and exclude column",
			`
with
  b as (select col1 from read_csv( 'AdBids.csv', delim='|', columns={'timestamp':'TIMESTAMP'})),
  i as (select col1 from read_csv( 'AdImpressions.csv', delim='|', columns={'timestamp':'TIMESTAMP'}))
select b.*, i.city, exclude(i.id), count(*), avg(b.bid_price) as bid_price from b join i on b.id = i.id
`,
			[]*ColumnRef{
				{IsStar: true, RelationName: "b"},
				{Name: "i.city"},
				{Name: "i.id", IsExclude: true},
				{Expr: "count_star()"},
				{Name: "bid_price", Expr: "avg(b.bid_price)"},
			},
		},
		{
			"CTEs and unions with join and exclude column",
			`
with
  b as (select col1 from read_csv( 'AdBids.csv', delim='|', columns={'timestamp':'TIMESTAMP'})),
  i as (select col1 from read_csv( 'AdImpressions.csv', delim='|', columns={'timestamp':'TIMESTAMP'}))
(select b.*, exclude(i.id) as bid_price from b join i on b.id = i.id) union all
(select i.city from b join i on b.id = i.id) union all
(select count(*), avg(b.bid_price) as bid_price from b join i on b.id = i.id)
`,
			[]*ColumnRef{
				{IsStar: true, RelationName: "b"},
				{Name: "i.id", IsExclude: true},
				{Name: "i.city"},
				{Expr: "count_star()"},
				{Name: "bid_price", Expr: "avg(b.bid_price)"},
			},
		},
	}

	for _, tt := range sqlVariations {
		t.Run(tt.title, func(t *testing.T) {
			ast, err := Parse(tt.sql)
			require.NoError(t, err)

			require.EqualValues(t, tt.columnRefs, ast.ExtractColumnRefs())
		})
	}
}

func TestAST_RewriteFunctionTableRefs(t *testing.T) {
	sqlVariations := []struct {
		title       string
		sql         string
		expectedSql string
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
		{
			"join with sub query",
			`
select * from
  read_csv( 'AdBids.csv', delim='|', columns={'timestamp':'TIMESTAMP'}) a join
  (select * from read_csv( 'AdImpressions.csv', delim='|', columns={'timestamp':'TIMESTAMP'}) i1 where i1.city='Bengaluru') i on a.id=i.id
where a='1' group by b limit 2`,
			`SELECT * FROM AdBids AS a INNER JOIN (SELECT * FROM AdImpressions AS i1 WHERE (i1.city = 'Bengaluru')) AS i ON ((a.id = i.id)) WHERE (a = '1') GROUP BY b LIMIT 2`,
		},
		{
			"replace with CTEs",
			`
with
  tbl2 as (select col1 from read_csv( 'AdBids.csv', delim='|', columns={'timestamp':'TIMESTAMP'})),
  tbl3 as (select col1 from read_csv( 'AdImpressions.csv', delim='|', columns={'timestamp':'TIMESTAMP'}))
select col1 from tbl2 join tbl3 on tbl2.id = tbl3.id
`,
			`WITH tbl2 AS (SELECT col1 FROM AdBids), tbl3 AS (SELECT col1 FROM AdImpressions)SELECT col1 FROM tbl2 INNER JOIN tbl3 ON ((tbl2.id = tbl3.id))`,
		},
		{
			"replace with CTEs and unions",
			`
with
  tbl2 as (select col1 from read_csv( 'AdBids_May.csv', delim='|', columns={'timestamp':'TIMESTAMP'})),
  tbl3 as (select col1 from read_csv( 'AdBids_June.csv', delim='|', columns={'timestamp':'TIMESTAMP'}))
select col1 from tbl2 union all select col1 from tbl3 union all select col1 from read_csv( 'AdBids_July.csv', delim='|', columns={'timestamp':'TIMESTAMP'})
`,
			`WITH tbl2 AS (SELECT col1 FROM AdBids_May), tbl3 AS (SELECT col1 FROM AdBids_June)((SELECT col1 FROM tbl2) UNION ALL (SELECT col1 FROM tbl3)) UNION ALL (SELECT col1 FROM AdBids_July)`,
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
			require.EqualValues(t, tt.expectedSql, actualSql)
		})
	}
}

func TestAST_RewriteBaseTableRefs(t *testing.T) {
	sqlVariations := []struct {
		title       string
		sql         string
		replace     []string
		expectedSql string
	}{
		{
			"simple table reference",
			`select * from AdBid a join "s3://data/AdImp.csv" i on a.id=i.id where a='1' group by b limit 2`,
			[]string{"AB", "AI"},
			`SELECT * FROM AB AS a INNER JOIN AI AS i ON ((a.id = i.id)) WHERE (a = '1') GROUP BY b LIMIT 2`,
		},
		{
			"table references with sub queries",
			`
select * from
  AdBid a join (select * from "s3://data/AdImp.csv" i1 where i1.city='Bengaluru') i on a.id=i.id
  where a='1' group by b limit 2`,
			[]string{"AB", "AI"},
			`SELECT * FROM AB AS a INNER JOIN (SELECT * FROM AI AS i1 WHERE (i1.city = 'Bengaluru')) AS i ON ((a.id = i.id)) WHERE (a = '1') GROUP BY b LIMIT 2`,
		},
		{
			"table references with CTEs",
			`
with
  tbl2 as (select col1 from AdBid a),
  tbl3 as (select col1 from "s3://data/AdImp.csv" i)
select col1 from tbl2 join tbl3 on tbl2.id = tbl3.id
`,
			[]string{"AB", "AI", "tbl2", "tbl3"},
			`WITH tbl2 AS (SELECT col1 FROM AI AS a), tbl3 AS (SELECT col1 FROM AB AS i)SELECT col1 FROM tbl2 INNER JOIN tbl3 ON ((tbl2.id = tbl3.id))`,
		},
		{
			"table references with CTEs and unions",
			`
with
  tbl2 as (select col1 from AdBid_May a),
  tbl3 as (select col1 from "s3://data/AdBid_June.csv" i)
select col1 from tbl2 union all select col1 from tbl3 union all select col1 from "s3://data/AdBid_July.csv"
`,
			[]string{"A_M", "A_J", "tbl2", "tbl3", "A_Jl"},
			`WITH tbl2 AS (SELECT col1 FROM A_J AS a), tbl3 AS (SELECT col1 FROM A_M AS i)((SELECT col1 FROM tbl2) UNION ALL (SELECT col1 FROM tbl3)) UNION ALL (SELECT col1 FROM A_Jl)`,
		},
	}

	for _, tt := range sqlVariations {
		t.Run(tt.title, func(t *testing.T) {
			ast, err := Parse(tt.sql)
			require.NoError(t, err)

			i := -1
			err = ast.RewriteTableRefs(func(table *TableRef) (*TableRef, bool) {
				i = i + 1
				return &TableRef{
					Name: tt.replace[i],
				}, true
			})

			actualSql, err := ast.Format()
			require.NoError(t, err)
			require.EqualValues(t, tt.expectedSql, actualSql)
		})
	}
}

func TestAST_RewriteLimit(t *testing.T) {
	sqlVariations := []struct {
		title       string
		sql         string
		limit       int
		offset      int
		expectedSql string
	}{
		{
			"InsertLimit_SELECT",
			`SELECT col1 FROM (SELECT col1 FROM tbl1) AS sub1 INNER JOIN (SELECT col1 FROM tbl1) AS sub2 ON (sub1.col1 = sub2.col1)`,
			100,
			0,
			`SELECT col1 FROM (SELECT col1 FROM tbl1) AS sub1 INNER JOIN (SELECT col1 FROM tbl1) AS sub2 ON ((sub1.col1 = sub2.col1)) LIMIT 100`,
		},
		{
			"UpdateLimit_SELECT",
			`SELECT col1 FROM (SELECT col1 FROM tbl1 LIMIT 2000) AS sub1 INNER JOIN (SELECT col1 FROM tbl1 LIMIT 2000) AS sub2 ON ((sub1.col1 = sub2.col1)) LIMIT 2000`,
			100,
			0,
			`SELECT col1 FROM (SELECT col1 FROM tbl1 LIMIT 2000) AS sub1 INNER JOIN (SELECT col1 FROM tbl1 LIMIT 2000) AS sub2 ON ((sub1.col1 = sub2.col1)) LIMIT 100`,
		},
		{
			"InsertLimit_WITH",
			`WITH tbl2 AS (SELECT col1 FROM tbl1), tbl3 AS (SELECT col1 FROM tbl1) SELECT col1 FROM tbl2 UNION ALL SELECT col1 FROM tbl3`,
			100,
			0,
			`WITH tbl2 AS (SELECT col1 FROM tbl1), tbl3 AS (SELECT col1 FROM tbl1)(SELECT col1 FROM tbl2) UNION ALL (SELECT col1 FROM tbl3) LIMIT 100`,
		},
		{
			"UpdateLimit_WITH",
			`WITH tbl2 AS (SELECT col1 FROM tbl1 LIMIT 2000), tbl3 AS (SELECT col1 FROM tbl1 LIMIT 2000)(SELECT col1 FROM tbl2 LIMIT 2000) UNION ALL (SELECT col1 FROM tbl3 LIMIT 2000) LIMIT 2000`,
			100,
			0,
			`WITH tbl2 AS (SELECT col1 FROM tbl1 LIMIT 2000), tbl3 AS (SELECT col1 FROM tbl1 LIMIT 2000)(SELECT col1 FROM tbl2 LIMIT 2000) UNION ALL (SELECT col1 FROM tbl3 LIMIT 2000) LIMIT 100`,
		},
		{
			"InsertLimit_SELECT_WHERE",
			`SELECT col1 FROM tbl1 WHERE col1 = 1 ORDER BY 1`,
			100,
			0,
			`SELECT col1 FROM tbl1 WHERE (col1 = 1) ORDER BY 1 LIMIT 100`,
		},
		{
			"UpdateLimit_SELECT_WHERE",
			`SELECT col1 FROM tbl1 WHERE (col1 = 1) ORDER BY 1 LIMIT 2000`,
			100,
			0,
			`SELECT col1 FROM tbl1 WHERE (col1 = 1) ORDER BY 1 LIMIT 100`,
		},
		{
			"UpdateLimit_args_?",
			`SELECT col1 FROM tbl1 WHERE col1 = ? ORDER BY 1 LIMIT 2000`,
			100,
			0,
			`SELECT col1 FROM tbl1 WHERE (col1 = $1) ORDER BY 1 LIMIT 100`,
		},
		{
			"UpdateLimit_args_$",
			`SELECT col1 FROM tbl1 WHERE col1 = $1 ORDER BY 1 LIMIT 2000`,
			100,
			0,
			`SELECT col1 FROM tbl1 WHERE (col1 = $1) ORDER BY 1 LIMIT 100`,
		},
		{
			"UpdateLimit_LIMIT_args",
			`SELECT col1 FROM tbl1 WHERE col1 = 1 ORDER BY 1 LIMIT ?`,
			100,
			0,
			`SELECT col1 FROM tbl1 WHERE (col1 = 1) ORDER BY 1 LIMIT 100`,
		},
		{
			"UpdateLimit_UNION",
			`SELECT col1 FROM tbl1 UNION ALL SELECT col1 FROM tbl1`,
			100,
			0,
			`(SELECT col1 FROM tbl1) UNION ALL (SELECT col1 FROM tbl1) LIMIT 100`,
		},
	}

	for _, tt := range sqlVariations {
		t.Run(tt.title, func(t *testing.T) {
			ast, err := Parse(tt.sql)
			require.NoError(t, err)

			err = ast.RewriteLimit(tt.limit, tt.offset)
			require.NoError(t, err)

			actualSql, err := ast.Format()
			require.NoError(t, err)
			require.EqualValues(t, tt.expectedSql, actualSql)
		})
	}
}

func TestAST_ExtractAnnotations(t *testing.T) {
	sqlVariations := []struct {
		title       string
		sql         string
		annotations map[string]*Annotation
	}{
		{
			"comments at the top",
			`
-- some random comment
-- @materialise_v1
-- @materialise_v2  :	true  
-- @materialise_v3  :	tr ue  
-- some other comment
select * from AdBids
`,
			map[string]*Annotation{
				"materialise_v1": {Key: "materialise_v1"},
				"materialise_v2": {Key: "materialise_v2", Value: "true"},
				"materialise_v3": {Key: "materialise_v3", Value: "tr ue"},
			},
		},
		{
			"comments in the middle",
			`
select
-- @measure: avg
-- @measure.format: usd
a,
-- @dimension
b
from AdBids
`,
			map[string]*Annotation{
				"measure":        {Key: "measure", Value: "avg"},
				"measure.format": {Key: "measure.format", Value: "usd"},
				"dimension":      {Key: "dimension"},
			},
		},
	}

	for _, tt := range sqlVariations {
		t.Run(tt.title, func(t *testing.T) {
			ast, err := Parse(tt.sql)
			require.NoError(t, err)

			require.EqualValues(t, tt.annotations, ast.ExtractAnnotations())
		})
	}
}
