---
title: PostgreSQL
description: Connect to data in PostgreSQL
sidebar_label: PostgreSQL
sidebar_position: 50
---

<!-- WARNING: There are links to this page in source code. If you move it, find and replace the links and consider adding a redirect in docusaurus.config.js. -->

## Overview

[PostgreSQL](https://www.postgresql.org/docs/current/intro-whatis.html) is an open-source object-relational database system known for its reliability, feature robustness, and performance. With support for advanced data types, full ACID compliance for transactional integrity, and an extensible architecture, PostgreSQL provides a highly scalable environment for managing diverse datasets, ranging from small applications to large-scale data warehouses. Its extensive SQL compliance, support for various programming languages, and strong community backing make it a versatile choice for a wide range of business intelligence and analytical applications. Rill supports natively connecting to and reading from PostgreSQL as a source by using either a [supported connection string](https://www.postgresql.org/docs/current/libpq-connect.html#LIBPQ-CONNSTRING) or [connection URI syntax](https://www.postgresql.org/docs/current/libpq-connect.html#LIBPQ-CONNSTRING-URIS).

As an example of a connection string:
```bash
host=localhost port=5432 dbname=postgres_db user=postgres_user password=postgres_pass
```

Using the same example, this would be an equivalent connection URI:
```bash
postgresql://postgres_user:postgres_pass@localhost:5432/postgres_db
```

<img src='/img/reference/connectors/postgres/postgresql.png' class='centered' />
<br />

## Local credentials

When using Rill Developer on your local machine, you will need to provide your credentials via a connector file. We would recommend not using plain text to create your file and instead use the `.env` file. For more details on your connector, see [connector YAML](/reference/project-files/connectors/#postgresql) for more details.

:::tip Updating the project environmental variable

If you've already deployed to Rill Cloud, you can either [push/pull the credential](/manage/project-management/variables-and-credentials#pushing-and-pulling-credentials-to--from-rill-cloud-via-the-cli) from the CLI with:
```
rill env push
rill env pull
```

Or, if its your first deployment, Rill will automatically deploy the .env into your Rill project.

:::

## Cloud deployment

Once a project with a PostgreSQL source has been deployed, Rill requires you to explicitly provide the connection string using the following command:

```
rill env configure
```
