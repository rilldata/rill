import type { MultiStepFormSchema } from "./types";

export const ducklakeSchema: MultiStepFormSchema = {
  $schema: "http://json-schema.org/draft-07/schema#",
  type: "object",
  title: "DuckLake",
  "x-category": "olap",
  "x-driver": "duckdb",
  "x-form-width": "wide",
  properties: {
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
    create_if_not_exists: {
      type: "boolean",
      title: "CREATE_IF_NOT_EXISTS",
      description:
        "Creates a new DuckLake if the specified one does not already exist.",
      default: true,
      "x-display": "toggle",
      "x-step": "connector",
      "x-advanced": true,
    },
    data_inlining_row_limit: {
      type: "integer",
      title: "DATA_INLINING_ROW_LIMIT",
      description: "The number of rows for which data inlining is used.",
      default: 0,
      "x-step": "connector",
      "x-advanced": true,
    },
    data_path: {
      type: "string",
      title: "DATA_PATH",
      description:
        "The storage location of the data files. Defaults to `metadata_file.files` for DuckDB files; required otherwise.",
      "x-step": "connector",
      "x-advanced": true,
    },
    encrypted: {
      type: "boolean",
      title: "ENCRYPTED",
      description: "Whether or not data is stored encrypted.",
      default: false,
      "x-display": "toggle",
      "x-step": "connector",
      "x-advanced": true,
    },
    meta_parameter_name: {
      type: "string",
      title: "META_PARAMETER_NAME",
      description: "Pass `PARAMETER_NAME` to the catalog server.",
      "x-step": "connector",
      "x-advanced": true,
    },
    metadata_catalog: {
      type: "string",
      title: "METADATA_CATALOG",
      description: "The name of the attached catalog database.",
      "x-placeholder": "__ducklake_metadata_ducklake_name",
      "x-step": "connector",
      "x-advanced": true,
    },
    metadata_parameters: {
      type: "string",
      title: "METADATA_PARAMETERS",
      description: "Map of parameters to pass to the catalog server.",
      "x-placeholder": "{}",
      "x-monospace": true,
      "x-step": "connector",
      "x-advanced": true,
    },
    metadata_path: {
      type: "string",
      title: "METADATA_PATH",
      description:
        "The connection string for connecting to the metadata catalog.",
      "x-step": "connector",
      "x-advanced": true,
    },
    metadata_schema: {
      type: "string",
      title: "METADATA_SCHEMA",
      description:
        "The schema in the catalog server in which to store the DuckLake tables.",
      default: "main",
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
      "x-step": "connector",
      "x-advanced": true,
    },
    override_data_path: {
      type: "boolean",
      title: "OVERRIDE_DATA_PATH",
      description:
        "If the path provided in `data_path` differs from the stored path and this option is set to true, the path is overridden.",
      default: true,
      "x-display": "toggle",
      "x-step": "connector",
      "x-advanced": true,
    },
    snapshot_time: {
      type: "string",
      title: "SNAPSHOT_TIME",
      description:
        "If provided, connect to DuckLake at a snapshot at a specified point in time.",
      "x-step": "connector",
      "x-advanced": true,
    },
    snapshot_version: {
      type: "string",
      title: "SNAPSHOT_VERSION",
      description:
        "If provided, connect to DuckLake at a specified snapshot id.",
      "x-step": "connector",
      "x-advanced": true,
    },
  },
  required: ["attach"],
};
