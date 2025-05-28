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

### `connector`

_[string]_ - Refers to the connector type  or [named connector](./connector.md#name) for the source.
 

### `sql`

_[string]_ - Raw SQL query to run against source _(required)_

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

  - **`table`** - _[string]_ - Name of the output table. If not specified, the model name is used. 

  - **`materialize`** - _[boolean]_ - Whether to materialize the model as a table or view 

  - **`connector`** - _[string]_ - Refers to the connector type for the output table. Can be `clickhouse` or `duckdb` and their named connector 

  - **`incremental_strategy`** - _[string]_ - Strategy to use for incremental updates. Can be 'append', 'merge' or 'partition_overwrite' 

  - **`unique_key`** - _[array of string]_ - List of columns that uniquely identify a row for merge strategy 

  - **`partition_by`** - _[string]_ - Column or expression to partition the table by 

  **Additional properties for `output` when `connector` is `clickhouse`**

  - **`type`** - _[string]_ - Type to materialize the model into. Can be 'TABLE', 'VIEW' or 'DICTIONARY' 

  - **`columns`** - _[string]_ - Column names and types. Can also include indexes. If unspecified, detected from the query. 

  - **`engine_full`** - _[string]_ - Full engine definition in SQL format. Can include partition keys, order, TTL, etc. 

  - **`engine`** - _[string]_ - Table engine to use. Default is MergeTree 

  - **`order_by`** - _[string]_ - ORDER BY clause. 

  - **`partition_by`** - _[string]_ - Partition BY clause. 

  - **`primary_key`** - _[string]_ - PRIMARY KEY clause. 

  - **`sample_by`** - _[string]_ - SAMPLE BY clause. 

  - **`ttl`** - _[string]_ - TTL settings for the table or columns. 

  - **`table_settings`** - _[string]_ - Table-specific settings. 

  - **`query_settings`** - _[string]_ - Settings used in insert/create table as select queries. 

  - **`distributed_settings`** - _[string]_ - Settings for distributed table. 

  - **`distributed_sharding_key`** - _[string]_ - Sharding key for distributed table. 

  - **`dictionary_source_user`** - _[string]_ - User for accessing the source dictionary table (used if type is DICTIONARY). 

  - **`dictionary_source_password`** - _[string]_ - Password for the dictionary source user. 

## Common Properties

### `name`

_[string]_ - Name is usually inferred from the filename, but can be specified manually. 

### `refs`

_[array of string]_ - List of resource references 

### `dev`

_[object]_ - Overrides any properties in development environment. 

### `prod`

_[object]_ - Overrides any properties in production environment. 

## Additional properties when `connector` is `athena` or [named connector](./connector.md#name) for athena

### `output_location`

_[string]_ - Output location for query results in S3. 

### `workgroup`

_[string]_ - AWS Athena workgroup to use for queries. 

### `region`

_[string]_ - AWS region to connect to Athena and the output location. 

## Additional properties when `connector` is `azure` or [named connector](./connector.md#name) of azure

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

## Additional properties when `connector` is `bigquery` or [named connector](./connector.md#name) of bigquery

### `project_id`

_[string]_ - ID of the BigQuery project. 

## Additional properties when `connector` is `duckdb` or [named connector](./connector.md#name) of duckdb

### `path`

_[string]_ - Path to the data source. 

### `format`

_[string]_ - Format of the data source (e.g., csv, json, parquet). 

### `pre_exec`

_[string]_ - refers to a SQL queries to run before the main query, available for DuckDB based models 

### `post_exec`

_[string]_ - refers to a SQL query that is run after the main query, available for DuckDB based models 

## Additional properties when `connector` is `gcs` or [named connector](./connector.md#name) of gcs

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

## Additional properties when `connector` is `local_file` or [named connector](./connector.md#name) of local_file

### `path`

_[string]_ - Path to the data source. 

### `format`

_[string]_ - Format of the data source (e.g., csv, json, parquet). 

## Additional properties when `connector` is `redshift` or [named connector](./connector.md#name) of redshift

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

## Additional properties when `connector` is `s3` or [named connector](./connector.md#name) of s3

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

## Additional properties when `connector` is `salesforce` or [named connector](./connector.md#name) of salesforce

### `soql`

_[string]_ - SOQL query to execute against the Salesforce instance. 

### `sobject`

_[string]_ - Salesforce object (e.g., Account, Contact) targeted by the query. 

### `queryAll`

_[boolean]_ - Whether to include deleted and archived records in the query (uses queryAll API). 