WITH minCTE as (
  SELECT 
  MIN(createdAt) as minValue,
  MAX(createdAt) as maxValue,
  MAX(createdAt) - MIN(createdAt) as width,
  MAX(createdAt) - MIN(createdAt) / 5 AS binSize
  from pages
),
-- bin the values here
binCTE AS (
  SELECT 
    floor(
      (10 * CAST(createdAt - (SELECT minValue FROM minCTE) as float)
      ) / 
      (select width from minCTE)
    )
    * ((select width from minCTE) / 10)
    + (select minValue FROM minCTE)
  AS histogram_bin, createdAt FROM pages
),
-- group and count
groupedCTE AS (SELECT 
  count(*) as c, 
  histogram_bin
FROM binCTE 
GROUP BY histogram_bin 
ORDER BY histogram_bin asc
),
-- remove the last bin
removeLastBinCTE AS (
    SELECT * from groupedCTE 
    WHERE
    histogram_bin != (SELECT maxValue from minCTE)
)
-- add one to the new last bin to get inclusion
SELECT 
    CASE WHEN (histogram_bin = (SELECT max(histogram_bin) FROM removeLastBinCTE)) THEN c+1 ELSE c END AS c, 
    histogram_bin
 from removeLastBinCTE
GROUP BY histogram_bin, c;