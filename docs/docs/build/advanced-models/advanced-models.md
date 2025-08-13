---
title: Create YAML Models
description: Create Advanced Models
sidebar_label: Create Advanced Models
sidebar_position: 00
---

<img src = '/img/build/advanced-models/advanced-model.png' class='rounded-gif' />
<br />

Unlike SQL models, YAML file models allow you to fine-tune a model to perform additional capabilities such as pre-exec, post-exec SQL, partitioning, and incremental modeling. This is an important addition to modeling, as it allows the user to customize the model's build process. In the case of partitions and incremental modeling, this will reduce the amount of data ingested into Rill at each interval and allow insight into specific issues per partition. Another use case is when using [multiple OLAP engines](/connect/olap/multiple-olap), this allows you to define where a SQL query is run. 

## When to use a YAML Models? 

For the majority of users on a DuckDB backed Rill project, YAML models are not required. When a project gets larger and refreshing the whole datasets becomes a time-consuming and costly task, we introduce incremental ingestion to help alleviate the problem. Along with incremental modeling, we use partitions to divide a dataset into smaller, more manageable partitonis. After enabling partitions, you will be able to refresh individual partitonis of data when required. 

Another use case is when using multiple OLAP engines. This allows you to specify where your SQL query is running. When both DuckDB and ClickHouse are enabled in a single environment, you will need to define `connector: duckdb/clickhouse` in the YAML to tell Rill where to run the SQL query, as well as define the `output` location. For more information, refer to the [YAML reference](/reference/project-files/advanced-models)

## Types of YAML Models


1. [Incremental Models](/build/advanced-models/incremental-models)
2. [Partitioned Models](/build/advanced-models/partitions)
3. [Staging Models](/build/advanced-models/staging)
4. [DuckDB `pre_exec`/`post_exec` Models](#duckdb-models-pre-exec-sql-post-exec)



## Creating a YAML Model
You can get started with an advanced model with the following code block: 

```yaml
#Model YAML
#Reference documentation: https://docs.rilldata.com/reference/project-files/advanced-models

type: model
connector: duckdb

sql: select * from <source>

output:
  connector: duckdb
  table: output_name
```

Please refer to [our reference documentation](../../reference/project-files/advanced-models) linked above for the available parameters to set in your model.

:::note

Currently, there isn't a UI button to start off with an advanced model YAML. Creating a model in Rill via the UI will always create a model.sql file. Instead, start with a blank file and rename it model_name.yaml and add the above sample code.

:::



## DuckDB Model's pre-exec, post-exec SQL

While we install a set of core libraries and extensions with our embedded DuckDB, there might be specific use cases where you want to add a different one. In order to do this, you will need to use the pre-exec parameter to ensure that everything is loaded before running your SQL query. 

Take the example of [`gsheets` community extension](https://duckdb.org/community_extensions/extensions/gsheets.html). In order to use this extension in Rill, you'll need to install and load the plugin. Once that's done you can define the secret and finally run the SQL. 

```yaml
pre_exec: |
    INSTALL gsheets FROM community; 
    LOAD gsheets; 
    CREATE TEMPORARY SECRET IF NOT EXISTS secret (TYPE gsheet, PROVIDER access_token, TOKEN '<your_token>');

sql: SELECT * FROM read_gsheet('https://docs.google.com/spreadsheets/d/<your_unique_ID>', headers=false);

```

:::tip Multiple queries to run? 
Like any SQL query, you can divide the queries with a semicolon to run multiple queries. This is available for both `pre_exec` and `post_exec`. The default `sql` parameter requires a single SELECT statement to run.
:::


Another example is attaching a database to DuckDB, running some queries against it then detaching said database. 

```yaml
pre_exec: ATTACH IF NOT EXISTS 'dbname=postgres host=localhost port=5432 user=postgres password=postgres' AS postgres_db (TYPE POSTGRES);
sql: SELECT * FROM postgres_query('postgres_db', 'SELECT * FROM USERS')
post_exec: DETACH DATABASE IF EXISTS postgres_db # Note : this is not mandatory but nice to have 
```

## Similar Considerations to Note

1. As with normal SQL models, materialization will be disabled by default and depending on your use-case setting this parameter to true may improve performance. For more information, check out [our model materialization notes.](../../reference/project-files/models#model-materialization)


2. The `pre_exec` and `post_exec` statements are run for every model execution and thus should be made idempotent.
A typical way is to use `IF NOT EXISTS` clauses for CREATE statements. Please refer to duckDB docs for exact definitions and verify if the statements are idempotent.