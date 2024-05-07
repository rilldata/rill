---
title: MySQL
description: Connect to data in MySQL
sidebar_label: MySQL
sidebar_position: 9
---

<!-- WARNING: There are links to this page in source code. If you move it, find and replace the links and consider adding a redirect in docusaurus.config.js. -->

## Overview

[MySQL](https://dev.mysql.com/doc/refman/8.0/en/introduction.html) is an open-source relational database management system (RDBMS) known for its reliability, performance, and ease of use. It is widely used for a variety of applications, from small to large enterprise projects, supporting structured data storage, retrieval, and management with SQL queries. MySQL offers a comprehensive ecosystem with support for advanced features, such as replication, transactions, and full-text indexing, making it a versatile choice for integrating with BI tools. Rill supports natively connecting to and reading from MySQL as a source by using the [Go MySQL Driver](https://github.com/go-sql-driver/mysql).

When connecting to MySQL, an appropriate Data Source Name (DSN) will need to be specified in the connector's configuration using the following syntax:

```bash

<username>:<password>@<protocol>(<hostname>:<port>)/<database_name>

```
- **username** and **password** should correspond to the user that Ril will be using to connect to MySQL
- **protocol** will typically be _tcp_ (unless otherwise specified)
- **hostname** and **port** should correspond to the respective ip address / hostname and port (default 3306) of your MySQL database
- **database_name** should correspond to the database in MySQL that you are using

![Connecting to MySQL](/img/reference/connectors/mysql/mysql.png)

## Local credentials

When using Rill Developer on your local machine (i.e. `rill start`), you have the option to specify a connection string when running Rill using the `--var` flag.
An example of passing the connection DSN to Rill via the terminal:

```bash
rill start --var connector.mysql.dsn="mysql_user:mysql_password@tcp(localhost:3306)/mysql_db"
```

Alternatively, you can include the connection string directly in the source YAML definition by adding the `database_url` parameter. 
An example of a source using this approach:

```yaml
type: "source"
connector: "mysql"
sql: "select * from my_table"
dsn: "mysql_user:mysql_password@tcp(localhost:3306)/mysql_db"
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

Once a project with a MySQL source has been deployed using `rill deploy`, Rill requires you to explicitly provide the connection string using the following command:

```
rill env configure
```

:::info

Note that you must `cd` into the Git repository that your project was deployed from before running `rill env configure`.

:::

:::tip Did you know?

If you've configured credentials locally already (in your `<RILL_HOME>/.env` file), you can use `rill env push` to [push these credentials](/build/credentials/credentials.md#rill-env-push) to your Rill Cloud project. This will allow other users to retrieve / reuse the same credentials automatically by running `rill env pull`.

:::