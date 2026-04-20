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
        "DuckDB `ATTACH` clause that points at your DuckLake catalog. Include the metadata backend, alias, and `DATA_PATH`.",
      "x-placeholder":
        "'ducklake:metadata.ducklake' AS my_ducklake (DATA_PATH 'datafiles')",
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
  },
  required: ["attach"],
};
