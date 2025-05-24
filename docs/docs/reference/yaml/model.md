---
note: GENERATED. DO NOT EDIT.
title: Model YAML
sidebar_position: 38
---

## Properties

### `type`

_[string]_ - Refers to the resource type and must be `model` _(required)_

### `refresh`

_[object]_ - Specifies the refresh schedule that Rill should follow to re-ingest and update the underlying model data 

  - **`cron`** - _[string]_ - A cron expression that defines the execution schedule 

  - **`time_zone`** - _[string]_ - Time zone to interpret the schedule in (e.g., 'UTC', 'America/Los_Angeles'). 

  - **`disable`** - _[boolean]_ - If true, disables the resource without deleting it. 

  - **`ref_update`** - _[boolean]_ - If true, allows the resource to run when a dependency updates. 

  - **`run_in_dev`** - _[boolean]_ - If true, allows the schedule to run in development mode. 

### `timeout`

_[string]_ - The maximum time to wait for model ingestion 

### `incremental`

_[boolean]_ - whether incremental modeling is required (optional) 

### `change_mode`

_[string]_ - Configure how changes to the model specifications are applied (optional). 'reset' will drop and recreate the model automatically, 'manual' will require a manual full or incremental refresh to apply changes, and 'patch' will switch to the new logic without re-processing historical data (only applies for incremental models). 

### `state`

_[oneOf]_ - Refers to the explicitly defined state of your model, cannot be used with partitions (optional) 

  - **option 1** - _[object]_ - Executes a raw SQL query against the project's data models.

    - **`sql`** - _[string]_ - Raw SQL query to run against existing models in the project. _(required)_

    - **`connector`** - _[string]_ - specifies the connector to use when running SQL or glob queries. 

  - **option 2** - _[object]_ - Executes a SQL query that targets a defined metrics view.

    - **`metrics_sql`** - _[string]_ - SQL query that targets a metrics view in the project _(required)_

  - **option 3** - _[object]_ - Calls a custom API defined in the project to compute data.

    - **`api`** - _[string]_ - Name of a custom API defined in the project. _(required)_

    - **`args`** - _[object]_ - Arguments to pass to the custom API. 

  - **option 4** - _[object]_ - Uses a file-matching pattern (glob) to query data from a connector.

    - **`glob`** - _[anyOf]_ - Defines the file path or pattern to query from the specified connector. _(required)_

      - **option 1** - _[string]_ - A simple file path/glob pattern as a string.

      - **option 2** - _[object]_ - An object-based configuration for specifying a file path/glob pattern with advanced options.

    - **`connector`** - _[string]_ - Specifies the connector to use with the glob input. 

  - **option 5** - _[object]_ - Uses the status of a resource as data.

    - **`resource_status`** - _[object]_ - Based on resource status _(required)_

      - **`where_error`** - _[boolean]_ - Indicates whether the condition should trigger when the resource is in an error state. 

### `partitions`

_[oneOf]_ - Refers to the how your data is partitioned, cannot be used with state. (optional) 

  - **option 1** - _[object]_ - Executes a raw SQL query against the project's data models.

    - **`sql`** - _[string]_ - Raw SQL query to run against existing models in the project. _(required)_

    - **`connector`** - _[string]_ - specifies the connector to use when running SQL or glob queries. 

  - **option 2** - _[object]_ - Executes a SQL query that targets a defined metrics view.

    - **`metrics_sql`** - _[string]_ - SQL query that targets a metrics view in the project _(required)_

  - **option 3** - _[object]_ - Calls a custom API defined in the project to compute data.

    - **`api`** - _[string]_ - Name of a custom API defined in the project. _(required)_

    - **`args`** - _[object]_ - Arguments to pass to the custom API. 

  - **option 4** - _[object]_ - Uses a file-matching pattern (glob) to query data from a connector.

    - **`glob`** - _[anyOf]_ - Defines the file path or pattern to query from the specified connector. _(required)_

      - **option 1** - _[string]_ - A simple file path/glob pattern as a string.

      - **option 2** - _[object]_ - An object-based configuration for specifying a file path/glob pattern with advanced options.

    - **`connector`** - _[string]_ - Specifies the connector to use with the glob input. 

  - **option 5** - _[object]_ - Uses the status of a resource as data.

    - **`resource_status`** - _[object]_ - Based on resource status _(required)_

      - **`where_error`** - _[boolean]_ - Indicates whether the condition should trigger when the resource is in an error state. 

### `sql`

_[string]_ - Raw SQL query to run against source _(required)_

### `materialize`

_[boolean]_ - models will be materialized in olap 

### `partitions_watermark`

_[string]_ - Refers to a customizable timestamp that can be set to check if an object has been updated (optional). 

### `partitions_concurrency`

_[integer]_ - Refers to the number of concurrent partitions that can be read at the same time (optional). 

### `stage`

_[object]_ - in the case of staging models, where an input source does not support direct write to the output and a staging table is required 

  - **`connector`** - _[string]_ - Refers to the connector type for the staging table _(required)_

### `output`

_[object]_ - to define the properties of output 

  - **`connector`** - _[string]_ - Refers to the connector type for the output table 

## Common Properties

### `name`

_[string]_ - Name is usually inferred from the filename, but can be specified manually. 

### `refs`

_[array of string]_ - List of resource references 

### `dev`

_[object]_ - Overrides any properties in development environment. 

### `prod`

_[object]_ - Overrides any properties in production environment. 

## One of Properties Options
- [athena](#athena)
- [azure](#azure)
- [bigquery](#bigquery)
- [clickhouse](#clickhouse)
- [duckdb](#duckdb)
- [gcs](#gcs)
- [https](#https)
- [local_file](#local_file)
- [motherduck](#motherduck)
- [mysql](#mysql)
- [pinot](#pinot)
- [postgres](#postgres)
- [redshift](#redshift)
- [s3](#s3)
- [salesforce](#salesforce)
- [snowflake](#snowflake)
- [sqlite](#sqlite)

## athena

### `connector`

_[string]_ - Must be either `athena` or a named Athena connector. _(required)_

### `output_location`

_[string]_ - Output location for query results in S3. 

### `workgroup`

_[string]_ - AWS Athena workgroup to use for queries. 

### `region`

_[string]_ - AWS region to connect to Athena and the output location. 

## azure

### `connector`

_[string]_ - Must be either `azure` or a named Azure connector. _(required)_

### `path`

_[string]_ - Path to the source 

### `account`

_[string]_ - Account identifier 

### `uri`

_[string]_ - Source URI 

### `extract`

_[object]_ - Arbitrary key-value pairs for extraction settings 

### `glob`

_[object]_ - Settings related to glob file matching. 

  - **`max_total_size`** - _[integer]_ - Maximum total size (in bytes) matched by glob 

  - **`max_objects_matched`** - _[integer]_ - Maximum number of objects matched by glob 

  - **`max_objects_listed`** - _[integer]_ - Maximum number of objects listed in glob 

  - **`page_size`** - _[integer]_ - Page size for glob listing 

### `batch_size`

_[string]_ - Size of a batch (e.g., '100MB') 

## bigquery

### `connector`

_[string]_ - Must be either `bigquery` or a named BigQuery connector. _(required)_

### `project_id`

_[string]_ - ID of the BigQuery project. 

## clickhouse

### `connector`

_[string]_ - Must be either `clickhouse` or a named ClickHouse connector. _(required)_

### `path`

_[string]_ - Path to the data source. 

### `format`

_[string]_ - Format of the data source (e.g., csv, json, parquet). 

## duckdb

### `connector`

_[string]_ - Must be either `duckdb` or a named DuckDB connector. _(required)_

### `path`

_[string]_ - Path to the data source. 

### `format`

_[string]_ - Format of the data source (e.g., csv, json, parquet). 

### `pre_exec`

_[string]_ - refers to a SQL queries to run before the main query, available for DuckDB based models 

### `post_exec`

_[string]_ - refers to a SQL query that is run after the main query, available for DuckDB based models 

## gcs

### `connector`

_[string]_ - Must be either `gcs` or a named GCS connector. _(required)_

### `path`

_[string]_ - Path to the source 

### `uri`

_[string]_ - Source URI 

### `extract`

_[object]_ - key-value pairs for extraction settings 

### `glob`

_[object]_ - Settings related to glob file matching. 

  - **`max_total_size`** - _[integer]_ - Maximum total size (in bytes) matched by glob 

  - **`max_objects_matched`** - _[integer]_ - Maximum number of objects matched by glob 

  - **`max_objects_listed`** - _[integer]_ - Maximum number of objects listed in glob 

  - **`page_size`** - _[integer]_ - Page size for glob listing 

### `batch_size`

_[string]_ - Size of a batch (e.g., '100MB') 

## https

### `connector`

_[string]_ - Must be either `https` or a named HTTPS connector. _(required)_

## local_file

### `connector`

_[string]_ - Must be either `local_file` or a named Local File connector. _(required)_

### `path`

_[string]_ - Path to the data source. 

### `format`

_[string]_ - Format of the data source (e.g., csv, json, parquet). 

## motherduck

### `connector`

_[string]_ - Must be either `motherduck` or a named MotherDuck connector. _(required)_

## mysql

### `connector`

_[string]_ - Must be either `mysql` or a named MySQL connector. _(required)_

## pinot

### `connector`

_[string]_ - Must be either `pinot` or a named Pinot connector. _(required)_

## postgres

### `connector`

_[string]_ - Must be either `postgres` or a named PostgreSQL connector. _(required)_

## redshift

### `connector`

_[string]_ - Must be either `redshift` or a named Redshift connector. _(required)_

### `output_location`

_[string]_ - S3 location where query results are stored. 

### `workgroup`

_[string]_ - Redshift Serverless workgroup to use. 

### `database`

_[string]_ - Name of the Redshift database. 

### `cluster_identifier`

_[string]_ - Identifier of the Redshift cluster. 

### `role_arn`

_[string]_ - ARN of the IAM role to assume for Redshift access. 

### `region`

_[string]_ - AWS region of the Redshift deployment. 

## s3

### `connector`

_[string]_ - Must be either `s3` or a named S3 connector. _(required)_

### `region`

_[string]_ - AWS region 

### `endpoint`

_[string]_ - AWS Endpoint 

### `path`

_[string]_ - Path to the source 

### `uri`

_[string]_ - Source URI 

### `extract`

_[object]_ - key-value pairs for extraction settings 

### `glob`

_[object]_ - Settings related to glob file matching. 

  - **`max_total_size`** - _[integer]_ - Maximum total size (in bytes) matched by glob 

  - **`max_objects_matched`** - _[integer]_ - Maximum number of objects matched by glob 

  - **`max_objects_listed`** - _[integer]_ - Maximum number of objects listed in glob 

  - **`page_size`** - _[integer]_ - Page size for glob listing 

### `batch_size`

_[string]_ - Size of a batch (e.g., '100MB') 

## salesforce

### `connector`

_[string]_ - Must be either `salesforce` or a named Salesforce connector. _(required)_

### `soql`

_[string]_ - SOQL query to execute against the Salesforce instance. 

### `sobject`

_[string]_ - Salesforce object (e.g., Account, Contact) targeted by the query. 

### `queryAll`

_[boolean]_ - Whether to include deleted and archived records in the query (uses queryAll API). 

## snowflake

### `connector`

_[string]_ - Must be either `snowflake` or a named Snowflake connector. _(required)_

## sqlite

### `connector`

_[string]_ - Must be either `sqlite` or a named SQLite connector. _(required)_