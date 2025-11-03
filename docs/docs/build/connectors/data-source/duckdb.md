---
title: External DuckDB 
description: Connect to external DuckDB databases and ingest data into Rill
sidebar_label: External DuckDB 
sidebar_position: 11
---

<!-- WARNING: There are links to this page in source code. If you move it, find and replace the links and consider adding a redirect in docusaurus.config.js. -->

While not recommended for production use, Rill allows you to `attach` external DuckDB databases to ingest data from them as a data source. This approach has several caveats and limitations during deployment and is primarily intended for local testing scenarios.

:::warning Local Development Only

There are several limitations with deployment to Rill Cloud, so we do not recommend this method for production environments. Key limitations include:

- Moving your DuckDB file into the `data/` folder within your project directory
- A size limitation of 100MB when deploying to Rill Cloud

[Contact us](/contact) if you have questions or encounter issues with these limitations.

:::

## Attaching an External DuckDB

In the default `connectors/duckdb.yaml` file, you can use the `init_sql` parameter to execute SQL statements during database initialization, such as attaching an external database to Rill's embedded DuckDB. For more details on the YAML configuration, see the [DuckDB reference page](/reference/project-files/connectors#duckdb).

```yaml
type: connector

driver: duckdb
managed: true

init_sql: 
  ATTACH '/path/to/your/duckdb.db' AS external_duckdb;
  INSTALL httpfs;
  LOAD httpfs;
```

## Importing Data to Your External DuckDB

After establishing a connection, you can import data through the connector UI. This process will write data from your attached database to [Rill's embedded DuckDB.](/build/connectors/olap/duckdb#rill-managed-duckdb)

```yaml
# Model YAML
# Reference documentation: https://docs.rilldata.com/reference/project-files/models

type: model
materialize: true

connector: duckdb

sql: SELECT * from external_duckdb.local_table
```
