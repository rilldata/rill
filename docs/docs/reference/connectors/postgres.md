---
title: PostgreSQL
description: Connect to data in PostgreSQL
sidebar_label: PostgreSQL
sidebar_position: 8
---

<!-- WARNING: There are links to this page in source code. If you move it, find and replace the links and consider adding a redirect in docusaurus.config.js. -->

## Overview

[PostgreSQL](https://www.postgresql.org/docs/current/intro-whatis.html) is an open-source object-relational database system known for its reliability, feature robustness, and performance. With support for advanced data types, full ACID compliance for transactional integrity, and extensible architecture, PostgreSQL provides a highly scalable environment for managing diverse datasets ranging from small applications to large-scale data warehouses. Its extensive SQL compliance, support for various programming languages, and strong community backing make it a versatile choice for a wide range of business intelligence and analytical applications. Rill supports natively connecting to and reading from PostgreSQL as a source by using either a [supported connection string](https://www.postgresql.org/docs/current/libpq-connect.html#LIBPQ-CONNSTRING) or [connection URI syntax](https://www.postgresql.org/docs/current/libpq-connect.html#LIBPQ-CONNSTRING-URIS).

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

Alternatively, you can include the connection string directly in the source YAML definition by adding the `database_url` parameter. 
An example of a source using this approach:

```yaml
type: "postgres"
sql: "select * from my_table"
database_url: "postgresql://postgres:postgres@localhost:5432/postgres"
```

:::warning Beware of committing credentials to Git

This second approach is generally not recommended outside of local development because it places the connection details (which may contain sensitive information like passwords!) in the source file, <u>which is committed to Git</u>.

:::

:::info Source Properties

For more information about available source properties / configurations, please refer to our reference documentation on [Source YAML](../../reference/project-files/index.md).

:::

:::tip Did you know?

If this project has already been deployed to Rill Cloud and credentials have been set for this source, you can use `rill env pull` to [pull these cloud credentials](/build/credentials/credentials.md#rill-env-pull) locally (into your local `.env` file). Please note that this may override any credentials that you have set locally for this source.

:::

## Cloud deployment

Once a project with a PostgreSQL source has been deployed using `rill deploy`, Rill requires you to explicitly provide the connection string using the following command:

```
rill env configure
```

:::info

Note that you must `cd` into the Git repository that your project was deployed from before running `rill env configure`.

:::

:::tip Did you know?

If you've configured credentials locally already (in your `<RILL_HOME>/.home` file), you can use `rill env push` to [push these credentials](/build/credentials/credentials.md#rill-env-push) to your Rill Cloud project. This will allow other users to retrieve / reuse the same credentials automatically by running `rill env pull`.

:::