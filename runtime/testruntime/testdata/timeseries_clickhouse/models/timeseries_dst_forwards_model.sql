SELECT
    'continuous' AS label,
    toDateTime('2023-03-10 00:00:00', 'UTC') + toIntervalMinute(number * 10) AS timestamp
FROM numbers(
    dateDiff('minute', toDateTime('2023-03-10 00:00:00', 'UTC'), toDateTime('2023-03-14 00:00:00', 'UTC')) / 10
)

UNION ALL
SELECT 'sparse_hour', toDateTime('2023-03-12 03:00:00', 'UTC')

UNION ALL
SELECT 'sparse_hour', toDateTime('2023-03-12 05:00:00', 'UTC')

UNION ALL
SELECT 'sparse_hour', toDateTime('2023-03-12 07:00:00', 'UTC')

UNION ALL
SELECT
    'sparse_day',
    toDateTime('2023-03-09 00:00:00', 'UTC') + toIntervalHour(number)
FROM numbers(
    dateDiff('hour', toDateTime('2023-03-09 00:00:00', 'UTC'), toDateTime('2023-03-11 00:00:00', 'UTC'))
)

UNION ALL
SELECT
    'sparse_day',
    toDateTime('2023-03-12 05:00:00', 'UTC') + toIntervalHour(number)
FROM numbers(
    dateDiff('hour', toDateTime('2023-03-12 05:00:00', 'UTC'), toDateTime('2023-03-13 00:00:00', 'UTC'))
);