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

In the default `connectors/duckdb.yaml` file, you can add `init_sql` or `attach` statements to attach an external database to Rill's embedded DuckDB. The `attach` statement runs before `init_sql`, allowing you to attach your database and execute subsequent initialization SQL queries. For more details on the YAML configuration, see the [DuckDB reference page](/reference/project-files/connectors#duckdb).

```yaml
# Connector YAML
# Reference documentation: https://docs.rilldata.com/reference/project-files/connectors
  
type: connector

driver: duckdb
managed: true

ATTACH: |
  '/path/to/your/duckdb.db' AS external_duckdb;
  use external_duckdb;

database_name: external_duckdb
schema: main

init_sql: |
  INSTALL httpfs;
  LOAD httpfs;
```

## Using DuckDB Extensions

DuckDB supports a wide variety of extensions that can enhance its functionality. To use extensions with Rill's embedded DuckDB, configure them in your connector's `init_sql` parameter.


### Popular Extensions

For a complete list of available extensions, see the [DuckDB Extensions documentation](https://duckdb.org/docs/extensions/overview).


## Importing Data to Your External DuckDB

After establishing a connection, you can import data through the connector UI. This process will write the data into your attached DuckDB database.

<img src='/img/connect/data-sources/create-model.png' class='rounded-gif' />
<br />

