WITH

Timeseries AS (
FROM episodes
SELECT 
  MAKE_DATE(
    CAST(STR_SPLIT(aired, ' ')[3] AS INT),
    CAST(STRFTIME(STRPTIME(STR_SPLIT(aired, ' ')[1],'%B'), '%-m') AS INT),
    CAST(REPLACE(STR_SPLIT(aired, ' ')[2], ',', '') AS INT)
    ) AS event_date,
  * EXCLUDE(aired),
)
,

ActorsTransformed AS (
SELECT 
  column0 AS aid, 
  column1 AS url, 
  column2 AS type, 
  column3 AS gender, 
  FROM actors 
  WHERE column0 != 'aid'
),

CharactersOrImpressions AS (
FROM impressions SELECT *, 'impressions' AS category
UNION  
FROM characters SELECT *, 'characters' AS category
)

FROM Timeseries a
LEFT JOIN appearances b ON a.epid = b.epid
LEFT JOIN CharactersOrImpressions c ON b.role = c.name
LEFT JOIN ActorsTransformed d on b.aid = d.aid
SELECT 
  a.event_date,
  a.sid AS season_id,
  CAST(a.sid AS VARCHAR) AS season_number,
  a.epid AS episode_id,
  CAST(epno AS VARCHAR) AS episode_number,
  b.tid AS appearance_id,
  CONCAT(
    'season:',CAST(a.sid AS VARCHAR), 
    ' episode:', CAST(epno AS VARCHAR), 
    ' segment:', CAST(tid AS VARCHAR)[9:] 
    ) AS appearance_number,
  CAST(tid AS VARCHAR)[9:] AS appearance_number,
  b.aid AS actor, 
  capacity AS actor_capacity,
  gender AS actor_gender,
  role AS actor_role,
  COALESCE(category, 'other') AS role_category,
  charid AS role_id,
  voice,

WHERE b.aid IS NOT NULL