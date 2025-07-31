---
note: GENERATED. DO NOT EDIT.
title: Connector YAML
sidebar_position: 35
---

When you add olap_connector to your rill.yaml file, you will need to set up a `<connector_name>.yaml` file in the 'connectors' directory. This file requires the following parameters,type and driver (see below for more parameter options). Rill will automatically test the connectivity to the OLAP engine upon saving the file. This can be viewed in the connectors tab in the UI.

:::tip Did you know?

Starting from Rill 0.46, you can directly create OLAP engines from the UI! Select + Add -> Data -> Connect an OLAP engine

:::


## Properties

### `type`

_[string]_ - Refers to the resource type and must be `connector` _(required)_

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
- [druid](#druid)
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
- [slack](#slack)
- [snowflake](#snowflake)
- [sqlite](#sqlite)

## athena

Configuration properties specific to the athena

### `driver`

_[string]_ - Refers to the driver type and must be driver `athena` _(required)_

### `aws_access_key_id`

_[string]_ - AWS Access Key ID used for authentication. Required when using static credentials directly or as base credentials for assuming a role. 

### `aws_secret_access_key`

_[string]_ - AWS Secret Access Key paired with the Access Key ID. Required when using static credentials directly or as base credentials for assuming a role. 

### `aws_access_token`

_[string]_ - AWS session token used with temporary credentials. Required only if the Access Key and Secret Key are part of a temporary session credentials. 

### `role_arn`

_[string]_ - ARN of the IAM role to assume. When specified, the SDK uses the base credentials to call STS AssumeRole and obtain temporary credentials scoped to this role. 

### `role_session_name`

_[string]_ - Session name to associate with the STS AssumeRole session. Used only if 'role_arn' is specified. Useful for identifying and auditing the session. 

### `external_id`

_[string]_ - External ID required by some roles when assuming them, typically for cross-account access. Used only if 'role_arn' is specified and the role's trust policy requires it. 

### `workgroup`

_[string]_ - Athena workgroup to use for query execution. Defaults to 'primary' if not specified. 

### `output_location`

_[string]_ - S3 URI where Athena query results should be stored (e.g., s3://your-bucket/athena/results/). Optional if the selected workgroup has a default result configuration. 

### `aws_region`

_[string]_ - AWS region where Athena and the result S3 bucket are located (e.g., us-east-1). Defaults to 'us-east-1' if not specified. 

### `allow_host_access`

_[boolean]_ - Allow the Athena client to access host environment configurations such as environment variables or local AWS credential files. Defaults to true, enabling use of credentials and settings from the host environment unless explicitly disabled. 

## azure

Configuration properties specific to the azure

### `driver`

_[string]_ - Refers to the driver type and must be driver `azure` _(required)_

### `azure_storage_account`

_[string]_ - Azure storage account name 

### `azure_storage_key`

_[string]_ - Azure storage access key 

### `azure_storage_sas_token`

_[string]_ - Optional azure SAS token for authentication 

### `azure_storage_connection_string`

_[string]_ - Optional azure connection string for storage account 

### `azure_storage_bucket`

_[string]_ - Name of the Azure Blob Storage container (equivalent to an S3 bucket) _(required)_

### `allow_host_access`

_[boolean]_ - Allow access to host environment configuration 

## bigquery

Configuration properties specific to the bigquery

### `driver`

_[string]_ - Refers to the driver type and must be driver `bigquery` _(required)_

### `google_application_credentials`

_[string]_ - Raw contents of the Google Cloud service account key (in JSON format) used for authentication. 

### `project_id`

_[string]_ - ID of the Google Cloud project to use for BigQuery operations. This can be omitted only if the project ID is included in the service account key. 

### `allow_host_access`

_[boolean]_ - Enable the BigQuery client to use credentials from the host environment when no service account JSON is provided. This includes Application Default Credentials from environment variables, local credential files, or the Google Compute Engine metadata server. Defaults to true, allowing seamless authentication in GCP environments. 

## clickhouse

Configuration properties specific to the clickhouse

### `driver`

_[string]_ - Refers to the driver type and must be driver `clickhouse` _(required)_

### `managed`

_[boolean]_ - `true` means Rill will provision the connector using the default provisioner. `false` disables automatic provisioning. 

### `mode`

_[string]_ - `read` - Controls the operation mode for the ClickHouse connection. Defaults to 'read' for safe operation with external databases. Set to 'readwrite' to enable model creation and table mutations. Note: When 'managed: true', this is automatically set to 'readwrite'. 

### `dsn`

_[string]_ - DSN(Data Source Name) for the ClickHouse connection 

### `username`

_[string]_ - Username for authentication 

### `password`

_[string]_ - Password for authentication 

### `host`

_[string]_ - Host where the ClickHouse instance is running 

### `port`

_[integer]_ - Port where the ClickHouse instance is accessible 

### `database`

_[string]_ - Name of the ClickHouse database within the cluster 

### `ssl`

_[boolean]_ - Indicates whether a secured SSL connection is required 

### `cluster`

_[string]_ - Cluster name, required for running distributed queries 

### `log_queries`

_[boolean]_ - Controls whether to log raw SQL queries 

### `settings_override`

_[string]_ - override the default settings used in queries. example `readonly = 1, session_timezone = 'UTC'` 

### `embed_port`

_[integer]_ - Port to run ClickHouse locally (0 for random port) 

### `can_scale_to_zero`

_[boolean]_ - Indicates if the database can scale to zero 

### `max_open_conns`

_[integer]_ - Maximum number of open connections to the database 

### `max_idle_conns`

_[integer]_ - Maximum number of idle connections in the pool 

### `dial_timeout`

_[string]_ - Timeout for dialing the ClickHouse server 

### `conn_max_lifetime`

_[string]_ - Maximum time a connection may be reused 

### `read_timeout`

_[string]_ - Maximum time for a connection to read data 

## druid

Configuration properties specific to the druid

### `driver`

_[string]_ - Refers to the driver type and must be driver `druid` _(required)_

### `dsn`

_[string]_ - Data Source Name (DSN) for connecting to Druid _(required)_

### `username`

_[string]_ - Username for authenticating with Druid 

### `password`

_[string]_ - Password for authenticating with Druid 

### `host`

_[string]_ - Hostname of the Druid coordinator or broker 

### `port`

_[integer]_ - Port number of the Druid service 

### `ssl`

_[boolean]_ - Enable SSL for secure connection 

### `log_queries`

_[boolean]_ - Log raw SQL queries sent to Druid 

### `max_open_conns`

_[integer]_ - Maximum number of open database connections (0 = default, -1 = unlimited) 

### `skip_version_check`

_[boolean]_ - Skip checking Druid version compatibility 

## duckdb

Configuration properties specific to the duckdb

### `driver`

_[string]_ - Refers to the driver type and must be driver `duckdb` _(required)_

### `pool_size`

_[integer]_ - Number of concurrent connections and queries allowed 

### `allow_host_access`

_[boolean]_ - Whether access to the local environment and file system is allowed 

### `cpu`

_[integer]_ - Number of CPU cores available to the database 

### `memory_limit_gb`

_[integer]_ - Amount of memory in GB available to the database 

### `read_write_ratio`

_[number]_ - Ratio of resources allocated to the read database; used to divide CPU and memory 

### `init_sql`

_[string]_ - is executed during database initialization. 

### `conn_init_sql`

_[string]_ - is executed when a new connection is initialized. 

### `secrets`

_[string]_ - Comma-separated list of other connector names to create temporary secrets for in DuckDB before executing a model. 

### `log_queries`

_[boolean]_ - Whether to log raw SQL queries executed through OLAP 

## gcs

Configuration properties specific to the gcs

### `driver`

_[string]_ - Refers to the driver type and must be driver `gcs` _(required)_

### `google_application_credentials`

_[string]_ - Google Cloud credentials JSON string 

### `bucket`

_[string]_ - Name of gcs bucket _(required)_

### `allow_host_access`

_[boolean]_ - Allow access to host environment configuration 

### `key_id`

_[string]_ - Optional S3-compatible Key ID when used in compatibility mode 

### `secret`

_[string]_ - Optional S3-compatible Secret when used in compatibility mode 

## https

Configuration properties specific to the https

### `driver`

_[string]_ - Refers to the driver type and must be driver `https` _(required)_

### `path`

_[string]_ - The full HTTPS URI to fetch data from _(required)_

### `headers`

_[object]_ - HTTP headers to include in the request 

## local_file

Configuration properties specific to the local_file

### `driver`

_[string]_ - Refers to the driver type and must be driver `local_file` _(required)_

### `dsn`

_[string]_ - Data Source Name (DSN) indicating the file path or location of the local file _(required)_

### `allow_host_access`

_[boolean]_ - Flag to indicate if access to host-level file paths is permitted 

## motherduck

Configuration properties specific to the motherduck

### `driver`

_[string]_ - Refers to the driver type and must be driver `motherduck` _(required)_

### `dsn`

_[string]_ - Data Source Name (DSN) specifying the MotherDuck connection endpoint _(required)_

### `token`

_[string]_ - Authentication token for accessing MotherDuck (secret) _(required)_

## mysql

Configuration properties specific to the mysql

### `driver`

_[string]_ - Refers to the driver type and must be driver `mysql` _(required)_

### `dsn`

_[string]_ - DSN(Data Source Name) for the mysql connection 

### `host`

_[string]_ - Hostname of the MySQL server 

### `port`

_[integer]_ - Port number for the MySQL server 

### `database`

_[string]_ - Name of the MySQL database 

### `user`

_[string]_ - Username for authentication 

### `password`

_[string]_ - Password for authentication 

### `ssl_mode`

_[string]_ - SSL mode can be DISABLED, PREFERRED or REQUIRED 

## pinot

Configuration properties specific to the pinot

### `driver`

_[string]_ - Refers to the driver type and must be driver `pinot` _(required)_

### `dsn`

_[string]_ - DSN(Data Source Name) for the Pinot connection _(required)_

### `username`

_[string]_ - Username for authenticating with Pinot 

### `password`

_[string]_ - Password for authenticating with Pinot 

### `broker_host`

_[string]_ - Hostname of the Pinot broker _(required)_

### `broker_port`

_[integer]_ - Port number for the Pinot broker 

### `controller_host`

_[string]_ - Hostname of the Pinot controller _(required)_

### `controller_port`

_[integer]_ - Port number for the Pinot controller 

### `ssl`

_[boolean]_ - Enable SSL connection to Pinot 

### `log_queries`

_[boolean]_ - Log raw SQL queries executed through Pinot 

### `max_open_conns`

_[integer]_ - Maximum number of open connections to the Pinot database 

## postgres

Configuration properties specific to the postgres

### `driver`

_[string]_ - Refers to the driver type and must be driver `postgres` _(required)_

### `dsn`

_[string]_ - DSN(Data Source Name) for the postgres connection 

### `host`

_[string]_ - Hostname of the Postgres server 

### `port`

_[string]_ - Port number for the Postgres server 

### `dbname`

_[string]_ - Name of the Postgres database 

### `user`

_[string]_ - Username for authentication 

### `password`

_[string]_ - Password for authentication 

### `sslmode`

_[string]_ - SSL mode can be disable, allow, prefer or require 

## redshift

Configuration properties specific to the redshift

### `driver`

_[string]_ - Refers to the driver type and must be driver `redshift` _(required)_

### `aws_access_key_id`

_[string]_ - AWS Access Key ID used for authenticating with Redshift. _(required)_

### `aws_secret_access_key`

_[string]_ - AWS Secret Access Key used for authenticating with Redshift. _(required)_

### `aws_access_token`

_[string]_ - AWS Session Token for temporary credentials (optional). 

### `region`

_[string]_ - AWS region where the Redshift cluster or workgroup is hosted (e.g., 'us-east-1'). 

### `database`

_[string]_ - Name of the Redshift database to query. _(required)_

### `workgroup`

_[string]_ - Workgroup name for Redshift Serverless, in case of provisioned Redshift clusters use 'cluster_identifier'. 

### `cluster_identifier`

_[string]_ - Cluster identifier for provisioned Redshift clusters, in case of Redshift Serverless use 'workgroup' . 

## s3

Configuration properties specific to the s3

### `driver`

_[string]_ - Refers to the driver type and must be driver `s3` _(required)_

### `aws_access_key_id`

_[string]_ - AWS Access Key ID used for authentication 

### `aws_secret_access_key`

_[string]_ - AWS Secret Access Key used for authentication 

### `aws_access_token`

_[string]_ - Optional AWS session token for temporary credentials 

### `bucket`

_[string]_ - Name of s3 bucket _(required)_

### `endpoint`

_[string]_ - Optional custom endpoint URL for S3-compatible storage 

### `region`

_[string]_ - AWS region of the S3 bucket 

### `allow_host_access`

_[boolean]_ - Allow access to host environment configuration 

### `retain_files`

_[boolean]_ - Whether to retain intermediate files after processing 

## salesforce

Configuration properties specific to the salesforce

### `driver`

_[string]_ - Refers to the driver type and must be driver `salesforce` _(required)_

### `username`

_[string]_ - Salesforce account username _(required)_

### `password`

_[string]_ - Salesforce account password (secret) 

### `key`

_[string]_ - Authentication key for Salesforce (secret) 

### `endpoint`

_[string]_ - Salesforce API endpoint URL _(required)_

### `client_id`

_[string]_ - Client ID used for Salesforce OAuth authentication 

## slack

Configuration properties specific to the slack

### `driver`

_[string]_ - Refers to the driver type and must be driver `slack` _(required)_

### `bot_token`

_[string]_ - Bot token used for authenticating Slack API requests _(required)_

## snowflake

Configuration properties specific to the snowflake

### `driver`

_[string]_ - Refers to the driver type and must be driver `snowflake` _(required)_

### `dsn`

_[string]_ - DSN (Data Source Name) for the Snowflake connection _(required)_

### `parallel_fetch_limit`

_[integer]_ - Maximum number of concurrent fetches during query execution 

## sqlite

Configuration properties specific to the sqlite

### `driver`

_[string]_ - Refers to the driver type and must be driver `sqlite` _(required)_

### `dsn`

_[string]_ - DSN(Data Source Name) for the sqlite connection _(required)_