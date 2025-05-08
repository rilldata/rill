CREATE DATABASE test_db;

DROP table all_datatypes;

CREATE TABLE all_datatypes (
        id                        INT PRIMARY KEY,
        boolean_col               BOOLEAN,
        int32_col                 INT,
        int64_col                 BIGINT,
        float_col                 REAL,  
        double_col                DOUBLE PRECISION,
        -- byte_array_col            VARBYTE(100), -- UNLOAD varbyte column "byte_array_col" is only supported for TEXT/CSV.
        string_col                VARCHAR(65535),
        decimal_col               DECIMAL(10,2),
        date_col                  DATE,
        -- time_col                  TIME, -- UNLOAD time without time zone column "time_col" is only supported for TEXT/CSV.
        -- timez_col                 TIMETZ, -- UNLOAD time with time zone column "timez_col" is only supported for TEXT/CSV.
        timestamp_col             TIMESTAMP,
        timestamptz_col           TIMESTAMPTZ,
        interval_year_month       INTERVAL YEAR TO MONTH,
        interval_day_second       INTERVAL DAY TO SECOND,
        -- uuid_col                  VARBYTE(16), -- UNLOAD varbyte column "uuid_col" is only supported for TEXT/CSV.
        list_int_col              SUPER,
        list_string_col           SUPER,
        map_col                   SUPER,
        struct_col                SUPER
);


INSERT INTO all_datatypes (
    id, 
    boolean_col, 
    int32_col, 
    int64_col, 
    float_col, 
    double_col, 
    -- byte_array_col, 
    string_col, 
    decimal_col, 
    date_col, 
    -- time_col, 
    -- timez_col, 
    timestamp_col, 
    timestamptz_col, 
    interval_year_month, 
    interval_day_second, 
    -- uuid_col, 
    list_int_col, 
    list_string_col,
    map_col, 
    struct_col
)
VALUES
-- 1. Non-zero values
(1, 
 TRUE, 
 123, 
 1234567890, 
 1.23, 
 123.456, 
-- FROM_HEX('deadbeef'),
 'Hello, world!', 
 99.99,
 DATE '2023-01-01', 
--  TIME '12:34:56', 
--  '12:34:56+01', 
 TIMESTAMP '2023-01-01 12:34:56', 
 TIMESTAMPTZ '2023-01-01 12:34:56+00',
 INTERVAL '2-6' YEAR TO MONTH, 
 INTERVAL '10 12:30:15' DAY TO SECOND,
--  FROM_HEX('550e8400e29b41d4a716446655440000'),
 JSON_PARSE('[1, 2, 3]'),
 JSON_PARSE('["a", "b", "c"]'),
 JSON_PARSE('{"key1": "value1", "key2": "value2"}'),
 JSON_PARSE('{"field1": 10, "field2": "value"}')
),

-- 2. Zero values
(2, 
 FALSE, 
 0, 
 0, 
 0.0, 
 0.0,
--  FROM_HEX('00'),
 '', 
 0.00,
 DATE '1970-01-01', 
--  TIME '00:00:00', 
--  '00:00:00+00',
 TIMESTAMP '1970-01-01 00:00:00', 
 TIMESTAMPTZ '1970-01-01 00:00:00+00',
 INTERVAL '0-0' YEAR TO MONTH, 
 INTERVAL '0 00:00:00' DAY TO SECOND,
--  FROM_HEX('00000000000000000000000000000000'),
 JSON_PARSE('[]'), 
 JSON_PARSE('[]'),
 JSON_PARSE('{}'), 
 JSON_PARSE('{}')
),

-- 3. Null values
(3, 
 NULL, 
 NULL, 
 NULL, 
 NULL, 
 NULL,
-- NULL, 
 NULL, 
 NULL,
 NULL, 
-- NULL, 
-- NULL,
 NULL, 
 NULL,
 NULL, 
 NULL,
--  NULL,
 NULL, 
 NULL, 
 NULL, 
 NULL
);




