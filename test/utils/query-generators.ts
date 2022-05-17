/**
 * query-generators.ts
 * -------------------
 * Utilities for generating queries in tests.
 */

/**
 * generates a CREATE TABLE AS (ctas) statement with the specified selectStatement.
 * @param table the table name
 * @param select_statement the select statement that populates the table
 * @param temp a boolean that specifies whether the table should be temporary or not
 * @returns a string that represents the templated CTAS query
 */
export function ctas(table: string, selectStatement: string, temp = true) {
    return `CREATE ${temp ? 'TEMPORARY' : ''} VIEW ${table} AS (${selectStatement})`
}

/**
 * Creates a temporary table that contains a single TIMESTAMP column, "ts", that contains a range of data between
 * start, end, and at the specified DuckDB interval.
 * @param table the table name you
 * @param start a timestamp string specifying the start of the series
 * @param end a timestamp string specifying the end of the series
 * @param interval an interval string specifying the interval size (e.g. "1 day")
 * @returns a string that represented a CTAS query that generates the time series.
 */
export function generateSeries(table: string, start: string, end: string, interval: string, addFauxCount = false) {
    return ctas(table, `
    SELECT generate_series as ts 
    ${addFauxCount ? ", 1 AS count" : ''}
    FROM generate_series(TIMESTAMP '${start}', TIMESTAMP '${end}', interval ${interval})
    `)
}
