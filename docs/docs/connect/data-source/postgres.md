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


## Connect to PostgreSQL

When using Rill Developer on your local machine (i.e., `rill start`), Connect to PostgreSQL via Add Data. This will automatically create the `postgres.yaml` file in your connectors/ folder and populate the `.env` file with `connector.postgres.*` parameters depending on if you inputted parameters or connection string.

```yaml
# Connector YAML
# Reference documentation: https://docs.rilldata.com/reference/project-files/connectors
  
type: connector

driver: postgres
host: "localhost"
port: "5432"
user: "postgres"
password: "{{ .env.connector.postgres.password }}"
dbname: "postgres"
```

:::tip Did you know?

If this project has already been deployed to Rill Cloud and credentials have been set for this connector, you can use `rill env pull` to [pull these cloud credentials](/connect/credentials.md#rill-env-pull) locally (into your local `.env` file). Please note that this may override any credentials you have set locally for this source.

:::

## Separating Dev and Prod Environments

When ingesting data locally, consider setting parameters in your connector file to limit how much data is retrieved, since costs can scale with the data source. This also helps other developers clone the project and iterate quickly by reducing ingestion time.

For more details, see our [Dev/Prod setup docs](/connect/templating).

## Cloud deployment

Once a project with a MySQL source has been deployed, Rill requires you to explicitly provide the connection string using the following command:

```
rill env configure
```


:::tip Did you know?

If you've already configured credentials locally (in your `<RILL_PROJECT_DIRECTORY>/.env` file), you can use `rill env push` to [push these credentials](/connect/credentials.md#rill-env-push) to your Rill Cloud project. This will allow other users to retrieve and reuse the same credentials automatically by running `rill env pull`.

:::