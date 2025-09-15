---
title: DuckDB as a Source
description: Connect to data in DuckDB locally 
sidebar_label: DuckDB 
sidebar_position: 11
---

<!-- WARNING: There are links to this page in source code. If you move it, find and replace the links and consider adding a redirect in docusaurus.config.js. -->


[DuckDB](https://duckdb.org/docs/) is a fast, lightweight database designed for data analysis. It runs directly within your application, making it perfect for analytical workloads.

Rill includes a [built-in DuckDB database](/connect/olap/duckdb) that stores all your ingested data, by default. If you already have data in your own DuckDB file and want to visualize your data in Rill, you can import your existing tables into Rill's database.


## Connecting to External DuckDB

In order to import your data into Rill, Rill will first need to be able to access the underlying DuckDB database. Once access has been established, the data will be read from your external database into the built-in DuckDB database in Rill.

If creating a new DuckDB source from the UI, you should provide the appropriate path to the DuckDB database file under **DB** and use the appropriate [DuckDB select statement](https://duckdb.org/docs/sql/statements/select.html) to read the table under **SQL**:

<img src='/img/reference/olap-engines/duckdb/duckdb.png' class='centered' />
<br />



## Cloud Deployment

When deploying to Rill Cloud, only the contents of the Rill data directory will be pushed. In other words, if your DuckDB file exists outside of your project folder, it will not upload to Rill Cloud. To ensure a smooth transition, move the DuckDB file into the data/ folder.



:::tip Live Connector (MotherDuck)

If you already have data in your MotherDuck instance and are thinking of using Rill as an application layer to visualize your data, we recommend using a live connector with MotherDuck.

Please review our [MotherDuck live connector docs](/connect/olap/motherduck) for more information. 
:::
