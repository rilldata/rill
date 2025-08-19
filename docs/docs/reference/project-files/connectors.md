---
note: GENERATED. DO NOT EDIT.
title: Connector YAML
sidebar_position: 31
---

Connector YAML files define how Rill connects to external data sources and OLAP engines. Each connector specifies a driver type and its required connection parameters.

## Available Connector Types

### _OLAP Engines_
- [**DuckDB**](#duckdb) - Embedded DuckDB engine (default)
- [**ClickHouse**](#clickhouse) - ClickHouse analytical database
- [**MotherDuck**](#motherduck) - MotherDuck cloud database
- [**Druid**](#druid) - Apache Druid
- [**Pinot**](#pinot) - Apache Pinot

### _Data Warehouses_
- [**Snowflake**](#snowflake) - Snowflake data warehouse
- [**BigQuery**](#bigquery) - Google BigQuery
- [**Redshift**](#redshift) - Amazon Redshift
- [**Athena**](#athena) - Amazon Athena

### _Databases_
- [**PostgreSQL**](#postgres) - PostgreSQL databases
- [**MySQL**](#mysql) - MySQL databases
- [**SQLite**](#sqlite) - SQLite databases

### _Cloud Storage_
- [**GCS**](#gcs) - Google Cloud Storage
- [**S3**](#s3) - Amazon S3 storage
- [**Azure**](#azure) - Azure Blob Storage

### _Other_
- [**HTTPS**](#https) - Public files via HTTP/HTTPS
- [**Salesforce**](#salesforce) - Salesforce data
- [**Slack**](#slack) - Slack data
- [**OpenAPI**](#openapi) - OpenAPI data

:::warning Security Recommendation
For all credential parameters (passwords, tokens, keys), use environment variables with the syntax `{{.env.connector.<connector_driver>.<parameter_name>}}`. This keeps sensitive data out of your YAML files and version control. See our [credentials documentation](/connect/credentials/) for complete setup instructions.
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

## Available Connector Types



### Athena




```yaml
type: connector                                  # Must be `connector` (required)
driver: athena                                   # Must be `athena` _(required)_

aws_access_key_id: AKIAIOSFODNN7EXAMPLE# AWS Access Key ID used for authentication. Required when using static credentials directly or as base credentials for assuming a role.
aws_secret_access_key: wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY# AWS Secret Access Key paired with the Access Key ID. Required when using static credentials directly or as base credentials for assuming a role.
aws_access_token: AKIAIOSFODNN7EXAMPLE# AWS session token used with temporary credentials. Required only if the Access Key and Secret Key are part of a temporary session credentials.
role_arn: arn:aws:iam::123456789012:role/MyRole# ARN of the IAM role to assume. When specified, the SDK uses the base credentials to call STS AssumeRole and obtain temporary credentials scoped to this role.
role_session_name: MySession         # Session name to associate with the STS AssumeRole session. Used only if 'role_arn' is specified. Useful for identifying and auditing the session.
external_id: MyExternalID            # External ID required by some roles when assuming them, typically for cross-account access. Used only if 'role_arn' is specified and the role's trust policy requires it.
workgroup: primary                   # Athena workgroup to use for query execution. Defaults to 'primary' if not specified.
output_location: s3://my-bucket/athena-output/# S3 URI where Athena query results should be stored (e.g., s3://your-bucket/athena/results/). Optional if the selected workgroup has a default result configuration.
aws_region: us-east-1                # AWS region where Athena and the result S3 bucket are located (e.g., us-east-1). Defaults to 'us-east-1' if not specified.
allow_host_access: true              # Allow the Athena client to access host environment configurations such as environment variables or local AWS credential files. Defaults to true, enabling use of credentials and settings from the host environment unless explicitly disabled.
```



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

### Azure




```yaml
type: connector                                  # Must be `connector` (required)
driver: azure                                   # Must be `azure` _(required)_

azure_storage_account: mystorageaccount# Azure storage account name
azure_storage_key: myaccesskey       # Azure storage access key
azure_storage_bucket: my-container   # Name of the Azure Blob Storage container (equivalent to an S3 bucket) _(required)_
azure_storage_sas_token: my-sas-token# Optional azure SAS token for authentication
azure_storage_connection_string: DefaultEndpointsProtocol=https;AccountName=mystorageaccount;AccountKey=myaccesskey;EndpointSuffix=core.windows.net# Optional azure connection string for storage account
allow_host_access: true              # Allow access to host environment configuratio
```



### `driver`

_[string]_ - Refers to the driver type and must be driver `azure` _(required)_

### `azure_storage_account`

_[string]_ - Azure storage account name 

### `azure_storage_key`

_[string]_ - Azure storage access key 

### `azure_storage_bucket`

_[string]_ - Name of the Azure Blob Storage container (equivalent to an S3 bucket) _(required)_

### `azure_storage_sas_token`

_[string]_ - Optional azure SAS token for authentication 

### `azure_storage_connection_string`

_[string]_ - Optional azure connection string for storage account 

### `allow_host_access`

_[boolean]_ - Allow access to host environment configuratio 

### BigQuery




```yaml
type: connector                                  # Must be `connector` (required)
driver: bigquery                                   # Must be `bigquery` _(required)_

google_application_credentials: {"type": "service_account", "project_id": "my-gcp-project"}# Raw contents of the Google Cloud service account key (in JSON format) used for authentication.
project_id: my-gcp-project           # Google Cloud project ID
dataset_id: my_dataset               # BigQuery dataset ID
location: US                         # BigQuery dataset location
allow_host_access: true              # Enable the BigQuery client to use credentials from the host environment when no service account JSON is provided. This includes Application Default Credentials from environment variables, local credential files, or the Google Compute Engine metadata server. Defaults to true, allowing seamless authentication in GCP environments.
```



### `driver`

_[string]_ - Refers to the driver type and must be driver `bigquery` _(required)_

### `google_application_credentials`

_[string]_ - Raw contents of the Google Cloud service account key (in JSON format) used for authentication. 

### `project_id`

_[string]_ - Google Cloud project ID 

### `dataset_id`

_[string]_ - BigQuery dataset ID 

### `location`

_[string]_ - BigQuery dataset location 

### `allow_host_access`

_[boolean]_ - Enable the BigQuery client to use credentials from the host environment when no service account JSON is provided. This includes Application Default Credentials from environment variables, local credential files, or the Google Compute Engine metadata server. Defaults to true, allowing seamless authentication in GCP environments. 

### ClickHouse




```yaml
type: connector                                  # Must be `connector` (required)
driver: clickhouse                                   # Must be `clickhouse` _(required)_

managed: true                        # `true` means Rill will provision the connector using the default provisioner. `false` disables automatic provisioning.
mode: readwrite                      # `read` - Controls the operation mode for the ClickHouse connection. Defaults to 'read' for safe operation with external databases. Set to 'readwrite' to enable model creation and table mutations. Note: When 'managed: true', this is automatically set to 'readwrite'.
dsn: clickhouse://localhost:9000/default# DSN(Data Source Name) for the ClickHouse connection
username: default                    # Username for authentication
password: mypassword                 # Password for authentication
host: localhost                      # Host where the ClickHouse instance is running
port: 9000                           # Port where the ClickHouse instance is accessible
database: default                    # Name of the ClickHouse database within the cluster
ssl: true                            # Indicates whether a secured SSL connection is required
cluster: my-cluster                  # Cluster name, required for running distributed queries
log_queries: true                    # Controls whether to log raw SQL queries
settings_override: readonly = 1, session_timezone = 'UTC'# override the default settings used in queries. example `readonly = 1, session_timezone = 'UTC'`
embed_port: 0                        # Port to run ClickHouse locally (0 for random port)
can_scale_to_zero: true              # Indicates if the database can scale to zero
max_open_conns: 10                   # Maximum number of open connections to the database
max_idle_conns: 10                   # Maximum number of idle connections in the pool
dial_timeout: 10s                    # Timeout for dialing the ClickHouse server
conn_max_lifetime: 10s               # Maximum time a connection may be reused
read_timeout: 10s                    # Maximum time for a connection to read data
```



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

### Druid




```yaml
type: connector                                  # Must be `connector` (required)
driver: druid                                   # Must be `druid` _(required)_

dsn: http://localhost:8082           # Data Source Name (DSN) for connecting to Druid _(required)_
username: admin                      # Username for authenticating with Druid
password: admin123                   # Password for authenticating with Druid
host: localhost                      # Hostname of the Druid coordinator or broker
port: 8082                           # Port number of the Druid service
ssl: true                            # Enable SSL for secure connection
log_queries: true                    # Log raw SQL queries sent to Druid
max_open_conns: 10                   # Maximum number of open database connections (0 = default, -1 = unlimited)
skip_version_check: true             # Skip checking Druid version compatibility
```



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

### DuckDB




```yaml
type: connector                                  # Must be `connector` (required)
driver: duckdb                                   # Must be `duckdb` _(required)_

pool_size: 10                        # Number of concurrent connections and queries allowed
allow_host_access: true              # Whether access to the local environment and file system is allowed
cpu: 10                              # Number of CPU cores available to the database
memory_limit_gb: 10                  # Amount of memory in GB available to the database
read_write_ratio: 0.5                # Ratio of resources allocated to the read database; used to divide CPU and memory
init_sql: CREATE TABLE IF NOT EXISTS users (id INTEGER PRIMARY KEY, name TEXT, email TEXT)# is executed during database initialization.
secrets: gcs,s3                      # Comma-separated list of other connector names to create temporary secrets for in DuckDB before executing a model.
log_queries: true                    # Whether to log raw SQL queries executed through OLAP
```



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

### `secrets`

_[string]_ - Comma-separated list of other connector names to create temporary secrets for in DuckDB before executing a model. 

### `log_queries`

_[boolean]_ - Whether to log raw SQL queries executed through OLAP 

### GCS




```yaml
type: connector                                  # Must be `connector` (required)
driver: gcs                                   # Must be `gcs` _(required)_

google_application_credentials: {"type": "service_account", "project_id": "my-project"}# Google Cloud credentials JSON string
bucket: my-gcs-bucket                # Name of gcs bucket _(required)_
allow_host_access: true              # Allow access to host environment configuration
key_id: AKIAIOSFODNN7EXAMPLE         # Optional S3-compatible Key ID when used in compatibility mode
secret: wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY# Optional S3-compatible Secret when used in compatibility mode
```



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

### HTTPS




```yaml
type: connector                                  # Must be `connector` (required)
driver: https                                   # Must be `https` _(required)_

path: https://api.example.com/data.csv# The full HTTPS URI to fetch data from _(required)_
headers: {"Authorization": "Bearer my-token"}# HTTP headers to include in the request
```



### `driver`

_[string]_ - Refers to the driver type and must be driver `https` _(required)_

### `path`

_[string]_ - The full HTTPS URI to fetch data from _(required)_

### `headers`

_[object]_ - HTTP headers to include in the request 

### MotherDuck




```yaml
type: connector                                  # Must be `connector` (required)
driver: duckdb                                   # Must be `duckdb` _(required)_

path: md:my_database                 # Path to your MD database _(required)_
schema_name: main                    # Define your schema if not main, uses main by default
init_sql: INSTALL 'motherduck';
LOAD 'motherduck';
SET motherduck_token= '{{ .env.motherduck_token }}'# SQL executed during database initialization. _(required)_
```



### `driver`

_[string]_ - Refers to the driver type and must be driver `duckdb` _(required)_

### `path`

_[string]_ - Path to your MD database _(required)_

### `schema_name`

_[string]_ - Define your schema if not main, uses main by default 

### `init_sql`

_[string]_ - SQL executed during database initialization. _(required)_

### MySQL




```yaml
type: connector                                  # Must be `connector` (required)
driver: mysql                                   # Must be `mysql` _(required)_

dsn: mysql://user:password@localhost:3306/mydatabase# DSN(Data Source Name) for the mysql connection
host: localhost                      # Hostname of the MySQL server
port: 3306                           # Port number for the MySQL server
database: mydatabase                 # Name of the MySQL database
user: myuser                         # Username for authentication
password: mypassword                 # Password for authentication
ssl_mode: PREFERRED                  # SSL mode can be DISABLED, PREFERRED or REQUIRED
```



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

### Pinot




```yaml
type: connector                                  # Must be `connector` (required)
driver: pinot                                   # Must be `pinot` _(required)_

dsn: pinot://localhost:8099          # DSN(Data Source Name) for the Pinot connection _(required)_
username: admin                      # Username for authenticating with Pinot
password: admin123                   # Password for authenticating with Pinot
broker_host: localhost               # Hostname of the Pinot broker _(required)_
broker_port: 8099                    # Port number for the Pinot broker
controller_host: localhost           # Hostname of the Pinot controller _(required)_
controller_port: 9000                # Port number for the Pinot controller
ssl: true                            # Enable SSL connection to Pinot
log_queries: true                    # Log raw SQL queries executed through Pinot
max_open_conns: 10                   # Maximum number of open connections to the Pinot database
```



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

### Postgres




```yaml
type: connector                                  # Must be `connector` (required)
driver: postgres                                   # Must be `postgres` _(required)_

dsn: postgresql://user:password@localhost:5432/mydatabase# DSN(Data Source Name) for the postgres connection
host: localhost                      # Hostname of the Postgres server
port: 5432                           # Port number for the Postgres server
dbname: mydatabase                   # Name of the Postgres database
user: postgres                       # Username for authentication
password: mypassword                 # Password for authentication
sslmode: prefer                      # SSL mode can be disable, allow, prefer or require
```



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

### Redshift




```yaml
type: connector                                  # Must be `connector` (required)
driver: redshift                                   # Must be `redshift` _(required)_

aws_access_key_id: AKIAIOSFODNN7EXAMPLE# AWS Access Key ID used for authenticating with Redshift. _(required)_
aws_secret_access_key: wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY# AWS Secret Access Key used for authenticating with Redshift. _(required)_
aws_access_token: AKIAIOSFODNN7EXAMPLE# AWS Session Token for temporary credentials (optional).
region: us-east-1                    # AWS region where the Redshift cluster or workgroup is hosted (e.g., 'us-east-1').
database: myredshiftdb               # Name of the Redshift database to query. _(required)_
workgroup: my-workgroup              # Workgroup name for Redshift Serverless, in case of provisioned Redshift clusters use 'cluster_identifier'.
cluster_identifier: my-cluster       # Cluster identifier for provisioned Redshift clusters, in case of Redshift Serverless use 'workgroup' .
```



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

### S3




```yaml
type: connector                                  # Must be `connector` (required)
driver: s3                                   # Must be `s3` _(required)_

aws_access_key_id: AKIAIOSFODNN7EXAMPLE# AWS Access Key ID used for authentication
aws_secret_access_key: wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY# AWS Secret Access Key used for authentication
aws_access_token: AKIAIOSFODNN7EXAMPLE# Optional AWS session token for temporary credentials
bucket: my-s3-bucket                 # Name of s3 bucket _(required)_
endpoint: https://s3.amazonaws.com   # Optional custom endpoint URL for S3-compatible storage
region: us-east-1                    # AWS region of the S3 bucket
allow_host_access: true              # Allow access to host environment configuration
retain_files: true                   # Whether to retain intermediate files after processing
```



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

### Salesforce




```yaml
type: connector                                  # Must be `connector` (required)
driver: salesforce                                   # Must be `salesforce` _(required)_

username: user@example.com           # Salesforce account username _(required)_
password: mypassword                 # Salesforce account password (secret)
key: mysecretkey                     # Authentication key for Salesforce (secret)
endpoint: https://login.salesforce.com# Salesforce API endpoint URL _(required)_
client_id: myclientid                # Client ID used for Salesforce OAuth authentication
```



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

### Slack




```yaml
type: connector                                  # Must be `connector` (required)
driver: slack                                   # Must be `slack` _(required)_

bot_token: xoxb-your-bot-token       # Bot token used for authenticating Slack API requests _(required)_
```



### `driver`

_[string]_ - Refers to the driver type and must be driver `slack` _(required)_

### `bot_token`

_[string]_ - Bot token used for authenticating Slack API requests _(required)_

### Snowflake




```yaml
type: connector                                  # Must be `connector` (required)
driver: snowflake                                   # Must be `snowflake` _(required)_

dsn: user:password@account/database/schema?warehouse=warehouse# DSN (Data Source Name) for the Snowflake connection _(required)_
parallel_fetch_limit: 10             # Maximum number of concurrent fetches during query execution
```



### `driver`

_[string]_ - Refers to the driver type and must be driver `snowflake` _(required)_

### `dsn`

_[string]_ - DSN (Data Source Name) for the Snowflake connection _(required)_

### `parallel_fetch_limit`

_[integer]_ - Maximum number of concurrent fetches during query execution 

### SQLite




```yaml
type: connector                                  # Must be `connector` (required)
driver: sqlite                                   # Must be `sqlite` _(required)_

dsn: file:./mydatabase.db            # DSN(Data Source Name) for the sqlite connection _(required)_
```



### `driver`

_[string]_ - Refers to the driver type and must be driver `sqlite` _(required)_

### `dsn`

_[string]_ - DSN(Data Source Name) for the sqlite connection _(required)_

### OpenAPI




```yaml
type: connector                                  # Must be `connector` (required)
driver: openapi                                   # Must be `openapi` _(required)_

api_key: sk-1234567890               # API key for connecting to OpenAI _(required)_
model: gpt-4o                        # The OpenAI model to use (e.g., 'gpt-4o')
base_url: https://api.openai.com/v1  # The base URL for the OpenAI API (e.g., 'https://api.openai.com/v1')
api_type: OPEN_AI                    # The type of OpenAI API to use
api_version: 2023-05-15              # The version of the OpenAI API to use (e.g., '2023-05-15'). Required when API Type is AZURE or AZURE_AD
```



### `driver`

_[string]_ - The driver type, must be set to "openapi" 

### `api_key`

_[string]_ - API key for connecting to OpenAI _(required)_

### `model`

_[string]_ - The OpenAI model to use (e.g., 'gpt-4o') 

### `base_url`

_[string]_ - The base URL for the OpenAI API (e.g., 'https://api.openai.com/v1') 

### `api_type`

_[string]_ - The type of OpenAI API to use 

### `api_version`

_[string]_ - The version of the OpenAI API to use (e.g., '2023-05-15'). Required when API Type is AZURE or AZURE_AD 