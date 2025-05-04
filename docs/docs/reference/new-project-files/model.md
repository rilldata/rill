---
note: GENERATED. DO NOT EDIT.
title: Model YAML
sidebar_position: 8
---



## Properties


**`type`**  - _[string]_ - Refers to the resource type and must be `model`  _(required)_

**`version`**  - _[integer]_ - Version of the parser to use for this file. Enables backwards compatibility for breaking changes. 

**`name`**  - _[string]_ - Name is usually inferred from the filename, but can be specified manually. 

**`namespace`**  - _[string]_ - Optional value to group resources by. Prepended to the resource name as `<namespace>/<name>`. 

**`refs`**  - _[array]_ - List of resource references, each as a string or map. 

     *option 1* - _[string]_ - A string reference like 'resource-name' or 'Kind/resource-name'.

     *option 2* - _[object]_ - An object reference with at least a 'name' and 'type'.

    - **`name`**  - _[string]_ -   _(required)_

    - **`type`**  - _[string]_ -  

 *option 1* - _[object]_ - 

**`connector`**  - _[string]_ -   _(required)_

**`aws_secret_access_key`**  - _[string]_ - AWS Secret Access Key for Athena access 

**`external_id`**  - _[string]_ - Optional External ID for assuming a role 

**`managed`**  - _[any of]_ - Boolean or map of properties 

**`role_arn`**  - _[string]_ - Optional AWS Role ARN to assume when accessing Athena 

**`role_session_name`**  - _[string]_ - Optional Session name when assuming the role 

**`allow_host_access`**  - _[boolean]_ - Allow access to host environment configuration 

**`aws_access_key_id`**  - _[string]_ - AWS Access Key ID for Athena access 

**`aws_access_token`**  - _[string]_ - Optional AWS session token for temporary credentials 

 *option 2* - _[object]_ - 

**`connector`**  - _[string]_ -   _(required)_

**`azure_storage_key`**  - _[string]_ - Azure storage access key 

**`azure_storage_sas_token`**  - _[string]_ - Optional azure SAS token for authentication 

**`managed`**  - _[any of]_ - Boolean or map of properties 

**`allow_host_access`**  - _[boolean]_ - Allow access to host environment configuration 

**`azure_storage_account`**  - _[string]_ - Azure storage account name 

**`azure_storage_connection_string`**  - _[string]_ - Optional azure connection string for storage account 

 *option 3* - _[object]_ - 

**`connector`**  - _[string]_ -   _(required)_

**`managed`**  - _[any of]_ - Boolean or map of properties 

**`allow_host_access`**  - _[boolean]_ - Allow access to host environment configuration 

**`google_application_credentials`**  - _[string]_ - Path to the Google Cloud credentials JSON file 

 *option 4* - _[object]_ - 

**`connector`**  - _[string]_ -   _(required)_

**`conn_max_lifetime`**  - _[string]_ - Maximum time a connection may be reused 

**`dial_timeout`**  - _[string]_ - Timeout for dialing the ClickHouse server 

**`max_idle_conns`**  - _[integer]_ - Maximum number of idle connections in the pool 

**`max_open_conns`**  - _[integer]_ - Maximum number of open connections to the database 

**`port`**  - _[integer]_ - Port where the ClickHouse instance is accessible 

**`read_timeout`**  - _[string]_ - Maximum time for a connection to read data 

**`settings_override`**  - _[string]_ - override the default settings used in queries. example `readonly = 1, session_timezone = 'UTC'` 

**`database`**  - _[string]_ - Name of the ClickHouse database within the cluster 

**`managed`**  - _[any of]_ - Boolean or map of properties 

**`password`**  - _[string]_ - Password for authentication 

**`username`**  - _[string]_ - Username for authentication 

**`cluster`**  - _[string]_ - Cluster name, required for running distributed queries 

**`host`**  - _[string]_ - Host where the ClickHouse instance is running 

**`ssl`**  - _[boolean]_ - Indicates whether a secured SSL connection is required 

**`can_scale_to_zero`**  - _[boolean]_ - Indicates if the database can scale to zero 

**`dsn`**  - _[string]_ - DSN(Data Source Name) for the ClickHouse connection 

**`embed_port`**  - _[integer]_ - Port to run ClickHouse locally (0 for random port) 

**`log_queries`**  - _[boolean]_ - Controls whether to log raw SQL queries 

 *option 5* - _[object]_ - 

**`connector`**  - _[string]_ -   _(required)_

**`dsn`**  - _[string]_ - Data Source Name (DSN) for connecting to Druid 

**`host`**  - _[string]_ - Hostname of the Druid coordinator or broker 

**`password`**  - _[string]_ - Password for authenticating with Druid 

**`ssl`**  - _[boolean]_ - Enable SSL for secure connection 

**`log_queries`**  - _[boolean]_ - Log raw SQL queries sent to Druid 

**`managed`**  - _[any of]_ - Boolean or map of properties 

**`max_open_conns`**  - _[integer]_ - Maximum number of open database connections (0 = default, -1 = unlimited) 

**`port`**  - _[integer]_ - Port number of the Druid service 

**`skip_version_check`**  - _[boolean]_ - Skip checking Druid version compatibility 

**`username`**  - _[string]_ - Username for authenticating with Druid 

 *option 6* - _[object]_ - 

**`connector`**  - _[string]_ -   _(required)_

**`allow_host_access`**  - _[boolean]_ - Whether access to the local environment and file system is allowed 

**`boot_queries`**  - _[string]_ - SQL to run when initializing a new connection, before extensions and defaults 

**`cpu`**  - _[integer]_ - Number of CPU cores available to the database 

**`init_sql`**  - _[string]_ - SQL to run when initializing a new connection, after extensions and defaults 

**`log_queries`**  - _[boolean]_ - Whether to log raw SQL queries executed through OLAP 

**`managed`**  - _[any of]_ - Boolean or map of properties 

**`memory_limit_gb`**  - _[integer]_ - Amount of memory in GB available to the database 

**`path`**  - _[string]_ - File path to the DuckDB database file 

**`pool_size`**  - _[integer]_ - Number of concurrent connections and queries allowed 

**`read_write_ratio`**  - _[number]_ - Ratio of resources allocated to the read database; used to divide CPU and memory 

 *option 7* - _[object]_ - 

**`connector`**  - _[string]_ -   _(required)_

**`allow_host_access`**  - _[boolean]_ - Allow access to host environment configuration 

**`google_application_credentials`**  - _[string]_ - Google Cloud credentials JSON string 

**`key_id`**  - _[string]_ - Optional S3-compatible Key ID when used in compatibility mode 

**`managed`**  - _[any of]_ - Boolean or map of properties 

**`secret`**  - _[string]_ - Optional S3-compatible Secret when used in compatibility mode 

 *option 8* - _[object]_ - 

**`connector`**  - _[string]_ -   _(required)_

**`headers`**  - _[object]_ - HTTP headers to include in the request 

**`managed`**  - _[any of]_ - Boolean or map of properties 

**`path`**  - _[string]_ - The full HTTPS URI to fetch data from 

 *option 9* - _[object]_ - 

**`connector`**  - _[string]_ -   _(required)_

**`allow_host_access`**  - _[boolean]_ - Flag to indicate if access to host-level file paths is permitted 

**`dsn`**  - _[string]_ - Data Source Name (DSN) indicating the file path or location of the local file 

**`managed`**  - _[any of]_ - Boolean or map of properties 

 *option 10* - _[object]_ - 

**`connector`**  - _[string]_ -   _(required)_

**`dsn`**  - _[string]_ - Data Source Name (DSN) specifying the MotherDuck connection endpoint 

**`managed`**  - _[any of]_ - Boolean or map of properties 

**`token`**  - _[string]_ - Authentication token for accessing MotherDuck (secret) 

 *option 11* - _[object]_ - 

**`connector`**  - _[string]_ -   _(required)_

**`dsn`**  - _[string]_ - DSN(Data Source Name) for the mysql connection 

**`managed`**  - _[any of]_ - Boolean or map of properties 

 *option 12* - _[object]_ - 

**`connector`**  - _[string]_ -   _(required)_

**`broker_host`**  - _[string]_ - Hostname of the Pinot broker 

**`broker_port`**  - _[integer]_ - Port number for the Pinot broker 

**`controller_port`**  - _[integer]_ - Port number for the Pinot controller 

**`dsn`**  - _[string]_ - DSN(Data Source Name) for the Pinot connection 

**`managed`**  - _[any of]_ - Boolean or map of properties 

**`max_open_conns`**  - _[integer]_ - Maximum number of open connections to the Pinot database 

**`password`**  - _[string]_ - Password for authenticating with Pinot 

**`ssl`**  - _[boolean]_ - Enable SSL connection to Pinot 

**`controller_host`**  - _[string]_ - Hostname of the Pinot controller 

**`log_queries`**  - _[boolean]_ - Log raw SQL queries executed through Pinot 

**`username`**  - _[string]_ - Username for authenticating with Pinot 

 *option 13* - _[object]_ - 

**`connector`**  - _[string]_ -   _(required)_

**`dsn`**  - _[string]_ - DSN(Data Source Name) for the postgres connection 

**`managed`**  - _[any of]_ - Boolean or map of properties 

 *option 14* - _[object]_ - 

**`connector`**  - _[string]_ -   _(required)_

**`aws_access_key_id`**  - _[string]_ - AWS access key ID for authentication 

**`aws_access_token`**  - _[string]_ - AWS session token for temporary credentials (optional) 

**`aws_secret_access_key`**  - _[string]_ - AWS secret access key for authentication 

**`managed`**  - _[any of]_ - Boolean or map of properties 

**`allow_host_access`**  - _[boolean]_ - Allow access to host environment configuration 

 *option 15* - _[object]_ - 

**`connector`**  - _[string]_ -   _(required)_

**`retain_files`**  - _[boolean]_ - Whether to retain intermediate files after processing 

**`allow_host_access`**  - _[boolean]_ - Allow access to host environment configuration 

**`aws_access_key_id`**  - _[string]_ - AWS Access Key ID used for authentication 

**`aws_access_token`**  - _[string]_ - Optional AWS session token for temporary credentials 

**`aws_secret_access_key`**  - _[string]_ - AWS Secret Access Key used for authentication 

**`endpoint`**  - _[string]_ - Optional custom endpoint URL for S3-compatible storage 

**`managed`**  - _[any of]_ - Boolean or map of properties 

**`region`**  - _[string]_ - AWS region of the S3 bucket 

 *option 16* - _[object]_ - 

**`connector`**  - _[string]_ -   _(required)_

**`client_id`**  - _[string]_ - Client ID used for Salesforce OAuth authentication 

**`endpoint`**  - _[string]_ - Salesforce API endpoint URL 

**`key`**  - _[string]_ - Authentication key for Salesforce (secret) 

**`managed`**  - _[any of]_ - Boolean or map of properties 

**`password`**  - _[string]_ - Salesforce account password (secret) 

**`username`**  - _[string]_ - Salesforce account username 

 *option 17* - _[object]_ - 

**`connector`**  - _[string]_ -   _(required)_

**`dsn`**  - _[string]_ - DSN (Data Source Name) for the Snowflake connection 

**`managed`**  - _[any of]_ - Boolean or map of properties 

**`parallel_fetch_limit`**  - _[integer]_ - Maximum number of concurrent fetches during query execution 

 *option 18* - _[object]_ - 

**`connector`**  - _[string]_ -   _(required)_

**`dsn`**  - _[string]_ - DSN(Data Source Name) for the sqlite connection 

**`managed`**  - _[any of]_ - Boolean or map of properties 

**`output`**  - _[object]_ - to define the properties of output 

  - **`connector`**  - _[string]_ - Refers to the connector type for the output table  _(required)_

**`partitions_watermark`**  - _[string]_ - Refers to a customizable timestamp that can be set to check if an object has been updated (optional). 

**`refresh`**  - _[object]_ - Specifies the refresh schedule that Rill should follow to re-ingest and update the underlying model data  

  - **`time_zone`**  - _[string]_ - Time zone to interpret the schedule in (e.g., 'UTC', 'America/Los_Angeles'). 

  - **`cron`**  - _[string]_ - A cron expression that defines the execution schedule 

  - **`disable`**  - _[boolean]_ - If true, disables the resource without deleting it. 

  - **`every`**  - _[string]_ - Run at a fixed interval using a Go duration string (e.g., '1h', '30m', '24h'). See: https://pkg.go.dev/time#ParseDuration 

  - **`ref_update`**  - _[boolean]_ - If true, allows the resource to run when a dependency updates. 

  - **`run_in_dev`**  - _[boolean]_ - If true, allows the schedule to run in development mode. 

**`stage`**  - _[object]_ - in the case of staging models, where an input source does not support direct write to the output and a staging table is required 

  - **`connector`**  - _[string]_ - Refers to the connector type for the staging table  _(required)_

**`incremental`**  - _[boolean]_ - whether incremental modeling is required (optional) 

**`materialize`**  - _[boolean]_ - models will be materialized in olap 

**`partitions`**  - _[object]_ - Refers to the how your data is partitioned, cannot be used with state. (optional) 

   *option 1* - 

   *option 2* - 

   *option 3* - 

   *option 4* - 

   *option 5* - 

**`partitions_concurrency`**  - _[integer]_ - Refers to the number of concurrent partitions that can be read at the same time (optional). 

**`sql`**  - _[string]_ - Raw SQL query to run against source 

**`state`**  - _[object]_ - Refers to the explicitly defined state of your model, cannot be used with partitions (optional) 

   *option 1* - 

   *option 2* - 

   *option 3* - 

   *option 4* - 

   *option 5* - 

**`timeout`**  - _[string]_ - The maximum time to wait for model ingestion 