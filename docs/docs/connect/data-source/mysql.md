---
title: MySQL
description: Connect to data in MySQL
sidebar_label: MySQL
sidebar_position: 45
---

<!-- WARNING: There are links to this page in source code. If you move it, find and replace the links and consider adding a redirect in docusaurus.config.js. -->

[MySQL](https://dev.mysql.com/doc/refman/8.0/en/introduction.html) is an open-source relational database management system (RDBMS) known for its reliability, performance, and ease of use. It is widely used for a variety of applications, from small to large enterprise projects, supporting structured data storage, retrieval, and management with SQL queries. MySQL offers a comprehensive ecosystem with support for advanced features such as replication, transactions, and full-text indexing, making it a versatile choice for integrating with BI tools. Rill supports natively connecting to and reading from MySQL as a source by using the [Go MySQL Driver](https://github.com/go-sql-driver/mysql).

When connecting to MySQL, an appropriate Data Source Name (DSN) must be specified in the connector's configuration using the following syntax:

```bash
<username>:<password>@<protocol>(<hostname>:<port>)/<database_name>
```
- **username** and **password** should correspond to the user that Rill will use to connect to MySQL
- **protocol** will typically be _tcp_ (unless otherwise specified)
- **hostname** and **port** should correspond to the respective IP address/hostname and port (default 3306) of your MySQL database
- **database_name** should correspond to the database in MySQL that you are using

<img src='/img/connect/connectors/mysql/mysql.png' class='centered' />
<br />

## Local credentials

When using Rill Developer on your local machine, you will need to provide your credentials via a connector file. We would recommend not using plain text to create your file and instead use the `.env` file. For more details on your connector, see [connector YAML](/reference/project-files/connectors#mysql) for more details.

:::tip Updating the project environmental variable

If you've already deployed to Rill Cloud, you can either [push/pull the credential](/manage/project-management/variables-and-credentials#pushing-and-pulling-credentials-to--from-rill-cloud-via-the-cli) from the CLI with:
```
rill env push
rill env pull
```
:::

## Separating Dev and Prod Environments

When ingesting data locally, consider setting parameters in your connector file to limit how much data is retrieved, since costs can scale with the data source. This also helps other developers clone the project and iterate quickly by reducing ingestion time.

For more details, see our [Dev/Prod setup docs](/connect/templating).

## Cloud deployment

Once a project with a MySQL source has been deployed, Rill requires you to explicitly provide the connection string using the following command:

```
rill env configure
```

