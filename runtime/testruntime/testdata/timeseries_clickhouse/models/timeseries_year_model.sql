SELECT *
FROM
(
    SELECT
        addMonths(toDateTime('2022-01-01 00:00:00', 'UTC'), number) AS timestamp
    FROM numbers(
        dateDiff('month',
            toDateTime('2022-01-01 00:00:00', 'UTC'),
            toDateTime('2025-12-01 00:00:00', 'UTC')
        ) + 1
    )
) a
CROSS JOIN
(
    SELECT
        1.0 AS clicks,
        'android' AS device,
        'Google' AS publisher,
        'Canada' AS country
) b;