---
note: GENERATED. DO NOT EDIT.
title: Source YAML
sidebar_position: 32
---

:::warning Deprecated Feature
**Sources have been deprecated** and are now considered "source models." While sources remain backward compatible, we recommend migrating to the new source model format for access to the latest features and improvements.

**Next steps:**
- Continue using sources if needed (backward compatible)
- Migrate to source models via the `type:model` parameter for existing projects
- See our [model YAML reference](advanced-models) for current documentation and best practices
:::

Source YAML files define data sources that can be imported into Rill projects.


## Properties

### `type`

_[string]_ - Refers to the resource type and must be `model` _(required)_

### `connector`

_[string]_ - Refers to the connector type for the source, see [connectors](/build/connect/) for more information _(required)_

### `uri`

_[string]_ - Refers to the URI of the remote connector you are using for the source. Rill also supports glob patterns as part of the URI for S3 and GCS (required for type: http, s3, gcs).

- `s3://your-org/bucket/file.parquet` — the s3 URI of your file
- `gs://your-org/bucket/file.parquet` — the gsutil URI of your file
- `https://data.example.org/path/to/file.parquet` — the web address of your file
 

### `path`

_[string]_ - Refers to the local path of the connector you are using for the source 

### `sql`

_[string]_ - Sets the SQL query to extract data from a SQL source 

### `region`

_[string]_ - Sets the cloud region of the S3 bucket or Athena 

### `endpoint`

_[string]_ - Overrides the S3 endpoint to connect to 

### `output_location`

_[string]_ - Sets the query output location and result files in Athena 

### `workgroup`

_[string]_ - Sets a workgroup for Athena connector 

### `project_id`

_[string]_ - Sets a project id to be used to run BigQuery jobs 

### `timeout`

_[string]_ - The maximum time to wait for source ingestion 

### `refresh`

_[object]_ - Specifies the refresh schedule that Rill should follow to re-ingest and update the underlying source data (optional).
- **cron** - a cron schedule expression, which should be encapsulated in single quotes, e.g. `* * * * *` (optional)
- **every** - a Go duration string, such as `24h` (optional)
 

  - **`cron`** - _[string]_ - A cron schedule expression, which should be encapsulated in single quotes, e.g. `* * * * *` 

  - **`every`** - _[string]_ - A Go duration string, such as `24h` 

### `db`

_[string]_ - Sets the database for motherduck connections and/or the path to the DuckDB/SQLite db file 

### `database_url`

_[string]_ - Postgres connection string that should be used 

### `duckdb`

_[object]_ - Specifies the raw parameters to inject into the DuckDB read_csv, read_json or read_parquet statement 

### `dsn`

_[string]_ - Used to set the Snowflake connection string 