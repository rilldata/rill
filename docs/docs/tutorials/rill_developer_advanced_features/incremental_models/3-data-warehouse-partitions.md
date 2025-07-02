---
title: "Incremental Model based on a state from Data Warehouses"
description:  "Getting Started with Partitions"
sidebar_label: "Data Warehouse: Incremental Stage Models"
sidebar_position: 12
tags:
  - Rill Developer
  - Advanced Features
---

Another advanced concept within Rill is using [Incremental Models](/build/advanced-models/incremental-models) for a SQL based source. For the most part, we suggest that you use a time column as your partitioning column. The main questions to ask are:
1. Do I have late arriving data? 
2. How many days back should I reasonably re-ingest based on my data pipelines?
3. How far back should my data exist in Rill historically? 

:::tip requirements
You will need to setup the connection to your data warehouse, depending on the connection please refer to [our documentation](https://docs.rilldata.com/reference/connectors/). 

In this example we use a DATE column as our defining state but depending on your data, you can use any defining column.

:::

## Understanding Partitions in Models

Here’s how it works at a high level:

- **Partition Definition**: Based on a SQL query, defines a key that allows you to increment your model by, usually a time column.
- **Execution Strategy**:
  - **Full Refresh**: Runs without incremental processing.
  - **Incremental Refresh**: Run incrementally based on the partitions defined, following the output connector's `incremental_strategy` (either append or merge for SQL connectors). 

:::tip Default Incremental Strategy 
We default to partition override for the partition strategy if nothing is defined. If you want to append instead, define the incremental_strategy in the output parameter.

```yaml
output:
  table: output_table_name
  incremental_strategy: append
```
:::

### Let's create a basic partitions model.

:::note Example
In this example, we are using a sample dataset that exists in Big Query: rilldata.ssb_100.date.
In this case our table is not getting updated, so instead we'll modify the SQL to show you how incremental works.
:::


1. Create a YAML file: `SQL_incremental_tutorial.yaml`

2. Use the following contents to create your own model.
```yaml
type: model
materialize: true

connector: "bigquery" #or "snowflake"

incremental: true
partitions:
  connector: duckdb
  sql: >
     {{ if dev }} 
        SELECT date_trunc('day', day) AS date_partition FROM generate_series(DATE '1992-01-01', DATE '1992-03-01', INTERVAL 1 DAY) AS ts(day)
     {{else}}
        SELECT date_trunc('day', day) AS start, date_trunc('day', start + INTERVAL 1 MONTH) AS end FROM generate_series(DATE '2024-01-01', DATE '2025-06-01', INTERVAL 1 MONTH) AS ts(day)
      UNION ALL 
        SELECT date_trunc('day', day) AS start, date_trunc('day', start + INTERVAL 1 DAY) AS end FROM generate_series(DATE '2025-06-01', CURRENT_DATE, INTERVAL 1 DAY) AS ts(day)
     {{end}}
sql: |
  SELECT *,
         PARSE_DATE('%Y%m%d', CAST(D_DATEKEY AS STRING)) AS DATE
  FROM rilldata.ssb_100.date
  {{if dev}}
    WHERE PARSE_DATE('%Y%m%d', CAST(D_DATEKEY AS STRING)) = PARSE_DATE('%Y-%m-%dT00:00:00Z','{{.partition.date_partition}}')
  {{else}}

  {{end}}

output:
  connector: duckdb
  # incremental_strategy: append 
  # By default we'll overwrite the partition defined above if changes are detected.
```

3. In the UI, try refreshing both incrementally and fully to see the difference in the model that loads. 



<img src = '/img/tutorials/advanced-models/data-warehouse-refresh.png' class='rounded-gif' />
<br />

There are two reasons why you'll see that the source data doesn't change on either refresh. 
1. Since we're using if dev, the partitions are static 3 days in the year of 1992. 
2. Even if we didn't use this, the number of partitions would increase by day but the recent partitions doesn't have any following data. 

To note in the logs, you'll want to look for the `Resolved model partitions` to see how many partitions are being detected. On a full refresh, you'll see each partition being loaded. On an incremental, you'll just see a single partition refreshed.

```bash
2025-07-01T17:23:57.608 INFO    Resolved model partitions       {"model": "SQL_increment_tutorial", "partitions": 3}
2025-07-01T17:24:00.614 INFO    Executed model partition        {"model": "SQL_increment_tutorial", "key": "86915540cd0c753cdd641e0f487cd7f6", "data": {"date_partition":"1992-01-01T00:00:00Z"}, "elapsed": "2.998785871s"}
2025-07-01T17:24:02.999 INFO    Executed model partition        {"model": "SQL_increment_tutorial", "key": "87883ba5a1792d2c2e818fb8b8de5c20", "data": {"date_partition":"1992-01-02T00:00:00Z"}, "elapsed": "2.377725743s"}
2025-07-01T17:24:05.285 INFO    Executed model partition        {"model": "SQL_increment_tutorial", "key": "d4a0bf3200df3675b043affa6882d0b2", "data": {"date_partition":"1992-01-03T00:00:00Z"}, "elapsed": "2.270095921s"}
2025-07-01T17:24:05.287 INFO    Reconciled resource     {"name": "SQL_increment_tutorial", "type": "Model", "elapsed": "7.904s"}
```


