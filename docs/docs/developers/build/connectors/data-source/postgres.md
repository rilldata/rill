---
title: PostgreSQL
description: Connect to data in PostgreSQL
sidebar_label: PostgreSQL
sidebar_position: 50
---

<!-- WARNING: There are links to this page in source code. If you move it, find and replace the links and consider adding a redirect in docusaurus.config.js. -->

## Overview

[PostgreSQL](https://www.postgresql.org/docs/current/intro-whatis.html) is an open-source object-relational database system known for its reliability, feature robustness, and performance. With support for advanced data types, full ACID compliance for transactional integrity, and an extensible architecture, PostgreSQL provides a highly scalable environment for managing diverse datasets, ranging from small applications to large-scale data warehouses. Its extensive SQL compliance, support for various programming languages, and strong community backing make it a versatile choice for a wide range of business intelligence and analytical applications. You can connect to and read from PostgreSQL databases using either a [supported connection string](https://www.postgresql.org/docs/current/libpq-connect.html#LIBPQ-CONNSTRING) or [connection URI syntax](https://www.postgresql.org/docs/current/libpq-connect.html#LIBPQ-CONNSTRING-URIS).

## Authentication Methods

To connect to PostgreSQL, you need to provide database connection credentials. Rill supports two approaches:

1. **Use Individual Parameters** (recommended for clarity)
2. **Use Connection String/URI** (alternative format)

:::tip Authentication Methods
Choose the method that best fits your setup. Both methods work for local development and Rill Cloud deployments.
:::

## Using the Add Data UI

When you add a PostgreSQL data model through the Rill UI, the process follows two steps:

1. **Configure Authentication** - Set up your PostgreSQL connector with database credentials
2. **Configure Data Model** - Define which table or query to execute

This two-step flow ensures your credentials are securely stored in the connector configuration, while your data model references remain clean and portable.

## Method 1: Individual Parameters (Recommended)

Using individual parameters provides clear, readable configuration for your PostgreSQL connection.

### Using the UI

1. Click **Add Data** in your Rill project
2. Select **PostgreSQL** as the data source type
3. In the authentication step:
   - Enter your host (e.g., `localhost` or your database server address)
   - Enter the port (default: `5432`)
   - Enter your database name
   - Enter your username
   - Enter your password
4. In the data model configuration step, enter your SQL query
5. Click **Create** to finalize

After the model YAML is generated, you can add additional [model settings](/developers/build/models/source-models) directly to the file.

### Manual Configuration

If you prefer to configure manually, create two files:

**Step 1: Create connector configuration**

Create `connectors/my_postgres.yaml`:

```yaml
type: connector
driver: postgres

host: "localhost"
port: "5432"
user: "postgres"
password: "{{ .env.connector.postgres.password }}"
dbname: "my_database"
```

**Step 2: Create model configuration**

Create `models/my_postgres_data.yaml`:

```yaml
type: model
connector: my_postgres

sql: SELECT * FROM my_table

# Add a refresh schedule
refresh:
  cron: "0 */6 * * *"
```

**Step 3: Add credentials to `.env`**

```bash
connector.postgres.password=your_password_here
```

:::tip Did you know?
If this project has already been deployed to Rill Cloud and credentials have been set for this connector, you can use `rill env pull` to [pull these cloud credentials](/developers/build/connectors/credentials#rill-env-pull) locally (into your local `.env` file). Please note that this may override any credentials you have set locally for this source.
:::

## Method 2: Connection String/URI

You can also use a connection string or URI format to configure your PostgreSQL connection.

### Connection String Format

```bash
host=localhost port=5432 dbname=postgres_db user=postgres_user password=postgres_pass
```

### Connection URI Format

```bash
postgresql://postgres_user:postgres_pass@localhost:5432/postgres_db
```

### Manual Configuration

**Step 1: Create connector configuration**

Create `connectors/my_postgres.yaml`:

```yaml
type: connector
driver: postgres

database_url: "{{ .env.connector.postgres.database_url }}"
```

**Step 2: Create model configuration**

Create `models/my_postgres_data.yaml`:

```yaml
type: model
connector: my_postgres

sql: SELECT * FROM my_table

# Add a refresh schedule
refresh:
  cron: "0 */6 * * *"
```

**Step 3: Add credentials to `.env`**

```bash
connector.postgres.database_url=postgresql://postgres_user:postgres_pass@localhost:5432/postgres_db
```

## Using PostgreSQL Data in Models

Once your connector is configured, you can reference PostgreSQL tables and run queries in your model configurations.

### Basic Example

```yaml
type: model
connector: my_postgres

sql: SELECT * FROM my_table

refresh:
  cron: "0 */6 * * *"
```

### Custom SQL Query

```yaml
type: model
connector: my_postgres

sql: |
  SELECT
    date_trunc('day', created_at) as event_date,
    status,
    COUNT(*) as order_count,
    SUM(total_amount) as revenue
  FROM orders
  WHERE created_at >= NOW() - INTERVAL '30 days'
  GROUP BY 1, 2

refresh:
  cron: "0 */6 * * *"
```

## Separating Dev and Prod Environments

When ingesting data locally, consider setting parameters in your connector file to limit how much data is retrieved, since costs can scale with the data source. This also helps other developers clone the project and iterate quickly by reducing ingestion time.

For more details, see our [Dev/Prod setup docs](/developers/build/connectors/templating).

## Deploy to Rill Cloud

When deploying a project to Rill Cloud, Rill requires you to explicitly provide the PostgreSQL connection credentials used in your project. Please refer to our [connector YAML reference docs](/reference/project-files/connectors#postgres) for more information.

If you subsequently add sources that require new credentials (or if you simply entered the wrong credentials during the initial deploy), you can update the credentials by pushing the `Deploy` button to update your project or by running the following command in the CLI:
```
rill env push
```

:::tip Did you know?
If you've already configured credentials locally (in your `<RILL_PROJECT_DIRECTORY>/.env` file), you can use `rill env push` to [push these credentials](/developers/build/connectors/credentials#rill-env-push) to your Rill Cloud project. This will allow other users to retrieve and reuse the same credentials automatically by running `rill env pull`.
:::
