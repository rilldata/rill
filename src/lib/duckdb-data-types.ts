/**
 * Provides mappings from duckdb's data types to conceptual types we use in the application:
 * CATEGORICALS, NUMERICS, and TIMESTAMPS.
 */
export const CATEGORICALS = new Set(['BYTE_ARRAY', 'VARCHAR', "CHAR", "BPCHAR", "TEXT", "STRING"]);

export const NUMERICS = new Set([
    'DOUBLE', 'DECIMAL', 'BIGINT', 'HUGEINT', 'SMALLINT', 'INTEGER', 'TINYINT', 'UBIGINT', 'UINTEGER', 'UTINYINT', 'INT1', 'FLOAT8', 'NUMERIC',
    'INT4', 'INT', 'SIGNED', 'SHORT', 'FLOAT']);

export const TIMESTAMPS = new Set(['TIMESTAMP', 'TIME', 'DATETIME', 'DATE']);

function setTypeTailwindStyles(list:string[], textClass: string, bgClass: string) {
    return list.reduce((acc, v) => {
        acc[v] = { textClass, bgClass };
        return acc;
    }, {});
}


export const DATA_TYPE_ICON_STYLES = {
    ...setTypeTailwindStyles(Array.from(CATEGORICALS), 'bg-sky-800', 'bg-sky-200'),
    ...setTypeTailwindStyles(Array.from(NUMERICS), 'bg-red-800', 'bg-red-200'),
    ...setTypeTailwindStyles(Array.from(TIMESTAMPS), 'bg-teal-800', 'bg-teal-200'),
}