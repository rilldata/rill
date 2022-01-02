export function histogramQuery(table, column, bins = 20) {
	return `
WITH minCTE as (
    SELECT 
    MIN(${column}) as minValue,
    MAX(${column}) as maxValue,
    MAX(${column}) - MIN(${column}) as width
    from ${table}
    ),
    binnedCTE AS (
        -- value is 0-totalBins
        SELECT floor((${bins} * (${column} - (SELECT minValue FROM minCTE))) 
        /  (select width from minCTE))
        -- multiply by bin width
        * ((select width from minCTE) / ${bins}) 
        -- move back to min
        + (select minValue FROM minCTE)
    AS histogram_bin
    FROM ${table}
    ),
    partitionedCTE AS (SELECT DISTINCT histogram_bin,
            COUNT(*) OVER (PARTITION BY histogram_bin) AS c 
    FROM binnedCTE
    ORDER BY histogram_bin asc
    ),
    removeLastBinCTE AS (
        SELECT * from partitionedCTE 
        WHERE
        histogram_bin != (SELECT maxValue from minCTE)
    )
    SELECT 
        -- add 1 to largest bin
        CASE WHEN (histogram_bin = (SELECT max(histogram_bin) FROM partitionedCTE)) THEN c+1 ELSE c END AS c, 
        histogram_bin
    from removeLastBinCTE
    GROUP BY histogram_bin, c;    
    `;
}
