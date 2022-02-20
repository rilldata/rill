import { extractCTEs, getCoreQuerySelectStatements } from "./model-structure";

const q1 = `
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
`
const cte1 = [
    {name: "cte1", substring: "SELECt * from tbl1 LIMIT 100", start: 20, end: 49 },
    {name: "cte2", substring: "SELECT * from cte1", start: 66, end: 85},
    {name: "cte3", substring: "select created_date, count(*) from tbl2 GROUP BY created_date", start: 102, end: 164},
]

const q2 = `
SELECt * from whatever;
`
const cte2 = [];

const q3 = `this is just a random string`;
const cte3 = [];

const q4 = `
with x AS (select * from whatever),
y AS (select dt from another_table),
whatever is next is what is next.
`

const cte4 = [
    {name: 'x', substring: 'select * from whatever', start: 12, end:34 },
    {name: 'y', substring: 'select dt from another_table', start: 43, end:71 },
]

const q5 = `
WITH x AS (WITH y as (select * from test) select * from y) select * from x)
SELECt * from x;
`

const cte5 = [
    {name: 'x', substring: 'WITH y as (select * from test) select * from y', start: 12, end: 58}
]

describe("extractCTEs", () => {
    /** NOTE: these tests assume the query is mostly valid. It will help with
     * a few cases where the query isn't, but this is always a requirement.
     */
    it('pulls out all the CTEs from a complex query', () => {
        // this query has multiple CTEs.
        expect(extractCTEs(q1)).toEqual(cte1);
        // this query doesn't have a cte.
        expect(extractCTEs(q2)).toEqual(cte2);
        // this query doesn't even technically work.
        expect(extractCTEs(q3)).toEqual(cte3);
        // this query is somewhat malformed after the CTEs,
        // but the CTEs can still be extracted.
        expect(extractCTEs(q4)).toEqual(cte4);
        // works with doubly-nested CTEs in that it ignores the nested CTEs.
        // one shouldn't even do this in practice but we'll still support it.
        expect(extractCTEs(q5)).toEqual(cte5);
    })
})