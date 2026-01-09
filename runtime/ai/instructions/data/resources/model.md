---
description: Detailed instructions and examples for developing model resources in Rill
---

# Instructions for developing a model in Rill

## Introduction

Models are resources that specify ETL or transformation logic, outputting a tabular dataset to one of the project's connectors. They are typically found near the root of the project's DAG, referencing only connectors and other models.

By default, models output data as a table with the same name as the model in the project's default OLAP connector. The core of a model is usually a `SELECT` SQL statement, which Rill executes as `CREATE TABLE <name> AS <SELECT statement>`. The SQL should be a plain SELECT query without a trailing semicolon.

Models in Rill are similar to models in dbt, but support additional advanced features:
- **Different input and output connectors:** Run a query in one database (e.g., BigQuery) and output results to another (e.g., DuckDB or ClickHouse).
- **Stateful incremental ingestion:** Track state and load only new or changed data.
- **Partition support:** Define explicit partitions (e.g., Hive-partitioned files in S3) for scalable, idempotent incremental runs.
- **Scheduled refresh:** Use cron expressions to automatically refresh data on a schedule.

### Model categories

When reasoning about a model, consider these attributes:

- **Source model**: References external data, typically reading from a SQL database or object store connector and writing to an OLAP connector.
- **Derived model**: References other models, usually performing joins or formatting columns to prepare denormalized tables for metrics views and dashboards.
- **Incremental model**: Contains logic for incrementally loading data, processing only new or changed records.
- **Partitioned model**: Loads data in well-defined increments (e.g., daily partitions), enabling scalability and idempotent incremental runs.
- **Materialized model**: Outputs a physical table rather than a SQL view.

### Performance considerations

Models are usually expensive resources that can take a long time to run. Create or edit them with caution.

**Exception:** Non-materialized models with the same input and output connector are cheap because they are created as SQL views rather than physical tables.

**Development tip:** Use a "dev partition" to limit data processed during development. This speeds up iteration and avoids unnecessary costs. See the partitions section below for details.

### Generating synthetic data for prototyping

When developing models for prototyping or demonstration purposes where external data sources are not yet available, generate a `SELECT` query that returns realistic synthetic data with these characteristics:
- Use realistic column names and data types that match typical business scenarios
- Always include a time/timestamp column for time-series analysis
- Generate 6-12 months of historical data with approximately 10,000 rows to enable meaningful analysis
- Space out timestamps realistically across the time period rather than clustering them
- Use realistic data distributions (e.g., varying quantities, diverse categories, plausible geographic distributions)

Only generate synthetic data when the user explicitly requests mock data or when required external sources don't exist in the project. If real data sources are available, always prefer using them.

## Materialization

The `materialize:` property controls whether a model creates a physical table or a SQL view:

- `materialize: true`: Creates a physical table. Use this for source models, expensive transformations, or when downstream queries need fast access.
- `materialize: false`: Creates a SQL view. The query re-executes on every access. Only suitable for lightweight transformations where input and output connectors are the same.

If `materialize` is omitted, it defaults to `true` for all cross-connector models and `false` for single-connector models (i.e. where the input and output connector is the same).

**Best practices:**
- Always materialize models that reference external data sources.
- Always materialize models that perform expensive joins or aggregations.
- Use views only for simple transformations on top of already-materialized tables.

## Incremental models

Incremental models process only new or changed data instead of reprocessing the entire dataset. This is essential for large datasets where full refreshes would be too slow or expensive.

### Incremental strategies

The `output.incremental_strategy` property controls how new data is merged with existing data:

- `partition_overwrite`: Entire partitions are replaced. This is the default strategy for partition-based incremental models.
- `merge`: New rows are merged based on `output.unique_key`. Use for upsert semantics.
- `append`: New rows are appended to the table. This is the default for state-based incremental models. Generally avoid this since retries will lead to duplicate data.

### Partition-based incremental models

Use the `partitions:` property to define explicit data partitions. Combined with `incremental: true`, Rill tracks which partitions have been processed to avoid duplicate processing. Example:

```yaml
type: model
incremental: true

partitions:
  glob:
    connector: s3
    path: s3://bucket/data/year=*/month=*/day=*/*.parquet

sql: SELECT * FROM read_parquet('{{ .partition.uri }}')
```

Each partition gets inserted using the `partition_overwrite` strategy by default. The `partition_overwrite` strategy overwrites partitions based on the column(s) described in `output.partition_by`. If `partition_by` is not explicitly specified, a column `__rill_partition` is injected into the table and used for partitioning.

### State-based incremental models

You can also do dbt-style state-based incremental ingestion using a watermark (e.g., the maximum timestamp already processed). This is discouraged since it is not idempotent (unlike partitions-based incremental ingestion).

Example:

```yaml
type: model
incremental: true

connector: bigquery
sql: |
  SELECT * FROM events
  {{ if incremental }}
  WHERE event_time > '{{ .state.max_time }}'::TIMESTAMP
  {{ end }}

state:
  sql: SELECT MAX(event_time) AS max_time FROM events
```

The `{{ if incremental }}` block ensures the filter only applies during incremental runs, not during the initial full load. The `state` query is evaluated and stored only after the first non-incremental run. Note that the `state` query runs against the project's default OLAP connector (e.g. DuckDB), which is the default output location for models.

## Partitions

Partitions enable scalable data ingestion by breaking large datasets into manageable chunks. They also enable idempotent reruns; if a partition fails, only that partition needs to be reprocessed.

### Glob-based partitions

Discover partitions from file paths in object storage:

```yaml
partitions:
  glob:
    connector: s3
    path: s3://bucket/data/year=*/month=*/day=*/*.parquet
```

Available template variables:
- `{{ .partition.uri }}`: Full URI of the matched file or directory
- `{{ .partition.path }}`: Path portion without the scheme/bucket prefix

By default, `glob:` matches files only, but you can pass `partition: directory` to have it emit leaf directory names instead.

### SQL-based partitions

Generate partitions using a SQL query:

```yaml
partitions:
  connector: bigquery
  sql: SELECT DISTINCT date_trunc('day', event_time) AS day FROM events
```

## Dev partitions (data limits in local development)

You can override properties in development using either a root-level `dev:` property or the `{{ if dev }}` templating function. 

Example using `dev:` property:
```yaml
dev:
  partitions:
    glob:
      path: s3://bucket/data/year=2025/month=12/day=01/*.parquet
      partition: directory
```

Example using the templating function:
```yaml
connector: bigquery
sql: |
  SELECT * FROM users {{ if dev }} LIMIT 10000 {{ end }}
```

Best practices for dev partitions:
- Always use for large external data sources to keep developer velocity high
- Use mainly for fact tables, not (small) dimension tables
- Don't use for derived models
- When possible, limit data by time range (e.g. 1 day of data) rather than a row limit to get a representative sample of data

## Referencing other models

Use `{{ ref "model_name" }}` to reference parent models in SQL statements that use templating:

```yaml
sql: SELECT * FROM {{ ref "events_raw" }} WHERE country = 'US'
```

**Note:** If your SQL statement contains no other templating, the `ref` function is optional for DuckDB SQL snippets; Rill can in that case invoke DuckDB's SQL parser to automatically detect model references. This does not apply for non-DuckDB SQL models.

## Refresh schedules

By default, models refresh when a parent model in the DAG is refreshed. For source models that don't have a parent model, you can configure a cron refresh schedule:
```yaml
refresh:
  cron: 0 * * * *
```

By default, cron refreshes are disabled in local development. If you need to test them locally, add `run_in_dev: true` under `refresh:`.

## Advanced concepts

### Staging connectors

When the input and output connectors cannot directly read from or write to each other, use a staging connector (typically an object store) as an intermediary:

```yaml
type: model

connector: snowflake
sql: SELECT * FROM events

stage:
  connector: s3
  path: s3://bucket/staging/

output:
  connector: clickhouse
```

This is useful when moving data between incompatible systems (e.g., Snowflake to ClickHouse).

NOTE: This is not a production-ready feature. You may see it in some projects, but you should never suggest it or use it in a project that hasn't adopted it independently.

### Change modes

The `change_mode:` property controls how Rill handles changes to model specifications:

- `reset`: Drop and recreate the model automatically (default).
- `manual`: Require a manual refresh to apply changes (user can choose to do an incremental or full refresh).
- `patch`: Switch to new logic without reprocessing historical data (incremental models only).

### Retry configuration

Configure automatic retries for transient failures:

```yaml
retry:
  attempts: 5
  delay: 10s
  exponential_backoff: true
  if_error_matches:
    - ".*Timeout.*"
    - ".*Bad Gateway.*"
```

This is configured by default for common errors, so only add an explicit `retry` clause if you need to support retries for special errors or long delays.

## Dialect-specific notes

### DuckDB

- **Model references:** When the SQL contains no other templating, `{{ ref "model" }}` is optional; Rill uses DuckDB's SQL parser to detect references.
- **Connector secrets:** By default, all compatible connectors are automatically mounted as DuckDB secrets. Use `create_secrets_from_connectors:` to explicitly control which connectors are available.
- **Pre/post execution:** Use `pre_exec:` and `post_exec:` for setup and teardown queries (e.g., attaching external databases). Some legacy projects configure DuckDB secrets here, but with the automatic secret creation referenced above, it is usually better to create separate connector files instead.
- **Cloud storage paths:** DuckDB can read directly from S3 (`s3://`) and GCS (`gs://`) paths in `read_parquet()`, `read_csv()`, and `read_json()` functions.
- **CSV options:** When reading CSV files, useful options include `auto_detect=true`, `header=true`, `ignore_errors=true`, `union_by_name=true`, and `all_varchar=true` for handling inconsistent schemas.
- **JSON files:** Use `read_json()` with `auto_detect=true` and `format='auto'` for flexible JSON ingestion, including gzipped files.

### ClickHouse

- **S3 credentials:** When using the `s3()` function, reference `.env` values directly using templating since ClickHouse lacks integrated secret management:
  ```yaml
  sql: SELECT * FROM s3('s3://bucket/path/*.parquet', '{{ .env.aws_access_key }}', '{{ .env.aws_secret_key }}', 'Parquet')
  ```
- **Required order_by:** The `output:` section must always include an `order_by` property for materialized ClickHouse tables.
- **MergeTree vs. ReplicatedMergeTree:** You don't need to configure `MergeTree` or `ReplicatedMergeTree` engines explicitly. Rill uses `MergeTree` for materialized models by default, and automatically switches to `ReplicatedMergeTree` (creating distributed tables) when connected to a Clickhouse cluster.
- **LowCardinality types:** Use `LowCardinality(String)` for string columns with limited distinct values (e.g., country, device_type, status) to improve storage and query performance.
- **TTL for data retention:** Use `output.ttl` to automatically expire old data and prevent unbounded growth in incremental models.
- **Performance indexes:** If performance is poor for models powering metrics views, add indexes via `output.columns` to improve query performance. Common index types include `bloom_filter` for high-cardinality columns and `set(N)` for low-cardinality columns.

### Other SQL connectors

- Connector properties should be added to the separate connector YAML file when possible. Some legacy models add properties directly in the model, but this is discouraged. (For example, `project_id` for BigQuery, `output_location` for Athena, or `dsn:` for Postgres.)

## Syntax

Here is a basic scaffold of a model's high-level structure:

```yaml
type: model

# Here goes common properties, like materialize, incremental, change_mode, partitions, etc.
# These are usually needed for source models, but not necessarily for derived models.
# For example:
materialize: true

# Here goes input properties, like connector, sql, pre_exec, etc.
# There's always at least one input property.
# For example:
connector: bigquery
sql: SELECT ...

# Here goes output properties, like connector, incremental_strategy, order_by, etc.
# This is usually omitted for derived models, and for source models that output to DuckDB when it is the default OLAP connector.
# For example:
output:
  connector: clickhouse
  order_by: event_time
```

## JSON Schema

Here is a full JSON schema for the model syntax:

```
{% json_schema_for_resource "model" %}
```

## Examples

### Simple model with mock data as a SQL file

```sql
-- models/mock_data.sql
SELECT now() AS time, 'Denmark' AS country, 1 AS revenue_usd
UNION ALL
SELECT now() AS time, 'United States' AS country, 2 AS revenue_usd
```

### Materialized model as a SQL file

```sql
-- models/events.sql
-- @materialize: true
SELECT * FROM 's3://bucket/path/to/file.parquet'
```

### Simple model as a YAML file

```yaml
# models/mock_data.yaml
type: model
materialize: false

sql: |
  SELECT now() AS time, 'Denmark' AS country, 1 AS revenue_usd
  UNION ALL
  SELECT now() AS time, 'United States' AS country, 2 AS revenue_usd
```

### S3 to DuckDB (Parquet files)

Assuming DuckDB is the default OLAP and there is an `s3.yaml` connector in the project:

```yaml
# models/events_raw.yaml
type: model
materialize: true

sql: |
  SELECT * FROM read_parquet('s3://my-bucket/data/events/*.parquet')
```

### S3 to DuckDB with explicit OLAP connector

Assuming there is an `s3.yaml` connector in the project:

```yaml
# models/orders.yaml
type: model
materialize: true

connector: duckdb
sql: |
  SELECT
    order_id,
    customer_id,
    order_date,
    total_amount
  FROM read_parquet('s3://my-bucket/orders/year=2025/month=*/*.parquet')

output:
  connector: duckdb
```

### GCS to DuckDB (JSON files)

Assuming DuckDB is the default OLAP and there is a `gcs.yaml` connector in the project:

```yaml
# models/commits.yaml
type: model
materialize: true

sql: |
  SELECT * FROM read_json(
    'gs://my-bucket/data/commits.json.gz',
    auto_detect=true,
    format='auto'
  )
```

### BigQuery to DuckDB

Assuming DuckDB is the default OLAP and there is a `bigquery.yaml` connector in the project:

```yaml
# models/orders.yaml
type: model

refresh:
  cron: 0 0 * * *

connector: bigquery
sql: |
  SELECT * FROM my_dataset.orders
  WHERE order_date >= DATE_SUB(CURRENT_DATE(), INTERVAL 90 DAY)
```

### Snowflake to DuckDB with dev data limit

Assuming DuckDB is the default OLAP and there is a `snowflake.yaml` connector in the project:

```yaml
# models/sales.yaml
type: model

refresh:
  cron: 0 6 * * *

connector: snowflake
sql: |
  SELECT * FROM staging.sales
  {{ if dev }} event_time >= '2025-01-01' AND event_time < '2025-02-01' {{ end }}
```

### MySQL to DuckDB

Assuming there is a `mysql.yaml` connector in the project:

```yaml
# models/users.yaml
type: model

refresh:
  cron: 0 * * * *

connector: mysql
sql: |
  SELECT
    id,
    email,
    created_at,
    status
  FROM users
  WHERE status = 'active'
```

### Local CSV file to DuckDB or Clickhouse

```yaml
# models/reference_data.yaml
type: model

connector: local_file
path: data/reference_data.csv
```

### HTTPS source (public Parquet file)

```yaml
# models/public_dataset.yaml
type: model

connector: https
uri: https://example.com/public/dataset.parquet
```

### Partition-based incremental S3 to DuckDB

Assuming DuckDB is the default OLAP and there is a `s3.yaml` connector in the project:

```yaml
# models/daily_events.yaml
type: model
incremental: true

partitions:
  glob:
    path: s3://my-bucket/events/year=*/month=*/day=*/*.parquet
    partition: directory

sql: SELECT * FROM read_parquet('{{ .partition.uri }}/*.parquet')
```

### Basic S3 to ClickHouse

Assuming Clickhouse is the default OLAP and there is an `s3.yaml` connector in the project:

```yaml
# models/events.yaml
type: model
materialize: true

sql: |
  SELECT * FROM s3(
    's3://my-bucket/events/year=*/month=*/day=*/*.parquet',
    '{{ .env.aws_access_key_id }}',
    '{{ .env.aws_secret_access_key }}',
    'Parquet'
  )

output:
  order_by: event_time
```

### Partition-based incremental S3 to ClickHouse

Assuming Clickhouse is the default OLAP and there is an `s3.yaml` connector in the project:

```yaml
# models/events.yaml
type: model
materialize: true
incremental: true

partitions:
  glob:
    connector: s3
    path: s3://my-bucket/events/year=*/month=*/day=*/*.parquet

sql: |
  SELECT * FROM s3(
    '{{ .partition.uri }}',
    '{{ .env.aws_access_key_id }}',
    '{{ .env.aws_secret_access_key }}',
    'Parquet'
  )

output:
  order_by: event_time
```

### ClickHouse with explicit column schema and indexes and TTL

Assuming Clickhouse is the default OLAP and there is an `s3.yaml` connector in the project:

```yaml
# models/impressions.yaml
type: model
materialize: true
incremental: true

partitions:
  glob:
    connector: s3
    path: s3://my-bucket/impressions/year=*/month=*/day=*/*.parquet
    partition: directory

sql: |
  SELECT
    '{{ .partition.uri }}' AS __partition,
    event_time,
    user_id,
    campaign_id,
    country,
    device_type,
    impressions,
    clicks,
    cost
  FROM s3(
    '{{ .partition.uri }}/*.parquet',
    '{{ .env.aws_access_key_id }}',
    '{{ .env.aws_secret_access_key }}',
    'Parquet'
  )

output:
  incremental_strategy: partition_overwrite
  partition_by: toYYYYMMDD(event_time)
  order_by: (event_time, user_id)
  ttl: event_time + INTERVAL 90 DAY DELETE
  columns: |
    (
      event_time DateTime,
      user_id LowCardinality(String),
      campaign_id LowCardinality(String),
      country LowCardinality(String),
      device_type LowCardinality(String),
      impressions UInt32,
      clicks UInt32,
      cost Float64,
      INDEX idx_campaign campaign_id TYPE bloom_filter GRANULARITY 4,
      INDEX idx_country country TYPE set(100) GRANULARITY 4
    )
```

### State-based incremental with merge strategy (upserts)

Assuming DuckDB is the default OLAP and there is a `bigquery.yaml` connector in the project:

```yaml
# models/users.yaml
type: model
incremental: true

connector: bigquery
sql: |
  SELECT * FROM users
  {{ if incremental }}
  WHERE updated_at > '{{ .state.max_updated_at }}'
  {{ end }}

state:
  sql: SELECT MAX(updated_at) as max_updated_at FROM users

output:
  incremental_strategy: merge
  unique_key: [user_id]
```

### Dev partitions for faster development

```yaml
# models/large_dataset.yaml
type: model
incremental: true

partitions:
  glob:
    path: s3://my-bucket/data/year=*/month=*/day=*/*.parquet
    partition: directory

dev:
  partitions:
    glob:
      path: s3://my-bucket/data/year=2025/month=01/day=01/*.parquet
      partition: directory

sql: SELECT * FROM read_parquet('{{ .partition.uri }}/*.parquet')
```

### Partition filtering with transform_sql

```yaml
# models/filtered_partitions.yaml
type: model
incremental: true

partitions:
  glob:
    path: s3://my-bucket/reports/y=*/m=*/d=*/h=*/*.parquet
    partition: directory
    transform_sql: |
      -- Only process partitions after a specific date
      SELECT uri, updated_on
      FROM {{ .table }}
      WHERE uri >= 's3://my-bucket/reports/y=2025/m=06/d=01/h=00'
      {{ if dev }}
      AND uri < 's3://my-bucket/reports/y=2025/m=07/d=01/h=00'
      {{ end }}
      ORDER BY uri ASC

sql: SELECT * FROM read_parquet('{{ .partition.uri }}/*.parquet')
```

### Long timeout for large data processing

```yaml
# models/large_historical.yaml
type: model
materialize: true
incremental: true
timeout: 72h

partitions:
  glob:
    path: s3://my-bucket/historical/year=*/*.parquet
    partition: directory

sql: SELECT * FROM read_parquet('{{ .partition.uri }}/*.parquet')
```

### Non-materialized derived model

```yaml
# models/enriched_orders.yaml
type: model
materialize: false

sql: |
  SELECT
    o.*,
    c.customer_name,
    c.customer_segment,
    p.product_category
  FROM {{ ref "orders" }} o
  LEFT JOIN {{ ref "customers" }} c ON o.customer_id = c.customer_id
  LEFT JOIN {{ ref "products" }} p ON o.product_id = p.product_id
```

### DuckDB model reading from S3 with CSV options

Assuming DuckDB is the default OLAP and there is a `s3.yaml` connector in the project:

```yaml
# models/csv_import.yaml
type: model
materialize: true
connector: duckdb

sql: |
  SELECT * FROM read_csv(
    's3://my-bucket/data/*.csv.gz',
    auto_detect=true,
    header=true,
    ignore_errors=true,
    union_by_name=true,
    all_varchar=true
  )
```

### SQL-based partitions (date range)

```yaml
# models/events_raw.yaml
type: model
incremental: true

partitions:
  connector: snowflake
  sql: SELECT DISTINCT date_trunc('day', event_time) AS day FROM events ORDER BY day

connector: snowflake
sql: |
  SELECT * FROM events
  WHERE date_trunc('day', event_time) = '{{ .partition.day }}'

output:
  incremental_strategy: partition_overwrite
```

### Change mode for incremental models

```yaml
# models/append_only.yaml
type: model
incremental: true
change_mode: patch  # Switch to new logic without reprocessing historical data

partitions:
  glob:
    path: s3://my-bucket/transactions/year=*/month=*/day=*/*.parquet
    partition: directory

sql: |
  SELECT
    event_time,
    country,
    costs_usd + profit_usd + tax_usd AS value_usd
  FROM read_parquet('{{ .partition.uri }}/*.parquet')
```
