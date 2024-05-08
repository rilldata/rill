select
    'continuous' as label,
    range as timestamp,
from range(TIMESTAMP '2023-03-10', TIMESTAMP '2023-03-14', INTERVAL 10 MINUTE)

union all
select 'sparse_hour' as label, '2023-03-12 03:00:00Z'::TIMESTAMP as timestamp
union all
select 'sparse_hour' as label, '2023-03-12 05:00:00Z'::TIMESTAMP as timestamp
union all
select 'sparse_hour' as label, '2023-03-12 07:00:00Z'::TIMESTAMP as timestamp

union all
select
    'sparse_day' as label,
    range as timestamp,
from range(TIMESTAMP '2023-03-09', TIMESTAMP '2023-03-11', INTERVAL 1 HOUR)
union all
select
    'sparse_day' as label,
    range as timestamp,
from range(TIMESTAMP '2023-03-12 05:00:00Z', TIMESTAMP '2023-03-13', INTERVAL 1 HOUR)