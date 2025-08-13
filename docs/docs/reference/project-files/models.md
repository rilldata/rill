---
title: Model YAML
sidebar_label: Model YAML
sidebar_position: 20
hide_table_of_contents: true
---

:::tip

Both regular models and source models can use the Model YAML specification described on this page. While [SQL models](/build/models/sql-models) are perfect for simple transformations, Model YAML files provide advanced capabilities for complex data processing scenarios.

**When to use Model YAML:**
- **Partitions** - Optimize performance with data partitioning strategies
- **Incremental models** - Process only new or changed data efficiently
- **Pre/post execution hooks** - Run custom logic before or after model execution
- **Staging** - Create intermediate tables for complex transformations
- **Output configuration** - Define specific output formats and destinations

Model YAML files give you fine-grained control over how your data is processed and transformed, making them ideal for production workloads and complex analytics pipelines.

:::

## Properties

**`type`** - Refers to the resource type and must be 'model' _(required)_

**`refresh`** - Specifies the refresh schedule that Rill should follow to re-ingest and update the underlying source data _(optional)_.
  - **`cron`** - A cron schedule expression, which should be encapsulated in single quotes, e.g., `'* * * * *'` _(optional)_
  - **`every`** - A Go duration string, such as `24h` ([docs](https://pkg.go.dev/time#ParseDuration)) _(optional)_

```yaml
refresh:
    cron: "0 8 * * *"
```

**`timeout`** — The maximum time to wait for model ingestion _(optional)_.

**`incremental`** - Set to `true` or `false` whether incremental modeling is required _(optional)_

**`state`** - Refers to the explicitly defined state of your model, cannot be used with `partitions` _(optional)_.
  - **`sql/glob`** - Refers to the location of the data depending on whether the data is cloud storage or a data warehouse.

**`partitions`** - Refers to how your data is partitioned, cannot be used with `state` _(optional)_.
  - **`connector`** - Refers to the connector that the partitions are using _(optional)_.
  - **`sql`** - Refers to the SQL query used to access the data in your data warehouse, use `sql` or `glob` _(optional)_.
  - **`glob`** - Refers to the location of the data in your cloud warehouse, use `sql` or `glob` _(optional)_.
    - **`path`** - In the case `glob` is selected, you will need to set the path of your source _(optional)_.
    - **`partition`** - In the case `glob` is selected, you can define how to partition the table: directory or hive _(optional)_.
    
```yaml
partitions:
  connector: duckdb
  sql: SELECT range AS num FROM range(0,10)
```

```yaml
partitions:
  glob:
    connector: [s3/gcs]
    path: [s3/gs]://path/to/file/**/*.parquet[.csv]
```

**`pre_exec`** – Refers to SQL queries to run before the main query, available for DuckDB-based models _(optional)_. Ensure `pre_exec` queries are idempotent. Use `IF NOT EXISTS` statements when applicable.

**`sql`** - Refers to the SQL query for your model _(required)_.

**`post_exec`** – Refers to a SQL query that is run after the main query, available for DuckDB-based models _(optional)_. Ensure `post_exec` queries are idempotent. Use `IF EXISTS` statements when applicable.

```yaml
pre_exec: ATTACH IF NOT EXISTS 'dbname=postgres host=localhost port=5432 user=postgres password=postgres' AS postgres_db (TYPE POSTGRES)

sql: SELECT * FROM postgres_query('postgres_db', 'SELECT * FROM USERS')

post_exec: DETACH DATABASE IF EXISTS postgres_db
```

**`partitions_watermark`** - Refers to a customizable timestamp that can be set to check if an object has been updated _(optional)_.

**`partitions_concurrency`** - Refers to the number of concurrent partitions that can be read at the same time _(optional)_.

**`stage`** - In the case of staging models, where an input source does not support direct write to the output and a staging table is required _(optional)_.
  - **`connector`** - Refers to the connector type for the staging table
  - **`path`** - Path of the temporary staging table

**`output`** - In the case of staging models, where the output needs to be defined where the staging table will write the temporary data _(optional)_.
  - **`connector`** - Refers to the connector type for the staging table _(optional)_.
  - **`incremental_strategy`** - Refers to how the incremental refresh will behave (merge or append) _(optional)_.
  - **`unique_key`** - Required if incremental_strategy is defined, refers to the unique column to use to merge _(optional)_.
  - **`materialize`** - Refers to the output table being materialized _(optional)_.
  - **`columns`** - Refers to a list of columns if you require to manually define column name and types _(optional)_.
  - **`engine_full`** - Refers to the ClickHouse engine specifications (ENGINE = ... PARTITION BY ... ORDER BY ... SETTINGS ...) _(optional)_.

**`materialize`** - Refers to the model being materialized as a table or not _(optional)_.
