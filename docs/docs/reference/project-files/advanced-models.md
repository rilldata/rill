---
title: Advanced Model YAML
sidebar_label: Advanced Model YAML
sidebar_position: 35
hide_table_of_contents: true
---

In some cases, advanced models will be required when implementing advanced features such as incremental partitioned models or staging models. 


## Properties

**`type`** - refers to the resource type and must be 'model'_(required)_ 

**`refresh`** - Specifies the refresh schedule that Rill should follow to re-ingest and update the underlying source data _(optional)_.
  - **`cron`** - a cron schedule expression, which should be encapsulated in single quotes, e.g. `'* * * * *'` _(optional)_
  - **`every`** - a Go duration string, such as `24h` ([docs](https://pkg.go.dev/time#ParseDuration)) _(optional)_
```
refresh:
    cron: "0 8 * * *"
```

**`timeout`** — The maximum time to wait for model ingestion _(optional)_.

**`incremental`** - set to `true` or `false` whether incremental modeling is required _(optional)_

**`state`** - refers to the explicitly defined state of your model, cannot be used with `partitions` _(optional)_.
  - **`sql/glob`** - refers to the location of the data depending if the data is cloud storage or a data warehouse.

**`partitions`** - refers to the how your data is partitioned, cannot be used with `state`.  _(optional)_.
  - **`connector`** - refers to the connector that the partitions is using _(optional)_.
  - **`sql`** - refers to the SQL query used to access the data in your data warehouse, use `sql` or `glob` _(optional)_.
  - **`glob`** - refers to the location of the data in your cloud warehouse, use `sql` or `glob` _(optional)_.
    - **`path`** - in the case `glob` is selected, you will need to set the path of your source _(optional)_. 
    - **`partition`** - in the case `glob` is selected, you can defined how to partition the table. directory or hive _(optional)_.
    
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

**`sql`** - refers to the SQL query for your model. _(required)_.

**`partitions_watermark`** - refers to a customizable timestamp that can be set to check if an object has been updated _(optional)_. 

**`partitions_concurrency`** - refers to the number of concurrent partitions that can be read at the same time _(optional)_. 

**`stage`** - in the case of staging models, where an input source does not support direct write to the output and a staging table is required _(optional)_. 
  - **`connector`** - refers to the connector type for the staging table
  - **`path`** - path of the temporary staging table

**`output`** - in the case of staging models, where the output needs to be defined where the staging table will write the temporary data _(optional)_. 
  - **`connector`** - refers to the connector type for the staging table  _(optional)_.
  - **`incremental_strategy`** - refers to how the incremental refresh will behave, (merge or append)  _(optional)_.
  - **`unique_key`** - required if incremental_stategy is defined, refers to the unique column to use to merge  _(optional)_.
  - **`materialize`** - refers to the output table being materialized  _(optional)_.

**`materialize`** - refers to the model being materialized as a table or not _(optional)_. 