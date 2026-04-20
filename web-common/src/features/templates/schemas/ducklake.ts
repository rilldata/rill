import type { MultiStepFormSchema } from "./types";

export const ducklakeSchema: MultiStepFormSchema = {
  $schema: "http://json-schema.org/draft-07/schema#",
  type: "object",
  title: "DuckLake",
  "x-category": "olap",
  "x-driver": "duckdb",
  "x-form-width": "wide",
  properties: {
    connection_mode: {
      type: "string",
      enum: ["sql", "parameters"],
      default: "sql",
      "x-display": "tabs",
      "x-enum-labels": ["ATTACH SQL", "Parameters"],
      "x-ui-only": true,
      "x-tab-group": {
        sql: ["attach"],
        parameters: ["catalog", "alias", "data_path"],
      },
      "x-step": "connector",
    },
    attach: {
      type: "string",
      title: "Attach clause",
      description:
        "DuckDB `ATTACH` clause that points at your DuckLake catalog. Include the metadata backend and `DATA_PATH`.",
      "x-placeholder":
        "'ducklake:duckdb_database.ducklake' (DATA_PATH 'other_data_path/', OVERRIDE_DATA_PATH true)",
      "x-monospace": true,
      "x-hint":
        "Supported metadata backends: DuckDB file, SQLite, Postgres, MySQL. Data path can be local or object storage (s3://, gs://, azure://).",
      "x-docs-url":
        "https://ducklake.select/docs/stable/duckdb/usage/connecting",
      "x-visible-if": { connection_mode: "sql" },
      "x-step": "connector",
    },
    mode: {
      type: "boolean",
      title: "Enable write mode",
      description:
        "Read-write mode allows Rill to drop, create, and modify tables, not just query them",
      default: false,
      "x-display": "toggle",
      "x-yaml-value": "readwrite",
      "x-step": "connector",
      "x-advanced": true,
    },

    // ── Parameters tab ──────────────────────────────────────────────
    // These fields compose into the `attach` string at submission time
    // (see composeDuckLakeAttach). They are UI-only and never written
    // to the generated YAML as standalone keys.
    catalog: {
      type: "string",
      title: "Metadata catalog",
      description:
        "The metadata identifier that follows `ducklake:`. For a DuckDB file, this is the file path (e.g. `duckdb_database.ducklake`). For an external catalog, use a connection string (e.g. `postgres:dbname=... host=...`).",
      "x-placeholder": "duckdb_database.ducklake",
      "x-monospace": true,
      "x-visible-if": { connection_mode: "parameters" },
      "x-ui-only": true,
      "x-step": "connector",
    },
    alias: {
      type: "string",
      title: "Alias",
      description:
        "Optional database alias used in queries. Inserted as `AS <alias>` in the generated `ATTACH` clause.",
      "x-placeholder": "my_ducklake",
      "x-visible-if": { connection_mode: "parameters" },
      "x-ui-only": true,
      "x-step": "connector",
    },
    data_path: {
      type: "string",
      title: "DATA_PATH",
      description: "The storage location of the data files.",
      "x-placeholder": "other_data_path/",
      "x-visible-if": { connection_mode: "parameters" },
      "x-ui-only": true,
      "x-step": "connector",
    },
    override_data_path: {
      type: "boolean",
      title: "OVERRIDE_DATA_PATH",
      description:
        "If the path provided in `DATA_PATH` differs from the stored path and this option is set to true, the path is overridden.",
      default: true,
      "x-display": "toggle",
      "x-visible-if": { connection_mode: "parameters" },
      "x-ui-only": true,
      "x-step": "connector",
      "x-advanced": true,
    },
    create_if_not_exists: {
      type: "boolean",
      title: "CREATE_IF_NOT_EXISTS",
      description:
        "Creates a new DuckLake if the specified one does not already exist.",
      default: true,
      "x-display": "toggle",
      "x-visible-if": { connection_mode: "parameters" },
      "x-ui-only": true,
      "x-step": "connector",
      "x-advanced": true,
    },
    data_inlining_row_limit: {
      type: "number",
      title: "DATA_INLINING_ROW_LIMIT",
      description: "The number of rows for which data inlining is used.",
      default: 0,
      "x-visible-if": { connection_mode: "parameters" },
      "x-ui-only": true,
      "x-step": "connector",
      "x-advanced": true,
    },
    encrypted: {
      type: "boolean",
      title: "ENCRYPTED",
      description: "Whether or not data is stored encrypted.",
      default: false,
      "x-display": "toggle",
      "x-visible-if": { connection_mode: "parameters" },
      "x-ui-only": true,
      "x-step": "connector",
      "x-advanced": true,
    },
    meta_parameter_name: {
      type: "string",
      title: "META_PARAMETER_NAME",
      description: "Pass `PARAMETER_NAME` to the catalog server.",
      "x-visible-if": { connection_mode: "parameters" },
      "x-ui-only": true,
      "x-step": "connector",
      "x-advanced": true,
    },
    metadata_catalog: {
      type: "string",
      title: "METADATA_CATALOG",
      description: "The name of the attached catalog database.",
      "x-placeholder": "__ducklake_metadata_ducklake_name",
      "x-visible-if": { connection_mode: "parameters" },
      "x-ui-only": true,
      "x-step": "connector",
      "x-advanced": true,
    },
    metadata_parameters: {
      type: "string",
      title: "METADATA_PARAMETERS",
      description: "Map of parameters to pass to the catalog server.",
      "x-placeholder": "{}",
      "x-monospace": true,
      "x-visible-if": { connection_mode: "parameters" },
      "x-ui-only": true,
      "x-step": "connector",
      "x-advanced": true,
    },
    metadata_path: {
      type: "string",
      title: "METADATA_PATH",
      description:
        "The connection string for connecting to the metadata catalog.",
      "x-visible-if": { connection_mode: "parameters" },
      "x-ui-only": true,
      "x-step": "connector",
      "x-advanced": true,
    },
    metadata_schema: {
      type: "string",
      title: "METADATA_SCHEMA",
      description:
        "The schema in the catalog server in which to store the DuckLake tables.",
      default: "main",
      "x-visible-if": { connection_mode: "parameters" },
      "x-ui-only": true,
      "x-step": "connector",
      "x-advanced": true,
    },
    automatic_migration: {
      type: "boolean",
      title: "AUTOMATIC_MIGRATION",
      description:
        "Automatically migrates the DuckLake catalog schema if the version does not match.",
      default: false,
      "x-display": "toggle",
      "x-visible-if": { connection_mode: "parameters" },
      "x-ui-only": true,
      "x-step": "connector",
      "x-advanced": true,
    },
    snapshot_time: {
      type: "string",
      title: "SNAPSHOT_TIME",
      description:
        "If provided, connect to DuckLake at a snapshot at a specified point in time.",
      "x-visible-if": { connection_mode: "parameters" },
      "x-ui-only": true,
      "x-step": "connector",
      "x-advanced": true,
    },
    snapshot_version: {
      type: "string",
      title: "SNAPSHOT_VERSION",
      description:
        "If provided, connect to DuckLake at a specified snapshot id.",
      "x-visible-if": { connection_mode: "parameters" },
      "x-ui-only": true,
      "x-step": "connector",
      "x-advanced": true,
    },
  },
  allOf: [
    {
      if: { properties: { connection_mode: { const: "sql" } } },
      then: { required: ["attach"] },
    },
    {
      if: { properties: { connection_mode: { const: "parameters" } } },
      then: { required: ["catalog"] },
    },
  ],
};
