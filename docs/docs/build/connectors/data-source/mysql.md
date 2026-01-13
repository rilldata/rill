---
title: MySQL
description: Connect to data in MySQL
sidebar_label: MySQL
sidebar_position: 40
---

<!-- WARNING: There are links to this page in source code. If you move it, find and replace the links and consider adding a redirect in docusaurus.config.js. -->

## Overview

[MySQL](https://dev.mysql.com/doc/refman/8.0/en/introduction.html) is an open-source relational database management system (RDBMS) known for its reliability, performance, and ease of use. It is widely used for a variety of applications, from small to large enterprise projects, supporting structured data storage, retrieval, and management with SQL queries. MySQL offers a comprehensive ecosystem with support for advanced features such as replication, transactions, and full-text indexing, making it a versatile choice for integrating with BI tools. You can connect to and read from MySQL databases directly.

When connecting to MySQL, you need to specify an appropriate Data Source Name (DSN) in the connector's configuration using the following syntax:

```bash
<scheme>://<user>:<password>@<host>:<port>/<database>
```

- **scheme**: The transport protocol to use. Use `mysql` for classic MySQL protocol connections and `mysqlx` for X Protocol connections.
- **user** and **password**: Should correspond to the user credentials that Rill will use to connect to MySQL.
- **host** and **port**: Should correspond to the IP address/hostname and port (default 3306) of your MySQL database.
- **database**: Should correspond to the database in MySQL that you are using.

For more details, see the [MySQL documentation on DSN formats](https://dev.mysql.com/doc/refman/8.4/en/connecting-using-uri-or-key-value-pairs.html#connecting-using-uri).

## Connect to MySQL

Create a connector with your credentials to connect to MySQL. Here's an example connector configuration file you can copy into your `connectors` directory to get started.

```yaml
type: connector 
driver: mysql 

host: "localhost"
port: 3306 
database: "mydatabase" 
user: "myusername" 
password: "{{ .env.MYSQL_PASSWORD }}"
ssl_mode: "DISABLED" 
```

:::tip Using the Add Data Form
You can also use the Add Data form in Rill Developer, which will automatically create the `mysql.yaml` file and populate the `.env` file with `connector.mysql.*` parameters based on the parameters or connection string you provide.
:::

## Separating Dev and Prod Environments

When ingesting data locally, consider setting parameters in your connector file to limit how much data is retrieved, since costs can scale with the data source. This also helps other developers clone the project and iterate quickly by reducing ingestion time.

For more details, see our [Dev/Prod setup docs](/build/connectors/templating).

## Deploy to Rill Cloud

When deploying a project to Rill Cloud, Rill requires you to explicitly provide the MySQL connection string used in your project. Please refer to our [connector YAML reference docs](/reference/project-files/connectors#mysql) for more information.

If you subsequently add sources that require new credentials (or if you simply entered the wrong credentials during the initial deploy), you can update the credentials by pushing the `Deploy` button to update your project or by running the following command in the CLI:
```
rill env push
```