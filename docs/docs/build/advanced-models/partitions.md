---
title: Partitioned Models
description: Create Partitioned Models
sidebar_label: Partitioned Models
sidebar_position: 03
---

## What are Partitions?

In Rill, partitions are a special type of state in which you can explicitly partition the model into parts. Depending on if your data is in cloud storage or a data warehouse, you can use the `glob` or `sql` parameters. This is useful when a specific partition is failing to ingest, you can specific to reload only that specific partition.


### Defining a Partition in a Model
Under the `partitions:` parameter, you will define the pattern in which your data is stored.

### SQL
When defining your SQL, it is important to understand the data that you are querying and creating a partition that makes sense. For example, possibly selecting a distinct customer_name per partition, or possibly partition the SQL by a chronological partition, such as month.

```yaml
partitions:
  sql: SELECT range AS num FROM range(0,10) #num is the partition variable and can be referenced as {{partition.num}}
  #sql: SELECT DISTINCT customer_name as cust_name from table #results in {{partition.cust_name}}
  ```

:::tip Using the SQL parition in the YAML
Depending on the column name of the partition, you can reference the partition using ` {{ .partition.<column_name> }}` in the model's SQL query.
```YAML
partitions:
  sql: SELECT range AS num FROM range(0,10)
sql: SELECT {{ .partition.num }} AS num, now() AS inserted_on
```
:::

### glob

When defining the glob pattern, you will need to consider whether you'd partition the data by folder or file. For information on glob patterns, see [glob patterns](/build/connect/glob-patterns).
In the first example, we are paritioning by each file with the suffix data.csv.
```yaml
partitions:
  glob: gs://my-bucket/**/*data.csv
  ```

If you'd prefer to partition it by folder your can add the partition parameter and define it as `directory`.
```yaml
glob:
  path: gs://rendo-test/**/*data.csv
  partition: directory #hive
```
:::tip Using the glob partition in the YAML
The glob partition has a predefined `{{ .partition.uri }}` reference to use in the model's SQL query.
```YAML
partitions:
  glob:
    connector: gcs
    path: gs://path/to/file/**/*.parquet
sql: SELECT * FROM read_parquet('{{ .partition.uri }}')
```
:::

### Viewing Partitions in Rill Developer

Once `partitions:` is defined in your model, a new button will appear in the right hand panel, `View Partitions`. When selecting this, a new UI will appear with all of your partitions and more information on each. Note that these can be sorted on all, pending, and errors.

<img src = '/img/build/advanced-models/partitions-developer.png' class='rounded-gif' />
<br />

You can sort the view on `all partitions`, `pending partitions` and `error partitions`. 
- **all partitions**: shows all the available paritions in the model.
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
:::note  Incremental not enabled
If you try to refresh a partition using the following command on a partitioned but not incremental model, you will experience the following error:
```
rill project refresh  --model <model_name> [--local] --partition ff7416f774dfb086006d0b4696c214e1
Error: can't refresh partitions on model "model_name" because it is not incremental
```
:::

You will need to enable [incremental modeling](incremental-partitioned-models.md) in order to individually refresh a partition. 