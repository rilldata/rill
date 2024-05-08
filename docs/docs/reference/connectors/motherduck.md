---
title: DuckDB / MotherDuck
description: Connect to data in DuckDB locally or MotherDuck
sidebar_label: DuckDB / MotherDuck
sidebar_position: 7
---

<!-- WARNING: There are links to this page in source code. If you move it, find and replace the links and consider adding a redirect in docusaurus.config.js. -->

## Overview

[DuckDB](https://duckdb.org/docs/) is an in-process SQL OLAP database management system designed for analytical workloads, aiming to be fast, reliable, and easy to integrate into data analysis applications. It supports standard SQL and operates directly on data in Pandas DataFrames, CSV files, and Parquet files, making it highly suitable for on-the-fly data analysis and machine learning projects. Rill supports natively connecting to and reading from a persisted DuckDB database that it has access to as a source by utilizing the [DuckDB Go driver](https://duckdb.org/docs/api/go.html).

[MotherDuck](https://motherduck.com/docs/getting-started/), on the other hand, is a managed DuckDB-in-the-cloud service, providing enhanced features for scalability, security, and collaboration within larger organizations. It offers advanced management tools, security features like access control and encryption, and support for concurrent access, enabling teams to leverage DuckDB's analytical capabilities at scale while ensuring data governance and security. Similarly, Rill supports natively connecting to and reading from Motherduck as a source by utilizing the [DuckDB Go driver](https://duckdb.org/docs/api/go.html)

![Connecting to DuckDB/MotherDuck](/img/reference/connectors/motherduck/motherduck.png)

## Connecting to DuckDB

As noted above, if you wish to connect to a persistent DuckDB database to read existing tables, Rill will first need to be able to access the underlying DuckDB database. As DuckDB is an _in-memory_ database that's primarily used for local use cases, credentials are not required (and you will typically use Rill Developer). However, if the database file is <u>included</u> in your Git repository, then Rill Cloud will also be able to serve your DuckDB sourced dashboards.

### Local credentials

If creating a new DuckDB source from the UI, you should pass in the appropriate path to the DuckDB database file under **DB** and use the appopriate [DuckDB select statement](https://duckdb.org/docs/sql/statements/select.html) to read in the table under **SQL**:

![Connecting to an existing DuckDB table](/img/reference/connectors/motherduck/duckdb_example.png)

On the other hand, if you are creating the source YAML file directly, the definition should look something like:

```yaml
type: "source"
connector: "duckdb"
sql: "SELECT * from <duckdb_table>"
db: "<path_to_duckdb_db_file>"
```

:::tip If deploying to Rill Cloud

If you plan to deploy a project containing a DuckDB source to Rill Cloud, it is recommended that you move the DuckDB database file to a `data` folder in your Rill project home directory. You can then use the relative path of the db file in your source definition (e.g. `data/test_duckdb.db`).

:::

### Cloud deployment

Once a project with a DuckDB source has been deployed using `rill deploy`, Rill Cloud will need to be able to have access to and retrieve the underlying persisted database file. In most cases, this means that the corresponding DuckDB database file should be included within a directory in your Git repository, which will allow you to specify a relative path in your source definition (from the project root).

:::warning When Using An External DuckDB Database

If the DuckDB database file is external to your Rill project directory, you will still be able to use the fully qualified path to read this SQLite database _locally_ using Rill Developer. However, when deployed to Rill Cloud, this source will throw an **error**.

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

:::tip Did you know?

If this project has already been deployed to Rill Cloud and credentials have been set for this source, you can use `rill env pull` to [pull these cloud credentials](/build/credentials/credentials.md#rill-env-pull) locally (into your local `.env` file). Please note that this may override any credentials that you have set locally for this source.

:::

### Cloud deployment

Once a project with a MotherDuck source has been deployed using `rill deploy`, Rill requires you to explicitly provide the motherduck token using the following command:

```
rill env configure
```

:::info

Note that you must `cd` into the Git repository that your project was deployed from before running `rill env configure`.

:::

:::tip Did you know?

If you've configured credentials locally already (in your `<RILL_PROJECT_DIRECTORY>/.env` file), you can use `rill env push` to [push these credentials](/build/credentials/credentials.md#rill-env-push) to your Rill Cloud project. This will allow other users to retrieve / reuse the same credentials automatically by running `rill env pull`.

:::