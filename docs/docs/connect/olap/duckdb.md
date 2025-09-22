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

## Connecting to External DuckDB

Along with our embedded DuckDB, Rill provides the ability to connect to external DuckDB database files. This allows you to leverage existing DuckDB databases inside of Rill.

### Configuration

<img src='/img/connect/connector/duckdb.png' class='rounded-gif' style={{width: '75%', display: 'block', margin: '0 auto'}}/>
<br />

### Important Considerations

:::warning File Location and Size

When connecting to external DuckDB files, consider the following:

- **File Location**: Ensure the DuckDB file is accessible from your Rill environment
- **File Size**: Large DuckDB files (>50GB) may impact performance
- **Permissions**: Ensure Rill has read/write access to the database file
- **Network Access**: For cloud deployments, ensure the file is accessible from the cloud environment

:::


Once connected, you can see all the tables in your external DuckDB database within Rill and create models, metrics views, and dashboards using this data.


## Using DuckDB Extensions

DuckDB supports a wide variety of extensions that can enhance its functionality. To use extensions with Rill's embedded DuckDB, configure them in your connector:

```yaml
# connectors/duckdb.yaml
type: connector
driver: duckdb
init_sql: |
  INSTALL httpfs;
  LOAD httpfs;
  INSTALL spatial;
  LOAD spatial;
```

### Popular Extensions

For a complete list of available extensions, see the [DuckDB Extensions documentation](https://duckdb.org/docs/extensions/overview).

## Multiple Engines 

While not recommended, it is possible in Rill to use multiple OLAP engines in a single project. For more information, see our page on [Using Multiple OLAP Engines](/connect/olap/multiple-olap).

## Additional Notes

- For dashboards powered by DuckDB, [measure definitions](/build/metrics-view/#measures) are required to follow standard [DuckDB SQL](https://duckdb.org/docs/sql/introduction) syntax.
- There is a known issue around creating a DuckDB source via the UI; you will need to create the YAML file manually.
- DuckDB supports most standard SQL functions and operators, making it easy to write complex analytical queries.
- For advanced analytics, consider using DuckDB's window functions, CTEs, and other advanced SQL features.