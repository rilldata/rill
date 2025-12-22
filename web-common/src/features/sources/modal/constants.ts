export const CONNECTION_TAB_OPTIONS: { value: string; label: string }[] = [
  { value: "parameters", label: "Enter parameters" },
  { value: "dsn", label: "Enter connection string" },
];

// pre-defined order for sources
export const SOURCES = [
  "athena",
  "azure",
  "bigquery",
  "gcs",
  "mysql",
  "postgres",
  "redshift",
  "s3",
  "salesforce",
  "snowflake",
  "sqlite",
  "https",
  "local_file",
];

export const OLAP_ENGINES = [
  "clickhousecloud",
  "clickhouse",
  "motherduck",
  "duckdb",
  "druid",
  "pinot",
];

export const ALL_CONNECTORS = [...SOURCES, ...OLAP_ENGINES];

// Connectors that support multi-step forms (connector -> source)
export const MULTI_STEP_CONNECTORS = [
  "gcs",
  "s3",
  "azure",
  "https",
  "postgres",
  "mysql",
  "snowflake",
  "bigquery",
  "redshift",
  "athena",
  "clickhousecloud",
  "clickhouse",
  "duckdb",
  "motherduck",
  "druid",
  "pinot",
];

export const FORM_HEIGHT_TALL = "max-h-[55.5rem] min-h-[38.5rem]";
export const FORM_HEIGHT_MEDIUM= "max-h-[47.5rem] min-h-[34.5rem]";
export const FORM_HEIGHT_DEFAULT = "max-h-[34.5rem] min-h-[34.5rem]";
export const MEDIUM_FORM_CONNECTORS = new Set([
  "clickhousecloud",
  "salesforce",
  "postgres",
  "s3",
  "mysql",
  "pinot",
]);
export const TALL_FORM_CONNECTORS = new Set([
  "clickhouse",
  "snowflake",
]);
