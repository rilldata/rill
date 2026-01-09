---
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
- **OpenAI**: Use `api_key` for AI-powered features
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
