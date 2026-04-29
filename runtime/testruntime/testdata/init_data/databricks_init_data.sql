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
