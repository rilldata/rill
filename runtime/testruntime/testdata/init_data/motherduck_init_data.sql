
-- TODO : uncomment uhugeint_col,bit_col once supported by duckdb go https://github.com/marcboeker/go-duckdb/blob/317bce30043ed8034997ad40b96e67212ddb752e/type.go#L51
-- TODO: uncomment union once we support it.
-- Create schema
CREATE SCHEMA IF NOT EXISTS integration_test;

-- Create table with all DuckDB-supported data types, including basic types in nested structures
CREATE TABLE integration_test.all_datatypes (
    -- Numeric Types
    tinyint_col TINYINT,
    smallint_col SMALLINT,
    integer_col INTEGER,
    bigint_col BIGINT,
    hugeint_col HUGEINT,
    utinyint_col UTINYINT,
    usmallint_col USMALLINT,
    uinteger_col UINTEGER,
    ubigint_col UBIGINT,
    -- uhugeint_col UHUGEINT,
    float_col FLOAT,
    double_col DOUBLE,
    decimal_col DECIMAL(18,3),

    -- Boolean Type
    boolean_col BOOLEAN,

    -- Character Types
    varchar_col VARCHAR,
    -- bit_col BIT,

    -- Binary Type
    blob_col BLOB,

    -- Date/Time Types
    date_col DATE,
    time_col TIME,
    timestamp_col TIMESTAMP,
    timestamp_s_col TIMESTAMP_S,
    timestamp_ms_col TIMESTAMP_MS,
    timestamp_ns_col TIMESTAMP_NS,
    timestamptz_col TIMESTAMP WITH TIME ZONE,
    interval_col INTERVAL,

    -- UUID Type
    uuid_col UUID,

    -- Nested Types with basic types inside
    list_int_col INTEGER[],
    list_float_col FLOAT[],
    list_varchar_col VARCHAR[],
    list_boolean_col BOOLEAN[],
    list_timestamp_col TIMESTAMP[],
    list_date_col DATE[],

    array_int_col INTEGER[3],
    array_float_col FLOAT[3],
    array_varchar_col VARCHAR[3],
    array_boolean_col BOOLEAN[3],
    array_date_col DATE[3],
    array_timestamp_col TIMESTAMP[3],

    map_int_col MAP(VARCHAR, INTEGER),
    map_float_col MAP(VARCHAR, FLOAT),
    map_varchar_col MAP(VARCHAR, VARCHAR),
    map_boolean_col MAP(VARCHAR, BOOLEAN),
    map_date_col MAP(VARCHAR, DATE),
    map_timestamp_col MAP(VARCHAR, TIMESTAMP),

    struct_col STRUCT(
        field_int INTEGER,
        field_float FLOAT,
        field_str VARCHAR,
        field_bool BOOLEAN,
        field_date DATE,
        field_timestamp TIMESTAMP
    )
    -- ,
    -- union_col UNION(
    --     int_val INTEGER,
    --     float_val FLOAT,
    --     str_val VARCHAR,
    --     bool_val BOOLEAN,
    --     date_val DATE,
    --     timestamp_val TIMESTAMP
    -- )
);

INSERT INTO integration_test.all_datatypes (
    tinyint_col, smallint_col, integer_col, bigint_col, hugeint_col, utinyint_col, usmallint_col, 
    uinteger_col, ubigint_col, 
    -- uhugeint_col, 
    float_col, double_col, decimal_col, boolean_col, 
    varchar_col, 
    -- bit_col, 
    blob_col, date_col, time_col, 
    timestamp_col, timestamp_s_col, timestamp_ms_col, 
    timestamp_ns_col, timestamptz_col, 
    interval_col, uuid_col, 
    list_int_col, list_float_col, list_varchar_col, list_boolean_col, 
    list_date_col, list_timestamp_col, 
    array_int_col, array_float_col, array_varchar_col, array_boolean_col, 
    array_date_col, array_timestamp_col, 
    map_int_col, map_float_col, map_varchar_col, map_boolean_col,
    map_date_col, map_timestamp_col, 
    struct_col
    -- , union_col
) VALUES 
-- Row 1: All values
(
    127, 32767, 2147483647, 9223372036854775807, 1234567890123456789, 255, 65535, 
    4294967295, 18446744073709551615, 
    -- 9223372036854775807, 
    3.14159, 2.71828, 1234.567, TRUE, 
    'Hello', 
    -- '101010'::BITSTRING, 
    BLOB 'Hello Blob', DATE '2025-04-29', TIME '12:34:56',
    TIMESTAMP '2025-04-29 12:34:56', TIMESTAMP_S '2025-04-29 12:34:56.123456', TIMESTAMP_MS '2025-04-29 12:34:56.123',
    TIMESTAMP_NS '2025-04-29 12:34:56.123456789', TIMESTAMP WITH TIME ZONE '2025-04-29 12:34:56+00:00', 
    INTERVAL '1 YEAR', UUID '550e8400-e29b-41d4-a716-446655440000', 
    [1, 2], [1.1, 2.2], ['A', 'B'], [TRUE, FALSE], 
    [DATE '2025-04-29', DATE '2025-05-01'], [TIMESTAMP '2025-04-29 12:34:56', TIMESTAMP '2025-05-01 12:34:56'],
    [1, 2, 3], [1.1, 2.2, 3.3], ['A', 'B', 'C'], [TRUE, FALSE, TRUE], 
    [DATE '2025-04-29', DATE '2025-05-01', DATE '2025-05-02'], [TIMESTAMP '2025-04-29 12:34:56', TIMESTAMP '2025-05-01 12:34:56',  TIMESTAMP '2025-05-02 12:34:56'],
    MAP{'key1':1,'key2': 2}, MAP{'key1': 1.1, 'key2': 2.2}, MAP{'key1': 'A', 'key2': 'B'}, MAP{'key1': TRUE, 'key2': FALSE},
    MAP{'key1': DATE '2025-04-29', 'key2': DATE '2025-05-01'}, MAP{'key1': TIMESTAMP '2025-04-29 12:34:56', 'key2': TIMESTAMP '2025-05-01 12:34:56'}, STRUCT_PACK(field_int := 123, field_float := 1.23, field_str := 'Struct Field', field_bool := TRUE, field_date := DATE '2025-04-29', field_timestamp := TIMESTAMP '2025-04-29 12:34:56')
    -- ,UNION_VALUE(int_val := 0)
  ),
-- Row 2: All zero values
(
    0, 0, 0, 0, 0, 0, 0, 
    0, 0, 
    -- 0, 
    0.0, 0.0, 0.0, FALSE, 
    '',
    --  '0'::BITSTRING, 
     BLOB '', DATE '1970-01-01', TIME '00:00:00', 
    TIMESTAMP '1970-01-01 00:00:00', TIMESTAMP_S '1970-01-01 00:00:00', TIMESTAMP_MS '1970-01-01 00:00:00', 
    TIMESTAMP_NS '1970-01-01 00:00:00', TIMESTAMP WITH TIME ZONE '1970-01-01 00:00:00+00:00', 
    INTERVAL '0 YEAR', UUID '00000000-0000-0000-0000-000000000000', 
    [0, 0], [0.0, 0.0], ['', ''], [FALSE, FALSE],
    [DATE '1970-01-01', DATE '1970-01-01'], [TIMESTAMP '1970-01-01 00:00:00', TIMESTAMP '1970-01-01 00:00:00'],
    [0, 0, 0], [0.0, 0.0, 0.0], ['', '', ''], [FALSE, FALSE, FALSE], 
    [DATE '1970-01-01', DATE '1970-01-01', DATE '1970-01-01'], [TIMESTAMP '1970-01-01 00:00:00', TIMESTAMP '1970-01-01 00:00:00', TIMESTAMP '1970-01-01 00:00:00'],
    MAP{'': 0},MAP{'': 0.0},MAP{'': ''},MAP{'': FALSE},
    MAP{'': DATE '1970-01-01'},MAP{'': TIMESTAMP '1970-01-01 00:00:00'},
    STRUCT_PACK(field_int := 0, field_str := '', field_bool := FALSE, field_float := 0.0, field_date := DATE '1970-01-01', field_timestamp := TIMESTAMP '1970-01-01 00:00:00')
    -- ,UNION_VALUE(str_val := '')
  ),
-- Row 3: All NULL values
(
    NULL, NULL, NULL, NULL, NULL, NULL, NULL, 
    NULL, NULL, 
    -- NULL, 
    NULL, NULL, NULL, NULL, 
    NULL, 
    -- NULL,
    NULL, NULL, NULL, 
    NULL, NULL, NULL, 
    NULL, NULL,
    NULL, NULL, 
    NULL, NULL, NULL, NULL,
    NULL, NULL,
    NULL, NULL, NULL, NULL, 
    NULL, NULL, 
    NULL, NULL, NULL, NULL,
    NULL, NULL, 
    NULL
    -- , NULL                          
);
