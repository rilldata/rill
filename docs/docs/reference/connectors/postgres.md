---
title: PostgreSQL
description: Connect to data in a PostgreSQL database
sidebar_label: PostgreSQL
sidebar_position: 7
---

<!-- WARNING: There are links to this page in source code. If you move it, find and replace the links and consider adding a redirect in docusaurus.config.js. -->

## Overview

PostgreSQL, often referred to as Postgres, is an open-source object-relational database system known for its reliability, feature robustness, and performance. With support for advanced data types, full ACID compliance for transactional integrity, and extensible architecture, Postgres provides a highly scalable environment for managing diverse datasets ranging from small applications to large-scale data warehouses. Its extensive SQL compliance, support for various programming languages, and strong community backing make it a versatile choice for a wide range of business intelligence and analytical applications. Rill supports natively connecting to and reading from Postgres as a source by using either a [supported connection string](https://www.postgresql.org/docs/current/libpq-connect.html#LIBPQ-CONNSTRING) or [connection URI syntax](https://www.postgresql.org/docs/current/libpq-connect.html#LIBPQ-CONNSTRING-URIS).

As an example of a connection string:
```bash
host=localhost port=5432 dbname=postgres_db user=postgres_user password=postgres_pass
```

Using the same example, this would be an equivalent connection URI:
```bash
postgresql://postgres_user:postgres_pass@localhost:5432/postgres_db
```

![Connecting to PostgreSQL](/img/reference/connectors/postgres/postgresql.png)

## Local credentials

When using Rill Developer on your local machine (i.e. `rill start`), you have the option to specify a connection string when running Rill using the `--var` flag.
An example of using this syntax in terminal:

```bash
rill start --var connector.postgres.database_url="postgresql://postgres:postgres@localhost:5432/postgres"
```

Alternatively, you can include the connection string directly in the source code by adding the `database_url` parameter. 
An example of a source using this approach:

```yaml
type: "postgres"
sql: "select * from my_table"
database_url: "postgresql://postgres:postgres@localhost:5432/postgres"
```

:::warning

This approach is generally not recommended outside of local development because it places the connection string (which may contain sensitive information like passwords!) in the source file, <u>which is committed to Git</u>.

:::

:::info Source Properties

For more information about available source properties / configurations, please refer to our reference documentation on [Source YAML](../../reference/project-files/index.md).

:::

## Cloud deployment

Once a project with a PostgreSQL source has been deployed using `rill deploy`, Rill requires you to explicitly provide the connection string using the following command:

```
rill env configure
```

:::info

Note that you must `cd` into the Git repository that your project was deployed from before running `rill env configure`.

:::