---
note: GENERATED. DO NOT EDIT.
title: Advanced Models YAML
sidebar_position: 37
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

_[object]_ - Connector YAML files define how Rill connects to external data sources and OLAP engines. Each connector specifies a driver type and its required connection parameters.
 

  - **`type`** - _[string]_ - Refers to the resource type and must be `connector` _(required)_

  - **`driver`** - _[string]_ - The type of connector, see [available connectors](#available-connector-types) (required) _(required)_

#### Option 1: Athena

**Type:** _[object]_

**Description:** Configuration properties specific to the athena

    - **`driver`** - _[string]_ - Refers to the driver type and must be driver `athena` _(required)_

    - **`aws_access_key_id`** - _[string]_ - AWS Access Key ID used for authentication. Required when using static credentials directly or as base credentials for assuming a role. 

    - **`aws_secret_access_key`** - _[string]_ - AWS Secret Access Key paired with the Access Key ID. Required when using static credentials directly or as base credentials for assuming a role. 

    - **`aws_access_token`** - _[string]_ - AWS session token used with temporary credentials. Required only if the Access Key and Secret Key are part of a temporary session credentials. 

    - **`role_arn`** - _[string]_ - ARN of the IAM role to assume. When specified, the SDK uses the base credentials to call STS AssumeRole and obtain temporary credentials scoped to this role. 

    - **`role_session_name`** - _[string]_ - Session name to associate with the STS AssumeRole session. Used only if 'role_arn' is specified. Useful for identifying and auditing the session. 

    - **`external_id`** - _[string]_ - External ID required by some roles when assuming them, typically for cross-account access. Used only if 'role_arn' is specified and the role's trust policy requires it. 

    - **`workgroup`** - _[string]_ - Athena workgroup to use for query execution. Defaults to 'primary' if not specified. 

    - **`output_location`** - _[string]_ - S3 URI where Athena query results should be stored (e.g., s3://your-bucket/athena/results/). Optional if the selected workgroup has a default result configuration. 

    - **`aws_region`** - _[string]_ - AWS region where Athena and the result S3 bucket are located (e.g., us-east-1). Defaults to 'us-east-1' if not specified. 

    - **`allow_host_access`** - _[boolean]_ - Allow the Athena client to access host environment configurations such as environment variables or local AWS credential files. Defaults to true, enabling use of credentials and settings from the host environment unless explicitly disabled. 

#### Option 2: Azure

**Type:** _[object]_

**Description:** Configuration properties specific to the azure

    - **`driver`** - _[string]_ - Refers to the driver type and must be driver `azure` _(required)_

    - **`azure_storage_account`** - _[string]_ - Azure storage account name 

    - **`azure_storage_key`** - _[string]_ - Azure storage access key 

    - **`azure_storage_sas_token`** - _[string]_ - Optional azure SAS token for authentication 

    - **`azure_storage_connection_string`** - _[string]_ - Optional azure connection string for storage account 

    - **`azure_storage_bucket`** - _[string]_ - Name of the Azure Blob Storage container (equivalent to an S3 bucket) _(required)_

    - **`allow_host_access`** - _[boolean]_ - Allow access to host environment configuration 

#### Option 3: BigQuery

**Type:** _[object]_

**Description:** Configuration properties specific to the bigquery

    - **`driver`** - _[string]_ - Refers to the driver type and must be driver `bigquery` _(required)_

    - **`google_application_credentials`** - _[string]_ - Raw contents of the Google Cloud service account key (in JSON format) used for authentication. 

    - **`project_id`** - _[string]_ - ID of the Google Cloud project to use for BigQuery operations. This can be omitted only if the project ID is included in the service account key. 

    - **`allow_host_access`** - _[boolean]_ - Enable the BigQuery client to use credentials from the host environment when no service account JSON is provided. This includes Application Default Credentials from environment variables, local credential files, or the Google Compute Engine metadata server. Defaults to true, allowing seamless authentication in GCP environments. 

#### Option 4: ClickHouse

**Type:** _[object]_

**Description:** Configuration properties specific to the clickhouse

    - **`driver`** - _[string]_ - Refers to the driver type and must be driver `clickhouse` _(required)_

    - **`managed`** - _[boolean]_ - `true` means Rill will provision the connector using the default provisioner. `false` disables automatic provisioning. 

    - **`mode`** - _[string]_ - `read` - Controls the operation mode for the ClickHouse connection. Defaults to 'read' for safe operation with external databases. Set to 'readwrite' to enable model creation and table mutations. Note: When 'managed: true', this is automatically set to 'readwrite'. 

    - **`dsn`** - _[string]_ - DSN(Data Source Name) for the ClickHouse connection 

    - **`username`** - _[string]_ - Username for authentication 

    - **`password`** - _[string]_ - Password for authentication 

    - **`host`** - _[string]_ - Host where the ClickHouse instance is running 

    - **`port`** - _[integer]_ - Port where the ClickHouse instance is accessible 

    - **`database`** - _[string]_ - Name of the ClickHouse database within the cluster 

    - **`ssl`** - _[boolean]_ - Indicates whether a secured SSL connection is required 

    - **`cluster`** - _[string]_ - Cluster name, required for running distributed queries 

    - **`log_queries`** - _[boolean]_ - Controls whether to log raw SQL queries 

    - **`settings_override`** - _[string]_ - override the default settings used in queries. example `readonly = 1, session_timezone = 'UTC'` 

    - **`embed_port`** - _[integer]_ - Port to run ClickHouse locally (0 for random port) 

    - **`can_scale_to_zero`** - _[boolean]_ - Indicates if the database can scale to zero 

    - **`max_open_conns`** - _[integer]_ - Maximum number of open connections to the database 

    - **`max_idle_conns`** - _[integer]_ - Maximum number of idle connections in the pool 

    - **`dial_timeout`** - _[string]_ - Timeout for dialing the ClickHouse server 

    - **`conn_max_lifetime`** - _[string]_ - Maximum time a connection may be reused 

    - **`read_timeout`** - _[string]_ - Maximum time for a connection to read data 

#### Option 5: Druid

**Type:** _[object]_

**Description:** Configuration properties specific to the druid

    - **`driver`** - _[string]_ - Refers to the driver type and must be driver `druid` _(required)_

    - **`dsn`** - _[string]_ - Data Source Name (DSN) for connecting to Druid _(required)_

    - **`username`** - _[string]_ - Username for authenticating with Druid 

    - **`password`** - _[string]_ - Password for authenticating with Druid 

    - **`host`** - _[string]_ - Hostname of the Druid coordinator or broker 

    - **`port`** - _[integer]_ - Port number of the Druid service 

    - **`ssl`** - _[boolean]_ - Enable SSL for secure connection 

    - **`log_queries`** - _[boolean]_ - Log raw SQL queries sent to Druid 

    - **`max_open_conns`** - _[integer]_ - Maximum number of open database connections (0 = default, -1 = unlimited) 

    - **`skip_version_check`** - _[boolean]_ - Skip checking Druid version compatibility 

#### Option 6: DuckDB

**Type:** _[object]_

**Description:** Configuration properties specific to the duckdb

    - **`driver`** - _[string]_ - Refers to the driver type and must be driver `duckdb` _(required)_

    - **`pool_size`** - _[integer]_ - Number of concurrent connections and queries allowed 

    - **`allow_host_access`** - _[boolean]_ - Whether access to the local environment and file system is allowed 

    - **`cpu`** - _[integer]_ - Number of CPU cores available to the database 

    - **`memory_limit_gb`** - _[integer]_ - Amount of memory in GB available to the database 

    - **`read_write_ratio`** - _[number]_ - Ratio of resources allocated to the read database; used to divide CPU and memory 

    - **`init_sql`** - _[string]_ - is executed during database initialization. 

    - **`conn_init_sql`** - _[string]_ - is executed when a new connection is initialized. 

    - **`secrets`** - _[string]_ - Comma-separated list of other connector names to create temporary secrets for in DuckDB before executing a model. 

    - **`log_queries`** - _[boolean]_ - Whether to log raw SQL queries executed through OLAP 

#### Option 7: GCS

**Type:** _[object]_

**Description:** Configuration properties specific to the gcs

    - **`driver`** - _[string]_ - Refers to the driver type and must be driver `gcs` _(required)_

    - **`google_application_credentials`** - _[string]_ - Google Cloud credentials JSON string 

    - **`bucket`** - _[string]_ - Name of gcs bucket _(required)_

    - **`allow_host_access`** - _[boolean]_ - Allow access to host environment configuration 

    - **`key_id`** - _[string]_ - Optional S3-compatible Key ID when used in compatibility mode 

    - **`secret`** - _[string]_ - Optional S3-compatible Secret when used in compatibility mode 

#### Option 8: HTTPS

**Type:** _[object]_

**Description:** Configuration properties specific to the https

    - **`driver`** - _[string]_ - Refers to the driver type and must be driver `https` _(required)_

    - **`path`** - _[string]_ - The full HTTPS URI to fetch data from _(required)_

    - **`headers`** - _[object]_ - HTTP headers to include in the request 

#### Option 9: MotherDuck

**Type:** _[object]_

**Description:** Configuration properties specific to the motherduck

    - **`driver`** - _[string]_ - Refers to the driver type and must be driver `duckdb` _(required)_

    - **`path`** - _[string]_ - Path to your MD database _(required)_

    - **`init_sql`** - _[string]_ - SQL executed during database initialization. _(required)_

#### Option 10: MySQL

**Type:** _[object]_

**Description:** Configuration properties specific to the mysql

    - **`driver`** - _[string]_ - Refers to the driver type and must be driver `mysql` _(required)_

    - **`dsn`** - _[string]_ - DSN(Data Source Name) for the mysql connection 

    - **`host`** - _[string]_ - Hostname of the MySQL server 

    - **`port`** - _[integer]_ - Port number for the MySQL server 

    - **`database`** - _[string]_ - Name of the MySQL database 

    - **`user`** - _[string]_ - Username for authentication 

    - **`password`** - _[string]_ - Password for authentication 

    - **`ssl_mode`** - _[string]_ - SSL mode can be DISABLED, PREFERRED or REQUIRED 

#### Option 11: Pinot

**Type:** _[object]_

**Description:** Configuration properties specific to the pinot

    - **`driver`** - _[string]_ - Refers to the driver type and must be driver `pinot` _(required)_

    - **`dsn`** - _[string]_ - DSN(Data Source Name) for the Pinot connection _(required)_

    - **`username`** - _[string]_ - Username for authenticating with Pinot 

    - **`password`** - _[string]_ - Password for authenticating with Pinot 

    - **`broker_host`** - _[string]_ - Hostname of the Pinot broker _(required)_

    - **`broker_port`** - _[integer]_ - Port number for the Pinot broker 

    - **`controller_host`** - _[string]_ - Hostname of the Pinot controller _(required)_

    - **`controller_port`** - _[integer]_ - Port number for the Pinot controller 

    - **`ssl`** - _[boolean]_ - Enable SSL connection to Pinot 

    - **`log_queries`** - _[boolean]_ - Log raw SQL queries executed through Pinot 

    - **`max_open_conns`** - _[integer]_ - Maximum number of open connections to the Pinot database 

#### Option 12: Postgres

**Type:** _[object]_

**Description:** Configuration properties specific to the postgres

    - **`driver`** - _[string]_ - Refers to the driver type and must be driver `postgres` _(required)_

    - **`dsn`** - _[string]_ - DSN(Data Source Name) for the postgres connection 

    - **`host`** - _[string]_ - Hostname of the Postgres server 

    - **`port`** - _[string]_ - Port number for the Postgres server 

    - **`dbname`** - _[string]_ - Name of the Postgres database 

    - **`user`** - _[string]_ - Username for authentication 

    - **`password`** - _[string]_ - Password for authentication 

    - **`sslmode`** - _[string]_ - SSL mode can be disable, allow, prefer or require 

#### Option 13: Redshift

**Type:** _[object]_

**Description:** Configuration properties specific to the redshift

    - **`driver`** - _[string]_ - Refers to the driver type and must be driver `redshift` _(required)_

    - **`aws_access_key_id`** - _[string]_ - AWS Access Key ID used for authenticating with Redshift. _(required)_

    - **`aws_secret_access_key`** - _[string]_ - AWS Secret Access Key used for authenticating with Redshift. _(required)_

    - **`aws_access_token`** - _[string]_ - AWS Session Token for temporary credentials (optional). 

    - **`region`** - _[string]_ - AWS region where the Redshift cluster or workgroup is hosted (e.g., 'us-east-1'). 

    - **`database`** - _[string]_ - Name of the Redshift database to query. _(required)_

    - **`workgroup`** - _[string]_ - Workgroup name for Redshift Serverless, in case of provisioned Redshift clusters use 'cluster_identifier'. 

    - **`cluster_identifier`** - _[string]_ - Cluster identifier for provisioned Redshift clusters, in case of Redshift Serverless use 'workgroup' . 

#### Option 14: S3

**Type:** _[object]_

**Description:** Configuration properties specific to the s3

    - **`driver`** - _[string]_ - Refers to the driver type and must be driver `s3` _(required)_

    - **`aws_access_key_id`** - _[string]_ - AWS Access Key ID used for authentication 

    - **`aws_secret_access_key`** - _[string]_ - AWS Secret Access Key used for authentication 

    - **`aws_access_token`** - _[string]_ - Optional AWS session token for temporary credentials 

    - **`bucket`** - _[string]_ - Name of s3 bucket _(required)_

    - **`endpoint`** - _[string]_ - Optional custom endpoint URL for S3-compatible storage 

    - **`region`** - _[string]_ - AWS region of the S3 bucket 

    - **`allow_host_access`** - _[boolean]_ - Allow access to host environment configuration 

    - **`retain_files`** - _[boolean]_ - Whether to retain intermediate files after processing 

#### Option 15: Salesforce

**Type:** _[object]_

**Description:** Configuration properties specific to the salesforce

    - **`driver`** - _[string]_ - Refers to the driver type and must be driver `salesforce` _(required)_

    - **`username`** - _[string]_ - Salesforce account username _(required)_

    - **`password`** - _[string]_ - Salesforce account password (secret) 

    - **`key`** - _[string]_ - Authentication key for Salesforce (secret) 

    - **`endpoint`** - _[string]_ - Salesforce API endpoint URL _(required)_

    - **`client_id`** - _[string]_ - Client ID used for Salesforce OAuth authentication 

#### Option 16: Slack

**Type:** _[object]_

**Description:** Configuration properties specific to the slack

    - **`driver`** - _[string]_ - Refers to the driver type and must be driver `slack` _(required)_

    - **`bot_token`** - _[string]_ - Bot token used for authenticating Slack API requests _(required)_

#### Option 17: Snowflake

**Type:** _[object]_

**Description:** Configuration properties specific to the snowflake

    - **`driver`** - _[string]_ - Refers to the driver type and must be driver `snowflake` _(required)_

    - **`dsn`** - _[string]_ - DSN (Data Source Name) for the Snowflake connection _(required)_

    - **`parallel_fetch_limit`** - _[integer]_ - Maximum number of concurrent fetches during query execution 

#### Option 18: SQLite

**Type:** _[object]_

**Description:** Configuration properties specific to the sqlite

    - **`driver`** - _[string]_ - Refers to the driver type and must be driver `sqlite` _(required)_

    - **`dsn`** - _[string]_ - DSN(Data Source Name) for the sqlite connection _(required)_

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

#### Option 1: SQL Query

**Type:** _[object]_

**Description:** Executes a raw SQL query against the project's data models.

    - **`sql`** - _[string]_ - Raw SQL query to run against existing models in the project. _(required)_

    - **`connector`** - _[string]_ - specifies the connector to use when running SQL or glob queries. 

#### Option 2: Metrics View Query

**Type:** _[object]_

**Description:** Executes a SQL query that targets a defined metrics view.

    - **`metrics_sql`** - _[string]_ - SQL query that targets a metrics view in the project _(required)_

#### Option 3: Custom API Call

**Type:** _[object]_

**Description:** Calls a custom API defined in the project to compute data.

    - **`api`** - _[string]_ - Name of a custom API defined in the project. _(required)_

    - **`args`** - _[object]_ - Arguments to pass to the custom API. 

#### Option 4: File Glob Query

**Type:** _[object]_

**Description:** Uses a file-matching pattern (glob) to query data from a connector.

    - **`glob`** - _[anyOf]_ - Defines the file path or pattern to query from the specified connector. _(required)_

      - **option 1** - _[string]_ - A simple file path/glob pattern as a string.

      - **option 2** - _[object]_ - An object-based configuration for specifying a file path/glob pattern with advanced options.

    - **`connector`** - _[string]_ - Specifies the connector to use with the glob input. 

#### Option 5: Resource Status Check

**Type:** _[object]_

**Description:** Uses the status of a resource as data.

    - **`resource_status`** - _[object]_ - Based on resource status _(required)_

      - **`where_error`** - _[boolean]_ - Indicates whether the condition should trigger when the resource is in an error state. 

### `partitions`

_[oneOf]_ - Refers to the how your data is partitioned, cannot be used with state. (optional) 

#### Option 1: SQL Query

**Type:** _[object]_

**Description:** Executes a raw SQL query against the project's data models.

    - **`sql`** - _[string]_ - Raw SQL query to run against existing models in the project. _(required)_

    - **`connector`** - _[string]_ - specifies the connector to use when running SQL or glob queries. 

#### Option 2: Metrics View Query

**Type:** _[object]_

**Description:** Executes a SQL query that targets a defined metrics view.

    - **`metrics_sql`** - _[string]_ - SQL query that targets a metrics view in the project _(required)_

#### Option 3: Custom API Call

**Type:** _[object]_

**Description:** Calls a custom API defined in the project to compute data.

    - **`api`** - _[string]_ - Name of a custom API defined in the project. _(required)_

    - **`args`** - _[object]_ - Arguments to pass to the custom API. 

#### Option 4: File Glob Query

**Type:** _[object]_

**Description:** Uses a file-matching pattern (glob) to query data from a connector.

    - **`glob`** - _[anyOf]_ - Defines the file path or pattern to query from the specified connector. _(required)_

      - **option 1** - _[string]_ - A simple file path/glob pattern as a string.

      - **option 2** - _[object]_ - An object-based configuration for specifying a file path/glob pattern with advanced options.

    - **`connector`** - _[string]_ - Specifies the connector to use with the glob input. 

#### Option 5: Resource Status Check

**Type:** _[object]_

**Description:** Uses the status of a resource as data.

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

## Additional properties when `connector` is [`athena`](./connectors#athena)

### `output_location`

_[string]_ - Output location for query results in S3. 

### `workgroup`

_[string]_ - AWS Athena workgroup to use for queries. 

### `region`

_[string]_ - AWS region to connect to Athena and the output location. 

## Additional properties when `connector` is [`azure`](./connectors#azure)

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

## Additional properties when `connector` is [`bigquery`](./connectors#bigquery)

### `project_id`

_[string]_ - ID of the BigQuery project. 

## Additional properties when `connector` is [`duckdb`](./connectors#duckdb)

### `path`

_[string]_ - Path to the data source. 

### `format`

_[string]_ - Format of the data source (e.g., csv, json, parquet). 

### `pre_exec`

_[string]_ - refers to SQL queries to run before the main query, available for DuckDB-based models. _(optional)_. Ensure `pre_exec` queries are idempotent. Use `IF NOT EXISTS` statements when applicable. 

### `post_exec`

_[string]_ - refers to a SQL query that is run after the main query, available for DuckDB-based models. _(optional)_. Ensure `post_exec` queries are idempotent. Use `IF EXISTS` statements when applicable. 

```yaml
pre_exec: ATTACH IF NOT EXISTS 'dbname=postgres host=localhost port=5432 user=postgres password=postgres' AS postgres_db (TYPE POSTGRES);
sql: SELECT * FROM postgres_query('postgres_db', 'SELECT * FROM USERS')
post_exec: DETACH DATABASE IF EXISTS postgres_db
```

## Additional properties when `connector` is [`gcs`](./connectors#gcs)

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

## Additional properties when `connector` is [`redshift`](./connectors#redshift)

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

## Additional properties when `connector` is [`s3`](./connectors#s3)

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

## Additional properties when `connector` is [`salesforce`](./connector#salesforce)

### `soql`

_[string]_ - SOQL query to execute against the Salesforce instance. 

### `sobject`

_[string]_ - Salesforce object (e.g., Account, Contact) targeted by the query. 

### `queryAll`

_[boolean]_ - Whether to include deleted and archived records in the query (uses queryAll API). 