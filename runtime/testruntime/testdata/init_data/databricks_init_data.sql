CREATE SCHEMA IF NOT EXISTS integration_test;

CREATE TABLE IF NOT EXISTS integration_test.all_datatypes (
    id                 INT,
    boolean_col        BOOLEAN,
    tinyint_col        TINYINT,
    smallint_col       SMALLINT,
    int32_col          INT,
    int64_col          BIGINT,
    float_col          FLOAT,
    double_col         DOUBLE,
    decimal_col        DECIMAL(18,6),
    string_col         STRING,
    varchar_col        VARCHAR(255),
    date_col           DATE,
    timestamp_col      TIMESTAMP,
    timestamp_ntz_col  TIMESTAMP_NTZ,
    binary_col         BINARY,
    array_col          ARRAY<INT>,
    map_col            MAP<STRING, STRING>,
    struct_col         STRUCT<city: STRING, zip: INT>
);

TRUNCATE TABLE integration_test.all_datatypes;

INSERT INTO integration_test.all_datatypes VALUES
(
    1, TRUE, 127, 32767, 2147483647, 9223372036854775807,
    3.14, 2.718, 456.789000,
    'Sample String', 'Large text data',
    '2024-03-26', '2024-03-26T14:30:00.000000', '2024-03-26T14:30:00.000000',
    X'48656C6C6F',
    ARRAY(1, 2, 3),
    MAP('city', 'New York'),
    NAMED_STRUCT('city', 'New York', 'zip', 10001)
),
(
    2, FALSE, 0, 0, 0, 0,
    0.0, 0.0, 0.000000,
    '', '',
    '1970-01-01', '1970-01-01T00:00:00.000000', '1970-01-01T00:00:00.000000',
    NULL,
    ARRAY(),
    MAP(),
    NAMED_STRUCT('city', '', 'zip', 0)
),
(
    3, NULL, NULL, NULL, NULL, NULL,
    NULL, NULL, NULL,
    NULL, NULL,
    NULL, NULL, NULL,
    NULL,
    NULL,
    NULL,
    NULL
);

-- The following tables are used for runtime/query tests.
-- Databricks does not support modeling, so data must be ingested offline before running these tests.

-- ad_bids table: used by toplist, aggregation, and comparison tests.
-- Requires loading the full AdBids dataset (~100K rows) from runtime/testruntime/testdata/ad_bids/data/AdBids.csv.gz.
CREATE TABLE IF NOT EXISTS integration_test.ad_bids (
    id        INT,
    timestamp TIMESTAMP,
    publisher STRING,
    domain    STRING,
    bid_price DOUBLE
);

-- timeseries_year table: monthly data from 2022-01 to 2025-12 used by timeseries year/IST/quarter grain tests.
CREATE OR REPLACE TABLE integration_test.timeseries_year AS
SELECT ts AS timestamp, 1.0 AS clicks, 'android' AS device, 'Google' AS publisher, 'Canada' AS country
FROM (SELECT EXPLODE(SEQUENCE(
    TIMESTAMP '2022-01-01 00:00:00',
    TIMESTAMP '2025-12-01 00:00:00',
    INTERVAL 1 MONTH
)) AS ts);

-- timeseries_dst_backwards table: 10-minute intervals around DST fall-back (Nov 2023) used by DST backwards tests.
CREATE OR REPLACE TABLE integration_test.timeseries_dst_backwards AS
WITH continuous AS (
    SELECT 'continuous' AS label, ts AS timestamp
    FROM (SELECT EXPLODE(SEQUENCE(
        TIMESTAMP '2023-11-03 00:00:00',
        TIMESTAMP '2023-11-06 23:50:00',
        INTERVAL 10 MINUTE
    )) AS ts)
),
sparse_hour AS (
    SELECT 'sparse_hour' AS label, ts AS timestamp
    FROM (SELECT EXPLODE(ARRAY(
        TIMESTAMP '2023-11-05 03:00:00',
        TIMESTAMP '2023-11-05 05:00:00',
        TIMESTAMP '2023-11-05 07:00:00'
    )) AS ts)
),
sparse_day AS (
    SELECT 'sparse_day' AS label, ts AS timestamp
    FROM (SELECT EXPLODE(SEQUENCE(
        TIMESTAMP '2023-11-02 00:00:00',
        TIMESTAMP '2023-11-03 23:50:00',
        INTERVAL 10 MINUTE
    )) AS ts)
    UNION ALL
    SELECT 'sparse_day', ts
    FROM (SELECT EXPLODE(SEQUENCE(
        TIMESTAMP '2023-11-05 05:00:00',
        TIMESTAMP '2023-11-05 23:50:00',
        INTERVAL 10 MINUTE
    )) AS ts)
)
SELECT * FROM continuous
UNION ALL SELECT * FROM sparse_hour
UNION ALL SELECT * FROM sparse_day;

-- timeseries_dst_forwards table: intervals around DST spring-forward (Mar 2023) used by DST forwards tests.
CREATE OR REPLACE TABLE integration_test.timeseries_dst_forwards AS
WITH continuous AS (
    SELECT 'continuous' AS label, ts AS timestamp
    FROM (SELECT EXPLODE(SEQUENCE(
        TIMESTAMP '2023-03-10 00:00:00',
        TIMESTAMP '2023-03-13 23:50:00',
        INTERVAL 10 MINUTE
    )) AS ts)
),
sparse_hour AS (
    SELECT 'sparse_hour' AS label, ts AS timestamp
    FROM (SELECT EXPLODE(ARRAY(
        TIMESTAMP '2023-03-12 03:00:00',
        TIMESTAMP '2023-03-12 05:00:00',
        TIMESTAMP '2023-03-12 07:00:00'
    )) AS ts)
),
sparse_day AS (
    SELECT 'sparse_day' AS label, ts AS timestamp
    FROM (SELECT EXPLODE(SEQUENCE(
        TIMESTAMP '2023-03-09 00:00:00',
        TIMESTAMP '2023-03-10 23:00:00',
        INTERVAL 1 HOUR
    )) AS ts)
    UNION ALL
    SELECT 'sparse_day', ts
    FROM (SELECT EXPLODE(SEQUENCE(
        TIMESTAMP '2023-03-12 05:00:00',
        TIMESTAMP '2023-03-12 23:00:00',
        INTERVAL 1 HOUR
    )) AS ts)
)
SELECT * FROM continuous
UNION ALL SELECT * FROM sparse_hour
UNION ALL SELECT * FROM sparse_day;

-- timeseries_gaps table: sparse 2019 data with gaps used by the having_clause timeseries test.
CREATE OR REPLACE TABLE integration_test.timeseries_gaps AS
SELECT 1.0 AS clicks, 3 AS imps, TIMESTAMP '2019-01-01 00:00:00' AS time, DATE '2019-01-01' AS day,
    'android' AS device, 'Google' AS publisher, 'google.com' AS domain, 25 AS latitude, 'Canada' AS country
UNION ALL
SELECT 1.0, 5, TIMESTAMP '2019-01-03 00:00:00', DATE '2019-01-03',
    'iphone', NULL, 'msn.com', NULL, NULL
UNION ALL
SELECT 1.0, 3, TIMESTAMP '2019-01-06 00:00:00', DATE '2019-01-06',
    'iphone', NULL, 'msn.com', NULL, NULL;
