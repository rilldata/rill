---
note: GENERATED. DO NOT EDIT.
title: Model YAML
sidebar_position: 8
---



## Properties


**`type`**  - _[string]_ - Refers to the resource type and must be `model`  _(required)_

**`name`**  - _[string]_ - Name is usually inferred from the filename, but can be specified manually. 

**`refs`**  - _[array of oneOf]_ - List of resource references, each as a string or map. 

  *option 1* - _[object]_ - An object reference with at least a `<name>` and `<type>`.

  - **`type`**  - _[string]_ - type of resource 

  - **`name`**  - _[string]_ - name of resource  _(required)_

  *option 2* - _[string]_ - A string reference like `<resource-name>` or `<type/resource-name>`.

**`refresh`**  - _[object]_ - Specifies the refresh schedule that Rill should follow to re-ingest and update the underlying model data  

  - **`cron`**  - _[string]_ - A cron expression that defines the execution schedule 

  - **`every`**  - _[string]_ - Run at a fixed interval using a Go duration string (e.g., '1h', '30m', '24h'). See: https://pkg.go.dev/time#ParseDuration 

  - **`time_zone`**  - _[string]_ - Time zone to interpret the schedule in (e.g., 'UTC', 'America/Los_Angeles'). 

  - **`disable`**  - _[boolean]_ - If true, disables the resource without deleting it. 

  - **`ref_update`**  - _[boolean]_ - If true, allows the resource to run when a dependency updates. 

  - **`run_in_dev`**  - _[boolean]_ - If true, allows the schedule to run in development mode. 

**`timeout`**  - _[string]_ - The maximum time to wait for model ingestion 

**`incremental`**  - _[boolean]_ - whether incremental modeling is required (optional) 

**`state`**  - _[oneOf]_ - Refers to the explicitly defined state of your model, cannot be used with partitions (optional) 

  *option 1* - _[object]_ 

  - **`sql`**  - _[string]_ - Raw SQL query to run against existing models in the project.  _(required)_

  - **`connector`**  - _[string]_ - specifies the connector to use when running SQL or glob queries. 

  *option 2* - _[object]_ 

  - **`metrics_sql`**  - _[string]_ - SQL query that targets a metrics view in the project  _(required)_

  *option 3* - _[object]_ 

  - **`api`**  - _[string]_ - Name of a custom API defined in the project.  _(required)_

  - **`args`**  - _[object]_ - Arguments to pass to the custom API. 

  *option 4* - _[object]_ 

  - **`glob`**  - _[anyOf]_ - Defines the file path or pattern to query from the specified connector.  _(required)_

    *option 1* - _[string]_ 

    *option 2* - _[object]_ 

  - **`connector`**  - _[string]_ - Specifies the connector to use with the glob input. 

  *option 5* - _[object]_ 

  - **`resource_status`**  - _[object]_ - Based on resource status  _(required)_

    - **`where_error`**  - _[boolean]_ - Indicates whether the condition should trigger when the resource is in an error state. 

**`partitions`**  - _[oneOf]_ - Refers to the how your data is partitioned, cannot be used with state. (optional) 

  *option 1* - _[object]_ 

  - **`sql`**  - _[string]_ - Raw SQL query to run against existing models in the project.  _(required)_

  - **`connector`**  - _[string]_ - specifies the connector to use when running SQL or glob queries. 

  *option 2* - _[object]_ 

  - **`metrics_sql`**  - _[string]_ - SQL query that targets a metrics view in the project  _(required)_

  *option 3* - _[object]_ 

  - **`api`**  - _[string]_ - Name of a custom API defined in the project.  _(required)_

  - **`args`**  - _[object]_ - Arguments to pass to the custom API. 

  *option 4* - _[object]_ 

  - **`glob`**  - _[anyOf]_ - Defines the file path or pattern to query from the specified connector.  _(required)_

    *option 1* - _[string]_ 

    *option 2* - _[object]_ 

  - **`connector`**  - _[string]_ - Specifies the connector to use with the glob input. 

  *option 5* - _[object]_ 

  - **`resource_status`**  - _[object]_ - Based on resource status  _(required)_

    - **`where_error`**  - _[boolean]_ - Indicates whether the condition should trigger when the resource is in an error state. 

**`sql`**  - _[string]_ - Raw SQL query to run against source 

**`materialize`**  - _[boolean]_ - models will be materialized in olap 

**`partitions_watermark`**  - _[string]_ - Refers to a customizable timestamp that can be set to check if an object has been updated (optional). 

**`partitions_concurrency`**  - _[integer]_ - Refers to the number of concurrent partitions that can be read at the same time (optional). 

**`stage`**  - _[object]_ - in the case of staging models, where an input source does not support direct write to the output and a staging table is required 

  - **`connector`**  - _[string]_ - Refers to the connector type for the staging table  _(required)_

**`output`**  - _[object]_ - to define the properties of output 

  - **`connector`**  - _[string]_ - Refers to the connector type for the output table 

**`dev`**  - _[object]_ - Overrides properties in development 

**`prod`**  - _[object]_ - Overrides properties in production 

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
- [snowflake](#snowflake)
- [sqlite](#sqlite)


### athena



**`connector`**  - _[string]_   _(required)_

**`output_location`**  - _[string]_ - Output location for query results in S3. 

**`workgroup`**  - _[string]_ - AWS Athena workgroup to use for queries. 

**`region`**  - _[string]_ - AWS region to connect to Athena and the output location. 

**`aws_access_key_id`**  - _[string]_ - AWS Access Key ID for Athena access 

**`aws_secret_access_key`**  - _[string]_ - AWS Secret Access Key for Athena access 

**`aws_access_token`**  - _[string]_ - Optional AWS session token for temporary credentials 

**`external_id`**  - _[string]_ - Optional External ID for assuming a role 

**`role_arn`**  - _[string]_ - Optional AWS Role ARN to assume when accessing Athena 

**`role_session_name`**  - _[string]_ - Optional Session name when assuming the role 

**`allow_host_access`**  - _[boolean]_ - Allow access to host environment configuration 

### azure



**`connector`**  - _[string]_   _(required)_

**`path`**  - _[string]_ - Path to the source 

**`account`**  - _[string]_ - Account identifier 

**`uri`**  - _[string]_ - Source URI 

**`extract`**  - _[object]_ - Arbitrary key-value pairs for extraction settings 

**`glob`**  - _[object]_ - Settings related to glob file matching. 

  - **`max_total_size`**  - _[integer]_ - Maximum total size (in bytes) matched by glob 

  - **`max_objects_matched`**  - _[integer]_ - Maximum number of objects matched by glob 

  - **`max_objects_listed`**  - _[integer]_ - Maximum number of objects listed in glob 

  - **`page_size`**  - _[integer]_ - Page size for glob listing 

**`batch_size`**  - _[string]_ - Size of a batch (e.g., '100MB') 

**`azure_storage_account`**  - _[string]_ - Azure storage account name 

**`azure_storage_key`**  - _[string]_ - Azure storage access key 

**`azure_storage_sas_token`**  - _[string]_ - Optional azure SAS token for authentication 

**`azure_storage_connection_string`**  - _[string]_ - Optional azure connection string for storage account 

**`allow_host_access`**  - _[boolean]_ - Allow access to host environment configuration 

### bigquery



**`connector`**  - _[string]_   _(required)_

**`project_id`**  - _[string]_ - ID of the BigQuery project. 

**`google_application_credentials`**  - _[string]_ - Path to the Google Cloud credentials JSON file 

**`allow_host_access`**  - _[boolean]_ - Allow access to host environment configuration 

### clickhouse



**`connector`**  - _[string]_   _(required)_

**`path`**  - _[string]_ - Path to the data source. 

**`format`**  - _[string]_ - Format of the data source (e.g., csv, json, parquet). 

**`managed`**  - _[boolean]_ - `true` means Rill will provision the connector using the default provisioner. `false` disables automatic provisioning. 

**`dsn`**  - _[string]_ - DSN(Data Source Name) for the ClickHouse connection 

**`username`**  - _[string]_ - Username for authentication 

**`password`**  - _[string]_ - Password for authentication 

**`host`**  - _[string]_ - Host where the ClickHouse instance is running 

**`port`**  - _[integer]_ - Port where the ClickHouse instance is accessible 

**`database`**  - _[string]_ - Name of the ClickHouse database within the cluster 

**`ssl`**  - _[boolean]_ - Indicates whether a secured SSL connection is required 

**`cluster`**  - _[string]_ - Cluster name, required for running distributed queries 

**`log_queries`**  - _[boolean]_ - Controls whether to log raw SQL queries 

**`settings_override`**  - _[string]_ - override the default settings used in queries. example `readonly = 1, session_timezone = 'UTC'` 

**`embed_port`**  - _[integer]_ - Port to run ClickHouse locally (0 for random port) 

**`can_scale_to_zero`**  - _[boolean]_ - Indicates if the database can scale to zero 

**`max_open_conns`**  - _[integer]_ - Maximum number of open connections to the database 

**`max_idle_conns`**  - _[integer]_ - Maximum number of idle connections in the pool 

**`dial_timeout`**  - _[string]_ - Timeout for dialing the ClickHouse server 

**`conn_max_lifetime`**  - _[string]_ - Maximum time a connection may be reused 

**`read_timeout`**  - _[string]_ - Maximum time for a connection to read data 

### druid



**`connector`**  - _[string]_   _(required)_

**`dsn`**  - _[string]_ - Data Source Name (DSN) for connecting to Druid 

**`username`**  - _[string]_ - Username for authenticating with Druid 

**`password`**  - _[string]_ - Password for authenticating with Druid 

**`host`**  - _[string]_ - Hostname of the Druid coordinator or broker 

**`port`**  - _[integer]_ - Port number of the Druid service 

**`ssl`**  - _[boolean]_ - Enable SSL for secure connection 

**`log_queries`**  - _[boolean]_ - Log raw SQL queries sent to Druid 

**`max_open_conns`**  - _[integer]_ - Maximum number of open database connections (0 = default, -1 = unlimited) 

**`skip_version_check`**  - _[boolean]_ - Skip checking Druid version compatibility 

### duckdb



**`connector`**  - _[string]_   _(required)_

**`path`**  - _[string]_ - Path to the data source. 

**`format`**  - _[string]_ - Format of the data source (e.g., csv, json, parquet). 

**`pre_exec`**  - _[string]_ - refers to a SQL queries to run before the main query, available for DuckDB based models 

**`post_exec`**  - _[string]_ - refers to a SQL query that is run after the main query, available for DuckDB based models 

**`pool_size`**  - _[integer]_ - Number of concurrent connections and queries allowed 

**`allow_host_access`**  - _[boolean]_ - Whether access to the local environment and file system is allowed 

**`cpu`**  - _[integer]_ - Number of CPU cores available to the database 

**`memory_limit_gb`**  - _[integer]_ - Amount of memory in GB available to the database 

**`read_write_ratio`**  - _[number]_ - Ratio of resources allocated to the read database; used to divide CPU and memory 

**`boot_queries`**  - _[string]_ - SQL to run when initializing a new connection, before extensions and defaults 

**`init_sql`**  - _[string]_ - SQL to run when initializing a new connection, after extensions and defaults 

**`log_queries`**  - _[boolean]_ - Whether to log raw SQL queries executed through OLAP 

### gcs



**`connector`**  - _[string]_   _(required)_

**`path`**  - _[string]_ - Path to the source 

**`uri`**  - _[string]_ - Source URI 

**`extract`**  - _[object]_ - key-value pairs for extraction settings 

**`glob`**  - _[object]_ - Settings related to glob file matching. 

  - **`max_total_size`**  - _[integer]_ - Maximum total size (in bytes) matched by glob 

  - **`max_objects_matched`**  - _[integer]_ - Maximum number of objects matched by glob 

  - **`max_objects_listed`**  - _[integer]_ - Maximum number of objects listed in glob 

  - **`page_size`**  - _[integer]_ - Page size for glob listing 

**`batch_size`**  - _[string]_ - Size of a batch (e.g., '100MB') 

**`google_application_credentials`**  - _[string]_ - Google Cloud credentials JSON string 

**`allow_host_access`**  - _[boolean]_ - Allow access to host environment configuration 

**`key_id`**  - _[string]_ - Optional S3-compatible Key ID when used in compatibility mode 

**`secret`**  - _[string]_ - Optional S3-compatible Secret when used in compatibility mode 

### https



**`connector`**  - _[string]_   _(required)_

**`path`**  - _[string]_ - The full HTTPS URI to fetch data from 

**`headers`**  - _[object]_ - HTTP headers to include in the request 

### local_file



**`connector`**  - _[string]_   _(required)_

**`path`**  - _[string]_ - Path to the data source. 

**`format`**  - _[string]_ - Format of the data source (e.g., csv, json, parquet). 

**`dsn`**  - _[string]_ - Data Source Name (DSN) indicating the file path or location of the local file 

**`allow_host_access`**  - _[boolean]_ - Flag to indicate if access to host-level file paths is permitted 

### motherduck



**`connector`**  - _[string]_   _(required)_

**`dsn`**  - _[string]_ - Data Source Name (DSN) specifying the MotherDuck connection endpoint 

**`token`**  - _[string]_ - Authentication token for accessing MotherDuck (secret) 

### mysql



**`connector`**  - _[string]_   _(required)_

**`dsn`**  - _[string]_ - DSN(Data Source Name) for the mysql connection 

### pinot



**`connector`**  - _[string]_   _(required)_

**`dsn`**  - _[string]_ - DSN(Data Source Name) for the Pinot connection 

**`username`**  - _[string]_ - Username for authenticating with Pinot 

**`password`**  - _[string]_ - Password for authenticating with Pinot 

**`broker_host`**  - _[string]_ - Hostname of the Pinot broker 

**`broker_port`**  - _[integer]_ - Port number for the Pinot broker 

**`controller_host`**  - _[string]_ - Hostname of the Pinot controller 

**`controller_port`**  - _[integer]_ - Port number for the Pinot controller 

**`ssl`**  - _[boolean]_ - Enable SSL connection to Pinot 

**`log_queries`**  - _[boolean]_ - Log raw SQL queries executed through Pinot 

**`max_open_conns`**  - _[integer]_ - Maximum number of open connections to the Pinot database 

### postgres



**`connector`**  - _[string]_   _(required)_

**`dsn`**  - _[string]_ - DSN(Data Source Name) for the postgres connection 

### redshift



**`connector`**  - _[string]_   _(required)_

**`output_location`**  - _[string]_ - S3 location where query results are stored. 

**`workgroup`**  - _[string]_ - Redshift Serverless workgroup to use. 

**`database`**  - _[string]_ - Name of the Redshift database. 

**`cluster_identifier`**  - _[string]_ - Identifier of the Redshift cluster. 

**`role_arn`**  - _[string]_ - ARN of the IAM role to assume for Redshift access. 

**`region`**  - _[string]_ - AWS region of the Redshift deployment. 

**`aws_access_key_id`**  - _[string]_ - AWS access key ID for authentication 

**`aws_secret_access_key`**  - _[string]_ - AWS secret access key for authentication 

**`aws_access_token`**  - _[string]_ - AWS session token for temporary credentials (optional) 

**`allow_host_access`**  - _[boolean]_ - Allow access to host environment configuration 

### s3



**`connector`**  - _[string]_   _(required)_

**`region`**  - _[string]_ - AWS region 

**`endpoint`**  - _[string]_ - AWS Endpoint 

**`path`**  - _[string]_ - Path to the source 

**`uri`**  - _[string]_ - Source URI 

**`extract`**  - _[object]_ - key-value pairs for extraction settings 

**`glob`**  - _[object]_ - Settings related to glob file matching. 

  - **`max_total_size`**  - _[integer]_ - Maximum total size (in bytes) matched by glob 

  - **`max_objects_matched`**  - _[integer]_ - Maximum number of objects matched by glob 

  - **`max_objects_listed`**  - _[integer]_ - Maximum number of objects listed in glob 

  - **`page_size`**  - _[integer]_ - Page size for glob listing 

**`batch_size`**  - _[string]_ - Size of a batch (e.g., '100MB') 

**`aws_access_key_id`**  - _[string]_ - AWS Access Key ID used for authentication 

**`aws_secret_access_key`**  - _[string]_ - AWS Secret Access Key used for authentication 

**`aws_access_token`**  - _[string]_ - Optional AWS session token for temporary credentials 

**`endpoint`**  - _[string]_ - Optional custom endpoint URL for S3-compatible storage 

**`region`**  - _[string]_ - AWS region of the S3 bucket 

**`allow_host_access`**  - _[boolean]_ - Allow access to host environment configuration 

**`retain_files`**  - _[boolean]_ - Whether to retain intermediate files after processing 

### salesforce



**`connector`**  - _[string]_   _(required)_

**`soql`**  - _[string]_ - SOQL query to execute against the Salesforce instance. 

**`sobject`**  - _[string]_ - Salesforce object (e.g., Account, Contact) targeted by the query. 

**`queryAll`**  - _[boolean]_ - Whether to include deleted and archived records in the query (uses queryAll API). 

**`username`**  - _[string]_ - Salesforce account username 

**`password`**  - _[string]_ - Salesforce account password (secret) 

**`key`**  - _[string]_ - Authentication key for Salesforce (secret) 

**`endpoint`**  - _[string]_ - Salesforce API endpoint URL 

**`client_id`**  - _[string]_ - Client ID used for Salesforce OAuth authentication 

### snowflake



**`connector`**  - _[string]_   _(required)_

**`dsn`**  - _[string]_ - DSN (Data Source Name) for the Snowflake connection 

**`parallel_fetch_limit`**  - _[integer]_ - Maximum number of concurrent fetches during query execution 

### sqlite



**`connector`**  - _[string]_   _(required)_

**`dsn`**  - _[string]_ - DSN(Data Source Name) for the sqlite connection 