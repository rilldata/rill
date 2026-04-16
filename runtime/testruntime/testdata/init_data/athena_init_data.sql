CREATE DATABASE integration_test;

USE integration_test;

CREATE EXTERNAL TABLE all_datatypes (
    id INT,
    boolean_col BOOLEAN,
    int32_col INT,
    int64_col BIGINT,
    float_col FLOAT,
    double_col DOUBLE,
    byte_array_col BINARY,
    fixed_len_byte_array_col BINARY,
    string_col STRING,
    decimal_col DECIMAL(10,2),
    date_col DATE,
    time_millis_col INT,
    time_micros_col BIGINT,
    timestamp_millis_col TIMESTAMP,
    timestamp_micros_col TIMESTAMP,
    uuid_col BINARY,
    list_int_col ARRAY<INT>,
    list_string_col ARRAY<STRING>,
    map_col MAP<STRING, INT>,
    struct_col STRUCT<
        field_int_col: INT,
        field_float_col: FLOAT,
        field_string_col: STRING
    >
)
STORED AS PARQUET
LOCATION 's3://integration-test.rilldata.com/parquet_test/'
TBLPROPERTIES ('parquet.compress'='SNAPPY');

-- Simple test tables for information schema tests (CTAS with PARQUET format)
CREATE TABLE foo
WITH (format = 'PARQUET', external_location = 's3://integration-test.rilldata.com/athena/foo/')
AS SELECT CAST('a' AS VARCHAR) AS bar, CAST(1 AS INT) AS baz
UNION ALL SELECT 'a', 2
UNION ALL SELECT 'b', 3
UNION ALL SELECT 'c', 4;

CREATE TABLE bar
WITH (format = 'PARQUET', external_location = 's3://integration-test.rilldata.com/athena/bar/')
AS SELECT CAST('a' AS VARCHAR) AS bar, CAST(1 AS INT) AS baz
UNION ALL SELECT 'a', 2
UNION ALL SELECT 'b', 3
UNION ALL SELECT 'c', 4;

CREATE TABLE foz
WITH (format = 'PARQUET', external_location = 's3://integration-test.rilldata.com/athena/foz/')
AS SELECT CAST('a' AS VARCHAR) AS bar, CAST(1 AS INT) AS baz
UNION ALL SELECT 'a', 2
UNION ALL SELECT 'b', 3
UNION ALL SELECT 'c', 4;

CREATE TABLE baz
WITH (format = 'PARQUET', external_location = 's3://integration-test.rilldata.com/athena/baz/')
AS SELECT CAST('a' AS VARCHAR) AS bar, CAST(1 AS INT) AS baz
UNION ALL SELECT 'a', 2
UNION ALL SELECT 'b', 3
UNION ALL SELECT 'c', 4;

CREATE VIEW model AS SELECT 1 AS col1, 2 AS col2, 3 AS col3;
