---
note: GENERATED. DO NOT EDIT.
title: Connector YAML
sidebar_position: 5
---



## Properties


**`type`**  - _[string]_ - Refers to the resource type and must be `connector`  _(required)_

**`name`**  - _[string]_ - Name is usually inferred from the filename, but can be specified manually. 

**`namespace`**  - _[string]_ - Optional value to group resources by. Prepended to the resource name as `<namespace>/<name>`. 

**`refs`**  - _[array]_ - List of resource references, each as a string or map. 

     *option 1* - _[string]_ - A string reference like 'resource-name' or 'Kind/resource-name'.

     *option 2* - _[object]_ - An object reference with at least a 'name' and 'type'.

    - **`name`**  - _[string]_ -   _(required)_

    - **`type`**  - _[string]_ -  

**`version`**  - _[integer]_ - Version of the parser to use for this file. Enables backwards compatibility for breaking changes. 

 *option 1* - _[object]_ - 

**`driver`**  - _[string]_ -   _(required)_

**`allow_host_access`**  - _[boolean]_ - Allow access to host environment configuration 

**`aws_access_key_id`**  - _[string]_ - AWS Access Key ID for Athena access 

**`aws_access_token`**  - _[string]_ - Optional AWS session token for temporary credentials 

**`aws_secret_access_key`**  - _[string]_ - AWS Secret Access Key for Athena access 

**`external_id`**  - _[string]_ - Optional External ID for assuming a role 

**`managed`**  - _[any of]_ - Boolean or map of properties 

**`role_arn`**  - _[string]_ - Optional AWS Role ARN to assume when accessing Athena 

**`role_session_name`**  - _[string]_ - Optional Session name when assuming the role 

 *option 2* - _[object]_ - 

**`driver`**  - _[string]_ -   _(required)_

**`azure_storage_sas_token`**  - _[string]_ - Optional azure SAS token for authentication 

**`managed`**  - _[any of]_ - Boolean or map of properties 

**`allow_host_access`**  - _[boolean]_ - Allow access to host environment configuration 

**`azure_storage_account`**  - _[string]_ - Azure storage account name 

**`azure_storage_connection_string`**  - _[string]_ - Optional azure connection string for storage account 

**`azure_storage_key`**  - _[string]_ - Azure storage access key 

 *option 3* - _[object]_ - 

**`driver`**  - _[string]_ -   _(required)_

**`google_application_credentials`**  - _[string]_ - Path to the Google Cloud credentials JSON file 

**`managed`**  - _[any of]_ - Boolean or map of properties 

**`allow_host_access`**  - _[boolean]_ - Allow access to host environment configuration 

 *option 4* - _[object]_ - 

**`driver`**  - _[string]_ -   _(required)_

**`conn_max_lifetime`**  - _[string]_ - Maximum time a connection may be reused 

**`dial_timeout`**  - _[string]_ - Timeout for dialing the ClickHouse server 

**`max_idle_conns`**  - _[integer]_ - Maximum number of idle connections in the pool 

**`max_open_conns`**  - _[integer]_ - Maximum number of open connections to the database 

**`cluster`**  - _[string]_ - Cluster name, required for running distributed queries 

**`port`**  - _[integer]_ - Port where the ClickHouse instance is accessible 

**`read_timeout`**  - _[string]_ - Maximum time for a connection to read data 

**`settings_override`**  - _[string]_ - override the default settings used in queries. example `readonly = 1, session_timezone = 'UTC'` 

**`can_scale_to_zero`**  - _[boolean]_ - Indicates if the database can scale to zero 

**`database`**  - _[string]_ - Name of the ClickHouse database within the cluster 

**`log_queries`**  - _[boolean]_ - Controls whether to log raw SQL queries 

**`password`**  - _[string]_ - Password for authentication 

**`dsn`**  - _[string]_ - DSN(Data Source Name) for the ClickHouse connection 

**`embed_port`**  - _[integer]_ - Port to run ClickHouse locally (0 for random port) 

**`host`**  - _[string]_ - Host where the ClickHouse instance is running 

**`managed`**  - _[any of]_ - Boolean or map of properties 

**`ssl`**  - _[boolean]_ - Indicates whether a secured SSL connection is required 

**`username`**  - _[string]_ - Username for authentication 

 *option 5* - _[object]_ - 

**`driver`**  - _[string]_ -   _(required)_

**`host`**  - _[string]_ - Hostname of the Druid coordinator or broker 

**`log_queries`**  - _[boolean]_ - Log raw SQL queries sent to Druid 

**`managed`**  - _[any of]_ - Boolean or map of properties 

**`password`**  - _[string]_ - Password for authenticating with Druid 

**`port`**  - _[integer]_ - Port number of the Druid service 

**`dsn`**  - _[string]_ - Data Source Name (DSN) for connecting to Druid 

**`max_open_conns`**  - _[integer]_ - Maximum number of open database connections (0 = default, -1 = unlimited) 

**`skip_version_check`**  - _[boolean]_ - Skip checking Druid version compatibility 

**`ssl`**  - _[boolean]_ - Enable SSL for secure connection 

**`username`**  - _[string]_ - Username for authenticating with Druid 

 *option 6* - _[object]_ - 

**`driver`**  - _[string]_ -   _(required)_

**`read_write_ratio`**  - _[number]_ - Ratio of resources allocated to the read database; used to divide CPU and memory 

**`allow_host_access`**  - _[boolean]_ - Whether access to the local environment and file system is allowed 

**`boot_queries`**  - _[string]_ - SQL to run when initializing a new connection, before extensions and defaults 

**`cpu`**  - _[integer]_ - Number of CPU cores available to the database 

**`init_sql`**  - _[string]_ - SQL to run when initializing a new connection, after extensions and defaults 

**`memory_limit_gb`**  - _[integer]_ - Amount of memory in GB available to the database 

**`log_queries`**  - _[boolean]_ - Whether to log raw SQL queries executed through OLAP 

**`managed`**  - _[any of]_ - Boolean or map of properties 

**`path`**  - _[string]_ - File path to the DuckDB database file 

**`pool_size`**  - _[integer]_ - Number of concurrent connections and queries allowed 

 *option 7* - _[object]_ - 

**`driver`**  - _[string]_ -   _(required)_

**`managed`**  - _[any of]_ - Boolean or map of properties 

**`secret`**  - _[string]_ - Optional S3-compatible Secret when used in compatibility mode 

**`allow_host_access`**  - _[boolean]_ - Allow access to host environment configuration 

**`google_application_credentials`**  - _[string]_ - Google Cloud credentials JSON string 

**`key_id`**  - _[string]_ - Optional S3-compatible Key ID when used in compatibility mode 

 *option 8* - _[object]_ - 

**`driver`**  - _[string]_ -   _(required)_

**`managed`**  - _[any of]_ - Boolean or map of properties 

**`path`**  - _[string]_ - The full HTTPS URI to fetch data from 

**`headers`**  - _[object]_ - HTTP headers to include in the request 

 *option 9* - _[object]_ - 

**`driver`**  - _[string]_ -   _(required)_

**`dsn`**  - _[string]_ - Data Source Name (DSN) indicating the file path or location of the local file 

**`managed`**  - _[any of]_ - Boolean or map of properties 

**`allow_host_access`**  - _[boolean]_ - Flag to indicate if access to host-level file paths is permitted 

 *option 10* - _[object]_ - 

**`driver`**  - _[string]_ -   _(required)_

**`dsn`**  - _[string]_ - Data Source Name (DSN) specifying the MotherDuck connection endpoint 

**`managed`**  - _[any of]_ - Boolean or map of properties 

**`token`**  - _[string]_ - Authentication token for accessing MotherDuck (secret) 

 *option 11* - _[object]_ - 

**`driver`**  - _[string]_ -   _(required)_

**`dsn`**  - _[string]_ - DSN(Data Source Name) for the mysql connection 

**`managed`**  - _[any of]_ - Boolean or map of properties 

 *option 12* - _[object]_ - 

**`driver`**  - _[string]_ -   _(required)_

**`password`**  - _[string]_ - Password for authenticating with Pinot 

**`broker_host`**  - _[string]_ - Hostname of the Pinot broker 

**`controller_host`**  - _[string]_ - Hostname of the Pinot controller 

**`log_queries`**  - _[boolean]_ - Log raw SQL queries executed through Pinot 

**`managed`**  - _[any of]_ - Boolean or map of properties 

**`max_open_conns`**  - _[integer]_ - Maximum number of open connections to the Pinot database 

**`ssl`**  - _[boolean]_ - Enable SSL connection to Pinot 

**`username`**  - _[string]_ - Username for authenticating with Pinot 

**`broker_port`**  - _[integer]_ - Port number for the Pinot broker 

**`controller_port`**  - _[integer]_ - Port number for the Pinot controller 

**`dsn`**  - _[string]_ - DSN(Data Source Name) for the Pinot connection 

 *option 13* - _[object]_ - 

**`driver`**  - _[string]_ -   _(required)_

**`managed`**  - _[any of]_ - Boolean or map of properties 

**`dsn`**  - _[string]_ - DSN(Data Source Name) for the postgres connection 

 *option 14* - _[object]_ - 

**`driver`**  - _[string]_ -   _(required)_

**`managed`**  - _[any of]_ - Boolean or map of properties 

**`allow_host_access`**  - _[boolean]_ - Allow access to host environment configuration 

**`aws_access_key_id`**  - _[string]_ - AWS access key ID for authentication 

**`aws_access_token`**  - _[string]_ - AWS session token for temporary credentials (optional) 

**`aws_secret_access_key`**  - _[string]_ - AWS secret access key for authentication 

 *option 15* - _[object]_ - 

**`driver`**  - _[string]_ -   _(required)_

**`aws_access_key_id`**  - _[string]_ - AWS Access Key ID used for authentication 

**`aws_access_token`**  - _[string]_ - Optional AWS session token for temporary credentials 

**`aws_secret_access_key`**  - _[string]_ - AWS Secret Access Key used for authentication 

**`endpoint`**  - _[string]_ - Optional custom endpoint URL for S3-compatible storage 

**`managed`**  - _[any of]_ - Boolean or map of properties 

**`region`**  - _[string]_ - AWS region of the S3 bucket 

**`retain_files`**  - _[boolean]_ - Whether to retain intermediate files after processing 

**`allow_host_access`**  - _[boolean]_ - Allow access to host environment configuration 

 *option 16* - _[object]_ - 

**`driver`**  - _[string]_ -   _(required)_

**`password`**  - _[string]_ - Salesforce account password (secret) 

**`username`**  - _[string]_ - Salesforce account username 

**`client_id`**  - _[string]_ - Client ID used for Salesforce OAuth authentication 

**`endpoint`**  - _[string]_ - Salesforce API endpoint URL 

**`key`**  - _[string]_ - Authentication key for Salesforce (secret) 

**`managed`**  - _[any of]_ - Boolean or map of properties 

 *option 17* - _[object]_ - 

**`driver`**  - _[string]_ -   _(required)_

**`bot_token`**  - _[string]_ - Bot token used for authenticating Slack API requests 

**`managed`**  - _[any of]_ - Boolean or map of properties 

 *option 18* - _[object]_ - 

**`driver`**  - _[string]_ -   _(required)_

**`dsn`**  - _[string]_ - DSN (Data Source Name) for the Snowflake connection 

**`managed`**  - _[any of]_ - Boolean or map of properties 

**`parallel_fetch_limit`**  - _[integer]_ - Maximum number of concurrent fetches during query execution 

 *option 19* - _[object]_ - 

**`driver`**  - _[string]_ -   _(required)_

**`dsn`**  - _[string]_ - DSN(Data Source Name) for the sqlite connection 

**`managed`**  - _[any of]_ - Boolean or map of properties 