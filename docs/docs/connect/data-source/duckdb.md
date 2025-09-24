---
title: External DuckDB File
description: Connect to external DuckDB databases and ingest data into Rill
sidebar_label: External DuckDB 
sidebar_position: 11
---

<!-- WARNING: There are links to this page in source code. If you move it, find and replace the links and consider adding a redirect in docusaurus.config.js. -->

[DuckDB](https://duckdb.org/docs/) is a fast, lightweight database designed for data analysis. While Rill includes a [built-in DuckDB database](/connect/olap/duckdb) that stores all your ingested data by default, you may want to import existing data from your own DuckDB files.

## Overview

While we support local usage of your [external DuckDB](/connect/olap/duckdb) as a "live connector", this approach isn't scalable when deploying your project to Rill Cloud. Instead, you can use the DuckDB connector to ingest the required tables into Rill's managed DuckDB database, ensuring your project works seamlessly in both local and cloud environments.

:::warning Location of local DuckDB

If your DuckDB file exists outside of your project folder, it will not be included in the deployment. To ensure a smooth deployment process, move your DuckDB file into the `data/` folder within your project directory. Also, there is a file limitation of 100MB. [Contact us](/contact) if you are having issues with deploying to Rill Cloud.

:::

## Connecting to External DuckDB

To import your existing DuckDB data into Rill, you'll need to establish a connection to your external DuckDB database. Once connected, the data will be read from your external database and ingested into Rill's built-in DuckDB database.

When creating a new DuckDB source from the UI, provide the appropriate path to your DuckDB database file in the **path** field. For a complete list of available properties, see our [reference documentation](/reference/project-files/connectors#external-duckdb).

<img src='/img/connect/data-sources/duckdb.png' class='rounded-gif' style={{width: '75%', display: 'block', margin: '0 auto'}}/>
<br />

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


## Creating a Model

After establishing the connection, create a model through the connector UI. This process will read data from your external DuckDB database and ingest it into Rill's managed DuckDB database, making it available for analysis and visualization.

<img src='/img/connect/data-sources/create-model.png' class='rounded-gif' />
<br />

## Cloud Deployment Considerations

When deploying to Rill Cloud, only the contents of your Rill project directory will be uploaded. If your DuckDB file exists outside of your project folder, it will not be included in the deployment. To ensure a smooth deployment process, move your DuckDB file into the `data/` folder within your project directory.

