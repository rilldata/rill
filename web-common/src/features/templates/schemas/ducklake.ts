import type { MultiStepFormSchema } from "./types";

export const ducklakeSchema: MultiStepFormSchema = {
  $schema: "http://json-schema.org/draft-07/schema#",
  type: "object",
  title: "DuckLake",
  "x-category": "olap",
  "x-driver": "duckdb",
  "x-form-width": "wide",
  properties: {
    // The Parameters tab UI is hidden for phase 1; we only expose the raw
    // ATTACH SQL path. The enum, tab-group, and parameter composer code are
    // preserved so the tab can be re-enabled later by flipping `x-hidden`.
    connection_mode: {
      type: "string",
      enum: ["sql", "parameters"],
      default: "sql",
      "x-display": "tabs",
      "x-enum-labels": ["ATTACH SQL", "Parameters"],
      "x-ui-only": true,
      "x-hidden": true,
      "x-tab-group": {
        sql: ["attach"],
        parameters: ["catalog_type", "alias", "data_path_type"],
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
      "x-custom-validator": "ducklake-attach",
      "x-step": "connector",
    },
    mode: {
      type: "boolean",
      title: "Read only",
      description:
        "Restrict Rill to read-only queries. Disable to let Rill drop, create, and modify tables.",
      default: true,
      "x-display": "toggle",
      "x-yaml-value": { true: "readonly", false: "readwrite" },
      "x-visible-if": { connection_mode: "parameters" },
      "x-step": "connector",
      "x-advanced": true,
    },

    // ── Parameters tab: catalog ──────────────────────────────
    // `catalog_type` picks the metadata backend; the fields under each
    // option compose into the `ducklake:<prefix>` portion of `attach`
    // (see composeDuckLakeAttach). All are x-ui-only and never emitted
    // as standalone YAML keys.
    catalog_type: {
      type: "string",
      title: "Metadata catalog",
      enum: ["duckdb", "sqlite", "postgres", "mysql"],
      default: "duckdb",
      "x-display": "select",
      "x-select-style": "rich",
      "x-visible-if": { connection_mode: "parameters" },
      "x-enum-labels": ["DuckDB file", "SQLite", "PostgreSQL", "MySQL"],
      "x-enum-descriptions": [
        "Store metadata in a local DuckDB file",
        "Store metadata in a SQLite file",
        "Store metadata in a PostgreSQL database",
        "Store metadata in a MySQL database",
      ],
      "x-ui-only": true,
      "x-grouped-fields": {
        duckdb: ["catalog_duckdb_path"],
        sqlite: ["catalog_sqlite_path"],
        postgres: [
          "catalog_postgres_dbname",
          "catalog_postgres_host",
          "catalog_postgres_port",
          "catalog_postgres_user",
          "catalog_postgres_password",
        ],
        mysql: [
          "catalog_mysql_database",
          "catalog_mysql_host",
          "catalog_mysql_port",
          "catalog_mysql_user",
          "catalog_mysql_password",
        ],
      },
      "x-step": "connector",
    },
    catalog_duckdb_path: {
      type: "string",
      title: "DuckDB file path",
      description:
        "Path to a DuckDB file that stores the DuckLake metadata. Created if it does not exist.",
      "x-placeholder": "catalog.ducklake",
      "x-monospace": true,
      "x-ui-only": true,
      "x-step": "connector",
    },
    catalog_sqlite_path: {
      type: "string",
      title: "SQLite file path",
      description:
        "Path to a SQLite file that stores the DuckLake metadata. Created if it does not exist.",
      "x-placeholder": "catalog.sqlite",
      "x-monospace": true,
      "x-ui-only": true,
      "x-step": "connector",
    },
    catalog_postgres_dbname: {
      type: "string",
      title: "Database",
      description: "Postgres database containing the DuckLake metadata tables.",
      "x-placeholder": "mydb",
      "x-ui-only": true,
      "x-step": "connector",
    },
    catalog_postgres_host: {
      type: "string",
      title: "Host",
      "x-placeholder": "localhost",
      "x-ui-only": true,
      "x-step": "connector",
    },
    catalog_postgres_port: {
      type: "string",
      title: "Port",
      pattern: "^\\d+$",
      errorMessage: { pattern: "Port must be a number" },
      "x-placeholder": "5432",
      "x-ui-only": true,
      "x-step": "connector",
    },
    catalog_postgres_user: {
      type: "string",
      title: "User",
      "x-placeholder": "postgres",
      "x-ui-only": true,
      "x-step": "connector",
    },
    catalog_postgres_password: {
      type: "string",
      title: "Password",
      description: "PostgreSQL password. Stored as a secret in `.env`.",
      "x-placeholder": "your_password",
      "x-secret": true,
      "x-env-var-name": "DUCKLAKE_CATALOG_POSTGRES_PASSWORD",
      "x-ui-only": true,
      "x-step": "connector",
    },
    catalog_mysql_database: {
      type: "string",
      title: "Database",
      description: "MySQL database containing the DuckLake metadata tables.",
      "x-placeholder": "mydb",
      "x-ui-only": true,
      "x-step": "connector",
    },
    catalog_mysql_host: {
      type: "string",
      title: "Host",
      "x-placeholder": "localhost",
      "x-ui-only": true,
      "x-step": "connector",
    },
    catalog_mysql_port: {
      type: "string",
      title: "Port",
      pattern: "^\\d+$",
      errorMessage: { pattern: "Port must be a number" },
      "x-placeholder": "3306",
      "x-ui-only": true,
      "x-step": "connector",
    },
    catalog_mysql_user: {
      type: "string",
      title: "User",
      "x-placeholder": "root",
      "x-ui-only": true,
      "x-step": "connector",
    },
    catalog_mysql_password: {
      type: "string",
      title: "Password",
      description: "MySQL password. Stored as a secret in `.env`.",
      "x-placeholder": "your_password",
      "x-secret": true,
      "x-env-var-name": "DUCKLAKE_CATALOG_MYSQL_PASSWORD",
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

    // ── Parameters tab: data path ────────────────────────────
    // `data_path_type` picks the storage backend. For object storage
    // (s3/gcs/azure) we require the matching Rill connector so that
    // DuckDB secret bridging (generateSecretSQL) has credentials to use.
    data_path_type: {
      type: "string",
      title: "Data path",
      enum: ["local", "s3", "gcs", "azure"],
      default: "local",
      "x-display": "select",
      "x-select-style": "rich",
      "x-visible-if": { connection_mode: "parameters" },
      "x-enum-labels": [
        "Local filesystem",
        "Amazon S3",
        "Google Cloud Storage",
        "Azure Blob Storage",
      ],
      "x-enum-descriptions": [
        "Store data files on the local filesystem",
        "Store data files in an S3 bucket",
        "Store data files in a GCS bucket",
        "Store data files in Azure Blob Storage",
      ],
      "x-ui-only": true,
      "x-required-driver": {
        s3: "s3",
        gcs: "gcs",
        azure: "azure",
      },
      "x-grouped-fields": {
        local: ["data_path_local"],
        s3: ["data_path_s3_info", "data_path_s3"],
        gcs: ["data_path_gcs_info", "data_path_gcs"],
        azure: ["data_path_azure_info", "data_path_azure"],
      },
      "x-step": "connector",
    },
    data_path_local: {
      type: "string",
      title: "Data path",
      description: "Local filesystem directory for DuckLake data files.",
      "x-placeholder": "data/",
      "x-monospace": true,
      "x-ui-only": true,
      "x-step": "connector",
    },
    data_path_s3_info: {
      type: "boolean",
      title: "S3 Connector Required",
      default: true,
      "x-informational": true,
      "x-ui-only": true,
      "x-step": "connector",
    },
    data_path_s3: {
      type: "string",
      title: "S3 URI",
      description: "S3 URI for the DuckLake data directory.",
      pattern: "^s3://[^/]+(/.*)?$",
      errorMessage: {
        pattern: "Must be an S3 URI (e.g. s3://bucket/path/)",
      },
      "x-placeholder": "s3://bucket/path/",
      "x-monospace": true,
      "x-ui-only": true,
      "x-step": "connector",
    },
    data_path_gcs_info: {
      type: "boolean",
      title: "GCS Connector Required",
      default: true,
      "x-informational": true,
      "x-ui-only": true,
      "x-step": "connector",
    },
    data_path_gcs: {
      type: "string",
      title: "GCS URI",
      description: "GCS URI for the DuckLake data directory.",
      pattern: "^gs://[^/]+(/.*)?$",
      errorMessage: {
        pattern: "Must be a GCS URI (e.g. gs://bucket/path/)",
      },
      "x-placeholder": "gs://bucket/path/",
      "x-monospace": true,
      "x-ui-only": true,
      "x-step": "connector",
    },
    data_path_azure_info: {
      type: "boolean",
      title: "Azure Connector Required",
      default: true,
      "x-informational": true,
      "x-ui-only": true,
      "x-step": "connector",
    },
    data_path_azure: {
      type: "string",
      title: "Azure URI",
      description: "Azure URI for the DuckLake data directory.",
      pattern: "^azure://.+",
      errorMessage: {
        pattern: "Must be an Azure URI (e.g. azure://container/path/)",
      },
      "x-placeholder": "azure://container/path/",
      "x-monospace": true,
      "x-ui-only": true,
      "x-step": "connector",
    },

    // ── Advanced parameters ──────────────────────────────────
    // These remain top-level (not in the tab group) so they live under
    // the collapsible "Advanced options" toggle. They use x-visible-if
    // to hide on the SQL tab; because they are x-ui-only, they survive
    // tab switches (the clear-hidden renderer skips x-ui-only fields).
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
      "x-placeholder": "0",
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
      "x-placeholder": "parameter_value",
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
    metadata_schema: {
      type: "string",
      title: "METADATA_SCHEMA",
      description:
        "The schema in the catalog server in which to store the DuckLake tables.",
      default: "main",
      "x-placeholder": "main",
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
      "x-placeholder": "2024-01-01 00:00:00",
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
      "x-placeholder": "1",
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
      if: {
        properties: {
          connection_mode: { const: "parameters" },
          catalog_type: { const: "duckdb" },
        },
      },
      then: { required: ["catalog_duckdb_path"] },
    },
    {
      if: {
        properties: {
          connection_mode: { const: "parameters" },
          catalog_type: { const: "sqlite" },
        },
      },
      then: { required: ["catalog_sqlite_path"] },
    },
    {
      if: {
        properties: {
          connection_mode: { const: "parameters" },
          catalog_type: { const: "postgres" },
        },
      },
      then: {
        required: ["catalog_postgres_dbname", "catalog_postgres_host"],
      },
    },
    {
      if: {
        properties: {
          connection_mode: { const: "parameters" },
          catalog_type: { const: "mysql" },
        },
      },
      then: {
        required: ["catalog_mysql_database", "catalog_mysql_host"],
      },
    },
  ],
};
