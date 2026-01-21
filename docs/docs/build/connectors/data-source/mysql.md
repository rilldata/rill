---
title: MySQL
description: Connect to data in MySQL
sidebar_label: MySQL
sidebar_position: 40
---

<!-- WARNING: There are links to this page in source code. If you move it, find and replace the links and consider adding a redirect in docusaurus.config.js. -->

## Overview

[MySQL](https://dev.mysql.com/doc/refman/8.0/en/introduction.html) is an open-source relational database management system (RDBMS) known for its reliability, performance, and ease of use. It is widely used for a variety of applications, from small to large enterprise projects, supporting structured data storage, retrieval, and management with SQL queries. MySQL offers a comprehensive ecosystem with support for advanced features such as replication, transactions, and full-text indexing, making it a versatile choice for integrating with BI tools. You can connect to and read from MySQL databases directly.

## Connection String Format

When connecting to MySQL, you need to specify an appropriate Data Source Name (DSN) using the following syntax:

```bash
<scheme>://<user>:<password>@<host>:<port>/<database>
```

- **scheme**: The transport protocol to use. Use `mysql` for classic MySQL protocol connections and `mysqlx` for X Protocol connections.
- **user** and **password**: Should correspond to the user credentials that Rill will use to connect to MySQL.
- **host** and **port**: Should correspond to the IP address/hostname and port (default 3306) of your MySQL database.
- **database**: Should correspond to the database in MySQL that you are using.

For more details, see the [MySQL documentation on DSN formats](https://dev.mysql.com/doc/refman/8.4/en/connecting-using-uri-or-key-value-pairs.html#connecting-using-uri).

## Using the Add Data UI

When you add a MySQL data model through the Rill UI, the process follows two steps:

1. **Configure Authentication** - Set up your MySQL connector with connection credentials (host, port, user, password, database)
2. **Configure Data Model** - Define which table or query to execute

This two-step flow ensures your credentials are securely stored in the connector configuration, while your data model references remain clean and portable.

---

## Connect to MySQL

### Using the UI

1. Click **Add Data** in your Rill project
2. Select **MySQL** as the data source type
3. In the authentication step:
   - Enter your MySQL host and port
   - Enter your database name
   - Enter your username and password
   - Configure SSL mode if needed
4. In the data model configuration step:
   - Enter your SQL query
   - Configure other model settings as needed
5. Click **Create** to finalize

The UI will automatically create both the connector file and model file for you.

### Manual Configuration

If you prefer to configure manually, create two files:

**Step 1: Create connector configuration**

Create `connectors/my_mysql.yaml`:

```yaml
type: connector
driver: mysql

host: "localhost"
port: 3306
database: "mydatabase"
user: "myusername"
password: "{{ .env.connector.mysql.password }}"
ssl_mode: "DISABLED"
```

**Step 2: Create model configuration**

Create `models/my_mysql_data.yaml`:

```yaml
type: model
connector: my_mysql

sql: SELECT * FROM my_table

# Add a refresh schedule
refresh:
  cron: "0 */6 * * *"
```

**Step 3: Add credentials to `.env`**

```bash
connector.mysql.password=your-secure-password
```

---

## Using MySQL Data in Models

Once your connector is configured, you can reference MySQL tables and run queries in your model configurations.

### Basic Example

```yaml
type: model
connector: my_mysql

sql: SELECT * FROM customers

refresh:
  cron: "0 */6 * * *"
```

### Custom SQL Query

```yaml
type: model
connector: my_mysql

sql: |
  SELECT
    DATE(created_at) as order_date,
    status,
    COUNT(*) as order_count,
    SUM(total_amount) as total_revenue
  FROM orders
  WHERE created_at >= DATE_SUB(CURDATE(), INTERVAL 30 DAY)
  GROUP BY 1, 2

refresh:
  cron: "0 */6 * * *"
```

---

## Separating Dev and Prod Environments

When ingesting data locally, consider setting parameters in your connector file to limit how much data is retrieved, since costs can scale with the data source. This also helps other developers clone the project and iterate quickly by reducing ingestion time.

For more details, see our [Dev/Prod setup docs](/build/connectors/templating).

---

## Deploy to Rill Cloud

When deploying a project to Rill Cloud, Rill requires you to explicitly provide the MySQL connection credentials used in your project. Please refer to our [connector YAML reference docs](/reference/project-files/connectors#mysql) for more information.

If you subsequently add sources that require new credentials (or if you simply entered the wrong credentials during the initial deploy), you can update the credentials by pushing the `Deploy` button to update your project or by running the following command in the CLI:
```
rill env push
```
