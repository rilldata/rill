---
title: DuckDB / MotherDuck
description: Connect to data in DuckDB locally or MotherDuck
sidebar_label: DuckDB / MotherDuck
sidebar_position: 6
---

<!-- WARNING: There are links to this page in source code. If you move it, find and replace the links and consider adding a redirect in docusaurus.config.js. -->

## Overview

[DuckDB](https://duckdb.org/docs/) is an in-process SQL OLAP database management system designed for analytical workloads, aiming to be fast, reliable, and easy to integrate into data analysis applications. It supports standard SQL and operates directly on data in Pandas DataFrames, CSV files, and Parquet files, making it highly suitable for on-the-fly data analysis and machine learning projects. As DuckDB is an in-memory database that only runs locally, *Rill Developer* supports natively connecting to and reading from a persisted DuckDB database as a source by utilizing the [DuckDB Go driver](https://duckdb.org/docs/api/go.html).

[MotherDuck](https://motherduck.com/docs/getting-started/), on the other hand, is a managed DuckDB-in-the-cloud service, providing enhanced features for scalability, security, and collaboration within larger organizations. It offers advanced management tools, security features like access control and encryption, and support for concurrent access, enabling teams to leverage DuckDB's analytical capabilities at scale while ensuring data governance and security. Similarly, Rill supports natively conecting to and reading from Motherduck as a source by utilizing the [DuckDB Go driver](https://duckdb.org/docs/api/go.html)

![Connecting to DuckDB/MotherDuck](/img/reference/connectors/motherduck/motherduck.png)

## Connecting to DuckDB

As noted above, if you wish to connect to a persistent DuckDB database so that you can use existing tables in Rill, <u>this has to be done from Rill Developer</u> and Rill Developer has to be run on the same instance / machine from where your DuckDB resides.

![Connecting to local DuckDB](/img/reference/connectors/motherduck/duckdb.png)

:::danger Rill Cloud

Given the architectures of both technologies, it is **not** possible to connect to your local DuckDB database from a deployed project in Rill Cloud. Attempting to do so will lead to unexpected behavior for your defined source (and result in errors for associated models and dashboards).

:::

## Connecting to MotherDuck

### Local credentials

When using Rill Developer on your local machine (i.e. `rill start`), Rill will use the `motherduck_token` configured in your environment variables to attempt to establish a connection with MotherDuck. If this is not defined, you will need to set this environment variable appropriately. 

```bash
export motherduck_token='<token>'
```

:::tip

An alternative option would be to set this line through your bash profile.

:::

:::info

For more information about authenticating with an appropriate service token, please refer to [MotherDuck's documentation](https://motherduck.com/docs/authenticating-to-motherduck/#using-the-service-token-to-connect).

:::

### Cloud deployment

Once a project with a MotherDuck source has been deployed using `rill deploy`, Rill requires you to explicitly provide the motherduck token using the following command:

```
rill env configure
```

:::info

Note that you must `cd` into the Git repository that your project was deployed from before running `rill env configure`.

:::
