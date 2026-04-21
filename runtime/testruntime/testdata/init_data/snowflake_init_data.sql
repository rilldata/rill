CREATE DATABASE integration_test;

USE DATABASE integration_test;
USE SCHEMA public; 

CREATE TABLE integration_test.public.all_datatypes (
    id                 INT PRIMARY KEY,
    boolean_col        BOOLEAN,
    tinyint_col        TINYINT,
    smallint_col       SMALLINT,
    int32_col          INT,
    int64_col          BIGINT,
    number_col         NUMBER(38,10),
    float_col          REAL,
    double_col         DOUBLE,
    decimal_col        DECIMAL(18,6),
    string_col         STRING,
    text_col           TEXT,
    date_col           DATE,
    time_col           TIME,
    timestamp_ntz_col  TIMESTAMP_NTZ,
    timestamp_ltz_col  TIMESTAMP_LTZ,
    timestamp_tz_col   TIMESTAMP_TZ,
    variant_col        VARIANT,
    array_col         ARRAY,
    object_col         OBJECT,
    binary_col         BINARY,
    geography_col      GEOGRAPHY,
    geometry_col       GEOMETRY
);


INSERT INTO integration_test.public.all_datatypes 
SELECT 
    1, TRUE, 127, 32767, 2147483647, 9223372036854775807, 
    12345.6789, 3.14, 2.718, 456.789, 
    'Sample String', 'Large text data', 
    '2024-03-26', '14:30:00', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, 
    PARSE_JSON('{"key": "value"}'), 
    ARRAY_CONSTRUCT(1, 2, 3), 
    OBJECT_CONSTRUCT('city', 'New York'), 
    TO_BINARY(HEX_ENCODE('Hello')), 
    TO_GEOGRAPHY('POINT(-122.4194 37.7749)'), 
    TO_GEOMETRY('LINESTRING(0 0, 1 1, 2 2)')

UNION ALL

SELECT 
    2, FALSE, 0, 0, 0, 0, 
    0.0, 0.0, 0.0, 0.0, 
    '', '', 
    '1970-01-01', '00:00:00', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, 
    PARSE_JSON('{}'), 
    ARRAY_CONSTRUCT(), 
    OBJECT_CONSTRUCT(), 
    NULL, 
    TO_GEOGRAPHY('POINT(0 0)'), 
    TO_GEOMETRY('POINT(0 0)')

UNION ALL

SELECT 
    3, NULL, NULL, NULL, NULL, NULL, 
    NULL, NULL, NULL, NULL, 
    NULL, NULL, 
    NULL, NULL, NULL, NULL, NULL, 
    NULL, 
    NULL, 
    NULL, 
    NULL,
    NULL,
    NULL;


-- The following tables are used for runtime/query tests.
-- Snowflake does not support modeling, so data must be ingested offline before running these tests.

-- ad_bids table: used by toplist, aggregation, and comparison tests.
-- Requires loading the full AdBids dataset (~100K rows) from runtime/testruntime/testdata/ad_bids/data/AdBids.csv.gz.
-- Upload the CSV to an internal stage and run the following (after creating the table below):
--
--   PUT file:///path/to/AdBids.csv.gz @my_stage;
--   COPY INTO integration_test.public.ad_bids
--     FROM @my_stage/AdBids.csv.gz
--     FILE_FORMAT = (TYPE = 'CSV' SKIP_HEADER = 1 FIELD_OPTIONALLY_ENCLOSED_BY = '"')
--     ON_ERROR = 'ABORT_STATEMENT';
--
CREATE OR REPLACE TABLE integration_test.public.ad_bids (
    id        INTEGER,
    timestamp TIMESTAMP_NTZ,
    publisher STRING,
    domain    STRING,
    bid_price FLOAT
);

-- timeseries_year table: monthly data from 2022-01 to 2025-12 used by timeseries year/IST/quarter grain tests.
-- 48 rows: SEQ4() produces 0..47, DATEADD(MONTH, 0..47, 2022-01-01) = 2022-01-01..2025-12-01.
CREATE OR REPLACE TABLE integration_test.public.timeseries_year AS
SELECT timestamp, 1.0 AS clicks, 'android' AS device, 'Google' AS publisher, 'Canada' AS country
FROM (
    SELECT DATEADD(MONTH, SEQ4(), TO_TIMESTAMP_NTZ('2022-01-01')) AS timestamp
    FROM TABLE(GENERATOR(ROWCOUNT => 48))
)
WHERE timestamp <= TO_TIMESTAMP_NTZ('2025-12-01');

-- timeseries_dst_backwards table: 10-minute intervals around DST fall-back (Nov 2023) used by DST backwards tests.
CREATE OR REPLACE TABLE integration_test.public.timeseries_dst_backwards AS
WITH continuous AS (
    -- 576 rows: 2023-11-03 00:00 to 2023-11-06 23:50 at 10-min intervals (5750 min / 10 + 1 = 576)
    SELECT 'continuous' AS label, DATEADD(MINUTE, SEQ4() * 10, TO_TIMESTAMP_NTZ('2023-11-03 00:00:00')) AS timestamp
    FROM TABLE(GENERATOR(ROWCOUNT => 576))
),
sparse_hour AS (
    SELECT 'sparse_hour' AS label, column1 AS timestamp
    FROM VALUES
        (TO_TIMESTAMP_NTZ('2023-11-05 03:00:00')),
        (TO_TIMESTAMP_NTZ('2023-11-05 05:00:00')),
        (TO_TIMESTAMP_NTZ('2023-11-05 07:00:00'))
),
sparse_day_part1 AS (
    -- 288 rows: 2023-11-02 00:00 to 2023-11-03 23:50 at 10-min intervals (2870 min / 10 + 1 = 288)
    SELECT 'sparse_day' AS label, DATEADD(MINUTE, SEQ4() * 10, TO_TIMESTAMP_NTZ('2023-11-02 00:00:00')) AS timestamp
    FROM TABLE(GENERATOR(ROWCOUNT => 288))
),
sparse_day_part2 AS (
    -- 114 rows: 2023-11-05 05:00 to 2023-11-05 23:50 at 10-min intervals (1130 min / 10 + 1 = 114)
    SELECT 'sparse_day' AS label, DATEADD(MINUTE, SEQ4() * 10, TO_TIMESTAMP_NTZ('2023-11-05 05:00:00')) AS timestamp
    FROM TABLE(GENERATOR(ROWCOUNT => 114))
)
SELECT * FROM continuous
UNION ALL SELECT * FROM sparse_hour
UNION ALL SELECT * FROM sparse_day_part1
UNION ALL SELECT * FROM sparse_day_part2;

-- timeseries_dst_forwards table: intervals around DST spring-forward (Mar 2023) used by DST forwards tests.
CREATE OR REPLACE TABLE integration_test.public.timeseries_dst_forwards AS
WITH continuous AS (
    -- 576 rows: 2023-03-10 00:00 to 2023-03-13 23:50 at 10-min intervals (5750 min / 10 + 1 = 576)
    SELECT 'continuous' AS label, DATEADD(MINUTE, SEQ4() * 10, TO_TIMESTAMP_NTZ('2023-03-10 00:00:00')) AS timestamp
    FROM TABLE(GENERATOR(ROWCOUNT => 576))
),
sparse_hour AS (
    SELECT 'sparse_hour' AS label, column1 AS timestamp
    FROM VALUES
        (TO_TIMESTAMP_NTZ('2023-03-12 03:00:00')),
        (TO_TIMESTAMP_NTZ('2023-03-12 05:00:00')),
        (TO_TIMESTAMP_NTZ('2023-03-12 07:00:00'))
),
sparse_day_part1 AS (
    -- 48 rows: 2023-03-09 00:00 to 2023-03-10 23:00 at 1-hour intervals (47 hours + 1 = 48)
    SELECT 'sparse_day' AS label, DATEADD(HOUR, SEQ4(), TO_TIMESTAMP_NTZ('2023-03-09 00:00:00')) AS timestamp
    FROM TABLE(GENERATOR(ROWCOUNT => 48))
),
sparse_day_part2 AS (
    -- 19 rows: 2023-03-12 05:00 to 2023-03-12 23:00 at 1-hour intervals (18 hours + 1 = 19)
    SELECT 'sparse_day' AS label, DATEADD(HOUR, SEQ4(), TO_TIMESTAMP_NTZ('2023-03-12 05:00:00')) AS timestamp
    FROM TABLE(GENERATOR(ROWCOUNT => 19))
)
SELECT * FROM continuous
UNION ALL SELECT * FROM sparse_hour
UNION ALL SELECT * FROM sparse_day_part1
UNION ALL SELECT * FROM sparse_day_part2;

-- timeseries_gaps table: sparse 2019 data with gaps used by the having_clause timeseries test.
CREATE OR REPLACE TABLE integration_test.public.timeseries_gaps AS
SELECT 1.0 AS clicks, 3 AS imps, TO_TIMESTAMP_NTZ('2019-01-01') AS ts, '2019-01-01'::DATE AS day,
    'android' AS device, 'Google' AS publisher, 'google.com' AS domain, 25 AS latitude, 'Canada' AS country
UNION ALL
SELECT 1.0, 5, TO_TIMESTAMP_NTZ('2019-01-03'), '2019-01-03'::DATE, 'iphone', NULL, 'msn.com', NULL, NULL
UNION ALL
SELECT 1.0, 3, TO_TIMESTAMP_NTZ('2019-01-06'), '2019-01-06'::DATE, 'iphone', NULL, 'msn.com', NULL, NULL;
