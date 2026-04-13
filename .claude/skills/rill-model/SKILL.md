---
name: rill-model
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
- `materialize: false`: Creates a SQL view. The query re-executes on every access. Only suitable for lightweight transformations where input and output connectors are the same that never reference external data.

If `materialize` is omitted, it defaults to `true` for all cross-connector models and `false` for single-connector models (i.e. where the input and output connector is the same).

In model files with a `.sql` extension, you can materialize by putting this on the first line of the file:
```sql
-- @materialize: true
```

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

By default, `glob:` matches files only, but you can pass `partition: directory` to have it emit leaf directory names instead. When you use `partition: directory`, the partition's URI will not include an asterisk, so you have to append that in the SQL query, e.g. `{{ .partition.uri }}/*.parquet`.

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
- Never try to "guess" a dev partition, use introspection tools like `list_bucket_files` (if available) to find a real directory you can use

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
The default `if_error_matches` values are:
`".*OvercommitTracker.*"`, `".*Bad Gateway.*"`, `".*Timeout.*"`, and `".*Connection refused.*"`.
If `retry.if_error_matches` is set, it overrides these defaults instead of adding to them.

```yaml
retry:
  attempts: 5
  delay: 10s
  exponential_backoff: true
  if_error_matches:
    - ".*OvercommitTracker.*"
    - ".*Bad Gateway.*"
    - ".*Timeout.*"
    - ".*Connection refused.*"
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

## Reference documentation

Here is a full JSON schema for the model syntax:

```
allOf:
    - properties:
        change_mode:
            description: Configure how changes to the model specifications are applied (optional). 'reset' will drop and recreate the model automatically, 'manual' will require a manual full or incremental refresh to apply changes, and 'patch' will switch to the new logic without re-processing historical data (only applies for incremental models).
            enum:
                - reset
                - manual
                - patch
            type: string
        connector:
            const: connector
            description: Refers to the resource type and is needed if setting an explicit OLAP engine. IE `clickhouse`
            type: string
        create_secrets_from_connectors:
            description: List of connector names for which temporary secrets should be created before executing the SQL. This allows DuckDB-based models to access cloud storage (S3, GCS, Azure) using credentials from named connectors.
            examples:
                - create_secrets_from_connectors: my_s3_connector
                - create_secrets_from_connectors:
                    - my_s3_connector
                    - my_other_s3_connector
            items:
                type: string
            type:
                - string
                - array
        incremental:
            description: whether incremental modeling is required (optional)
            type: boolean
        materialize:
            description: models will be materialized in olap
            type: boolean
        output:
            allOf:
                - if:
                    properties:
                        connector:
                            const: clickhouse
                    required:
                        - connector
                    title: Additional properties for `output` when `connector` is `clickhouse`
                  then:
                    properties:
                        columns:
                            description: Column names and types. Can also include indexes. If unspecified, detected from the query.
                            type: string
                        dictionary_source_password:
                            description: Password for the dictionary source user.
                            type: string
                        dictionary_source_user:
                            description: User for accessing the source dictionary table (used if type is DICTIONARY).
                            type: string
                        distributed_settings:
                            description: Settings for distributed table.
                            type: string
                        distributed_sharding_key:
                            description: Sharding key for distributed table.
                            type: string
                        engine:
                            description: Table engine to use. Default is MergeTree
                            type: string
                        engine_full:
                            description: Full engine definition in SQL format. Can include partition keys, order, TTL, etc.
                            type: string
                        order_by:
                            description: ORDER BY clause.
                            type: string
                        partition_by:
                            description: Partition BY clause.
                            type: string
                        primary_key:
                            description: PRIMARY KEY clause.
                            type: string
                        query_settings:
                            description: Settings used in insert/create table as select queries.
                            type: string
                        sample_by:
                            description: SAMPLE BY clause.
                            type: string
                        table_settings:
                            description: Table-specific settings.
                            type: string
                        ttl:
                            description: TTL settings for the table or columns.
                            type: string
                        type:
                            description: Type to materialize the model into. Can be 'TABLE', 'VIEW' or 'DICTIONARY'
                            enum:
                                - TABLE
                                - VIEW
                                - DICTIONARY
                            type: string
            description: to define the properties of output
            properties:
                connector:
                    description: Refers to the connector type for the output table. Can be `clickhouse` or `duckdb` and their named connector
                    type: string
                incremental_strategy:
                    description: Strategy to use for incremental updates. Can be 'append', 'merge' or 'partition_overwrite'
                    enum:
                        - append
                        - merge
                        - partition_overwrite
                    type: string
                materialize:
                    description: Whether to materialize the model as a table or view
                    type: boolean
                partition_by:
                    description: Column or expression to partition the table by
                    type: string
                table:
                    description: Name of the output table. If not specified, the model name is used.
                    type: string
                unique_key:
                    description: List of columns that uniquely identify a row for merge strategy
                    items:
                        type: string
                    type: array
            type: object
        partitions:
            $ref: '#/definitions/data_properties'
            description: Refers to the how your data is partitioned, cannot be used with state. (optional)
            examples:
                - partitions:
                    glob: gcs://my_bucket/y=*/m=*/d=*/*.parquet
                - partitions:
                    connector: duckdb
                    sql: SELECT range AS num FROM range(0,10)
        partitions_concurrency:
            description: Refers to the number of concurrent partitions that can be read at the same time (optional).
            type: integer
        partitions_watermark:
            description: Refers to a customizable timestamp that can be set to check if an object has been updated (optional).
            type: string
        post_exec:
            description: Refers to a SQL query that is run after the main query, available for DuckDB-based and ClickHouse-based models. (optional). Ensure post_exec queries are idempotent. Use IF EXISTS statements when applicable.
            examples:
                - post_exec: DETACH DATABASE IF EXISTS postgres_db
            type: string
        pre_exec:
            description: Refers to SQL queries to run before the main query, available for DuckDB-based and ClickHouse-based models. (optional). Ensure pre_exec queries are idempotent. Use IF NOT EXISTS statements when applicable.
            examples:
                - pre_exec: ATTACH IF NOT EXISTS 'dbname=postgres host=localhost port=5432 user=postgres password=postgres' AS postgres_db (TYPE POSTGRES)
            type: string
        refresh:
            $ref: '#/definitions/schedule_properties'
            description: Specifies the refresh schedule that Rill should follow to re-ingest and update the underlying model data
            examples:
                - refresh:
                    cron: '* * * * *'
        retry:
            description: Refers to the retry configuration for the model. (optional)
            examples:
                - retry:
                    attempts: 5
                    delay: 10s
                    exponential_backoff: true
                    if_error_matches:
                        - .*OvercommitTracker.*
                        - .*Bad Gateway.*
                        - .*Timeout.*
                        - .*Connection refused.*
            properties:
                attempts:
                    description: The number of attempts to retry the model.
                    type: integer
                delay:
                    description: The delay between attempts.
                    type: string
                exponential_backoff:
                    description: Whether to use exponential backoff.
                    type: boolean
                if_error_matches:
                    description: The error messages to match.
                    items:
                        type: string
                    type: array
            type: object
        sql:
            description: Raw SQL query to run against source
            type: string
        stage:
            additionalProperties: true
            description: in the case of staging models, where an input source does not support direct write to the output and a staging table is required
            examples:
                - stage:
                    connector: s3
                    path: s3://my_bucket/my_staging_table
            properties:
                connector:
                    description: Refers to the connector type for the staging table
                    type: string
                path:
                    description: Refers to the path to the staging table
                    type: string
            required:
                - connector
            type: object
        state:
            $ref: '#/definitions/data_properties'
            description: Refers to the explicitly defined state of your model, cannot be used with partitions (optional)
            examples:
                - state:
                    sql: SELECT MAX(date) as max_date
        tests:
            description: Define data quality tests for the model. Each test must have a `name` and either an `assert` expression or a `sql` query. An `assert` test passes when no rows violate the condition. A `sql` test passes when the query returns zero rows.
            examples:
                - tests:
                    - assert: revenue >= 0
                      name: assert_positive_revenue
                    - assert: id IS NOT NULL
                      name: no_null_ids
                - tests:
                    - name: row_count_check
                      sql: SELECT 'fail' WHERE (SELECT COUNT(*) FROM my_model) = 0
            items:
                properties:
                    assert:
                        description: A SQL boolean expression applied to each row of the model. The test passes if no rows violate the condition (i.e., all rows satisfy `assert`). Cannot be combined with `sql`.
                        type: string
                    connector:
                        description: The connector to use when executing the test query. Defaults to the model's connector.
                        type: string
                    name:
                        description: A unique name for the test.
                        type: string
                    sql:
                        description: A SQL query that returns rows representing test failures. The test passes if the query returns zero rows. Cannot be combined with `assert`.
                        type: string
                required:
                    - name
                type: object
            type: array
        timeout:
            description: The maximum time to wait for model ingestion
            type: string
        type:
            const: model
            description: Refers to the resource type and must be `model`
            type: string
      required:
        - type
        - sql
      title: Properties
      type: object
    - $ref: '#/definitions/common_properties'
    - allOf:
        - if:
            properties:
                connector:
                    const: clickhouse
            required:
                - connector
            title: Additional properties for `output` when `connector` is `clickhouse`
          then:
            properties:
                columns:
                    description: Column names and types. Can also include indexes. If unspecified, detected from the query.
                    type: string
                dictionary_source_password:
                    description: Password for the dictionary source user.
                    type: string
                dictionary_source_user:
                    description: User for accessing the source dictionary table (used if type is DICTIONARY).
                    type: string
                distributed_settings:
                    description: Settings for distributed table.
                    type: string
                distributed_sharding_key:
                    description: Sharding key for distributed table.
                    type: string
                engine:
                    description: Table engine to use. Default is MergeTree
                    type: string
                engine_full:
                    description: Full engine definition in SQL format. Can include partition keys, order, TTL, etc.
                    type: string
                order_by:
                    description: ORDER BY clause.
                    type: string
                partition_by:
                    description: Partition BY clause.
                    type: string
                primary_key:
                    description: PRIMARY KEY clause.
                    type: string
                query_settings:
                    description: Settings used in insert/create table as select queries.
                    type: string
                sample_by:
                    description: SAMPLE BY clause.
                    type: string
                table_settings:
                    description: Table-specific settings.
                    type: string
                ttl:
                    description: TTL settings for the table or columns.
                    type: string
                type:
                    description: Type to materialize the model into. Can be 'TABLE', 'VIEW' or 'DICTIONARY'
                    enum:
                        - TABLE
                        - VIEW
                        - DICTIONARY
                    type: string
      required:
        - type
        - sql
      type: object
    - $ref: '#/definitions/common_properties'
    - allOf:
        - if:
            properties:
                connector:
                    const: athena
            required:
                - connector
            title: Additional properties when `connector` is `athena` or [named connector](./connectors#athena) for athena
          then:
            $ref: '#/definitions/models/definitions/athena'
        - if:
            properties:
                connector:
                    const: azure
            required:
                - connector
            title: Additional properties when `connector` is `azure` or [named connector](./connectors#azure) of azure
          then:
            $ref: '#/definitions/models/definitions/azure'
        - if:
            properties:
                connector:
                    const: bigquery
            required:
                - connector
            title: Additional properties when `connector` is `bigquery` or [named connector](./connectors#bigquery) of bigquery
          then:
            $ref: '#/definitions/models/definitions/bigquery'
        - if:
            properties:
                connector:
                    const: duckdb
            required:
                - connector
            title: Additional properties when `connector` is `duckdb` or [named connector](./connectors#duckdb) of duckdb
          then:
            $ref: '#/definitions/models/definitions/duckdb'
        - if:
            properties:
                connector:
                    const: gcs
            required:
                - connector
            title: Additional properties when `connector` is `gcs` or [named connector](./connectors#gcs) of gcs
          then:
            $ref: '#/definitions/models/definitions/gcs'
        - if:
            properties:
                connector:
                    const: local_file
            required:
                - connector
            title: Additional properties when `connector` is `local_file` or [named connector](/developers/build/connectors/data-source/local-file) of local_file
          then:
            $ref: '#/definitions/models/definitions/local_file'
        - if:
            properties:
                connector:
                    const: redshift
            required:
                - connector
            title: Additional properties when `connector` is `redshift` or [named connector](./connectors#redshift) of redshift
          then:
            $ref: '#/definitions/models/definitions/redshift'
        - if:
            properties:
                connector:
                    const: s3
            required:
                - connector
            title: Additional properties when `connector` is `s3` or [named connector](./connectors#s3) of s3
          then:
            $ref: '#/definitions/models/definitions/s3'
        - if:
            properties:
                connector:
                    const: salesforce
            required:
                - connector
            title: Additional properties when `connector` is `salesforce` or [named connector](./connectors#salesforce) of salesforce
          then:
            $ref: '#/definitions/models/definitions/salesforce'
      type: object
definitions:
    athena:
        properties:
            output_location:
                description: Output location for query results in S3.
                type: string
            region:
                description: AWS region to connect to Athena and the output location.
                type: string
            workgroup:
                description: AWS Athena workgroup to use for queries.
                type: string
        type: object
    azure:
        properties:
            account:
                description: Account identifier
                type: string
            batch_size:
                description: Size of a batch (e.g., '100MB')
                type: string
            extract:
                additionalProperties: true
                description: Arbitrary key-value pairs for extraction settings
                type: object
            glob:
                description: Settings related to glob file matching.
                properties:
                    max_objects_listed:
                        description: Maximum number of objects listed in glob
                        type: integer
                    max_objects_matched:
                        description: Maximum number of objects matched by glob
                        type: integer
                    max_total_size:
                        description: Maximum total size (in bytes) matched by glob
                        type: integer
                    page_size:
                        description: Page size for glob listing
                        type: integer
                type: object
            path:
                description: Path to the source
                type: string
            uri:
                description: Source URI
                type: string
        type: object
    bigquery:
        properties:
            project_id:
                description: ID of the BigQuery project.
                type: string
        type: object
    duckdb:
        properties:
            create_secrets_from_connectors:
                description: List of connector names for which temporary secrets should be created before executing the SQL.
                type:
                    - string
                    - array
            format:
                description: Format of the data source (e.g., csv, json, parquet).
                type: string
            path:
                description: Path to the data source.
                type: string
            post_exec:
                description: refers to a SQL query that is run after the main query, available for DuckDB-based and ClickHouse-based models. _(optional)_. Ensure `post_exec` queries are idempotent. Use `IF EXISTS` statements when applicable.
                examples:
                    - post_exec: DETACH DATABASE IF EXISTS postgres_db
                      pre_exec: ATTACH IF NOT EXISTS 'dbname=postgres host=localhost port=5432 user=postgres password=postgres' AS postgres_db (TYPE POSTGRES);
                      sql: SELECT * FROM postgres_query('postgres_db', 'SELECT * FROM USERS')
                type: string
            pre_exec:
                description: refers to SQL queries to run before the main query, available for DuckDB-based and ClickHouse-based models. _(optional)_. Ensure `pre_exec` queries are idempotent. Use `IF NOT EXISTS` statements when applicable.
                type: string
        type: object
    gcs:
        properties:
            batch_size:
                description: Size of a batch (e.g., '100MB')
                type: string
            extract:
                additionalProperties: true
                description: key-value pairs for extraction settings
                type: object
            glob:
                description: Settings related to glob file matching.
                properties:
                    max_objects_listed:
                        description: Maximum number of objects listed in glob
                        type: integer
                    max_objects_matched:
                        description: Maximum number of objects matched by glob
                        type: integer
                    max_total_size:
                        description: Maximum total size (in bytes) matched by glob
                        type: integer
                    page_size:
                        description: Page size for glob listing
                        type: integer
                type: object
            path:
                description: Path to the source
                type: string
            uri:
                description: Source URI
                type: string
        type: object
    local_file:
        properties:
            format:
                description: Format of the data source (e.g., csv, json, parquet).
                type: string
            invalidate_on_change:
                description: When true, the model will be invalidated and re-processed if the source file changes.
                type: boolean
            path:
                description: Path to the data source.
                type: string
        type: object
    redshift:
        properties:
            cluster_identifier:
                description: Identifier of the Redshift cluster.
                type: string
            database:
                description: Name of the Redshift database.
                type: string
            output_location:
                description: S3 location where query results are stored.
                type: string
            region:
                description: AWS region of the Redshift deployment.
                type: string
            role_arn:
                description: ARN of the IAM role to assume for Redshift access.
                type: string
            workgroup:
                description: Redshift Serverless workgroup to use.
                type: string
        type: object
    s3:
        properties:
            batch_size:
                description: Size of a batch (e.g., '100MB')
                type: string
            endpoint:
                description: AWS Endpoint
                type: string
            extract:
                additionalProperties: true
                description: key-value pairs for extraction settings
                type: object
            glob:
                description: Settings related to glob file matching.
                properties:
                    max_objects_listed:
                        description: Maximum number of objects listed in glob
                        type: integer
                    max_objects_matched:
                        description: Maximum number of objects matched by glob
                        type: integer
                    max_total_size:
                        description: Maximum total size (in bytes) matched by glob
                        type: integer
                    page_size:
                        description: Page size for glob listing
                        type: integer
                type: object
            path:
                description: Path to the source
                type: string
            region:
                description: AWS region
                type: string
            uri:
                description: Source URI
                type: string
        type: object
    salesforce:
        properties:
            queryAll:
                description: Whether to include deleted and archived records in the query (uses queryAll API).
                type: boolean
            sobject:
                description: Salesforce object (e.g., Account, Contact) targeted by the query.
                type: string
            soql:
                description: SOQL query to execute against the Salesforce instance.
                type: string
        type: object
description: |4
    This file is used to define YAML models. For more information on our SQL models, see the [SQL models](/developers/build/models/) documentation.
    :::tip

    Both regular models and source models can use the Model YAML specification described on this page. While [SQL models](/developers/build/models) are perfect for simple transformations, Model YAML files provide advanced capabilities for complex data processing scenarios.

    **When to use Model YAML:**
    - **Partitions** - Optimize performance with data partitioning strategies
    - **Incremental models** - Process only new or changed data efficiently
    - **Pre/post execution hooks** - Run custom logic before or after model execution
    - **Staging** - Create intermediate tables for complex transformations
    - **Output configuration** - Define specific output formats and destinations

    Model YAML files give you fine-grained control over how your data is processed and transformed, making them ideal for production workloads and complex analytics pipelines.

    :::
examples:
    - connector: bigquery
      incremental: true
      output:
        connector: duckdb
      sql: "SELECT ... FROM events \n  {{ if incremental }} \n      WHERE event_time > '{{.state.max_date}}' \n  {{end}}\n"
      state:
        sql: SELECT MAX(date) as max_date
      type: model
    - output:
        connector: duckdb
        incremental_strategy: append
      partitions:
        glob:
            connector: gcs
            path: gs://rilldata-public/github-analytics/Clickhouse/2025/*/commits_*.parquet
      sql: SELECT * FROM read_parquet('{{ .partition.uri }}')
      type: model
    - incremental: true
      output:
        connector: duckdb
        incremental_strategy: append
      partitions:
        glob:
            partition: directory
            path: gs://rilldata-public/github-analytics/Clickhouse/2025/*/*
      refresh:
        cron: 0 8 * * *
      sql: "SELECT * \n  FROM read_parquet('gs://rilldata-public/{{ .partition.path }}/commits_*.parquet') \n  WHERE '{{ .partition.path }}' IS NOT NULL\n"
      type: model
    - connector: snowflake
      incremental: true
      output:
        connector: clickhouse
      partitions:
        connector: duckdb
        sql: SELECT range as day FROM range(TIMESTAMPTZ '2024-01-01', now(), INTERVAL 1 DAY)
      sql: SELECT * FROM events WHERE date_trunc('day', event_time) = '{{ .partition.day }}'
      stage:
        connector: s3
        path: s3://bucket/temp-data
      type: model
id: models
title: Models YAML
type: object
```