-- Removed interval datatype(INTERVAL_MONTH_DAY_NANO) as is not supported by arrow to parquet code. https://github.com/rilldata/arrow/blob/v15.0.2/go/parquet/pqarrow/schema.go#L406
-- Removed range datatype from struct because range inside struct is not supported https://github.com/googleapis/google-cloud-go/blob/main/bigquery/value.go#L899
CREATE OR REPLACE TABLE  `rilldata.integration_test.all_datatypes` (
    int_col INT64,
    float_col FLOAT64,
    numeric_col NUMERIC,
    bignumeric_col BIGNUMERIC,
    bool_col BOOL,
    string_col STRING,
    bytes_col BYTES,
    date_col DATE,
    datetime_col DATETIME,
    time_col TIME,
    timestamp_col TIMESTAMP,
    json_col JSON,
    geography_col GEOGRAPHY,
    range_date_col RANGE<DATE>,
    range_datetime_col RANGE<DATETIME>,
    range_timestamp_col RANGE<TIMESTAMP>,
    array_int_col ARRAY<INT64>,
    array_float_col ARRAY<FLOAT64>,
    array_numeric_col ARRAY<NUMERIC>,
    array_bignumeric_col ARRAY<BIGNUMERIC>,
    array_bool_col ARRAY<BOOL>,
    array_string_col ARRAY<STRING>,
    array_bytes_col ARRAY<BYTES>,
    array_date_col ARRAY<DATE>,
    array_datetime_col ARRAY<DATETIME>,
    array_time_col ARRAY<TIME>,
    array_timestamp_col ARRAY<TIMESTAMP>,
    array_json_col ARRAY<JSON>,
    array_geography_col ARRAY<GEOGRAPHY>,
    array_range_date_col ARRAY<RANGE<DATE>>,
    array_range_datetime_col ARRAY<RANGE<DATETIME>>,
    array_range_timestamp_col ARRAY<RANGE<TIMESTAMP>>,
    array_struct_col ARRAY<STRUCT<
        field_int INT64,
        field_float FLOAT64,
        field_numeric NUMERIC,
        field_bignumeric BIGNUMERIC,
        field_bool BOOL,
        field_string STRING,
        field_bytes BYTES,
        field_date DATE,
        field_datetime DATETIME,
        field_time TIME,
        field_timestamp TIMESTAMP,
        field_json JSON,
        field_geography GEOGRAPHY
    >>,
    struct_col STRUCT<
        field_int INT64,
        field_float FLOAT64,
        field_numeric NUMERIC,
        field_bignumeric BIGNUMERIC,
        field_bool BOOL,
        field_string STRING,
        field_bytes BYTES,
        field_date DATE,
        field_datetime DATETIME,
        field_time TIME,
        field_timestamp TIMESTAMP,
        field_json JSON,
        field_geography GEOGRAPHY,
        field_array_int ARRAY<INT64>,
        field_array_float ARRAY<FLOAT64>,
        field_array_numeric ARRAY<NUMERIC>,
        field_array_bignumeric ARRAY<BIGNUMERIC>,
        field_array_bool ARRAY<BOOL>,
        field_array_string ARRAY<STRING>,
        field_array_bytes ARRAY<BYTES>,
        field_array_date ARRAY<DATE>,
        field_array_datetime ARRAY<DATETIME>,
        field_array_time ARRAY<TIME>,
        field_array_timestamp ARRAY<TIMESTAMP>,
        field_array_json ARRAY<JSON>,
        field_array_geography ARRAY<GEOGRAPHY>
    >
);


INSERT INTO `rilldata.integration_test.all_datatypes` VALUES
    (1, 1.1, Cast(123.45 as NUMERIC), Cast(99999999999999999999999999999999999999.99 as BIGNUMERIC), TRUE, 'sample1', b'abc', DATE(2023,1,1), DATETIME(2023,1,1,12,34,56), TIME(12,34,56), TIMESTAMP("2023-01-01 12:34:56 UTC"), PARSE_JSON('{"key": "value1"}'), ST_GEOGPOINT(1, 2), Range(DATE(2023,1,1),DATE(2023,2,1)), Range(DATETIME(2023,1,1,12,34,56),DATETIME(2024,1,1,12,34,56)), Range(TIMESTAMP("2023-01-01 12:34:56 UTC"),TIMESTAMP("2024-01-01 12:34:56 UTC")), [1], [1.1], [Cast(123.45 as NUMERIC)], [Cast(99999999999999999999999999999999999999.99 as BIGNUMERIC)], [TRUE], ['sample1'], [b'abc'], [DATE(2023,1,1)], [DATETIME(2023,1,1,12,34,56)], [TIME(12,34,56)], [TIMESTAMP("2023-01-01 12:34:56 UTC")], [PARSE_JSON('{"key": "value1"}')], [ST_GEOGPOINT(1, 2)], [Range(DATE(2023,1,1),DATE(2023,2,1))], [Range(DATETIME(2023,1,1,12,34,56),DATETIME(2024,1,1,12,34,56))], [Range(TIMESTAMP("2023-01-01 12:34:56 UTC"),TIMESTAMP("2024-01-01 12:34:56 UTC"))], [(1, 1.1, Cast(123.45 as NUMERIC), Cast(99999999999999999999999999999999999999.99 as BIGNUMERIC), TRUE, 'sample1', b'abc', DATE(2023,1,1), DATETIME(2023,1,1,12,34,56), TIME(12,34,56), TIMESTAMP("2023-01-01 12:34:56 UTC"), PARSE_JSON('{"key": "value1"}'), ST_GEOGPOINT(1, 2))],(1, 1.1, Cast(123.45 as NUMERIC), Cast(99999999999999999999999999999999999999.99 as BIGNUMERIC), TRUE, 'sample1', b'abc', DATE(2023,1,1), DATETIME(2023,1,1,12,34,56), TIME(12,34,56), TIMESTAMP("2023-01-01 12:34:56 UTC"), PARSE_JSON('{"key": "value1"}'), ST_GEOGPOINT(1, 2), [1], [1.1], [Cast(123.45 as NUMERIC)], [Cast(99999999999999999999999999999999999999.99 as BIGNUMERIC)], [TRUE], ['sample1'], [b'abc'], [DATE(2023,1,1)], [DATETIME(2023,1,1,12,34,56)], [TIME(12,34,56)], [TIMESTAMP("2023-01-01 12:34:56 UTC")], [PARSE_JSON('{"key": "value1"}')], [ST_GEOGPOINT(1, 2)] )),
    (0, 0.0, Cast(0.0 as NUMERIC), Cast(0.0 as BIGNUMERIC), FALSE, '', b'', DATE(1970,1,1), DATETIME(1970,1,1,00,00,00), TIME(00,00,00), TIMESTAMP("1970-01-01 00:00:00 UTC"), PARSE_JSON('{}'), ST_GEOGPOINT(0, 0), Range(DATE(1970,1,1),DATE(1970,1,2)), Range(DATETIME(1970,1,1,00,00,00),DATETIME(1970,1,1,00,00,01)), Range(TIMESTAMP("1970-01-01 00:00:00 UTC"),TIMESTAMP("1970-01-01 00:00:01 UTC")), [], [], [], [], [], [], [], [], [], [], [], [], [], [], [], [], [],(0, 0.0, Cast(0.0 as NUMERIC), Cast(0.0 as BIGNUMERIC), FALSE, '', b'', DATE(1970,1,1), DATETIME(1970,1,1,00,00,00), TIME(00,00,00), TIMESTAMP("1970-01-01 00:00:00 UTC"), PARSE_JSON('{}'), ST_GEOGPOINT(0, 0), [], [], [], [], [], [], [], [], [], [], [], [], [])),
    (NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL,NULL);



-- Below is for all datatype but some of them is not supported so not using it for now.
CREATE OR REPLACE TABLE  `rilldata.integration_test.all_datatypes` (
    int_col INT64,
    float_col FLOAT64,
    numeric_col NUMERIC,
    bignumeric_col BIGNUMERIC,
    bool_col BOOL,
    string_col STRING,
    bytes_col BYTES,
    date_col DATE,
    datetime_col DATETIME,
    time_col TIME,
    timestamp_col TIMESTAMP,
    json_col JSON,
    geography_col GEOGRAPHY,
    interval_col INTERVAL,
    range_date_col RANGE<DATE>,
    range_datetime_col RANGE<DATETIME>,
    range_timestamp_col RANGE<TIMESTAMP>,
    array_int_col ARRAY<INT64>,
    array_float_col ARRAY<FLOAT64>,
    array_numeric_col ARRAY<NUMERIC>,
    array_bignumeric_col ARRAY<BIGNUMERIC>,
    array_bool_col ARRAY<BOOL>,
    array_string_col ARRAY<STRING>,
    array_bytes_col ARRAY<BYTES>,
    array_date_col ARRAY<DATE>,
    array_datetime_col ARRAY<DATETIME>,
    array_time_col ARRAY<TIME>,
    array_timestamp_col ARRAY<TIMESTAMP>,
    array_json_col ARRAY<JSON>,
    array_geography_col ARRAY<GEOGRAPHY>,
    array_interval_col ARRAY<INTERVAL>,
    array_range_date_col ARRAY<RANGE<DATE>>,
    array_range_datetime_col ARRAY<RANGE<DATETIME>>,
    array_range_timestamp_col ARRAY<RANGE<TIMESTAMP>>,
    array_struct_col ARRAY<STRUCT<
        field_int INT64,
        field_float FLOAT64,
        field_numeric NUMERIC,
        field_bignumeric BIGNUMERIC,
        field_bool BOOL,
        field_string STRING,
        field_bytes BYTES,
        field_date DATE,
        field_datetime DATETIME,
        field_time TIME,
        field_timestamp TIMESTAMP,
        field_json JSON,
        field_geography GEOGRAPHY,
        field_interval INTERVAL,
        field_range_date RANGE<DATE>,
        field_range_datetime RANGE<DATETIME>,
        field_range_timestamp RANGE<TIMESTAMP>
    >>,
    struct_col STRUCT<
        field_int INT64,
        field_float FLOAT64,
        field_numeric NUMERIC,
        field_bignumeric BIGNUMERIC,
        field_bool BOOL,
        field_string STRING,
        field_bytes BYTES,
        field_date DATE,
        field_datetime DATETIME,
        field_time TIME,
        field_timestamp TIMESTAMP,
        field_json JSON,
        field_geography GEOGRAPHY,
        field_interval INTERVAL,
        field_range_date RANGE<DATE>,
        field_range_datetime RANGE<DATETIME>,
        field_range_timestamp RANGE<TIMESTAMP>,
        field_array_int ARRAY<INT64>,
        field_array_float ARRAY<FLOAT64>,
        field_array_numeric ARRAY<NUMERIC>,
        field_array_bignumeric ARRAY<BIGNUMERIC>,
        field_array_bool ARRAY<BOOL>,
        field_array_string ARRAY<STRING>,
        field_array_bytes ARRAY<BYTES>,
        field_array_date ARRAY<DATE>,
        field_array_datetime ARRAY<DATETIME>,
        field_array_time ARRAY<TIME>,
        field_array_timestamp ARRAY<TIMESTAMP>,
        field_array_json ARRAY<JSON>,
        field_array_geography ARRAY<GEOGRAPHY>,
        field_array_interval ARRAY<INTERVAL>,
        field_array_range_date ARRAY<RANGE<DATE>>,
        field_array_range_datetime ARRAY<RANGE<DATETIME>>,
        field_array_range_timestamp ARRAY<RANGE<TIMESTAMP>>
    >
);


INSERT INTO `rilldata.integration_test.all_datatypes` VALUES
    (1, 1.1, Cast(123.45 as NUMERIC), Cast(99999999999999999999999999999999999999.99 as BIGNUMERIC), TRUE, 'sample1', b'abc', DATE(2023,1,1), DATETIME(2023,1,1,12,34,56), TIME(12,34,56), TIMESTAMP("2023-01-01 12:34:56 UTC"), PARSE_JSON('{"key": "value1"}'), ST_GEOGPOINT(1, 2), INTERVAL 1 DAY, Range(DATE(2023,1,1),DATE(2023,2,1)), Range(DATETIME(2023,1,1,12,34,56),DATETIME(2024,1,1,12,34,56)), Range(TIMESTAMP("2023-01-01 12:34:56 UTC"),TIMESTAMP("2024-01-01 12:34:56 UTC")), [1], [1.1], [Cast(123.45 as NUMERIC)], [Cast(99999999999999999999999999999999999999.99 as BIGNUMERIC)], [TRUE], ['sample1'], [b'abc'], [DATE(2023,1,1)], [DATETIME(2023,1,1,12,34,56)], [TIME(12,34,56)], [TIMESTAMP("2023-01-01 12:34:56 UTC")], [PARSE_JSON('{"key": "value1"}')], [ST_GEOGPOINT(1, 2)], [INTERVAL 1 DAY], [Range(DATE(2023,1,1),DATE(2023,2,1))], [Range(DATETIME(2023,1,1,12,34,56),DATETIME(2024,1,1,12,34,56))], [Range(TIMESTAMP("2023-01-01 12:34:56 UTC"),TIMESTAMP("2024-01-01 12:34:56 UTC"))], [(1, 1.1, Cast(123.45 as NUMERIC), Cast(99999999999999999999999999999999999999.99 as BIGNUMERIC), TRUE, 'sample1', b'abc', DATE(2023,1,1), DATETIME(2023,1,1,12,34,56), TIME(12,34,56), TIMESTAMP("2023-01-01 12:34:56 UTC"), PARSE_JSON('{"key": "value1"}'), ST_GEOGPOINT(1, 2), INTERVAL 1 DAY, Range(DATE(2023,1,1),DATE(2023,2,1)), Range(DATETIME(2023,1,1,12,34,56),DATETIME(2024,1,1,12,34,56)), Range(TIMESTAMP("2023-01-01 12:34:56 UTC"),TIMESTAMP("2024-01-01 12:34:56 UTC")))],(1, 1.1, Cast(123.45 as NUMERIC), Cast(99999999999999999999999999999999999999.99 as BIGNUMERIC), TRUE, 'sample1', b'abc', DATE(2023,1,1), DATETIME(2023,1,1,12,34,56), TIME(12,34,56), TIMESTAMP("2023-01-01 12:34:56 UTC"), PARSE_JSON('{"key": "value1"}'), ST_GEOGPOINT(1, 2), INTERVAL 1 DAY, Range(DATE(2023,1,1),DATE(2023,2,1)), Range(DATETIME(2023,1,1,12,34,56),DATETIME(2024,1,1,12,34,56)), Range(TIMESTAMP("2023-01-01 12:34:56 UTC"),TIMESTAMP("2024-01-01 12:34:56 UTC")), [1], [1.1], [Cast(123.45 as NUMERIC)], [Cast(99999999999999999999999999999999999999.99 as BIGNUMERIC)], [TRUE], ['sample1'], [b'abc'], [DATE(2023,1,1)], [DATETIME(2023,1,1,12,34,56)], [TIME(12,34,56)], [TIMESTAMP("2023-01-01 12:34:56 UTC")], [PARSE_JSON('{"key": "value1"}')], [ST_GEOGPOINT(1, 2)], [INTERVAL 1 DAY], [Range(DATE(2023,1,1),DATE(2023,2,1))], [Range(DATETIME(2023,1,1,12,34,56),DATETIME(2024,1,1,12,34,56))], [Range(TIMESTAMP("2023-01-01 12:34:56 UTC"),TIMESTAMP("2024-01-01 12:34:56 UTC"))])),
    (0, 0.0, Cast(0.0 as NUMERIC), Cast(0.0 as BIGNUMERIC), FALSE, '', b'', DATE(1970,1,1), DATETIME(1970,1,1,00,00,00), TIME(00,00,00), TIMESTAMP("1970-01-01 00:00:00 UTC"), PARSE_JSON('{}'), ST_GEOGPOINT(0, 0), INTERVAL 0 DAY, Range(DATE(1970,1,1),DATE(1970,1,2)), Range(DATETIME(1970,1,1,00,00,00),DATETIME(1970,1,1,00,00,01)), Range(TIMESTAMP("1970-01-01 00:00:00 UTC"),TIMESTAMP("1970-01-01 00:00:01 UTC")), [], [], [], [], [], [], [], [], [], [], [], [], [], [], [], [], [], [],(0, 0.0, Cast(0.0 as NUMERIC), Cast(0.0 as BIGNUMERIC), FALSE, '', b'', DATE(1970,1,1), DATETIME(1970,1,1,00,00,00), TIME(00,00,00), TIMESTAMP("1970-01-01 00:00:00 UTC"), PARSE_JSON('{}'), ST_GEOGPOINT(0, 0), INTERVAL 0 DAY, Range(DATE(1970,1,1),DATE(1970,1,2)), Range(DATETIME(1970,1,1,00,00,00),DATETIME(1970,1,1,00,00,01)), Range(TIMESTAMP("1970-01-01 00:00:00 UTC"),TIMESTAMP("1970-01-01 00:00:01 UTC")), [], [], [], [], [], [], [], [], [], [], [], [], [], [], [], [], [])),
    (NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL,NULL);

CREATE OR REPLACE TABLE `rilldata.integration_test.foo` (bar STRING, baz INT64);
INSERT INTO `rilldata.integration_test.foo` VALUES ('a', 1), ('a', 2), ('b', 3), ('c', 4);

CREATE OR REPLACE TABLE `rilldata.integration_test.bar` (bar STRING, baz INT64);
INSERT INTO `rilldata.integration_test.bar` VALUES ('a', 1), ('a', 2), ('b', 3), ('c', 4);

CREATE OR REPLACE TABLE `rilldata.integration_test.foz` (bar STRING, baz INT64);
INSERT INTO `rilldata.integration_test.foz` VALUES ('a', 1), ('a', 2), ('b', 3), ('c', 4);

CREATE OR REPLACE TABLE `rilldata.integration_test.baz` (bar STRING, baz INT64);
INSERT INTO `rilldata.integration_test.baz` VALUES ('a', 1), ('a', 2), ('b', 3), ('c', 4);

CREATE OR REPLACE VIEW `rilldata.integration_test.model` AS SELECT 1 AS col1, 2 AS col2, 3 AS col3;

-- The following tables are used for runtime/query tests.
-- BigQuery does not support modeling, so data must be ingested offline before running these tests.

-- ad_bids table: used by toplist, aggregation, and comparison tests.
-- Requires loading the full AdBids dataset (~100K rows) from runtime/testruntime/testdata/ad_bids/data/AdBids.csv.gz.
-- Run the following command from the repository root to load the data (after creating the table below):
--
--   bq load --source_format=CSV --skip_leading_rows=1 --replace \
--     rilldata:integration_test.ad_bids \
--     runtime/testruntime/testdata/ad_bids/data/AdBids.csv.gz \
--     id:INTEGER,timestamp:TIMESTAMP,publisher:STRING,domain:STRING,bid_price:FLOAT64
--
CREATE OR REPLACE TABLE `rilldata.integration_test.ad_bids` (
    id        INT64,
    timestamp TIMESTAMP,
    publisher STRING,
    domain    STRING,
    bid_price FLOAT64
);

-- timeseries_year table: monthly data from 2022-01 to 2025-12 used by timeseries year/IST/quarter grain tests.
CREATE OR REPLACE TABLE `rilldata.integration_test.timeseries_year` AS
SELECT ts AS timestamp, 1.0 AS clicks, 'android' AS device, 'Google' AS publisher, 'Canada' AS country
FROM UNNEST(GENERATE_TIMESTAMP_ARRAY(
    TIMESTAMP '2022-01-01 00:00:00 UTC',
    TIMESTAMP '2025-12-01 00:00:00 UTC',
    INTERVAL 1 MONTH
)) AS ts;

-- timeseries_dst_backwards table: 10-minute intervals around DST fall-back (Nov 2023) used by DST backwards tests.
CREATE OR REPLACE TABLE `rilldata.integration_test.timeseries_dst_backwards` AS
WITH continuous AS (
    SELECT 'continuous' AS label, ts AS timestamp
    FROM UNNEST(GENERATE_TIMESTAMP_ARRAY(
        TIMESTAMP '2023-11-03 00:00:00 UTC',
        TIMESTAMP '2023-11-06 23:50:00 UTC',
        INTERVAL 10 MINUTE
    )) AS ts
),
sparse_hour AS (
    SELECT 'sparse_hour' AS label, ts AS timestamp
    FROM UNNEST(ARRAY[
        TIMESTAMP '2023-11-05 03:00:00 UTC',
        TIMESTAMP '2023-11-05 05:00:00 UTC',
        TIMESTAMP '2023-11-05 07:00:00 UTC'
    ]) AS ts
),
sparse_day AS (
    SELECT 'sparse_day' AS label, ts AS timestamp
    FROM UNNEST(GENERATE_TIMESTAMP_ARRAY(
        TIMESTAMP '2023-11-02 00:00:00 UTC',
        TIMESTAMP '2023-11-03 23:50:00 UTC',
        INTERVAL 10 MINUTE
    )) AS ts
    UNION ALL
    SELECT 'sparse_day', ts
    FROM UNNEST(GENERATE_TIMESTAMP_ARRAY(
        TIMESTAMP '2023-11-05 05:00:00 UTC',
        TIMESTAMP '2023-11-05 23:50:00 UTC',
        INTERVAL 10 MINUTE
    )) AS ts
)
SELECT * FROM continuous
UNION ALL SELECT * FROM sparse_hour
UNION ALL SELECT * FROM sparse_day;

-- timeseries_dst_forwards table: intervals around DST spring-forward (Mar 2023) used by DST forwards tests.
CREATE OR REPLACE TABLE `rilldata.integration_test.timeseries_dst_forwards` AS
WITH continuous AS (
    SELECT 'continuous' AS label, ts AS timestamp
    FROM UNNEST(GENERATE_TIMESTAMP_ARRAY(
        TIMESTAMP '2023-03-10 00:00:00 UTC',
        TIMESTAMP '2023-03-13 23:50:00 UTC',
        INTERVAL 10 MINUTE
    )) AS ts
),
sparse_hour AS (
    SELECT 'sparse_hour' AS label, ts AS timestamp
    FROM UNNEST(ARRAY[
        TIMESTAMP '2023-03-12 03:00:00 UTC',
        TIMESTAMP '2023-03-12 05:00:00 UTC',
        TIMESTAMP '2023-03-12 07:00:00 UTC'
    ]) AS ts
),
sparse_day AS (
    SELECT 'sparse_day' AS label, ts AS timestamp
    FROM UNNEST(GENERATE_TIMESTAMP_ARRAY(
        TIMESTAMP '2023-03-09 00:00:00 UTC',
        TIMESTAMP '2023-03-10 23:00:00 UTC',
        INTERVAL 1 HOUR
    )) AS ts
    UNION ALL
    SELECT 'sparse_day', ts
    FROM UNNEST(GENERATE_TIMESTAMP_ARRAY(
        TIMESTAMP '2023-03-12 05:00:00 UTC',
        TIMESTAMP '2023-03-12 23:00:00 UTC',
        INTERVAL 1 HOUR
    )) AS ts
)
SELECT * FROM continuous
UNION ALL SELECT * FROM sparse_hour
UNION ALL SELECT * FROM sparse_day;

-- timeseries_gaps table: sparse 2019 data with gaps used by the having_clause timeseries test.
CREATE OR REPLACE TABLE `rilldata.integration_test.timeseries_gaps` AS
SELECT 1.0 AS clicks, 3 AS imps, TIMESTAMP '2019-01-01 00:00:00 UTC' AS time, DATE '2019-01-01' AS day,
    'android' AS device, 'Google' AS publisher, 'google.com' AS domain, 25 AS latitude, 'Canada' AS country
UNION ALL
SELECT 1.0, 5, TIMESTAMP '2019-01-03 00:00:00 UTC', DATE '2019-01-03',
    'iphone', NULL, 'msn.com', NULL, NULL
UNION ALL
SELECT 1.0, 3, TIMESTAMP '2019-01-06 00:00:00 UTC', DATE '2019-01-06',
    'iphone', NULL, 'msn.com', NULL, NULL;
