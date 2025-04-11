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



