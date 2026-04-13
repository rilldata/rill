---
name: rill-connector
description: Detailed instructions and examples for developing connector resources in Rill
---

# Instructions for developing a connector in Rill

## Introduction

Connectors are resources that contain credentials and settings for connecting to external systems. They are typically found at the root of the project's DAG, providing access to data sources and services that power downstream resources like models and metrics views.

Connectors are usually lightweight resources. When reconciled, they validate the connection to the external system but do not move or process data. The main exception is managed OLAP connectors (with `managed: true`), which trigger database provisioning.

### Driver capabilities

Each connector uses a **driver** that implements one or more capabilities:

- **OLAP database**: Can power metrics views and dashboards (e.g., `duckdb`, `clickhouse`)
- **SQL database**: Can run SQL queries as model inputs (e.g., `postgres`, `bigquery`, `snowflake`)
- **Information schema**: Can list tables and their schemas (e.g., `duckdb`, `bigquery`)
- **Object store**: Can list, read, and write flat files (e.g., `s3`, `gcs`)
- **Notifier**: Can send notifications and alerts (e.g., `slack`)
- **AI**: Can generate embeddings or responses (e.g., `openai`)

## Core Concepts

### Naming conventions

Connectors are typically named after their driver (e.g., a file `connectors/duckdb.yaml` creates a connector named `duckdb`). Use descriptive names when you have multiple connectors of the same type:
- `connectors/s3_data.yaml` and `connectors/s3_staging.yaml`
- `connectors/clickhouse_prod.yaml` and `connectors/clickhouse_dev.yaml`

### Secrets management

Always store sensitive credentials in `.env` and reference them using template syntax:

```yaml
type: connector
driver: s3
aws_access_key_id: "{{ .env.aws_access_key_id }}"
aws_secret_access_key: "{{ .env.aws_secret_access_key }}"
```

NOTE: Some legacy projects use the deprecated `.vars` instead of `.env`.

### Managed connectors

OLAP connectors can be provisioned automatically by Rill using `managed: true`. This is supported for `duckdb` and `clickhouse` drivers:

```yaml
type: connector
driver: duckdb
managed: true
```

When a managed connector is reconciled, Rill provisions the database infrastructure. The user is billed for the CPU, memory, and disk usage of the provisioned database.

### Access modes

Control read/write access using the `mode` property:

- `mode: read`: Prevents Rill models from writing to this connector
- `mode: readwrite`: Allows both reading and writing (default for managed connectors)

Use `mode: read` when connecting to external databases with pre-existing tables to prevent unintended modifications.

### Dev/Prod configuration

Use `dev:` and `prod:` blocks for environment-specific settings:

```yaml
type: connector
driver: clickhouse
mode: readwrite

# Use a local database in development to prevent overwriting data in the production cluster when iterating on changes.
dev:
  managed: true

prod:
  host: "{{ .env.clickhouse_host }}"
  port: 9440
  username: "{{ .env.clickhouse_user }}"
  password: "{{ .env.clickhouse_password }}"
  ssl: true
```

## Driver-Specific Notes

### DuckDB

DuckDB is Rill's default embedded OLAP database. Key properties:

- `managed: true`: Rill provisions and manages the database
- `init_sql`: SQL to run at startup (install extensions, create secrets, attach databases)

For MotherDuck (cloud DuckDB), use the `path` property with `md:` prefix:

```yaml
type: connector
driver: duckdb
path: "md:my_database"
token: "{{ .env.motherduck_token }}"
```

### ClickHouse

ClickHouse can be user-managed or Rill-managed. Key properties:

- `managed: true`: Rill provisions and manages an empty Clickhouse cluster. If set, don't set any other connector properties.
- `host`, `port`, `username`, `password`: Connection credentials
- `database`: Target database name
- `ssl: true`: Required for ClickHouse Cloud
- `cluster`: Cluster name for multi-node Clickhouse clusters
- `dsn`: Alternative connection string format (format: `clickhouse://host:port?username=<username>&...`)

Common ports:
- `8443`: HTTPS native protocol (ClickHouse Cloud)
- `9440`: Secure native protocol
- `9000`: Native protocol (non-SSL)

### S3

AWS S3 and S3-compatible storage. Key properties:

- `aws_access_key_id`, `aws_secret_access_key`: AWS credentials
- `region`: AWS region
- `endpoint`: Custom endpoint for S3-compatible services (R2, MinIO, GCS interop)
- `path_prefixes`: A list of bucket paths that the connector can access, e.g. `[s3://my-bucket]`; useful for improving bucket introspection

### GCS

Google Cloud Storage. Key properties:

- `google_application_credentials`: Service account JSON (must be a literal JSON string value)
- `key_id`: HMAC key ID to use instead of a service account JSON; required for direct use with DuckDB and Clickhouse through S3 compatibility
- `secret`: HMAC secret to use instead of a service account JSON; required for direct use with DuckDB and Clickhouse through S3 compatibility

### BigQuery

Google BigQuery. Key properties:

- `project_id`: GCP project ID
- `google_application_credentials`: Service account JSON

### Snowflake

Snowflake data warehouse. Key properties:

- `account`, `user`, `privateKey`, `database`, `schema`, `warehouse`, `role`: Connection parameters
- `dsn`: Connection string to use instead of separate connection parameters

### Postgres

PostgreSQL database. Key properties:

- `host`, `port`, `user`, `password`, `dbname`: Connection parameters
- `sslmode`: SSL mode (`disable`, `require`, `verify-full`, etc.)

### Druid

Apache Druid. Can be configured via host/port or DSN:

- `host`, `port`, `username`, `password`, `ssl`: Direct connection
- `dsn`: Full connection string

### Redshift

Amazon Redshift. Key properties:

- `aws_access_key_id`, `aws_secret_access_key`: AWS credentials
- `workgroup`: Redshift Serverless workgroup name
- `region`: AWS region
- `database`: Database name

### Athena

Amazon Athena. Key properties:

- `aws_access_key_id`, `aws_secret_access_key`: AWS credentials
- `workgroup`: Redshift Serverless workgroup name
- `region`: AWS region
- `output_location`: S3 path in format `s3://bucket/path` to store temporary query results in (Athena only)

### Other drivers

- **Slack**: Use `bot_token` for alert notifications
- **OpenAI** or **Claude**: Use `api_key` for AI-powered features
- **HTTPS**: Simple connector for public HTTP sources
- **Pinot**: Use `broker_host`, `controller_host`, `username`, `password`

## Examples

### DuckDB: Managed

Explicit:

```yaml
# connectors/duckdb.yaml
type: connector
driver: duckdb
managed: true
```

or relying on defaults:

```yaml
# connectors/duckdb.yaml
type: connector
driver: duckdb
```

### DuckDB: With init_sql for S3 secrets

```yaml
# connectors/duckdb.yaml
type: connector
driver: duckdb

init_sql: |
  CREATE SECRET IF NOT EXISTS s3 (
    TYPE S3,
    KEY_ID '{{ .env.aws_access_key_id }}',
    SECRET '{{ .env.aws_secret_access_key }}',
    REGION 'us-east-1'
  )
```

This is now deprecated in favor of creating a dedicated `s3.yaml` connector file, which Rill will automatically load and create as a secret in DuckDB.

### DuckDB: With extensions

```yaml
# connectors/duckdb.yaml
type: connector
driver: duckdb

init_sql: |
  INSTALL spatial;
  LOAD spatial;
```

### DuckDB: MotherDuck database with existing tables

```yaml
# connectors/motherduck.yaml
type: connector
driver: duckdb
path: "md:my_database"
token: "{{ .env.motherduck_token }}"
schema_name: main
mode: read
```

### ClickHouse: Cloud with SSL

```yaml
# connectors/clickhouse.yaml
type: connector
driver: clickhouse
host: "abc123.us-east-1.aws.clickhouse.cloud"
port: 8443
username: "default"
password: "{{ .env.clickhouse_password }}"
database: "default"
ssl: true
```

### ClickHouse: Readwrite with cluster

```yaml
# connectors/clickhouse.yaml
type: connector
driver: clickhouse
mode: readwrite

host: "{{ .env.clickhouse_host }}"
port: 9440
username: "{{ .env.clickhouse_user }}"
password: "{{ .env.clickhouse_password }}"
database: "default"
cluster: "my_cluster"
ssl: true
```

### ClickHouse: Dev/prod configuration

```yaml
# connectors/clickhouse.yaml
type: connector
driver: clickhouse
mode: readwrite

dev:
  managed: true

prod:
  host: "{{ .env.clickhouse_host }}"
  port: 9440
  username: "{{ .env.clickhouse_user }}"
  password: "{{ .env.clickhouse_password }}"
  database: "default"
  cluster: "{{ .env.clickhouse_cluster }}"
  ssl: true
```

### S3: Basic with credentials and region

```yaml
# connectors/s3.yaml
type: connector
driver: s3
aws_access_key_id: "{{ .env.aws_access_key_id }}"
aws_secret_access_key: "{{ .env.aws_secret_access_key }}"
region: us-west-2
```

### S3: Cloudflare R2 (S3-compatible)

```yaml
# connectors/r2.yaml
type: connector
driver: s3
aws_access_key_id: "{{ .env.r2_access_key_id }}"
aws_secret_access_key: "{{ .env.r2_secret_access_key }}"
endpoint: "https://{{ .env.r2_account_id }}.r2.cloudflarestorage.com"
region: auto
```

### GCS: Minimal (uses default credentials)

```yaml
# connectors/gcs.yaml
type: connector
driver: gcs
```

### GCS: With explicit credentials

```yaml
# connectors/gcs.yaml
type: connector
driver: gcs
google_application_credentials: "{{ .env.gcs_service_account_json }}"
```

### BigQuery

```yaml
# connectors/bigquery.yaml
type: connector
driver: bigquery
project_id: "my-gcp-project"
google_application_credentials: "{{ .env.bigquery_service_account_json }}"
```

### Snowflake: Basic with DSN

```yaml
# connectors/snowflake.yaml
type: connector
driver: snowflake
dsn: "{{ .env.snowflake_dsn }}"
```

### Postgres

```yaml
# connectors/postgres.yaml
type: connector
driver: postgres
host: "{{ .env.postgres_host }}"
port: 5432
user: "{{ .env.postgres_user }}"
password: "{{ .env.postgres_password }}"
dbname: "analytics"
sslmode: require
```

### Druid: Host-based

```yaml
# connectors/druid.yaml
type: connector
driver: druid
host: "{{ .env.druid_host }}"
port: 8888
username: "{{ .env.druid_user }}"
password: "{{ .env.druid_password }}"
ssl: true
```

### Redshift: Serverless

```yaml
# connectors/redshift.yaml
type: connector
driver: redshift
aws_access_key_id: "{{ .env.aws_access_key_id }}"
aws_secret_access_key: "{{ .env.aws_secret_access_key }}"
workgroup: "my-workgroup"
region: us-east-1
database: "analytics"
```

### OpenAI

```yaml
# connectors/openai.yaml
type: connector
driver: openai
api_key: "{{ .env.openai_api_key }}"
```

### Claude

```yaml
# connectors/claude.yaml
type: connector
driver: claude
api_key: "{{ .env.claude_api_key }}"
```

### Slack

```yaml
# connectors/slack.yaml
type: connector
driver: slack
bot_token: "{{ .env.slack_bot_token }}"
```

### Pinot

```yaml
# connectors/pinot.yaml
type: connector
driver: pinot
broker_host: "{{ .env.pinot_broker_host }}"
controller_host: "{{ .env.pinot_controller_host }}"
username: "{{ .env.pinot_user }}"
password: "{{ .env.pinot_password }}"
ssl: true
```

## Reference documentation

Here is a full JSON schema for the connector syntax:

```
allOf:
    - properties:
        type:
            const: connector
            description: Refers to the resource type and must be `connector`
            type: string
      required:
        - type
      title: Properties
      type: object
    - $ref: '#/definitions/common_properties'
    - oneOf:
        - examples:
            - allow_host_access: true
              aws_access_key_id: '{{ .env.AWS_ACCESS_KEY_ID }}'
              aws_access_token: '{{ .env.AWS_ACCESS_TOKEN }}'
              aws_region: us-east-1
              aws_secret_access_key: '{{ .env.AWS_SECRET_ACCESS_KEY }}'
              driver: athena
              external_id: MyExternalID
              output_location: s3://my-bucket/athena-output/
              role_arn: arn:aws:iam::123456789012:role/MyRole
              role_session_name: MySession
              type: connector
              workgroup: primary
          properties:
            allow_host_access:
                description: Allow the Athena client to access host environment configurations such as environment variables or local AWS credential files. Defaults to true, enabling use of credentials and settings from the host environment unless explicitly disabled.
                type: boolean
            aws_access_key_id:
                description: AWS Access Key ID used for authentication. Required when using static credentials directly or as base credentials for assuming a role.
                type: string
            aws_access_token:
                description: AWS session token used with temporary credentials. Required only if the Access Key and Secret Key are part of a temporary session credentials.
                type: string
            aws_region:
                description: AWS region where Athena and the result S3 bucket are located (e.g., us-east-1). Defaults to 'us-east-1' if not specified.
                type: string
            aws_secret_access_key:
                description: AWS Secret Access Key paired with the Access Key ID. Required when using static credentials directly or as base credentials for assuming a role.
                type: string
            driver:
                const: athena
                description: Refers to the driver type and must be driver `athena`
                type: string
            external_id:
                description: External ID required by some roles when assuming them, typically for cross-account access. Used only if 'role_arn' is specified and the role's trust policy requires it.
                type: string
            output_location:
                description: S3 URI where Athena query results should be stored (e.g., s3://your-bucket/athena/results/). Optional if the selected workgroup has a default result configuration.
                type: string
            role_arn:
                description: ARN of the IAM role to assume. When specified, the SDK uses the base credentials to call STS AssumeRole and obtain temporary credentials scoped to this role.
                type: string
            role_session_name:
                description: Session name to associate with the STS AssumeRole session. Used only if 'role_arn' is specified. Useful for identifying and auditing the session.
                type: string
            workgroup:
                description: Athena workgroup to use for query execution. Defaults to 'primary' if not specified.
                type: string
          required:
            - driver
          title: Athena
          type: object
        - examples:
            - azure_storage_account: mystorageaccount
              azure_storage_key: '{{ .env.AZURE_STORAGE_KEY }}'
              driver: azure
              type: connector
          properties:
            allow_host_access:
                description: Allow access to host environment configuration
                type: boolean
            azure_storage_account:
                description: Azure storage account name
                type: string
            azure_storage_connection_string:
                description: Optional azure connection string for storage account
                type: string
            azure_storage_key:
                description: Azure storage access key
                type: string
            azure_storage_sas_token:
                description: Optional azure SAS token for authentication
                type: string
            driver:
                const: azure
                description: Refers to the driver type and must be driver `azure`
                type: string
            path_prefixes:
                description: |
                    A list of container or virtual directory prefixes that this connector is allowed to access.
                    Useful when different containers or paths use different credentials, allowing the system
                    to route access through the appropriate connector based on the blob path.
                    Example: `azure://my-bucket/`, ` azure://my-bucket/path/` ,`azure://my-bucket/path/prefix`
                type:
                    - string
                    - array
          required:
            - driver
            - azure_storage_account
            - azure_storage_key
          title: Azure
          type: object
        - examples:
            - allow_host_access: true
              driver: bigquery
              google_application_credentials: '{{ .env.GOOGLE_APPLICATION_CREDENTIALS }}'
              project_id: my-project-id
              type: connector
          properties:
            allow_host_access:
                description: Enable the BigQuery client to use credentials from the host environment when no service account JSON is provided. This includes Application Default Credentials from environment variables, local credential files, or the Google Compute Engine metadata server. Defaults to true, allowing seamless authentication in GCP environments.
                type: boolean
            dataset_id:
                description: BigQuery dataset ID
                type: string
            driver:
                const: bigquery
                description: Refers to the driver type and must be driver `bigquery`
                type: string
            google_application_credentials:
                description: Raw contents of the Google Cloud service account key (in JSON format) used for authentication.
                type: string
            location:
                description: BigQuery dataset location
                type: string
            project_id:
                description: Google Cloud project ID
                type: string
          required:
            - driver
          title: BigQuery
          type: object
        - examples:
            - cluster: mycluster
              database: mydatabase
              driver: clickhouse
              host: localhost
              managed: false
              mode: readwrite
              password: '{{ .env.CLICKHOUSE_PASSWORD }}'
              port: 9000
              ssl: true
              type: connector
              username: myusername
          properties:
            can_scale_to_zero:
                description: Indicates if the database can scale to zero
                type: boolean
            cluster:
                description: Cluster name, required for running distributed queries
                type: string
            conn_max_lifetime:
                description: Maximum time a connection may be reused
                type: string
            database:
                description: Name of the ClickHouse database within the cluster
                type: string
            dial_timeout:
                description: Timeout for dialing the ClickHouse server
                type: string
            driver:
                const: clickhouse
                description: Refers to the driver type and must be driver `clickhouse`
                type: string
            dsn:
                description: DSN(Data Source Name) for the ClickHouse connection
                type: string
            embed_port:
                description: Port to run ClickHouse locally (0 for random port)
                type: integer
            host:
                description: Host where the ClickHouse instance is running
                type: string
            log_queries:
                description: Controls whether to log raw SQL queries
                type: boolean
            managed:
                description: '`true` means Rill will provision the connector using the default provisioner. `false` disables automatic provisioning.'
                type: boolean
            max_idle_conns:
                description: Maximum number of idle connections in the pool
                type: integer
            max_open_conns:
                description: Maximum number of open connections to the database
                type: integer
            mode:
                description: '`read` - Controls the operation mode for the ClickHouse connection. Defaults to ''read'' for safe operation with external databases. Set to ''readwrite'' to enable model creation and table mutations. Note: When ''managed: true'', this is automatically set to ''readwrite''.'
                type: string
            password:
                description: Password for authentication
                type: string
            port:
                description: Port where the ClickHouse instance is accessible
                type: integer
            query_settings:
                description: query settings to be set on dashboard queries. `query_settings_override` takes precedence over these settings and if set these are ignored. Each setting must be separated by a comma. Example `max_threads = 8, max_memory_usage = 10000000000`
                type: string
            query_settings_override:
                description: override the default settings used in queries. Changing the default settings can lead to incorrect query results and is generally not recommended. If you need to add settings, use `query_settings`
                type: string
            read_timeout:
                description: Maximum time for a connection to read data
                type: string
            ssl:
                description: Indicates whether a secured SSL connection is required
                type: boolean
            username:
                description: Username for authentication
                type: string
          required:
            - driver
          title: ClickHouse
          type: object
        - examples:
            - driver: druid
              host: localhost
              password: '{{ .env.DRUID_PASSWORD }}'
              port: 8082
              ssl: true
              type: connector
              username: myusername
          properties:
            driver:
                const: druid
                description: Refers to the driver type and must be driver `druid`
                type: string
            dsn:
                description: Data Source Name (DSN) for connecting to Druid
                type: string
            host:
                description: Hostname of the Druid coordinator or broker
                type: string
            log_queries:
                description: Log raw SQL queries sent to Druid
                type: boolean
            max_open_conns:
                description: Maximum number of open database connections (0 = default, -1 = unlimited)
                type: integer
            password:
                description: Password for authenticating with Druid
                type: string
            port:
                description: Port number of the Druid service
                type: integer
            skip_version_check:
                description: Skip checking Druid version compatibility
                type: boolean
            ssl:
                description: Enable SSL for secure connection
                type: boolean
            username:
                description: Username for authenticating with Druid
                type: string
          required:
            - driver
            - dsn
          title: Druid
          type: object
        - examples:
            - allow_host_access: true
              cpu: 4
              driver: duckdb
              init_sql: INSTALL httpfs; LOAD httpfs;
              log_queries: true
              memory_limit_gb: 16
              mode: readwrite
              pool_size: 5
              read_write_ratio: 0.7
              type: connector
          properties:
            allow_host_access:
                description: Whether access to local environment and file system is allowed
                type: boolean
            attach:
                description: Full ATTACH statement to attach a DuckDB database
                type: string
            boot_queries:
                description: Deprecated - Use init_sql instead
                type: string
            conn_init_sql:
                description: SQL executed when a new connection is initialized
                type: string
            cpu:
                description: Number of CPU cores available to the database
                minimum: 1
                type: integer
            create_secrets_from_connectors:
                description: List of connector names for which temporary secrets should be created before executing the SQL.
                type:
                    - string
                    - array
            database_name:
                description: Name of the attached DuckDB database (auto-detected if not set)
                type: string
            driver:
                const: duckdb
                description: Must be "duckdb"
                type: string
            init_sql:
                description: SQL executed during database initialization
                type: string
            log_queries:
                description: Whether to log raw SQL queries executed through OLAP
                type: boolean
            memory_limit_gb:
                description: Amount of memory in GB available to the database
                minimum: 1
                type: integer
            mode:
                default: read
                description: Set the mode for the DuckDB connection.
                enum:
                    - read
                    - readwrite
                type: string
            path:
                description: Path to external DuckDB database
                type: string
            pool_size:
                description: Number of concurrent connections and queries allowed
                minimum: 1
                type: integer
            read_write_ratio:
                default: 0.5
                description: Ratio of resources allocated to read vs write operations
                maximum: 1
                minimum: 0
                type: number
            schema_name:
                description: Default schema used by the DuckDB database
                type: string
          required:
            - driver
          title: DuckDB
          type: object
        - examples:
            - driver: duckdb
              mode: read
              path: /path/to/my-duckdb-database.db
              type: connector
          properties:
            driver:
                const: duckdb
                description: Refers to the driver type and must be driver `duckdb`
                type: string
            mode:
                default: read
                description: Set the mode for the DuckDB connection.
                enum:
                    - read
                    - readwrite
                type: string
            path:
                description: Path to the DuckDB database
                type: string
          required:
            - driver
            - db
          title: External DuckDB
          type: object
        - examples:
            - driver: gcs
              google_application_credentials: '{{ .env.GOOGLE_APPLICATION_CREDENTIALS }}'
              type: connector
          properties:
            allow_host_access:
                description: Allow access to host environment configuration
                type: boolean
            driver:
                const: gcs
                description: Refers to the driver type and must be driver `gcs`
                type: string
            google_application_credentials:
                description: Google Cloud credentials JSON string
                type: string
            key_id:
                description: Optional S3-compatible Key ID when used in compatibility mode
                type: string
            path_prefixes:
                description: "A list of bucket path prefixes that this connector is allowed to access. \nUseful when different buckets or bucket prefixes use different credentials, \nallowing the system to select the appropriate connector based on the bucket path.\nExample: `gs://my-bucket/`, ` gs://my-bucket/path/` ,`gs://my-bucket/path/prefix`\n"
                type:
                    - string
                    - array
            secret:
                description: Optional S3-compatible Secret when used in compatibility mode
                type: string
          required:
            - driver
            - bucket
          title: GCS
          type: object
        - examples:
            - driver: https
              headers:
                Authorization: Bearer {{ .env.HTTPS_TOKEN }}
              type: connector
          properties:
            driver:
                const: https
                description: Refers to the driver type and must be driver `https`
                type: string
            headers:
                additionalProperties:
                    type: string
                description: HTTP headers to include in the request
                type: object
            path_prefixes:
                description: |
                    A list of HTTP/HTTPS URL prefixes that this connector is allowed to access.
                    Useful when different URL namespaces use different credentials, enabling the
                    system to choose the appropriate connector based on the URL path.
                    Example: `https://example.com/`, ` https://example.com/path/` ,`https://example.com/path/prefix`
                type:
                    - string
                    - array
          required:
            - driver
            - path
          title: HTTPS
          type: object
        - examples:
            - driver: duckdb
              path: md:my_database
              schema_name: my_schema
              token: '{{ .env.MOTHERDUCK_TOKEN }}'
              type: connector
          properties:
            create_secrets_from_connectors:
                description: List of connector names for which temporary secrets should be created before executing the SQL.
                type:
                    - string
                    - array
            driver:
                const: duckdb
                description: Refers to the driver type and must be driver `duckdb`.
                type: string
            init_sql:
                description: SQL executed during database initialization.
                type: string
            mode:
                default: read
                description: Set the mode for the MotherDuck connection. By default, it is set to 'read' which allows only read operations. Set to 'readwrite' to enable model creation and table mutations.
                enum:
                    - read
                    - readwrite
                type: string
            path:
                description: Path to your MD database
                type: string
            schema_name:
                description: Define your schema if not main, uses main by default
                type: string
            token:
                description: MotherDuck token
                type: string
          required:
            - driver
            - path
            - token
          title: MotherDuck
          type: object
        - examples:
            - database: mydb
              driver: mysql
              host: localhost
              password: '{{ .env.MYSQL_PASSWORD }}'
              port: 3306
              ssl-mode: preferred
              type: connector
              user: user
            - driver: mysql
              dsn: '{{ .env.MYSQL_DSN }}'
              type: connector
          properties:
            database:
                description: Name of the MySQL database
                type: string
            driver:
                description: Refers to the driver type and must be driver `mysql`
                type: string
            dsn:
                description: |
                    **Data Source Name (DSN)** for the MySQL connection, provided in [MySQL URI format](https://dev.mysql.com/doc/refman/8.4/en/connecting-using-uri-or-key-value-pairs.html#connecting-using-uri).
                    The DSN must follow the standard MySQL URI scheme:
                    ```text
                    mysql://user:password@host:3306/my-db
                    ```
                    Rules for special characters in password:
                    - The following characters are allowed [unescaped in the URI](https://datatracker.ietf.org/doc/html/rfc3986#section-2.3): `~` `.` `_` `-`
                    - All other special characters must be percent-encoded (`%XX` format).
                    ```text
                    mysql://user:pa%40ss@localhost:3306/my-db   # password contains '@'
                    mysql://user:pa%3Ass@localhost:3306/my-db   # password contains ':'
                    ```
                type: string
            host:
                description: Hostname of the MySQL server
                type: string
            password:
                description: Password for authentication
                type: string
            port:
                description: Port number for the MySQL server
                type: integer
            ssl-mode:
                description: 'ssl mode options: `disabled`, `preferred`, or `required`.'
                type: string
            user:
                description: Username for authentication
                type: string
          required:
            - driver
          title: MySQL
          type: object
        - examples:
            - api_key: '{{ .env.OPENAI_API_KEY }}'
              api_type: openai
              api_version: "2023-05-15"
              base_url: https://api.openai.com/v1
              driver: openai
              model: gpt-4o
              type: connector
          properties:
            api_key:
                description: API key for connecting to OpenAI
                type: string
            api_type:
                description: The type of OpenAI API to use
                type: string
            api_version:
                description: The version of the OpenAI API to use (e.g., '2023-05-15'). Required when API Type is AZURE or AZURE_AD
                type: string
            base_url:
                description: The base URL for the OpenAI API (e.g., 'https://api.openai.com/v1')
                type: string
            driver:
                description: The driver type, must be set to "openai"
                type: string
            model:
                description: The OpenAI model to use (e.g., 'gpt-4o')
                type: string
          required:
            - api_key
          title: OpenAI
          type: object
        - examples:
            - api_key: '{{ .env.claude_api_key }}'
              driver: claude
              model: claude-opus-4-5
              type: connector
          properties:
            api_key:
                description: API key for connecting to Claude
                type: string
            base_url:
                description: The base URL for the Claude API
                type: string
            driver:
                description: The driver type, must be set to "claude"
                type: string
            max_tokens:
                description: Maximum number of tokens in the response (e.g., 8192)
                type: number
            model:
                description: The Claude model to use (e.g., 'claude-opus-4-5')
                type: string
            temperature:
                description: Sampling temperature to use (e.g., 0.0)
                type: number
          required:
            - api_key
          title: Claude
          type: object
        - examples:
            - api_key: '{{ .env.gemini_api_key }}'
              driver: gemini
              model: gemini-2.5-pro-preview-05-06
              type: connector
          properties:
            api_key:
                description: API key for connecting to Gemini
                type: string
            driver:
                description: The driver type, must be set to "gemini"
                type: string
            include_thoughts:
                description: Whether to include thinking/reasoning in the response
                type: boolean
            max_output_tokens:
                description: Maximum number of tokens in the response (e.g., 8192)
                type: number
            model:
                description: The Gemini model to use (e.g., 'gemini-2.5-pro-preview-05-06')
                type: string
            temperature:
                description: Sampling temperature to use (0.0-2.0)
                type: number
            thinking_level:
                description: Level of 'thinking' for the model's response (e.g., 'MINIMAL', 'LOW', 'MEDIUM', 'HIGH'). Default is 'LOW'.
                type: string
            top_k:
                description: Top-K sampling parameter
                type: number
            top_p:
                description: Nucleus sampling parameter
                type: number
          required:
            - api_key
          title: Gemini
          type: object
        - examples:
            - broker_host: localhost
              broker_port: 9000
              controller_host: localhost
              controller_port: 9000
              driver: pinot
              log_queries: true
              max_open_conns: 100
              password: '{{ .env.PINOT_PASSWORD }}'
              ssl: true
              timeout_ms: 30000
              type: connector
              username: myusername
          properties:
            broker_host:
                description: Hostname of the Pinot broker
                type: string
            broker_port:
                description: Port number for the Pinot broker
                type: integer
            controller_host:
                description: Hostname of the Pinot controller
                type: string
            controller_port:
                description: Port number for the Pinot controller
                type: integer
            driver:
                description: Refers to the driver type and must be driver `pinot`
                type: string
            dsn:
                description: DSN(Data Source Name) for the Pinot connection
                type: string
            log_queries:
                description: Log raw SQL queries executed through Pinot
                type: boolean
            max_open_conns:
                description: Maximum number of open connections to the Pinot database
                type: integer
            password:
                description: Password for authenticating with Pinot
                type: string
            ssl:
                description: Enable SSL connection to Pinot
                type: boolean
            timeout_ms:
                description: Query timeout in milliseconds
                type: integer
            username:
                description: Username for authenticating with Pinot
                type: string
          required:
            - driver
            - dsn
            - broker_host
            - controller_host
          title: Pinot
          type: object
        - examples:
            - catalog: default_catalog
              database: my_database
              driver: starrocks
              host: starrocks-fe.example.com
              password: '{{ .env.STARROCKS_PASSWORD }}'
              port: 9030
              ssl: false
              type: connector
              username: analyst
          properties:
            catalog:
                default: default_catalog
                description: StarRocks catalog name (for external catalogs like Iceberg, Hive)
                type: string
            database:
                description: StarRocks database name
                type: string
            driver:
                const: starrocks
                description: Refers to the driver type and must be driver `starrocks`
                type: string
            dsn:
                description: DSN (Data Source Name) for the StarRocks connection. Follows MySQL protocol format.
                type: string
            host:
                description: StarRocks FE (Frontend) server hostname
                type: string
            password:
                description: Password for authentication
                type: string
            port:
                default: 9030
                description: MySQL protocol port of StarRocks FE
                type: integer
            ssl:
                default: false
                description: Enable SSL/TLS encryption
                type: boolean
            username:
                description: Username for authentication
                type: string
          required:
            - driver
          title: StarRocks
          type: object
        - examples:
            - dbname: mydatabase
              driver: postgres
              host: localhost
              password: '{{ .env.POSTGRES_PASSWORD }}'
              port: 5432
              sslmode: prefer
              type: connector
              user: myusername
            - driver: postgres
              dsn: '{{ .env.POSTGRES_DSN }}'
              type: connector
          properties:
            dbname:
                description: Name of the Postgres database
                type: string
            driver:
                description: Refers to the driver type and must be driver `postgres`
                type: string
            dsn:
                description: |
                    **Data Source Name (DSN)** for the PostgreSQL connection, provided in
                    [PostgreSQL connection string format](https://www.postgresql.org/docs/current/libpq-connect.html#LIBPQ-CONNSTRING).
                    PostgreSQL supports both **key=value format** and **URI format**.

                    key=value format example:
                    ```text
                    user=user password=password host=host port=5432 dbname=mydb
                    ```
                    Rules for key=value format for special characters:
                    - To write an empty value, or a value containing spaces, `=`, single quotes, or backslashes, surround it with single quotes.
                    - Single quotes and backslashes inside a value must be escaped with a backslash (`\'` and `\\`).

                    URI format example:
                    ```text
                    postgres://user:password@host:5432/mydb
                    ```

                    Rules for URI format:
                    - The following characters are allowed [unescaped in the URI](https://datatracker.ietf.org/doc/html/rfc3986#section-2.3): `~` `.` `_` `-`
                    - All other special characters must be percent-encoded (`%XX` format).

                    Examples (URI format with encoded characters):
                    ```text
                    postgres://user:pa%40ss@localhost:5432/my-db   # '@' is encoded as %40
                    postgres://user:pa%3Ass@localhost:5432/my-db   # ':' is encoded as %3A
                    ```
                type: string
            host:
                description: Hostname of the Postgres server
                type: string
            password:
                description: Password for authentication
                type: string
            port:
                description: Port number for the Postgres server
                type: string
            sslmode:
                description: 'ssl mode options: `disable`, `allow`, `prefer` or `require`.'
                type: string
            user:
                description: Username for authentication
                type: string
          required:
            - driver
          title: Postgres
          type: object
        - examples:
            - dbname: postgres
              driver: postgres
              host: aws-0-us-east-1.pooler.supabase.com
              password: '{{ .env.SUPABASE_PASSWORD }}'
              port: 5432
              sslmode: require
              type: connector
              user: postgres.[your-project-ref]
            - driver: postgres
              dsn: '{{ .env.SUPABASE_DSN }}'
              type: connector
          properties:
            dbname:
                description: Name of the Supabase database
                type: string
            driver:
                description: Refers to the driver type and must be driver `postgres`
                type: string
            dsn:
                description: |
                    **Data Source Name (DSN)** for the Supabase connection, provided in
                    [PostgreSQL connection string format](https://www.postgresql.org/docs/current/libpq-connect.html#LIBPQ-CONNSTRING).
                    Supabase uses PostgreSQL under the hood, so all PostgreSQL connection formats are supported.

                    URI format example:
                    ```text
                    postgresql://postgres.[ref]:[password]@aws-0-[region].pooler.supabase.com:5432/postgres
                    ```
                type: string
            host:
                description: Hostname of the Supabase database (e.g. aws-0-us-east-1.pooler.supabase.com)
                type: string
            password:
                description: Password for authentication
                type: string
            port:
                description: Port number for the Supabase database
                type: string
            sslmode:
                description: 'ssl mode options: `disable`, `allow`, `prefer` or `require`.'
                type: string
            user:
                description: Username for authentication (e.g. postgres.[your-project-ref])
                type: string
          required:
            - driver
          title: Supabase
          type: object
        - examples:
            - aws_access_key_id: '{{ .env.AWS_ACCESS_KEY_ID }}'
              aws_access_token: '{{ .env.AWS_ACCESS_TOKEN }}'
              aws_secret_access_key: '{{ .env.AWS_SECRET_ACCESS_KEY }}'
              cluster_identifier: my-cluster-identifier
              database: mydatabase
              driver: redshift
              region: us-east-1
              type: connector
              workgroup: my-workgroup
          properties:
            aws_access_key_id:
                description: AWS Access Key ID used for authenticating with Redshift.
                type: string
            aws_access_token:
                description: AWS Session Token for temporary credentials (optional).
                type: string
            aws_secret_access_key:
                description: AWS Secret Access Key used for authenticating with Redshift.
                type: string
            cluster_identifier:
                description: Cluster identifier for provisioned Redshift clusters, in case of Redshift Serverless use 'workgroup' .
                type: string
            database:
                description: Name of the Redshift database to query.
                type: string
            driver:
                description: Refers to the driver type and must be driver `redshift`
                type: string
            region:
                description: AWS region where the Redshift cluster or workgroup is hosted (e.g., 'us-east-1').
                type: string
            workgroup:
                description: Workgroup name for Redshift Serverless, in case of provisioned Redshift clusters use 'cluster_identifier'.
                type: string
          required:
            - driver
            - aws_access_key_id
            - aws_secret_access_key
            - database
          title: Redshift
          type: object
        - examples:
            - aws_access_key_id: '{{ .env.AWS_ACCESS_KEY_ID }}'
              aws_access_token: '{{ .env.AWS_ACCESS_TOKEN }}'
              aws_secret_access_key: '{{ .env.AWS_SECRET_ACCESS_KEY }}'
              driver: s3
              endpoint: https://my-s3-endpoint.com
              region: us-east-1
              type: connector
          properties:
            allow_host_access:
                description: Allow access to host environment configuration
                type: boolean
            aws_access_key_id:
                description: AWS Access Key ID used for authentication
                type: string
            aws_access_token:
                description: Optional AWS session token for temporary credentials
                type: string
            aws_external_id:
                description: External ID for cross-account role assumption
                type: string
            aws_role_arn:
                description: ARN of the IAM role to assume for accessing S3 resources
                type: string
            aws_role_session_name:
                description: Session name to use when assuming the IAM role
                type: string
            aws_secret_access_key:
                description: AWS Secret Access Key used for authentication
                type: string
            driver:
                description: Refers to the driver type and must be driver `s3`
                type: string
            endpoint:
                description: Optional custom endpoint URL for S3-compatible storage
                type: string
            path_prefixes:
                description: |
                    A list of bucket path prefixes that this connector is allowed to access.
                    Useful when different buckets or bucket prefixes use different credentials,
                    allowing the system to select the appropriate connector based on the bucket path.
                    Example: `s3://my-bucket/`, ` s3://my-bucket/path/` ,`s3://my-bucket/path/prefix`
                type:
                    - string
                    - array
            region:
                description: AWS region of the S3 bucket
                type: string
          required:
            - driver
            - bucket
          title: S3
          type: object
        - examples:
            - client_id: my-client-id
              driver: salesforce
              endpoint: https://login.salesforce.com
              key: '{{ .env.SALESFORCE_KEY }}'
              password: '{{ .env.SALESFORCE_PASSWORD }}'
              type: connector
              username: myusername
          properties:
            client_id:
                description: Client ID used for Salesforce OAuth authentication
                type: string
            driver:
                description: Refers to the driver type and must be driver `salesforce`
                type: string
            endpoint:
                description: Salesforce API endpoint URL
                type: string
            key:
                description: Authentication key for Salesforce (secret)
                type: string
            password:
                description: Salesforce account password (secret)
                type: string
            username:
                description: Salesforce account username
                type: string
          required:
            - driver
            - username
            - endpoint
            - client_id
          title: Salesforce
          type: object
        - examples:
            - bot_token: '{{ .env.SLACK_BOT_TOKEN }}'
              driver: slack
              type: connector
          properties:
            bot_token:
                description: Bot token used for authenticating Slack API requests
                type: string
            driver:
                description: Refers to the driver type and must be driver `slack`
                type: string
          required:
            - driver
            - bot_token
          title: Slack
          type: object
        - examples:
            - account: my_account_identifier
              database: my_db
              driver: snowflake
              parallel_fetch_limit: 2
              privateKey: '{{ .env.SNOWFLAKE_PRIVATE_KEY }}'
              role: my_role
              schema: my_schema
              type: connector
              user: my_user
              warehouse: my_wh
            - driver: snowflake
              dsn: '{{ .env.SNOWFLAKE_DSN }}'
              parallel_fetch_limit: 2
              type: connector
          properties:
            account:
                description: Snowflake account identifier. To find your Snowflake account identifier, look at your Snowflake account URL. The account identifier is everything before .snowflakecomputing.com
                type: string
            authenticator:
                description: Optional authenticator type (e.g., SNOWFLAKE_JWT).
                type: string
            database:
                description: Name of the Snowflake database.
                type: string
            driver:
                const: snowflake
                description: Refers to the driver type and must be driver `snowflake`
                type: string
            dsn:
                description: "DSN (Data Source Name) for the Snowflake connection.\n\nThis is intended for **advanced configuration** where you want to specify\nproperties that are not explicitly defined above.  \nIt can only be used when the other connection fields (account, user, password,\ndatabase, schema, warehouse, role, authenticator, privateKey) are **not used**.\n\nFor details on private key generation and encoding, see the `privateKey` property.\n"
                type: string
            parallel_fetch_limit:
                description: Maximum number of concurrent fetches during query execution.
                type: integer
            password:
                description: Password for the Snowflake connection. _(deprecated, use privateKey instead)_
                type: string
            privateKey:
                description: |
                    Private key for JWT authentication.
                    :::tip
                    Private key must be generated as a **PKCS#8 (nocrypt) key**, since the Snowflake Go driver
                    only supports unencrypted private keys. After generating, it must be **base64 URL encoded**.

                    Example commands to generate and encode:

                    ```bash
                    # Generate a 2048-bit unencrypted PKCS#8 private key
                    openssl genrsa 2048 | openssl pkcs8 -topk8 -inform PEM -out rsa_key.p8 -nocrypt

                    # Convert URL safe format for Snowflake
                    cat rsa_key.p8 | grep -v "\----" | tr -d '\n' | tr '+/' '-_'
                    ```
                    See: https://docs.snowflake.com/en/guide/key-pair-auth
                    :::
                type: string
            role:
                description: Snowflake role to use.
                type: string
            schema:
                description: Schema within the database to use.
                type: string
            user:
                description: Username for the Snowflake connection.
                type: string
            warehouse:
                description: Compute warehouse to use for queries.
                type: string
          required:
            - type
            - driver
          title: Snowflake
          type: object
        - examples:
            - driver: sqlite
              dsn: file:mydatabase.db
              type: connector
          properties:
            driver:
                description: Refers to the driver type and must be driver `sqlite`
                type: string
            dsn:
                description: DSN(Data Source Name) for the sqlite connection
                type: string
          required:
            - driver
            - dsn
          title: SQLite
          type: object
description: |
    Connector YAML files define how Rill connects to external data sources and OLAP engines. Each connector specifies a driver type and its required connection parameters.

    ## Available Connector Types

    ### _OLAP Engines_
    - [**ClickHouse**](#clickhouse) - ClickHouse analytical database
    - [**Druid**](#druid) - Apache Druid
    - [**DuckDB**](#duckdb) - Embedded DuckDB engine (default)
    - [**External DuckDB**](#external-duckdb) - External DuckDB database
    - [**MotherDuck**](#motherduck) - MotherDuck cloud database
    - [**Pinot**](#pinot) - Apache Pinot
    - [**StarRocks**](#starrocks) - StarRocks analytical database

    ### _Data Warehouses_
    - [**Athena**](#athena) - Amazon Athena
    - [**BigQuery**](#bigquery) - Google BigQuery
    - [**Redshift**](#redshift) - Amazon Redshift
    - [**Snowflake**](#snowflake) - Snowflake data warehouse

    ### _Databases_
    - [**MySQL**](#mysql) - MySQL databases
    - [**PostgreSQL**](#postgres) - PostgreSQL databases
    - [**SQLite**](#sqlite) - SQLite databases
    - [**Supabase**](#supabase) - Supabase (managed PostgreSQL)

    ### _Object Storage_
    - [**Azure**](#azure) - Azure Blob Storage
    - [**GCS**](#gcs) - Google Cloud Storage
    - [**S3**](#s3) - Amazon S3 storage

    ### Service Integrations
    - [**Claude**](#claude) - Claude connector for chat with your own API key
    - [**OpenAI**](#openai) - OpenAI connector for chat with your own API key
    - [**Gemini**](#gemini) - Gemini connector for chat with your own API key
    - [**Slack**](#slack) - Slack data

    ### _Other_
    - [**HTTPS**](#https) - Public files via HTTP/HTTPS
    - [**Salesforce**](#salesforce) - Salesforce data

    :::warning Security Recommendation
    For all credential parameters (passwords, tokens, keys), use environment variables with the syntax `{{ .env.KEY_NAME }}`. This keeps sensitive data out of your YAML files and version control. See our [credentials documentation](/developers/build/connectors/credentials/) for complete setup instructions.
    :::
id: connectors
title: Connector YAML
type: object
```