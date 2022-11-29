SELECT 
  to_timestamp(CAST(ts AS INT)) AS event_datetime, 
  CASE WHEN light THEN 'light' ELSE 'dark' END AS light, 
  CASE WHEN motion THEN 'activity' ELSE 'motionless' END AS motion,
  * EXCLUDE (ts, light, motion)
FROM iot_telemetry_data