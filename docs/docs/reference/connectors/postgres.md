---
title: Postgres
description: Connect to data in a Postgres server
sidebar_label: Postgres
sidebar_position: 70
---

<!-- WARNING: There are links to this page in source code. If you move it, find and replace the links and consider adding a redirect in docusaurus.config.js. -->

## How to configure credentials in Rill

Rill utilizes a PostgreSQL connection string to retrieve the necessary connection parameters for establishing a connection with PostgreSQL. For detailed information on connection strings, please consult the [PostgreSQL documentation](https://www.postgresql.org/docs/current/libpq-connect.html#LIBPQ-CONNSTRING).
How you configure postgres connection string depends on whether you are developing a project locally using `rill start` or are setting up a deployment using `rill deploy`.

### Configure credentials for local development

When working on a local project, you have the option to specify a connection string when running Rill using the `--var` flag.
An example of using this syntax in terminal:
```
rill start --var connector.postgres.database_url="postgresql://postgres:postgres@localhost:5432/postgres"
```

Alternatively, you can include the connection string directly in the source code by adding the `database_url` parameter. 
An example of a source using this approach:
```
type: "postgres"
sql: "select * from my_table"
database_url: "postgresql://postgres:postgres@localhost:5432/postgres"
```
This approach is less recommended because it places the connection string (which may contain sensitive information like passwords) in the source file, which is committed to Git. For more information, please refer to the documentation on [sources](../../reference/project-files/index.md).

### Configure credentials for deployments on Rill Cloud

Once a project having a Postgres source has been deployed using `rill deploy`, Rill requires you to explicitly provide the connection string using following command:
```
rill env configure
```
Note that you must `cd` into the Git repository that your project was deployed from before running `rill env configure`.
