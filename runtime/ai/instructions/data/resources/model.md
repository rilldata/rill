---
description: Detailed instructions and examples for developing model resources in Rill
---

# Instructions for developing a model in Rill

## Introduction

Models are resources that specify ETL or transformation logic, outputting a tabular dataset to one of the project's connectors. They are typically found near the root of the project's DAG, referencing only connectors and other models.

By default, models output data as a table with the same name as the model in the project's default OLAP connector. The core of a model is usually a `SELECT` SQL statement, which Rill executes as `CREATE TABLE <name> AS <SELECT statement>`.

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

### State-based incremental models

Use the `state:` property to track a high-water mark (e.g., the maximum timestamp already processed):

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
  sql: SELECT MAX(event_time) as max_time FROM {{ ref "self" }}
```

The `{{ if incremental }}` block ensures the filter only applies during incremental runs, not during the initial full load. The `state` query is evaluated and stored only after the first non-incremental run. Note that the `state` query runs against the project's default OLAP connector (e.g. DuckDB), which is the default output location for models.

### Partition-based incremental models

Use the `partitions:` property to define explicit data partitions. Combined with `incremental: true`, Rill tracks which partitions have been processed:

```yaml
type: model
incremental: true

partitions:
  glob:
    connector: s3
    path: s3://bucket/data/year=*/month=*/day=*/*.parquet

sql: SELECT * FROM read_parquet('{{ .partition.uri }}')
```

### Incremental strategies

The `output.incremental_strategy:` property controls how new data is merged with existing data:

- `partition_overwrite`: Entire partitions are replaced. This is the default strategy.
- `merge`: New rows are merged based on `output.unique_key`. Use for upsert semantics.
- `append`: New rows are appended to the table. Generally avoid this since retries will lead to duplicate data.

The `partition_overwrite` strategy overwrites partitions based on the column(s) described in `output.partition_by`. If that is not explicitly specified, it injects a column `__rill_partition` into the table and partitions by it (default).

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
- `{{ .partition.path }}`: Path portion without the connector prefix

### SQL-based partitions

Generate partitions using a SQL query:

```yaml
partitions:
  connector: duckdb
  sql: SELECT range AS day FROM range(TIMESTAMPTZ '2024-01-01', now(), INTERVAL 1 DAY)
```

### Dev partitions

During development, limit processing to a subset of partitions to speed up iteration:

```yaml
dev:
  partitions:
    glob:
      path: s3://bucket/data/year=2025/month=12/day=01/*.parquet
```

## Referencing other models

Use `{{ ref "model_name" }}` to reference parent models in SQL statements that use templating:

```yaml
sql: SELECT * FROM {{ ref "events_raw" }} WHERE country = 'US'
```

**Note:** If your SQL statement contains no other templating, the `ref` function is optional for DuckDB SQL snippets; Rill can in that case invoke DuckDB's SQL parser to automatically detects model references. This does not apply for non-DuckDB SQL models.

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

- **Model references:** When SQL contains no other templating, `{{ ref "model" }}` is optional; Rill uses DuckDB's SQL parser to detect references.
- **Connector secrets:** By default, all compatible connectors are automatically mounted as DuckDB secrets. Use `create_secrets_from_connectors:` to explicitly control which connectors are available.
- **Pre/post execution:** Use `pre_exec:` and `post_exec:` for setup and teardown queries (e.g., attaching external databases). Some legacy projects configure DuckDB secrets here, but that is usually not necessary now given the automatic secret creation referenced above.
- **Cloud storage paths:** DuckDB can read directly from S3 (`s3://`) and GCS (`gs://`) paths in `read_parquet()`, `read_csv()`, and `read_json()` functions.
- **CSV options:** When reading CSV files, useful options include `auto_detect=true`, `header=true`, `ignore_errors=true`, `union_by_name=true`, and `all_varchar=true` for handling inconsistent schemas.
- **JSON files:** Use `read_json()` with `auto_detect=true` and `format='auto'` for flexible JSON ingestion, including gzipped files.

### ClickHouse

- **S3 credentials:** When using the `s3()` function, reference `.env` values directly using templating since ClickHouse lacks integrated secret management:
  ```yaml
  sql: SELECT * FROM s3('{{ .env.s3_bucket }}/path/*.parquet', '{{ .env.aws_access_key }}', '{{ .env.aws_secret_key }}', 'Parquet')
  ```
- **Required order_by:** The `output:` section must always include an `order_by` clause for materialized ClickHouse tables.
- **Performance indexes:** For models powering metrics views, consider adding indexes via `output.columns` to improve query performance. Common index types include `bloom_filter` for high-cardinality columns and `set(N)` for low-cardinality columns.
- **TTL for data retention:** Use `output.ttl` to automatically expire old data and prevent unbounded growth in incremental models.
- **LowCardinality types:** Use `LowCardinality(String)` for string columns with limited distinct values (e.g., country, device_type, status) to improve storage and query performance.
- **ReplicatedMergeTree:** For production deployments, use `ReplicatedMergeTree()` engine. A common pattern is to conditionally use `MergeTree()` in dev and `ReplicatedMergeTree()` in prod via templating.
- **Projections:** Define projections in `output.columns` for pre-aggregated views that accelerate common query patterns.

### BigQuery

- **project_id:** Always specify `project_id` to indicate which GCP project contains the data.
- **run_in_dev:** Set `refresh.run_in_dev: true` if you want the model to refresh during local development; otherwise refreshes only happen in production.

### Snowflake

- **Dev data limits:** Use `{{ if dev }} LIMIT N {{ end }}` to limit data during development for faster iteration.
- **Prod-only refresh:** Use the `prod:` block to configure refresh schedules that only apply in production.

### Athena

- **Required configuration:** Athena models require `workgroup`, `region`, and `output_location` parameters.
- **Timestamp handling:** Use `from_iso8601_timestamp()` when comparing timestamps stored as strings in state queries.

### MySQL / Postgres

- **DSN configuration:** Use environment variable references for database connection strings: `dsn: "{{ .env.mysql_dsn }}"` or configure credentials in a connector YAML file.

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
models: 
    title: Models YAML
    id: models
    type: object
    description: |

      This file is used to define YAML models. For more information on our SQL models, see the [SQL models](/build/models/) documentation.
      :::tip

      Both regular models and source models can use the Model YAML specification described on this page. While [SQL models](/build/models) are perfect for simple transformations, Model YAML files provide advanced capabilities for complex data processing scenarios.

      **When to use Model YAML:**
      - **Partitions** - Optimize performance with data partitioning strategies
      - **Incremental models** - Process only new or changed data efficiently
      - **Pre/post execution hooks** - Run custom logic before or after model execution
      - **Staging** - Create intermediate tables for complex transformations
      - **Output configuration** - Define specific output formats and destinations

      Model YAML files give you fine-grained control over how your data is processed and transformed, making them ideal for production workloads and complex analytics pipelines.

      :::
    allOf:
      - title: Properties
        type: object
        properties:
          type:
            type: string
            const: model
            description: Refers to the resource type and must be `model`
          refresh:
            $ref: '#/definitions/schedule_properties'
            description: Specifies the refresh schedule that Rill should follow to re-ingest and update the underlying model data 
            examples: 
              - refresh:
                  cron: "* * * * *"
          connector:
            type: string
            const: connector
            description: Refers to the resource type and is needed if setting an explicit OLAP engine. IE `clickhouse`
          sql:
            type: string
            description: Raw SQL query to run against source
          pre_exec:
            type: string
            description: Refers to SQL queries to run before the main query, available for DuckDB-based models. (optional). Ensure pre_exec queries are idempotent. Use IF NOT EXISTS statements when applicable.
            examples: 
              - pre_exec: ATTACH IF NOT EXISTS 'dbname=postgres host=localhost port=5432 user=postgres password=postgres' AS postgres_db (TYPE POSTGRES)
          post_exec:
            type: string
            description: Refers to a SQL query that is run after the main query, available for DuckDB-based models. (optional). Ensure post_exec queries are idempotent. Use IF EXISTS statements when applicable.
            examples: 
              - post_exec: DETACH DATABASE IF EXISTS postgres_db
          retry:
            type: object
            description: Refers to the retry configuration for the model. (optional)
            properties:
              attempts:
                type: integer
                description: The number of attempts to retry the model.
              delay:
                type: string
                description: The delay between attempts.
              exponential_backoff:
                type: boolean
                description: Whether to use exponential backoff.
              if_error_matches:
                type: array
                items:
                  type: string
                description: The error messages to match.
            examples:
              - retry:
                  attempts: 5
                  delay: 10s
                  exponential_backoff: true
                  if_error_matches:
                    - ".*OvercommitTracker.*"
                    - ".*Timeout.*"
                    - ".*Bad Gateway.*"
          timeout:
            type: string
            description: The maximum time to wait for model ingestion
          incremental:
            type: boolean
            description: whether incremental modeling is required (optional)
          change_mode:
            type: string
            enum:
              - reset
              - manual
              - patch
            description: Configure how changes to the model specifications are applied (optional). 'reset' will drop and recreate the model automatically, 'manual' will require a manual full or incremental refresh to apply changes, and 'patch' will switch to the new logic without re-processing historical data (only applies for incremental models).
          state:
            $ref: '#/definitions/data_properties'
            description:  Refers to the explicitly defined state of your model, cannot be used with partitions (optional)
            examples: 
              - state:
                  sql: SELECT MAX(date) as max_date
          partitions:
            $ref: '#/definitions/data_properties'
            description: Refers to the how your data is partitioned, cannot be used with state. (optional)
            examples: 
              - partitions:
                  glob: gcs://my_bucket/y=*/m=*/d=*/*.parquet
              - partitions:
                  connector: duckdb
                  sql: SELECT range AS num FROM range(0,10)
          materialize:
            type: boolean
            description: models will be materialized in olap
          partitions_watermark:
            type: string
            description: Refers to a customizable timestamp that can be set to check if an object has been updated (optional).
          partitions_concurrency:
            type: integer
            description: Refers to the number of concurrent partitions that can be read at the same time (optional).
          stage:
            type: object
            properties:
              connector:
                type: string
                description: Refers to the connector type for the staging table
              path:
                type: string
                description: Refers to the path to the staging table

            required:
              - connector
            description: in the case of staging models, where an input source does not support direct write to the output and a staging table is required
            examples: 
              - stage:
                  connector: s3
                  path: s3://my_bucket/my_staging_table

            additionalProperties: true
          output:
            type: object
            description: to define the properties of output
            properties:
              table:
                type: string
                description: Name of the output table. If not specified, the model name is used.
              materialize:
                type: boolean
                description: Whether to materialize the model as a table or view
              connector:
                type: string
                description: Refers to the connector type for the output table. Can be `clickhouse` or `duckdb` and their named connector 
              incremental_strategy:
                type: string
                enum:
                  - append
                  - merge
                  - partition_overwrite
                description: Strategy to use for incremental updates. Can be 'append', 'merge' or 'partition_overwrite'
              unique_key:
                type: array
                items:
                  type: string
                description: List of columns that uniquely identify a row for merge strategy
              partition_by:
                type: string
                description: Column or expression to partition the table by
            allOf:
              - if:
                  title: Additional properties for `output` when `connector` is `clickhouse`
                  properties:
                    connector:
                      const: clickhouse
                  required:
                    - connector
                then:
                  properties:
                    type:
                      type: string
                      description: Type to materialize the model into. Can be 'TABLE', 'VIEW' or 'DICTIONARY'
                      enum:
                        - TABLE
                        - VIEW
                        - DICTIONARY
                    columns:
                      type: string
                      description: Column names and types. Can also include indexes. If unspecified, detected from the query.
                    engine_full:
                      type: string
                      description: Full engine definition in SQL format. Can include partition keys, order, TTL, etc.
                    engine:
                      type: string
                      description: Table engine to use. Default is MergeTree
                    order_by:
                      type: string
                      description: ORDER BY clause.
                    partition_by:
                      type: string
                      description: Partition BY clause.
                    primary_key:
                      type: string
                      description: PRIMARY KEY clause.
                    sample_by:
                      type: string
                      description: SAMPLE BY clause.
                    ttl:
                      type: string
                      description: TTL settings for the table or columns.
                    table_settings:
                      type: string
                      description: Table-specific settings.
                    query_settings:
                      type: string
                      description: Settings used in insert/create table as select queries.
                    distributed_settings:
                      type: string
                      description: Settings for distributed table.
                    distributed_sharding_key:
                      type: string
                      description: Sharding key for distributed table.
                    dictionary_source_user:
                      type: string
                      description: User for accessing the source dictionary table (used if type is DICTIONARY).
                    dictionary_source_password:
                      type: string
                      description: Password for the dictionary source user.
        required:
          - type
          - sql
      - $ref: '#/definitions/common_properties'
      - type: object
        allOf:
              - if:
                  title: Additional properties for `output` when `connector` is `clickhouse`
                  properties:
                    connector:
                      const: clickhouse
                  required:
                    - connector
                then:
                  properties:
                    type:
                      type: string
                      description: Type to materialize the model into. Can be 'TABLE', 'VIEW' or 'DICTIONARY'
                      enum:
                        - TABLE
                        - VIEW
                        - DICTIONARY
                    columns:
                      type: string
                      description: Column names and types. Can also include indexes. If unspecified, detected from the query.
                    engine_full:
                      type: string
                      description: Full engine definition in SQL format. Can include partition keys, order, TTL, etc.
                    engine:
                      type: string
                      description: Table engine to use. Default is MergeTree
                    order_by:
                      type: string
                      description: ORDER BY clause.
                    partition_by:
                      type: string
                      description: Partition BY clause.
                    primary_key:
                      type: string
                      description: PRIMARY KEY clause.
                    sample_by:
                      type: string
                      description: SAMPLE BY clause.
                    ttl:
                      type: string
                      description: TTL settings for the table or columns.
                    table_settings:
                      type: string
                      description: Table-specific settings.
                    query_settings:
                      type: string
                      description: Settings used in insert/create table as select queries.
                    distributed_settings:
                      type: string
                      description: Settings for distributed table.
                    distributed_sharding_key:
                      type: string
                      description: Sharding key for distributed table.
                    dictionary_source_user:
                      type: string
                      description: User for accessing the source dictionary table (used if type is DICTIONARY).
                    dictionary_source_password:
                      type: string
                      description: Password for the dictionary source user.
        required:
          - type
          - sql
      - $ref: '#/definitions/common_properties'
      - type: object
        allOf:
          - if:
              title: Additional properties when `connector` is `athena` or [named connector](./connectors#athena) for athena
              properties:
                connector:
                  const: athena
              required:
                - connector
            then:
              $ref: '#/definitions/models/definitions/athena'
          - if:
              title: Additional properties when `connector` is `azure` or [named connector](./connectors#azure) of azure
              properties:
                connector:
                  const: azure
              required:
                - connector
            then:
              $ref: '#/definitions/models/definitions/azure'
          - if:
              title: Additional properties when `connector` is `bigquery` or [named connector](./connectors#bigquery) of bigquery
              properties:
                connector:
                  const: bigquery
              required:
                - connector
            then:
              $ref: '#/definitions/models/definitions/bigquery'
          - if:
              title: Additional properties when `connector` is `duckdb` or [named connector](./connectors#duckdb) of duckdb
              properties:
                connector:
                  const: duckdb
              required:
                - connector
            then:
              $ref: '#/definitions/models/definitions/duckdb'
          - if:
              title: Additional properties when `connector` is `gcs` or [named connector](./connectors#gcs) of gcs
              properties:
                connector:
                  const: gcs
              required:
                - connector
            then:
              $ref: '#/definitions/models/definitions/gcs'
          - if:
              title: Additional properties when `connector` is `local_file` or [named connector](/build/connectors/data-source/local-file) of local_file
              properties:
                connector:
                  const: local_file
              required:
                - connector
            then:
              $ref: '#/definitions/models/definitions/local_file'
          - if:
              title: Additional properties when `connector` is `redshift` or [named connector](./connectors#redshift) of redshift
              properties:
                connector:
                  const: redshift
              required:
                - connector
            then:
              $ref: '#/definitions/models/definitions/redshift'
          - if:
              title: Additional properties when `connector` is `s3` or [named connector](./connectors#s3) of s3
              properties:
                connector:
                  const: s3
              required:
                - connector
            then:
              $ref: '#/definitions/models/definitions/s3'
          - if:
              title: Additional properties when `connector` is `salesforce` or [named connector](./connectors#salesforce) of salesforce
              properties:
                connector:
                  const: salesforce
              required:
                - connector
            then:
              $ref: '#/definitions/models/definitions/salesforce'
    definitions:
      athena:
        type: object
        properties:
          output_location:
            type: string
            description: Output location for query results in S3.
          workgroup:
            type: string
            description: AWS Athena workgroup to use for queries.
          region:
            type: string
            description: AWS region to connect to Athena and the output location.
      azure:
        type: object
        properties:
          path:
            type: string
            description: Path to the source
          account:
            type: string
            description: Account identifier
          uri:
            type: string
            description: Source URI
          extract:
            type: object
            description: Arbitrary key-value pairs for extraction settings
            additionalProperties: true
          glob:
            type: object
            description: Settings related to glob file matching.
            properties:
              max_total_size:
                type: integer
                description: Maximum total size (in bytes) matched by glob
              max_objects_matched:
                type: integer
                description: Maximum number of objects matched by glob
              max_objects_listed:
                type: integer
                description: Maximum number of objects listed in glob
              page_size:
                type: integer
                description: Page size for glob listing
          batch_size:
            type: string
            description: 'Size of a batch (e.g., ''100MB'')'
      bigquery:
        type: object
        properties:
          project_id:
            type: string
            description: ID of the BigQuery project.
      duckdb:
        type: object
        properties:
          path:
            type: string
            description: Path to the data source.
          format:
            type: string
            description: 'Format of the data source (e.g., csv, json, parquet).'
          pre_exec:
            type: string
            description: 'refers to SQL queries to run before the main query, available for DuckDB-based models. _(optional)_. Ensure `pre_exec` queries are idempotent. Use `IF NOT EXISTS` statements when applicable.'
          post_exec:
            type: string
            description: 'refers to a SQL query that is run after the main query, available for DuckDB-based models. _(optional)_. Ensure `post_exec` queries are idempotent. Use `IF EXISTS` statements when applicable.'
            examples:
            - pre_exec: ATTACH IF NOT EXISTS 'dbname=postgres host=localhost port=5432 user=postgres password=postgres' AS postgres_db (TYPE POSTGRES);
              sql: SELECT * FROM postgres_query('postgres_db', 'SELECT * FROM USERS')
              post_exec: DETACH DATABASE IF EXISTS postgres_db 
          create_secrets_from_connectors:
            type:
              - string
              - array
            description: List of connector names for which temporary secrets should be created before executing the SQL.
      gcs:
        type: object
        properties:
          path:
            type: string
            description: Path to the source
          uri:
            type: string
            description: Source URI
          extract:
            type: object
            description: key-value pairs for extraction settings
            additionalProperties: true
          glob:
            type: object
            description: Settings related to glob file matching.
            properties:
              max_total_size:
                type: integer
                description: Maximum total size (in bytes) matched by glob
              max_objects_matched:
                type: integer
                description: Maximum number of objects matched by glob
              max_objects_listed:
                type: integer
                description: Maximum number of objects listed in glob
              page_size:
                type: integer
                description: Page size for glob listing
          batch_size:
            type: string
            description: 'Size of a batch (e.g., ''100MB'')'
      local_file:
        type: object
        properties:
          path:
            type: string
            description: Path to the data source.
          format:
            type: string
            description: 'Format of the data source (e.g., csv, json, parquet).'
      redshift:
        type: object
        properties:
          output_location:
            type: string
            description: S3 location where query results are stored.
          workgroup:
            type: string
            description: Redshift Serverless workgroup to use.
          database:
            type: string
            description: Name of the Redshift database.
          cluster_identifier:
            type: string
            description: Identifier of the Redshift cluster.
          role_arn:
            type: string
            description: ARN of the IAM role to assume for Redshift access.
          region:
            type: string
            description: AWS region of the Redshift deployment.
      s3:
        type: object
        properties:
          region:
            type: string
            description: AWS region
          endpoint:
            type: string
            description: AWS Endpoint
          path:
            type: string
            description: Path to the source
          uri:
            type: string
            description: Source URI
          extract:
            type: object
            description: key-value pairs for extraction settings
            additionalProperties: true
          glob:
            type: object
            description: Settings related to glob file matching.
            properties:
              max_total_size:
                type: integer
                description: Maximum total size (in bytes) matched by glob
              max_objects_matched:
                type: integer
                description: Maximum number of objects matched by glob
              max_objects_listed:
                type: integer
                description: Maximum number of objects listed in glob
              page_size:
                type: integer
                description: Page size for glob listing
          batch_size:
            type: string
            description: 'Size of a batch (e.g., ''100MB'')'
      salesforce:
        type: object
        properties:
          soql:
            type: string
            description: SOQL query to execute against the Salesforce instance.
          sobject:
            type: string
            description: Salesforce object (e.g., Account, Contact) targeted by the query.
          queryAll:
            type: boolean
            description: Whether to include deleted and archived records in the query (uses queryAll API).
    examples: 
      - ### Incremental model 
        type: model
        incremental: true
        connector: bigquery 

        state:
          sql: SELECT MAX(date) as max_date

        sql: |
              SELECT ... FROM events 
                {{ if incremental }} 
                    WHERE event_time > '{{.state.max_date}}' 
                {{end}}
        output:
          connector: duckdb 

      - ### Partitioned model 
        type: model

        partitions:
          glob:
            connector: gcs
            path: gs://rilldata-public/github-analytics/Clickhouse/2025/*/commits_*.parquet

        sql: SELECT * FROM read_parquet('{{ .partition.uri }}')

        output:
          connector: duckdb
          incremental_strategy: append

      - ### Partitioned Incremental model 
        type: model

        incremental: true
        refresh:
            cron: "0 8 * * *"

        partitions:
          glob:
            path: gs://rilldata-public/github-analytics/Clickhouse/2025/*/*
            partition: directory
          
        sql: |
          SELECT * 
            FROM read_parquet('gs://rilldata-public/{{ .partition.path }}/commits_*.parquet') 
            WHERE '{{ .partition.path }}' IS NOT NULL


        output:
          connector: duckdb
          incremental_strategy: append

      - ### Staging model 
        type: model 
        connector: snowflake

        # Use DuckDB to generate a range of days from 1st Jan to today
        partitions:
          connector: duckdb
          sql: SELECT range as day FROM range(TIMESTAMPTZ '2024-01-01', now(), INTERVAL 1 DAY)

        # Don't reload previously ingested partitions on every refresh
        incremental: true

        # Query Snowflake for all events belonging to the current partition
        sql: SELECT * FROM events WHERE date_trunc('day', event_time) = '{{ .partition.day }}'

        # Since ClickHouse can't ingest from Snowflake or vice versa, we use S3 as a temporary staging connector
        stage:
          connector: s3
          path: s3://bucket/temp-data

        # Produce the final output into ClickHouse, requires a clickhouse.yaml connector defined.
        output:
          connector: clickhouse
```

## Examples

### Simple model as a SQL file

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

## Source Models by Connector

### S3 to DuckDB (Parquet files)

```yaml
# models/events_raw.yaml
type: model
materialize: true

sql: |
  SELECT * FROM read_parquet('s3://my-bucket/data/events/*.parquet')
```

### S3 to DuckDB with explicit connector

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
```

### GCS to DuckDB (JSON files)

```yaml
# models/commits.yaml
type: model
materialize: true

connector: duckdb
sql: |
  SELECT * FROM read_json(
    'gs://my-bucket/data/commits.json.gz',
    auto_detect=true,
    format='auto'
  )
```

### GCS to DuckDB (Parquet files)

```yaml
# models/analytics.yaml
type: model

materialize: true
sql: |
  SELECT * FROM read_parquet('gs://my-bucket/analytics/events/*.parquet')
```

### BigQuery to DuckDB

```yaml
# models/orders.yaml
type: model
materialize: true

connector: bigquery
sql: |
  SELECT * FROM my_dataset.orders
  WHERE order_date >= DATE_SUB(CURRENT_DATE(), INTERVAL 90 DAY)

refresh:
  cron: '0 8 * * *'
  run_in_dev: true
```

### BigQuery to DuckDB with complex query

```yaml
# models/user_metrics.yaml
type: model
materialize: true

refresh:
  cron: '0 */6 * * *'
  run_in_dev: true

connector: bigquery
sql: |
  SELECT
    user_id,
    COUNT(*) AS total_orders,
    SUM(order_total) AS lifetime_value,
    MIN(order_date) AS first_order_date,
    MAX(order_date) AS last_order_date
  FROM analytics.orders
  WHERE order_status = 'completed'
  GROUP BY user_id
```

### Snowflake to DuckDB

```yaml
# models/sales.yaml
type: model
materialize: true
connector: snowflake

sql: |
  SELECT * FROM staging.sales
  {{ if dev }} LIMIT 10000 {{ end }}

prod:
  refresh:
    cron: '0 6 * * *'
```

### MySQL to DuckDB

```yaml
# models/users.yaml
type: model
connector: mysql
dsn: "{{ .env.mysql_dsn }}"

sql: |
  SELECT
    id,
    email,
    created_at,
    status
  FROM users
  WHERE status = 'active'

refresh:
  every: "10m"
```

### Postgres to DuckDB

```yaml
# models/products.yaml
type: model
connector: postgres

sql: |
  SELECT
    id,
    name,
    category,
    price,
    created_at
  FROM public.products
  WHERE active = true

refresh:
  cron: '0 * * * *'

output:
  connector: duckdb
  table: products
```

### Athena to DuckDB

```yaml
# models/clickstream.yaml
type: model
connector: athena
workgroup: primary
region: us-east-1
output_location: "s3://my-athena-results/queries/"

sql: |
  SELECT
    session_id,
    user_id,
    page_url,
    event_type,
    event_timestamp
  FROM analytics.clickstream
  WHERE event_date >= date_add('day', -30, current_date)

refresh:
  cron: '0,30 * * * *'
```

### Local CSV file to DuckDB

```yaml
# models/reference_data.yaml
type: model
materialize: true
connector: local_file
path: data/reference_data.csv
```

### HTTPS source (public Parquet file)

```yaml
# models/public_dataset.yaml
type: model
connector: https
uri: "https://example.com/public/dataset.parquet"
```

---

## ClickHouse Output Models

### Basic S3 to ClickHouse

```yaml
# models/events.yaml
type: model
materialize: true

partitions:
  glob:
    connector: s3
    path: s3://my-bucket/events/year=*/month=*/day=*/*.parquet

sql: SELECT * FROM read_parquet('{{ .partition.uri }}')

output:
  connector: clickhouse
  order_by: event_time
```

### ClickHouse with explicit column schema and indexes

```yaml
# models/impressions.yaml
type: model
materialize: true
incremental: true

partitions:
  glob:
    connector: s3
    path: s3://my-bucket/impressions/dt=*/*.parquet
    partition: directory

sql: SELECT * FROM read_parquet('{{ .partition.uri }}/*.parquet')

output:
  connector: clickhouse
  incremental_strategy: partition_overwrite
  order_by: (event_time, user_id)
  partition_by: toYYYYMMDD(event_time)
  columns: |
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
```

### ClickHouse with ReplicatedMergeTree (dev/prod)

```yaml
# models/analytics.yaml
type: model
materialize: true
incremental: true

partitions:
  glob:
    connector: s3
    path: s3://my-bucket/analytics/dt=*/*.parquet
    partition: directory

sql: SELECT * FROM read_parquet('{{ .partition.uri }}/*.parquet')

output:
  connector: clickhouse
  incremental_strategy: partition_overwrite
  order_by: (event_time, user_id)
  partition_by: toYYYYMMDD(event_time)
  engine: "{{ if dev }}MergeTree(){{ else }}ReplicatedMergeTree(){{ end }}"
```

### ClickHouse with TTL for data retention

```yaml
# models/logs.yaml
type: model
materialize: true
incremental: true

partitions:
  glob:
    path: s3://my-bucket/logs/dt=*/*.parquet
    partition: directory

sql: SELECT * FROM read_parquet('{{ .partition.uri }}/*.parquet')

output:
  connector: clickhouse
  order_by: log_time
  partition_by: toYYYYMM(log_time)
  ttl: log_time + INTERVAL 90 DAY
```

### ClickHouse with projections for pre-aggregation

```yaml
# models/metrics.yaml
type: model
materialize: true
incremental: true

partitions:
  glob:
    path: s3://my-bucket/metrics/dt=*/*.parquet
    partition: directory

sql: SELECT * FROM read_parquet('{{ .partition.uri }}/*.parquet')

output:
  connector: clickhouse
  order_by: (metric_time, dimension_1)
  partition_by: toYYYYMMDD(metric_time)
  columns: |
    metric_time DateTime,
    dimension_1 LowCardinality(String),
    dimension_2 LowCardinality(String),
    value Float64,
    PROJECTION daily_totals (
      SELECT
        toDate(metric_time) AS date,
        dimension_1,
        sum(value) AS total_value
      GROUP BY date, dimension_1
    )
```

---

## Incremental Models

### State-based incremental (high-water mark)

```yaml
# models/events_incremental.yaml
type: model
incremental: true
connector: bigquery

sql: |
  SELECT * FROM raw.events
  {{ if incremental }}
    WHERE event_time > TIMESTAMP('{{ .state.max_time }}')
  {{ end }}

state:
  sql: SELECT MAX(event_time) as max_time FROM {{ ref "self" }}

refresh:
  cron: '*/15 * * * *'
```

### Partition-based incremental (S3 glob)

```yaml
# models/daily_events.yaml
type: model
incremental: true

partitions:
  glob:
    path: s3://my-bucket/events/year=*/month=*/day=*/*.parquet
    partition: directory

sql: SELECT * FROM read_parquet('{{ .partition.uri }}/*.parquet')

output:
  incremental_strategy: partition_overwrite
```

### Incremental with merge strategy (upserts)

```yaml
# models/user_profiles.yaml
type: model
incremental: true
connector: snowflake

sql: |
  SELECT * FROM staging.user_profiles
  {{ if incremental }}
    WHERE updated_at > TO_DATE('{{ .state.max_updated }}', 'YYYY-MM-DD')
  {{ end }}

state:
  sql: SELECT MAX(updated_at) as max_updated FROM user_profiles

output:
  connector: duckdb
  incremental_strategy: merge
  unique_key: [user_id]
```

### Incremental with multiple unique keys

```yaml
# models/exchange_rates.yaml
type: model
incremental: true

partitions:
  glob:
    path: s3://my-bucket/exchange/dt=*/*.csv.gz
    partition: directory

sql: |
  SELECT
    date_export,
    source_currency,
    target_currency,
    exchange_rate
  FROM read_csv(
    '{{ .partition.uri }}/*.csv.gz',
    auto_detect=true,
    header=true
  )

output:
  incremental_strategy: merge
  unique_key: [date_export, source_currency, target_currency]
```

### Athena incremental with state

```yaml
# models/bid_responses.yaml
type: model
connector: athena
workgroup: primary
region: us-east-1
output_location: "s3://my-athena-results/queries/"

incremental: true
state:
  sql: SELECT max(max_created_at) as max_timestamp FROM bid_responses

sql: |
  SELECT
    campaign_id,
    COUNT(*) as total_bids,
    AVG(price) as avg_price,
    MAX(created_at) as max_created_at,
    date_trunc('hour', created_at) as hour
  FROM analytics.bid_responses
  WHERE created_at >= date_add('day', -30, current_date)
    {{ if incremental }}
      AND created_at > from_iso8601_timestamp('{{ .state.max_timestamp }}')
    {{ end }}
    {{ if dev }}
      AND created_at >= date_add('day', -3, current_date)
    {{ end }}
  GROUP BY campaign_id, date_trunc('hour', created_at)

refresh:
  cron: '15,45 * * * *'
```

---

## Advanced Patterns

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

sql: SELECT * FROM read_parquet('{{ .partition.uri }}/*.parquet')
```

### transform_sql for partition filtering

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
      ORDER BY uri ASC

sql: SELECT * FROM read_parquet('{{ .partition.uri }}/*.parquet')
```

### Scheduled refresh with cron

```yaml
# models/daily_metrics.yaml
type: model
materialize: true
connector: bigquery

sql: SELECT * FROM analytics.daily_metrics

refresh:
  cron: '0 6 * * *'  # Daily at 6 AM
  run_in_dev: false
```

### Scheduled refresh with interval

```yaml
# models/live_metrics.yaml
type: model
connector: mysql
dsn: "{{ .env.mysql_dsn }}"

sql: SELECT * FROM metrics_summary

refresh:
  every: "5m"
```

### Prod/dev SQL differentiation

```yaml
# models/sales_data.yaml
type: model
connector: snowflake

sql: |
  SELECT * FROM staging.sales
  {{ if dev }}
    WHERE sale_date >= DATEADD(day, -7, CURRENT_DATE())
    LIMIT 50000
  {{ else }}
    WHERE sale_date >= DATEADD(day, -365, CURRENT_DATE())
  {{ end }}

prod:
  refresh:
    cron: '30 5 * * *'
```

### Retry configuration for transient errors

```yaml
# models/external_api_data.yaml
type: model
materialize: true

partitions:
  glob:
    path: s3://my-bucket/api-exports/dt=*/*.parquet
    partition: directory

sql: SELECT * FROM read_parquet('{{ .partition.uri }}/*.parquet')

retry:
  attempts: 5
  delay: 30s
  exponential_backoff: true
  if_error_matches:
    - ".*Timeout.*"
    - ".*Connection reset.*"
    - ".*Bad Gateway.*"
```

### Long timeout for large data processing

```yaml
# models/large_historical.yaml
type: model
materialize: true
incremental: true

partitions:
  glob:
    path: s3://my-bucket/historical/year=*/*.parquet
    partition: directory

sql: SELECT * FROM read_parquet('{{ .partition.uri }}/*.parquet')

timeout: 2h
```

### Model referencing other models

```yaml
# models/enriched_orders.yaml
type: model

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

### Non-materialized view (derived model)

```yaml
# models/recent_orders.yaml
type: model
materialize: false

sql: |
  SELECT * FROM {{ ref "orders" }}
  WHERE order_date >= CURRENT_DATE - INTERVAL 30 DAY
```

### DuckDB model reading from S3 with CSV options

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
# models/daily_snapshot.yaml
type: model
incremental: true

partitions:
  connector: duckdb
  sql: SELECT range AS day FROM range(TIMESTAMPTZ '2024-01-01', now(), INTERVAL 1 DAY)

connector: snowflake
sql: |
  SELECT * FROM events
  WHERE date_trunc('day', event_time) = '{{ .partition.day }}'

output:
  connector: duckdb
  incremental_strategy: partition_overwrite
```

### ClickHouse model using S3 table function

```yaml
# models/s3_direct.yaml
type: model
materialize: true
connector: clickhouse

sql: |
  SELECT * FROM s3(
    '{{ .env.s3_bucket }}/data/*.parquet',
    '{{ .env.aws_access_key_id }}',
    '{{ .env.aws_secret_access_key }}',
    'Parquet'
  )

output:
  connector: clickhouse
  order_by: event_time
```

### Change mode for incremental models

```yaml
# models/append_only.yaml
type: model
incremental: true
change_mode: patch  # Switch to new logic without reprocessing historical data

partitions:
  glob:
    path: s3://my-bucket/events/dt=*/*.parquet
    partition: directory

sql: SELECT * FROM read_parquet('{{ .partition.uri }}/*.parquet')
```

### ClickHouse to ClickHouse (same connector)

```yaml
# models/aggregated_metrics.yaml
type: model
connector: clickhouse
database: analytics

sql: |
  SELECT
    toStartOfHour(event_time) AS hour,
    dimension,
    count() AS event_count,
    sum(value) AS total_value
  FROM events
  WHERE event_date >= today() - 7
  GROUP BY hour, dimension

output:
  connector: clickhouse
```