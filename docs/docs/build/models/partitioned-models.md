---
title: Partitioned Models
description: Create Partitioned Models
sidebar_label: Partitioned Models
sidebar_position: 25
---

## What are Partitions?

In Rill, partitions are a special type of state that allows you to explicitly partition the model into parts. Depending on whether your data is in cloud storage or a data warehouse, you can use the `glob` or `sql` parameters. This is useful when a specific partition is failing to ingest; you can specify to reload only that specific partition.

### Defining a Partition in a Model
Under the `partitions:` parameter, you will define the pattern in which your data is stored. Both SQL and glob patterns support [templating](/build/connectors/templating) and can be used to separate `dev` and `prod` instances.

### SQL
When defining your SQL partitions, it is important to understand the data that you are querying and creating a partition that makes sense. For example, you might select a distinct customer_name per partition, or partition the SQL by a chronological partition, such as month.

#### Using DuckDB for Partition Queries

By default, partition queries use DuckDB (Rill's embedded OLAP engine):

```yaml
partitions:
  sql: SELECT range AS num FROM range(0,100) #num is the partition variable and can be referenced as {{partition.num}}
  #sql: SELECT DISTINCT customer_name as cust_name from table #results in {{partition.cust_name}}
dev:
  partitions:
    sql: SELECT range AS num FROM range(0,10)
sql: SELECT * from table where column = {{partition.num}}
```

#### Using Other Connectors for Partition Queries

You can query partitions directly from data sources like **Athena**, **BigQuery**, **MySQL**, **Postgres**, **Redshift**, or **Snowflake** by specifying a `connector` in the `partitions` section. This is particularly useful when:
- You want to leverage native partitioning features (like BigQuery's `_PARTITIONTIME`)
- You need to query large tables that benefit from the warehouse's optimization

**Data Warehouses**

Athena:

```yaml
type: model

partitions:
  connector: athena
  sql: |
    SELECT DISTINCT year, month
    FROM s3_data_partitioned
    WHERE year >= 2024

connector: athena
sql: |
  SELECT * FROM s3_data_partitioned
  WHERE year = {{ .partition.year }}
    AND month = {{ .partition.month }}

output:
  connector: duckdb
```

BigQuery:

```yaml
type: model

partitions:
  connector: bigquery
  sql: |
    SELECT DISTINCT _PARTITIONTIME AS partition_time
    FROM `project.dataset.table`
    WHERE TIMESTAMP_TRUNC(_PARTITIONTIME, MONTH) = TIMESTAMP("2025-08-01")

connector: bigquery
sql: |
  SELECT * FROM `project.dataset.table`
  WHERE _PARTITIONTIME = '{{ .partition.partition_time }}'

output:
  connector: duckdb  
```

Redshift:

```yaml
type: model

partitions:
  connector: redshift
  sql: |
    SELECT DISTINCT DATE_TRUNC('month', transaction_date) AS month
    FROM transactions
    WHERE transaction_date >= '2024-01-01'

connector: redshift
sql: |
  SELECT * FROM transactions
  WHERE DATE_TRUNC('month', transaction_date) = '{{ .partition.month }}'

output:
  connector: duckdb
```

Snowflake:

```yaml
type: model
connector: snowflake

partitions:
  connector: snowflake
  sql: |
    select 
      DISTINCT date_trunc('YEAR', release_date) as "year" 
    from 
      rillqa.public.horror_movies 
    where "year" > '1999-01-01' limit 3

sql: select * from rillqa.public.horror_movies where date_trunc('YEAR', release_date) = '{{ .partition.year }}'

output:
  connector: duckdb  
```

**OLTP Databases**

MySQL:

```yaml
type: model

partitions:
  connector: mysql
  sql: |
    SELECT DISTINCT DATE(order_date) AS order_day
    FROM orders
    WHERE order_date >= '2025-01-01'

connector: mysql
sql: |
  SELECT * FROM orders
  WHERE DATE(order_date) = '{{ .partition.order_day }}'

output:
  connector: duckdb
```

Postgres:

```yaml
type: model

partitions:
  connector: postgres
  sql: |
    SELECT DISTINCT DATE_TRUNC('day', created_at) AS partition_day
    FROM events
    WHERE created_at >= '2025-01-01'

connector: postgres
sql: |
  SELECT * FROM events
  WHERE DATE_TRUNC('day', created_at) = '{{ .partition.partition_day }}'

output:
  connector: duckdb
```

:::tip Why use multiple connectors?

Using Athena, BigQuery, MySQL, Postgres, Redshift, or Snowflake for partition discovery and data extraction, then outputting to DuckDB, gives you:
- **Best of both worlds**: Leverage your warehouse's partitioning and scale for extraction
- **Fast dashboards**: DuckDB provides extremely fast query performance for end-user dashboards
- **Cost optimization**: Only query what you need from your warehouse, reducing scan costs

:::

:::tip Using the SQL partition in the YAML
Depending on the column name of the partition, you can reference the partition using `{{ .partition.<column_name> }}` in the model's SQL query.
```yaml
partitions:
  sql: SELECT range AS num FROM range(0,10)
sql: SELECT {{ .partition.num }} AS num, now() AS inserted_on {{if dev}} limit 1000 {{end}}
```
:::

### glob

When defining the glob pattern, you will need to consider whether you'd partition the data by folder or file.
In the first example, we are partitioning by each file with the suffix data.csv.
```yaml
partitions:
  glob: gs://my-bucket/y=2025/m=03/d=15/*data.csv
  #glob: gs://my-bucket/{{if dev}}y=2025/m=03/d=15{{else}}**{{end}}/*data.csv
```

Or, you can define each glob separately.
```yaml
partitions:
    glob:
      path: 'gs://my-bucket/**/*.parquet'

dev:
  partitions:
    glob:
     path: 'gs://my-bucket/y=2025/m=03/d=15/*.parquet'
```

If you'd prefer to partition it by folder, you can add the partition parameter and define it as `directory`.
```yaml
glob:
  path: gs://rendo-test/**/*data.csv
  partition: directory #hive
```

:::tip Using the glob partition in the YAML
The glob partition has a predefined `{{ .partition.uri }}` reference to use in the model's SQL query.
```yaml
partitions:
  glob:
    connector: gcs
    path: gs://path/to/file/**/*.parquet
sql: SELECT * FROM read_parquet('{{ .partition.uri }}')
```
:::

### Viewing Partitions in Rill Developer

Once `partitions:` is defined in your model, a new button will appear in the right-hand panel: `View Partitions`. When selecting this, a new UI will appear with all of your partitions and more information on each. Note that these can be sorted by all, pending, and errors.

<img src='/img/build/advanced-models/partitions-developer.png' class='rounded-gif' />
<br />

You can sort the view by `all partitions`, `pending partitions`, and `error partitions`.
- **all partitions**: shows all the available partitions in the model.
- **pending partitions**: shows the partitions that are waiting to be processed.
- **error partitions**: displays any partitions that errored during ingestion.

### Viewing Partitions in the CLI
Likewise to the UI, you can view the partitions of a model within the CLI. 

```
rill project partitions 
List partitions for a model

Usage:
  rill project partitions [<project>] <model> [flags]

Flags:
      --project string      Project Name
      --path string         Project directory (default ".")
      --model string        Model Name
      --pending             Only fetch pending partitions
      --errored             Only fetch errored partitions
      --local               Target locally running Rill
      --page-size uint32    Number of partitions to return per page (default 50)
      --page-token string   Pagination token
```

If running locally, you will need to add the `--local` flag to the command.
```bash
rill project partitions model_name [--local]
  KEY (10)                           DATA        EXECUTED ON            ELAPSED   ERROR  
 ---------------------------------- ----------- ---------------------- --------- ------- 
  ff7416f774dfb086006d0b4696c214e1   {"num":0}   2024-11-12T22:48:49Z   95ms     
  ...
```

### Refreshing Partitions via the CLI 
:::note Incremental not enabled
If you try to refresh a partition using the following command on a partitioned but not incremental model, you will experience the following error:
```
rill project refresh  --model <model_name> [--local] --partition ff7416f774dfb086006d0b4696c214e1
Error: can't refresh partitions on model "model_name" because it is not incremental
```
:::

You will need to enable [incremental modeling](/build/models/incremental-partitioned-models) in order to individually refresh a partition. 