CREATE DATABASE testDB;
GO

USE testDB;
GO

CREATE TABLE all_datatypes (
    -- Integer types
    bit_col BIT,
    tinyint_col TINYINT,
    smallint_col SMALLINT,
    int_col INT,
    bigint_col BIGINT,

    -- Floating point types
    real_col REAL,
    float_col FLOAT,

    -- Decimal types
    decimal_col DECIMAL(10,2),
    numeric_col NUMERIC(10,2),
    money_col MONEY,
    smallmoney_col SMALLMONEY,

    -- String types
    char_col CHAR(10),
    varchar_col VARCHAR(255),
    text_col TEXT,
    nchar_col NCHAR(10),
    nvarchar_col NVARCHAR(255),
    ntext_col NTEXT,

    -- Binary types
    binary_col BINARY(10),
    varbinary_col VARBINARY(255),

    -- Date and Time types
    date_col DATE,
    time_col TIME,
    datetime_col DATETIME,
    datetime2_col DATETIME2,
    smalldatetime_col SMALLDATETIME,
    datetimeoffset_col DATETIMEOFFSET,

    -- Other types
    uniqueidentifier_col UNIQUEIDENTIFIER,
    xml_col XML
);
GO

INSERT INTO all_datatypes (
    bit_col, tinyint_col, smallint_col, int_col, bigint_col,
    real_col, float_col,
    decimal_col, numeric_col, money_col, smallmoney_col,
    char_col, varchar_col, text_col, nchar_col, nvarchar_col, ntext_col,
    binary_col, varbinary_col,
    date_col, time_col, datetime_col, datetime2_col, smalldatetime_col, datetimeoffset_col,
    uniqueidentifier_col, xml_col
) VALUES
(
    -- Row 1: All NULL values
    NULL, NULL, NULL, NULL, NULL,
    NULL, NULL,
    NULL, NULL, NULL, NULL,
    NULL, NULL, NULL, NULL, NULL, NULL,
    NULL, NULL,
    NULL, NULL, NULL, NULL, NULL, NULL,
    NULL, NULL
),
(
    -- Row 2: Non-NULL values
    1, 127, 32767, 2147483647, 999999999999999999,
    1.5, 2.5,
    3.30, 4.40, 5.5000, 6.6000,
    'C', 'VarChar', 'Text content', N'NChar', N'NVarChar', N'NText content',
    0x42696E617279000000, 0x566172, -- 'Binary\0\0\0' and 'Var'
    '2024-02-14', '12:34:56', '2025-02-14 12:34:56', '2025-02-14 12:34:56.1234567', '2025-02-14 12:35:00', '2025-02-14 12:34:56.1234567 +05:30',
    '6F9619FF-8B86-D011-B42D-00CF4FC964FF', '<root><element>value</element></root>'
),
(
    -- Row 3: Zero/empty values
    0, 0, 0, 0, 0,
    0.0, 0.0,
    0.00, 0.00, 0.0000, 0.0000,
    '', '', '', N'', N'', N'',
    0x00000000000000000000, 0x00,
    '1970-01-01', '00:00:00', '1970-01-01 00:00:00', '1970-01-01 00:00:00', '1970-01-02 00:00:00', '1970-01-01 00:00:00 +00:00',
    '00000000-0000-0000-0000-000000000000', '<empty/>'
);
GO
