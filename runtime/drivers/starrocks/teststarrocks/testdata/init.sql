-- StarRocks Test Database Initialization
-- Version: 4.0.3
-- This script creates test tables with all supported StarRocks data types

CREATE DATABASE IF NOT EXISTS test_db;
USE test_db;

-- Table 1: All basic data types (Duplicate Key Model)
CREATE TABLE IF NOT EXISTS all_types (
    -- Primary key
    id INT NOT NULL,

    -- Numeric types
    bool_col BOOLEAN,
    tinyint_col TINYINT,
    smallint_col SMALLINT,
    int_col INT,
    bigint_col BIGINT,
    largeint_col LARGEINT,
    float_col FLOAT,
    double_col DOUBLE,
    decimal_col DECIMAL(18, 4),

    -- String types
    char_col CHAR(10),
    varchar_col VARCHAR(255),
    string_col STRING,

    -- Date/Time types
    date_col DATE,
    datetime_col DATETIME,

    -- Semi-structured types
    json_col JSON,
    array_col ARRAY<INT>,
    map_col MAP<STRING, INT>,
    struct_col STRUCT<name STRING, age INT>
)
DUPLICATE KEY(id)
DISTRIBUTED BY HASH(id) BUCKETS 1
PROPERTIES ("replication_num" = "1");

-- Table 2: Aggregate types (Aggregate Key Model)
-- HLL, BITMAP, PERCENTILE are only available in aggregate tables
CREATE TABLE IF NOT EXISTS aggregate_types (
    -- Key column
    id INT NOT NULL,
    dt DATE NOT NULL,

    -- Aggregate columns with special types
    hll_col HLL HLL_UNION,
    bitmap_col BITMAP BITMAP_UNION,
    count_col BIGINT SUM DEFAULT "0"
)
AGGREGATE KEY(id, dt)
DISTRIBUTED BY HASH(id) BUCKETS 1
PROPERTIES ("replication_num" = "1");

-- Table 3: Binary type test (StarRocks 4.0+)
CREATE TABLE IF NOT EXISTS binary_types (
    id INT NOT NULL,
    binary_col VARBINARY(255),
    blob_col VARBINARY(65535)
)
DUPLICATE KEY(id)
DISTRIBUTED BY HASH(id) BUCKETS 1
PROPERTIES ("replication_num" = "1");

-- Insert test data into all_types
INSERT INTO all_types VALUES
(1, true, 127, 32767, 2147483647, 9223372036854775807, 170141183460469231731687303715884105727,
 3.14, 3.141592653589793, 12345.6789,
 'char_val', 'varchar_value', 'string_value',
 '2024-01-15', '2024-01-15 10:30:00',
 '{"key": "value", "num": 123}',
 [1, 2, 3, 4, 5],
 map{"key1": 1, "key2": 2},
 named_struct("name", "John", "age", 30)),
(2, false, -128, -32768, -2147483648, -9223372036854775808, -170141183460469231731687303715884105728,
 -3.14, -3.141592653589793, -12345.6789,
 'char_2', 'varchar_2', 'string_2',
 '2024-06-20', '2024-06-20 15:45:30',
 '{"nested": {"array": [1,2,3]}}',
 [10, 20, 30],
 map{"a": 100, "b": 200},
 named_struct("name", "Jane", "age", 25)),
(3, NULL, NULL, NULL, NULL, NULL, NULL,
 NULL, NULL, NULL,
 NULL, NULL, NULL,
 NULL, NULL,
 NULL,
 NULL,
 NULL,
 NULL);

-- Insert test data into aggregate_types
INSERT INTO aggregate_types VALUES
(1, '2024-01-01', hll_hash('user1'), to_bitmap(100), 10),
(1, '2024-01-01', hll_hash('user2'), to_bitmap(101), 20),
(2, '2024-01-02', hll_hash('user3'), to_bitmap(200), 30);

-- Insert test data into binary_types
INSERT INTO binary_types VALUES
(1, x'48454C4C4F', x'576F726C64'),
(2, x'0102030405', x'AABBCCDDEE'),
(3, NULL, NULL);

-- Table 4: String encoding test (UTF-8, special characters)
CREATE TABLE IF NOT EXISTS string_encoding_test (
    id INT NOT NULL,
    ascii_col VARCHAR(255),
    unicode_col VARCHAR(255),
    emoji_col VARCHAR(255),
    korean_col VARCHAR(255),
    chinese_col VARCHAR(255),
    japanese_col VARCHAR(255)
)
DUPLICATE KEY(id)
DISTRIBUTED BY HASH(id) BUCKETS 1
PROPERTIES ("replication_num" = "1");

INSERT INTO string_encoding_test VALUES
(1, 'Hello World', 'Héllo Wörld', '😀🎉🚀', '안녕하세요', '你好世界', 'こんにちは'),
(2, 'Test 123', 'Tëst 456', '👍👎', '테스트', '测试', 'テスト');

-- Table 5: Complex types (ARRAY, MAP, STRUCT)
CREATE TABLE IF NOT EXISTS complex_types (
    id INT NOT NULL,
    array_col ARRAY<INT>,
    map_col MAP<STRING, INT>,
    struct_col STRUCT<name STRING, age INT>
)
DUPLICATE KEY(id)
DISTRIBUTED BY HASH(id) BUCKETS 1
PROPERTIES ("replication_num" = "1");

INSERT INTO complex_types VALUES
(1, [1, 2, 3], map{'a': 1, 'b': 2}, row('John', 30)),
(2, [4, 5], map{'c': 3}, row('Jane', 25)),
(3, NULL, NULL, NULL);

-- Table 6: Boundary values test
CREATE TABLE IF NOT EXISTS boundary_values (
    id INT NOT NULL,
    tinyint_min TINYINT,
    tinyint_max TINYINT,
    smallint_min SMALLINT,
    smallint_max SMALLINT,
    int_min INT,
    int_max INT,
    bigint_min BIGINT,
    bigint_max BIGINT,
    empty_string VARCHAR(255),
    whitespace_string VARCHAR(255)
)
DUPLICATE KEY(id)
DISTRIBUTED BY HASH(id) BUCKETS 1
PROPERTIES ("replication_num" = "1");

INSERT INTO boundary_values VALUES
(1, -128, 127, -32768, 32767, -2147483648, 2147483647, -9223372036854775808, 9223372036854775807, '', '   ');

-- Table 7: Special characters test
CREATE TABLE IF NOT EXISTS special_chars (
    id INT NOT NULL,
    newline_col VARCHAR(255),
    tab_col VARCHAR(255),
    quote_col VARCHAR(255),
    emoji_col VARCHAR(255),
    sql_injection_col VARCHAR(255)
)
DUPLICATE KEY(id)
DISTRIBUTED BY HASH(id) BUCKETS 1
PROPERTIES ("replication_num" = "1");

INSERT INTO special_chars VALUES
(1, 'line1\nline2', 'col1\tcol2', 'it''s a "test"', '😀🎉', 'SELECT * FROM users; DROP TABLE--');

-- Table 8: High-precision DECIMAL test (DECIMAL32, DECIMAL64, DECIMAL128)
-- StarRocks uses different internal types based on precision:
-- - DECIMAL(1-9, S)   → DECIMAL32
-- - DECIMAL(10-18, S) → DECIMAL64
-- - DECIMAL(19-38, S) → DECIMAL128
CREATE TABLE IF NOT EXISTS decimal_precision_test (
    id INT NOT NULL,
    decimal32_col DECIMAL(9, 4),
    decimal64_col DECIMAL(18, 6),
    decimal128_col DECIMAL(38, 10)
)
DUPLICATE KEY(id)
DISTRIBUTED BY HASH(id) BUCKETS 1
PROPERTIES ("replication_num" = "1");

INSERT INTO decimal_precision_test VALUES
(1, 12345.6789, 123456789012.345678, 12345678901234567890123456.7890123456),
(2, -99999.9999, -999999999999.999999, -99999999999999999999999999.9999999999),
(3, 0.0001, 0.000001, 0.0000000001);

-- Table 9: Ad Bids table for metricsview tests
-- This table mirrors the structure used in other OLAP driver tests (ClickHouse, DuckDB)
CREATE TABLE IF NOT EXISTS ad_bids (
    id INT NOT NULL,
    timestamp DATETIME NOT NULL,
    publisher VARCHAR(255),
    domain VARCHAR(255),
    bid_price DOUBLE
)
DUPLICATE KEY(id)
DISTRIBUTED BY HASH(id) BUCKETS 1
PROPERTIES ("replication_num" = "1");

-- Ad bids data is loaded from AdBids.csv.gz via LOAD DATA LOCAL INFILE in teststarrocks.go
