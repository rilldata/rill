SELECT
    'continuous' AS label,
    toDateTime('2023-11-03 00:00:00', 'UTC') + toIntervalMinute(number * 10) AS timestamp
FROM numbers(
    dateDiff('minute', toDateTime('2023-11-03 00:00:00', 'UTC'), toDateTime('2023-11-07 00:00:00', 'UTC')) / 10
)

UNION ALL
SELECT 'sparse_hour', toDateTime('2023-11-05 03:00:00', 'UTC')

UNION ALL
SELECT 'sparse_hour', toDateTime('2023-11-05 05:00:00', 'UTC')

UNION ALL
SELECT 'sparse_hour', toDateTime('2023-11-05 07:00:00', 'UTC')

UNION ALL
SELECT
    'sparse_day',
    toDateTime('2023-11-02 00:00:00', 'UTC') + toIntervalMinute(number * 10)
FROM numbers(
    dateDiff('minute', toDateTime('2023-11-02 00:00:00', 'UTC'), toDateTime('2023-11-04 00:00:00', 'UTC')) / 10
)

UNION ALL
SELECT
    'sparse_day',
    toDateTime('2023-11-05 05:00:00', 'UTC') + toIntervalMinute(number * 10)
FROM numbers(
    dateDiff('minute', toDateTime('2023-11-05 05:00:00', 'UTC'), toDateTime('2023-11-06 00:00:00', 'UTC')) / 10
);