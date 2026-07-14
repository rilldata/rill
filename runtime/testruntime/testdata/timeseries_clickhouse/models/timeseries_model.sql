SELECT 1.0 AS clicks, 3 as imps, TIMESTAMP '2019-01-01 00:00:00' AS time, DATE '2019-01-01' as day, 'android' AS device, 'Google' AS publisher, 'google.com' AS domain, 25 as latitude, 'Canada' as country
UNION ALL
SELECT 1.0 AS clicks, 5 as imps, TIMESTAMP '2019-01-02 00:00:00' AS time, DATE '2019-01-02' as day, 'iphone' AS device, null AS publisher, 'msn.com' AS domain, NULL as latitude, NULL as country
