select * from (select generate_series as timestamp from generate_series(TIMESTAMP '2022-01-01 00:00:00', TIMESTAMP '2025-12-01 00:00:00', INTERVAL '1' MONTH)) a
cross join 
(SELECT 1.0 AS clicks,  'android' AS device, 'Google' AS publisher, 'Canada' as country) b 