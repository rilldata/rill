WITH minCTE as (
  SELECT 
  MIN(createdAt) as minValue,
  MAX(createdAt) as maxValue,
  MAX(createdAt) - MIN(createdAt) as width
  from articles
),
binnedCTE AS (
    -- value is 0-totalBins
    SELECT floor((20.0 * (createdAt - (SELECT minValue FROM minCTE))) 
    /  (select width from minCTE))
    -- multiply by bin width
    * ((select width from minCTE) / 20.0) 
    -- move back to min
    + (select minValue FROM minCTE)
  AS histogram_bin
  FROM articles
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
    CASE WHEN (histogram_bin = (SELECT max(histogram_bin) FROM partitionedCTE)) THEN c+1 ELSE c END AS count, 
    epoch_ms(histogram_bin)
 from removeLastBinCTE
GROUP BY histogram_bin, c;