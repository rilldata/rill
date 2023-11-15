select
    date_part('minute', range) as minute,
  date_part('hour', range) as hour,
  date_part('day', range) as day,
  range as timestamp,
from range(TIMESTAMP '2023-11-01', TIMESTAMP '2023-11-10', INTERVAL 10 MINUTE)