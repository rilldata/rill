---
title: DuckDB
description: Power Rill dashboards using DuckDB (default)
sidebar_label: DuckDB
sidebar_position: 10
---

[DuckDB](https://duckdb.org/why_duckdb.html) is an in-memory, columnar SQL database designed for analytical (OLAP) workloads, offering high-speed data processing and analysis. Its columnar storage model and vectorized query execution make it highly efficient for OLAP tasks, enabling fast aggregation, filtering, and joins on large datasets.

## Rill Managed DuckDB

By default, Rill includes DuckDB as an embedded OLAP engine that ingests data from [data sources](/build/connectors) and powers your dashboards. When you start a new project, you'll see a `connectors/duckdb.yaml` file alongside other project files. No additional configuration is needed to use DuckDB with Rill Developer or Rill Cloud.

```yaml
type: connector

driver: duckdb
managed: true
```

:::tip Performance Considerations

DuckDB is an excellent analytical engine but can face performance challenges as data size grows significantly. As a general guideline, we recommend keeping your data size in DuckDB **under 50GB** along with other [performance recommendations](/guides/performance). For larger datasets, Rill still provides excellent performance but may require additional backend optimizations. [Contact us](/contact) if you need assistance with large-scale deployments.

:::

## Live Connect to External DuckDB

Rill also supports connecting to external DuckDB database files as a "live connector". This allows you to leverage existing DuckDB databases within Rill to create metrics views and dashboards.

:::warning Local Development Only

This setup is designed for local development and testing only. It will not deploy to Rill Cloud under most circumstances because:

- Rill Cloud can only access files within your project directory
- If your DuckDB file is outside the project folder, it cannot be bundled for deployment
- Files larger than 100MB will fail to deploy due to upload size limits

For production deployments, consider using our [external DuckDB data source](/build/connectors/data-source/duckdb) to ingest your data instead.

:::

### Configuration

Using the UI, select the DuckDB icon under the OLAP section to add a new DuckDB connector. Any existing connectors with data models will need to be refreshed to ingest the data into your external DuckDB. 

```yaml
type: connector

driver: duckdb
path: '/path/to/main.db'
```

### Setting the Default OLAP Connection

Creating a connection to MotherDuck will automatically add the `olap_connector` property in your project's [rill.yaml](/reference/project-files/rill-yaml) and change the default OLAP engine to `duckdb`.

```yaml
olap_connector: duckdb
```


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

## Multiple OLAP Engines

While not recommended, Rill supports using multiple OLAP engines in a single project. For more information, see [Using Multiple OLAP Engines](/build/connectors/olap/multiple-olap).

## Additional Notes

- For dashboards powered by DuckDB, [measure definitions](/build/metrics-view/#measures) are required to follow standard [DuckDB SQL](https://duckdb.org/docs/sql/introduction) syntax.
- There is a known issue around creating a DuckDB source via the UI; you will need to create the YAML file manually.