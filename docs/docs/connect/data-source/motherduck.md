---
title: DuckDB / MotherDuck
description: Connect to data in DuckDB locally or MotherDuck
sidebar_label: DuckDB / MotherDuck
sidebar_position: 11
---

<!-- WARNING: There are links to this page in source code. If you move it, find and replace the links and consider adding a redirect in docusaurus.config.js. -->

:::tip Live Connector vs. Ingesting into Rill's Embedded DuckDB

When deciding to use MotherDuck, you'll need to decide whether or not you'll use it as a source connector or a live connector. There are different nuances and use cases for both. 

Please review this documentation as well as our [MotherDuck live connector docs](/connect/olap/motherduck) for more information. 
:::

## Overview



[DuckDB](https://duckdb.org/docs/) is an in-process SQL OLAP database management system designed for analytical workloads, aiming to be fast, reliable, and easy to integrate into data analysis applications. It supports standard SQL and operates directly on data in Pandas DataFrames, CSV files, and Parquet files, making it highly suitable for on-the-fly data analysis and machine learning projects. Rill supports natively connecting to and reading from a persisted DuckDB database that it has access to as a source by utilizing the [DuckDB Go driver](https://duckdb.org/docs/api/go.html).

[MotherDuck](https://motherduck.com/docs/getting-started/) is a managed DuckDB-in-the-cloud service, providing enhanced features for scalability, security, and collaboration within larger organizations. It offers advanced management tools, security features like access control and encryption, and support for concurrent access, enabling teams to leverage DuckDB's analytical capabilities at scale while ensuring data governance and security. Similarly, Rill supports natively connecting to and reading from MotherDuck as a source by utilizing the [DuckDB Go driver](https://duckdb.org/docs/api/go.html).



## Connecting to External DuckDB/MotherDuck

As noted above, if you wish to connect to a persistent DuckDB database to read existing tables, Rill will first need to be able to access the underlying DuckDB or MotherDuck database. Once access has been established, the data will be read from your external database into the built-in DuckDB database in Rill.

### Local credentials (DuckDB)

If creating a new DuckDB source from the UI, you should provide the appropriate path to the DuckDB database file under **DB** and use the appropriate [DuckDB select statement](https://duckdb.org/docs/sql/statements/select.html) to read in the table under **SQL**:

<img src='/img/reference/olap-engines/duckdb/duckdb.png' class='centered' />
<br />

### Local credentials (MotherDuck)

When using Rill Developer on your local machine (i.e., `rill start`), Rill will use the `motherduck_token` configured in your environment variables to attempt to establish a connection with MotherDuck. If this is not defined, you will need to set this environment variable appropriately. 

<img src='/img/reference/connectors/motherduck/motherduck.png' class='centered' />
<br />

Alternatively, you can create a connector using our [connector YAML reference docs](/reference/project-files/connectors#motherduck-as-a-source).

:::tip Credentials 
Don't forget to set up your MotherDuck token! 

```bash
export motherduck_token='<token>'
```

or via the `.env` file.

```yaml
motherduck_token='token'
```
:::

### Cloud deployment

Once a project with a DuckDB source has been deployed, Rill Cloud will need to be able to access and retrieve the underlying persisted database file. In most cases, this means that the corresponding DuckDB database file should be included within a directory in your Git repository, which will allow you to specify a relative path in your source definition (from the project root). Or, for MotherDuck, the motherduck_token must be defined.


:::tip If deploying to Rill Cloud

If you plan to deploy a project containing a DuckDB source to Rill Cloud, it is recommended that you move the DuckDB database file to a `data` folder in your Rill project home directory. You can then use the relative path of the db file in your source definition (e.g., `data/test_duckdb.db`).

:::
