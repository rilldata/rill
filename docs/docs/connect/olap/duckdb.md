---
title: DuckDB
description: Power Rill dashboards using DuckDB (default)
sidebar_label: DuckDB
sidebar_position: 10
---

[DuckDB](https://duckdb.org/why_duckdb.html) is an in-memory, columnar SQL database designed for analytical (OLAP) workloads, offering high-speed data processing and analysis. Its columnar storage model and vectorized query execution make it highly efficient for OLAP tasks, enabling fast aggregation, filtering, and joins on large datasets. DuckDB's ease of integration with data science tools and its ability to run directly within analytical environments like Python and R, without the need for a separate server, make it an attractive choice for OLAP applications seeking simplicity and performance.

By default, Rill includes DuckDB as an embedded OLAP engine that is used to ingest data from [sources](/connect) and power your dashboards. Nothing more needs to be done if you wish to power your dashboards on Rill Developer or Rill Cloud. 

However, you may need to add additional extensions into the embedded DuckDB Engine. To do so, you'll need to define the [Connector YAML](/reference/project-files/connectors#duckdb) and use `init_sql` to install/load/set `extension_name`.

:::tip Optimizing performance on DuckDB

DuckDB is a very useful analytical engine but can start to hit performance and scale challenges as the size of ingested data grows significantly. As a general rule of thumb, we recommend keeping the size of data in DuckDB **under 50GB** along with some other [performance recommendations](/guides/performance). For larger volumes of data, Rill still promises great performance but additional backend optimizations will be needed. [Please contact us](/contact)!

:::

### Multiple Engines 

While not recommended, it is possible in Rill to use multiple OLAP engines in a single project. For more information, see our page on [Using Multiple OLAP Engines](/connect/olap/multiple-olap).

## Additional Notes

- For dashboards powered by DuckDB, [measure definitions](/build/metrics-view/#measures) are required to follow standard [DuckDB SQL](https://duckdb.org/docs/sql/introduction) syntax.
- There is a known issue around creating a DuckDB source via the UI; you will need to create the YAML file manually.