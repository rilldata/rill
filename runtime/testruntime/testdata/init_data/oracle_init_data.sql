CREATE TABLE all_datatypes (
    id NUMBER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    number_col NUMBER,
    binary_float_col BINARY_FLOAT,
    binary_double_col BINARY_DOUBLE,
    integer_col INTEGER,
    smallint_col SMALLINT,
    varchar2_col VARCHAR2(255),
    nvarchar2_col NVARCHAR2(255),
    char_col CHAR(10),
    nchar_col NCHAR(10),
    clob_col CLOB,
    blob_col BLOB,
    raw_col RAW(100),
    date_col DATE,
    timestamp_col TIMESTAMP,
    timestamp_tz_col TIMESTAMP WITH TIME ZONE,
    boolean_col NUMBER(1),
    json_col CLOB CONSTRAINT json_col_json CHECK (json_col IS JSON)
);

INSERT INTO all_datatypes (number_col, binary_float_col, binary_double_col, integer_col, smallint_col, varchar2_col, nvarchar2_col, char_col, nchar_col, clob_col, blob_col, raw_col, date_col, timestamp_col, timestamp_tz_col, boolean_col, json_col)
VALUES (
    42.5, 3.14, 2.718281828, 1234567, 123,
    'Hello World', N'Unicode Text', 'ABCD', N'XY',
    'This is a CLOB text field for testing.',
    UTL_RAW.CAST_TO_RAW('binary data'),
    UTL_RAW.CAST_TO_RAW('raw bytes'),
    TO_DATE('2024-02-14', 'YYYY-MM-DD'),
    TO_TIMESTAMP('2025-02-14 12:34:56.789', 'YYYY-MM-DD HH24:MI:SS.FF3'),
    TO_TIMESTAMP_TZ('2025-02-14 12:34:56.789 +05:30', 'YYYY-MM-DD HH24:MI:SS.FF3 TZH:TZM'),
    1,
    '{"key": "value", "number": 42}'
);

INSERT INTO all_datatypes (number_col, binary_float_col, binary_double_col, integer_col, smallint_col, varchar2_col, nvarchar2_col, char_col, nchar_col, clob_col, blob_col, raw_col, date_col, timestamp_col, timestamp_tz_col, boolean_col, json_col)
VALUES (
    NULL, NULL, NULL, NULL, NULL,
    NULL, NULL, NULL, NULL,
    NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL
);

INSERT INTO all_datatypes (number_col, binary_float_col, binary_double_col, integer_col, smallint_col, varchar2_col, nvarchar2_col, char_col, nchar_col, clob_col, blob_col, raw_col, date_col, timestamp_col, timestamp_tz_col, boolean_col, json_col)
VALUES (
    0, 0, 0, 0, 0,
    '', N'', '          ', N'          ',
    '',
    NULL, NULL,
    TO_DATE('1970-01-01', 'YYYY-MM-DD'),
    TO_TIMESTAMP('1970-01-01 00:00:00.000', 'YYYY-MM-DD HH24:MI:SS.FF3'),
    TO_TIMESTAMP_TZ('1970-01-01 00:00:00.000 +00:00', 'YYYY-MM-DD HH24:MI:SS.FF3 TZH:TZM'),
    0,
    '{}'
);
