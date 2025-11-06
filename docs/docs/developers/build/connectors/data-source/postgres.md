---
title: PostgreSQL
description: Connect to data in PostgreSQL
sidebar_label: PostgreSQL
sidebar_position: 50
---

<!-- WARNING: There are links to this page in source code. If you move it, find and replace the links and consider adding a redirect in docusaurus.config.js. -->

## Overview

[PostgreSQL](https://www.postgresql.org/docs/current/intro-whatis.html) is an open-source object-relational database system known for its reliability, feature robustness, and performance. With support for advanced data types, full ACID compliance for transactional integrity, and an extensible architecture, PostgreSQL provides a highly scalable environment for managing diverse datasets, ranging from small applications to large-scale data warehouses. Its extensive SQL compliance, support for various programming languages, and strong community backing make it a versatile choice for a wide range of business intelligence and analytical applications. You can connect to and read from PostgreSQL databases using either a [supported connection string](https://www.postgresql.org/docs/current/libpq-connect.html#LIBPQ-CONNSTRING) or [connection URI syntax](https://www.postgresql.org/docs/current/libpq-connect.html#LIBPQ-CONNSTRING-URIS).

As an example of a connection string:
```bash
host=localhost port=5432 dbname=postgres_db user=postgres_user password=postgres_pass
```

Using the same example, this would be an equivalent connection URI:
```bash
postgresql://postgres_user:postgres_pass@localhost:5432/postgres_db
```


## Connect to PostgreSQL

Create a connector with your credentials to connect to PostgreSQL. Here's an example connector configuration file you can copy into your `connectors` directory to get started:

```yaml
type: connector

driver: postgres
host: "localhost"
port: "5432"
user: "postgres"
password: "{{ .env.connector.postgres.password }}"
dbname: "postgres"
```

:::tip Using the Add Data Form
You can also use the Add Data form in Rill Developer, which will automatically create the `postgres.yaml` file and populate the `.env` file with `connector.postgres.*` parameters based on the parameters or connection string you provide.
:::

## Separating Dev and Prod Environments

When ingesting data locally, consider setting parameters in your connector file to limit how much data is retrieved, since costs can scale with the data source. This also helps other developers clone the project and iterate quickly by reducing ingestion time.

For more details, see our [Dev/Prod setup docs](/build/connectors/templating).

## Deploy to Rill Cloud

When deploying your project to Rill Cloud, you must explicitly provide the PostgreSQL connection string. If these credentials exist in your `.env` file, they'll be pushed with your project automatically.

To manually configure your environment variables, run:
```bash
rill env configure
```