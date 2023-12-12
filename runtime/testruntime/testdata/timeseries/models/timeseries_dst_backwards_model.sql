select
    'continuous' as label,
    range as timestamp,
from range(TIMESTAMP '2023-11-03', TIMESTAMP '2023-11-07', INTERVAL 10 MINUTE)

union all
select 'sparse_hour' as label, '2023-11-05 03:00:00Z'::TIMESTAMP as timestamp
union all
select 'sparse_hour' as label, '2023-11-05 05:00:00Z'::TIMESTAMP as timestamp
union all
select 'sparse_hour' as label, '2023-11-05 07:00:00Z'::TIMESTAMP as timestamp

union all
select
    'sparse_day' as label,
    range as timestamp,
from range(TIMESTAMP '2023-11-02', TIMESTAMP '2023-11-04', INTERVAL 10 MINUTE)
union all
select
    'sparse_day' as label,
    range as timestamp,
from range(TIMESTAMP '2023-11-05 05:00:00Z', TIMESTAMP '2023-11-06', INTERVAL 10 MINUTE)
