---
title: Create Advanced Models
description: Create Advanced Models
sidebar_label: Create Advanced Models
sidebar_position: 00
---

<img src = '/img/build/advanced-models/advanced-model.png' class='rounded-gif' />
<br />

Unlike SQL models, YAML file models allow the ability to fine tune a model to perform additional capabilities such as pre-exec, post-exec SQL, partitioning, and incremental modeling. This is an imporant addition to modeling as it allows the user to customize the model's method of building. In the case of partitions and incremental modeling, this will reduce the amount of data ingested into Rill at each interval and allow insight into specific issues per partition. Another use case is when using [multiple OLAP engines](../connect/multiple-connectors.md), this allows you to define where a SQL query is run. 

## When to use Advanced Models? 

For the majority of users on a DuckDB backed Rill project, advanced models are not required. When a project gets larger and refreshing the whole datasets becomes a time consuming and costly task, we introduce incremental ingestion to help alleviate the problem. Along with incremental modeling, we use partitions to divide a dataset into smaller to manage datasets. When enabling partitions, you are able to refresh single sections of data if required. 

Another use case is when using multiple OLAP engines. This allows you to specify where you SQL is running. While we do not officially support ClickHouse modeling, this is available behind a feature flag `clickhouseModeling`. When both DuckDB and ClickHouse are enabled in a single environment, you will need to define `connector: duckdb/clickhouse` in the yaml to tell Rill where to run the SQL query.


## Types of Advanced Models

1. [Incremental Models](./incremental-models)
2. [Partitioned Models](./partitions)
3. [Staging Models](./staging)
4. [DuckDB `pre_exec`/`post_exec` Models](#duckdb-models-pre-exec-sql-post-exec)


## Creating an Advanced Model
You can get started with an advanced model with the following code block: 

```yaml
#Model YAML
#Reference documentation: https://docs.rilldata.com/reference/project-files/advanced-models

type: model
connector: duckdb

sql: select * from <source>
```

Please refer to [our reference documentation](../../reference/project-files/advanced-models) linked above for the available parameters to set in your model.

:::note

Currently there isn't a UI button to start off with an advanced model YAML. Creating a model in Rill will always create a model.sql file. 

:::



## DuckDB Model's pre-exec, sql, post-exec 

While we install a set of core libraries and extensions with our embed DuckDB, there might be specific use-cases where you might want to add a different one. In order to do this, you will need to use the pre-exec parameter to ensure that everything is loaded before running your SQL query. 

Take the example of [`gsheets` community extension](https://duckdb.org/community_extensions/extensions/gsheets.html). In order to use this extension in Rill, you'll need to install and load the plugin. Once that's done you can define the secret and finally run the SQL. 

```yaml
pre_exec: INSTALL gsheets FROM community; LOAD gsheets; CREATE SECRET (TYPE gsheet, PROVIDER access_token, TOKEN '<your_token>');

sql: SELECT * FROM read_gsheet('https://docs.google.com/spreadsheets/d/<your_unique_ID>', headers=false);

```

:::tip Multiple queries to run? 
Like any SQL query, you can divide the queries with a semi-colon to run multiple queries. This is available for both `pre_exec` and `post_exec`. The default `sql` parameter requires a single SELECT statement to run.
:::


Another example is attaching a database to DuckDB, running some queries against it then detaching said database. 

```yaml
pre_exec: ATTACH 'dbname=postgres host=localhost port=5432 user=postgres password=postgres' AS postgres_db (TYPE POSTGRES);
sql: SELECT * FROM postgres_query('postgres_db', 'SELECT * FROM USERS')
post_exec: DETACH postgres_db # Note : this is not mandatory but nice to have 
```

## Similar Considerations to Note

As with normal SQL models, materialization will be disabled by default and depending on your use-case setting this parameter to true may improve performance. For more information, check out [our model materialization notes.](../../reference/project-files/models#model-materialization)