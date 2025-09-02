---
title: DuckDB 
description: Connect to data in DuckDB locally 
sidebar_label: DuckDB 
sidebar_position: 11
---

<!-- WARNING: There are links to this page in source code. If you move it, find and replace the links and consider adding a redirect in docusaurus.config.js. -->

[DuckDB](https://duckdb.org/docs/) is an in-process SQL OLAP database management system designed for analytical workloads, aiming to be fast, reliable, and easy to integrate into data analysis applications. It supports standard SQL and operates directly on data in Pandas DataFrames, CSV files, and Parquet files, making it highly suitable for on-the-fly data analysis and machine learning projects. Rill supports natively connecting to and reading from a persisted DuckDB database that it has access to as a source by utilizing the [DuckDB Go driver](https://duckdb.org/docs/api/go.html).


## Connecting to External DuckDB

As noted above, if you wish to connect to a persistent DuckDB database to read existing tables, Rill will first need to be able to access the underlying DuckDB or MotherDuck database. Once access has been established, the data will be read from your external database into the built-in DuckDB database in Rill.

### Local credentials (DuckDB)

If creating a new DuckDB source from the UI, you should provide the appropriate path to the DuckDB database file under **DB** and use the appropriate [DuckDB select statement](https://duckdb.org/docs/sql/statements/select.html) to read in the table under **SQL**:

<img src='/img/connect/olap-engines/duckdb/duckdb.png' class='centered' />
<br />

:::warning RILL's DATA DIRECTORY 

When deploying to Rill Cloud, only the contents of the Rill data directory will be pushed. In other words, if you DuckDB file exists outside of your project folder, it will not upload to Rill Cloud. To ensure a smooth transition, move the DuckDB file into the data/ folder.

:::


## Cloud deployment

Once a project with a PostgreSQL source has been deployed, Rill requires you to explicitly provide the connection string using the following command:

```
rill env configure
```


:::tip Live Connector (MotherDuck) vs. Connecting to local DuckDB

If you already have data in your MotherDuck instance or local DuckDB, and/or are thinking of using Rill as an application layer over an existing database, we recommend using a live connector with MotherDuck.

Please review this documentation as well as our [MotherDuck live connector docs](/connect/olap/motherduck) for more information. 
:::
