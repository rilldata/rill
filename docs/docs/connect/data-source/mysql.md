---
title: MySQL
description: Connect to data in MySQL
sidebar_label: MySQL
sidebar_position: 9
---

<!-- WARNING: There are links to this page in source code. If you move it, find and replace the links and consider adding a redirect in docusaurus.config.js. -->

## Overview

[MySQL](https://dev.mysql.com/doc/refman/8.0/en/introduction.html) is an open-source relational database management system (RDBMS) known for its reliability, performance, and ease of use. It is widely used for a variety of applications, from small to large enterprise projects, supporting structured data storage, retrieval, and management with SQL queries. MySQL offers a comprehensive ecosystem with support for advanced features such as replication, transactions, and full-text indexing, making it a versatile choice for integrating with BI tools. Rill supports natively connecting to and reading from MySQL.

When connecting to MySQL, an appropriate Data Source Name (DSN) must be specified in the connector's configuration using the following syntax:

```bash
<scheme>://<user>:<password>@<host>:<port>/<database>
```
- **scheme** The transport protocol to use. Use `mysql` for classic MySQL protocol connections and  `mysqlx` for X Protocol connections.
- **user** and **password** should correspond to the user that Rill will use to connect to MySQL
- **host** and **port** should correspond to the respective IP address/hostname and port (default 3306) of your MySQL database
- **database** should correspond to the database in MySQL that you are using

For more details, see the [MySQL documentation on DSN formats](https://dev.mysql.com/doc/refman/8.4/en/connecting-using-uri-or-key-value-pairs.html#connecting-using-uri).

<img src='/img/reference/connectors/mysql/mysql.png' class='centered' />
<br />

## Local credentials

When using Rill Developer on your local machine (i.e., `rill start`), you have the option to specify a connection string when running Rill using the `--env` flag.
An example of passing the connection DSN to Rill via the terminal:

```bash
rill start --env connector.mysql.dsn="mysql://mysql_user:mysql_password@localhost:3306/mysql_db"
```

Alternatively, you can include the connection string directly in the source YAML definition by adding the `dsn` parameter.
An example of a source using this approach:

```yaml
type: "model"
connector: "mysql"
sql: "select * from my_table"
dsn: "mysql://mysql_user:mysql_password@localhost:3306/mysql_db"
```

:::warning Beware of committing credentials to Git

This second approach is generally not recommended outside of local development because it places the connection details (which may contain sensitive information like passwords) in the source file, <u>which is committed to Git</u>.

:::

:::info Source Properties

For more information about available source properties and configurations, please refer to our reference documentation on [Source YAML](/reference/project-files/index.md).

:::

:::tip Did you know?

If this project has already been deployed to Rill Cloud and credentials have been set for this source, you can use `rill env pull` to [pull these cloud credentials](/connect/credentials.md#rill-env-pull) locally (into your local `.env` file). Please note that this may override any credentials you have set locally for this source.

:::

## Separating Dev and Prod Environments

When ingesting data locally, consider setting parameters in your connector file to limit how much data is retrieved, since costs can scale with the data source. This also helps other developers clone the project and iterate quickly by reducing ingestion time.

For more details, see our [Dev/Prod setup docs](/connect/templating).

## Cloud deployment

Once a project with a MySQL source has been deployed, Rill requires you to explicitly provide the connection string using the following command:

```
rill env configure
```

:::info

Note that you must `cd` into the Git repository from which your project was deployed before running `rill env configure`.

:::

:::tip Did you know?

If you've already configured credentials locally (in your `<RILL_PROJECT_DIRECTORY>/.env` file), you can use `rill env push` to [push these credentials](/connect/credentials.md#rill-env-push) to your Rill Cloud project. This will allow other users to retrieve and reuse the same credentials automatically by running `rill env pull`.

:::