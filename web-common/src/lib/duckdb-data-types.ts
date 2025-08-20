/**
 * Provides mappings from duckdb's data types to conceptual types we use in the application:
 * CATEGORICALS, NUMERICS, and TIMESTAMPS.
 */

export const INTEGERS = new Set([
  // Rill runtime types
  "CODE_INT8",
  "CODE_INT16",
  "CODE_INT32",
  "CODE_INT64",
  "CODE_INT128",
  "CODE_UINT8",
  "CODE_UINT16",
  "CODE_UINT32",
  "CODE_UINT64",
  "CODE_UINT128",

  // DuckDB native types
  "BIGINT",
  "HUGEINT",
  "SMALLINT",
  "INTEGER",
  "TINYINT",
  "UBIGINT",
  "UINTEGER",
  "UTINYINT",
  "INT1",
  "INT4",
  "INT",
  "SIGNED",
  "SHORT",
]);

export const FLOATS = new Set([
  // Rill runtime types
  "CODE_FLOAT32",
  "CODE_FLOAT64",

  // DuckDB native types
  "DOUBLE",
  "DECIMAL",
  "FLOAT8",
  "NUMERIC",
  "FLOAT",
]);
export const DATES = new Set(["CODE_DATE", "DATE"]);
export const NUMERICS = new Set([...INTEGERS, ...FLOATS]);
export const BOOLEANS = new Set(["CODE_BOOL", "BOOLEAN", "BOOL", "LOGICAL"]);
export const TIMESTAMPS = new Set([
  // Rill runtime types
  "CODE_TIMESTAMP",
  "CODE_TIME",

  // DuckDB native types
  "TIMESTAMP",
  "TIME",
  "DATETIME",
  "TIMESTAMPTZ",
  "TIMESTAMP WITH TIME ZONE",
  "TIMESTAMP_S",
  "TIMESTAMP_MS",
  "TIMESTAMP_NS",

  ...DATES,
]);
export const INTERVALS = new Set(["INTERVAL"]);
export const NESTED = new Set(["STRUCT", "MAP", "LIST"]);

export function isList(type: string) {
  return type?.includes("[]");
}

// decimal values don't quite match a simple FLOATS.has(type) function,
// so we need this one.
export function isFloat(type: string) {
  return FLOATS.has(type) || type?.startsWith("DECIMAL");
}

export function isStruct(type: string) {
  return type?.startsWith("STRUCT");
}

export function isNested(type: string) {
  return (
    type === "JSON" ||
    isList(type) ||
    isStruct(type) ||
    [...NESTED].some((typeDef) => type?.startsWith(typeDef))
  );
}

export const STRING_LIKES = new Set([
  // Rill runtime types
  "CODE_STRING",
  "CODE_BYTES",
  "CODE_UUID",

  // DuckDB native types
  "UUID",
  "BYTE_ARRAY",
  "VARCHAR",
  "CHAR",
  "BPCHAR",
  "TEXT",
  "STRING",
]);

export const CATEGORICALS = new Set([...BOOLEANS, ...STRING_LIKES]);
export const ANY_TYPES = new Set([
  ...NUMERICS,
  ...BOOLEANS,
  ...TIMESTAMPS,
  ...INTERVALS,
  ...CATEGORICALS,
]);

export const TypesMap = new Map([
  [INTEGERS, "INTEGERS"],
  [FLOATS, "FLOATS"],
  [NUMERICS, "NUMERICS"],
  [BOOLEANS, "BOOLEANS"],
  [TIMESTAMPS, "TIMESTAMPS"],
  [INTERVALS, "INTERVALS"],
  [CATEGORICALS, "CATEGORICALS"],
  [ANY_TYPES, "ANY"],
]);

export interface Interval {
  months: number;
  days: number;
  micros: number;
}
interface ColorTokens {
  textClass: string;
  bgClass: string;
  vizFillClass: string;
  vizStrokeClass: string;
}

export const CATEGORICAL_TOKENS: ColorTokens = {
  textClass: "text-sky-800",
  bgClass: "bg-sky-200",
  vizFillClass: "fill-sky-800 ",
  vizStrokeClass: "stroke-sky-800",
};

export const NUMERIC_TOKENS: ColorTokens = {
  textClass: "text-red-800",
  bgClass: "bg-red-200",
  vizFillClass: "fill-red-300",
  vizStrokeClass: "stroke-red-300",
};

export const TIMESTAMP_TOKENS: ColorTokens = {
  textClass: "text-teal-800",
  bgClass: "bg-teal-200",
  vizFillClass: "fill-teal-500",
  vizStrokeClass: "stroke-teal-500",
};

export const NESTED_TOKENS: ColorTokens = {
  textClass: "text-gray-800",
  bgClass: "bg-gray-200",
  vizFillClass: "fill-gray-800",
  vizStrokeClass: "stroke-gray-800",
};

export const INTERVAL_TOKENS: ColorTokens = TIMESTAMP_TOKENS;

function setTypeTailwindStyles(
  list: string[],
  // a tailwind class, for now.
  colorTokens: ColorTokens,
) {
  return list.reduce((acc, v) => {
    acc[v] = { ...colorTokens };
    return acc;
  }, {});
}

export const DATA_TYPE_COLORS = {
  ...setTypeTailwindStyles(Array.from(CATEGORICALS), CATEGORICAL_TOKENS),
  ...setTypeTailwindStyles(Array.from(NUMERICS), NUMERIC_TOKENS),
  ...setTypeTailwindStyles(Array.from(TIMESTAMPS), TIMESTAMP_TOKENS),
  ...setTypeTailwindStyles(Array.from(INTERVALS), INTERVAL_TOKENS),
  ...setTypeTailwindStyles(Array.from(BOOLEANS), CATEGORICAL_TOKENS),
  ...setTypeTailwindStyles(Array.from(NESTED), NESTED_TOKENS),
};

/**
 * These are the intervals that are used in the rollup timegrain estimation.
 * These intervals get templated into the query as a duckdb INTERVAL (e.g. INTERVAL 1 hour).
 */
export enum PreviewRollupInterval {
  ms = "1 millisecond",
  second = "1 second",
  minute = "1 minute",
  hour = "1 hour",
  day = "1 day",
  month = "1 month",
  year = "1 year",
}
