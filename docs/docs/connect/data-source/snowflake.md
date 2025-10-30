---
title: Snowflake
sidebar_label: Snowflake
sidebar_position: 60
hide_table_of_contents: true
---

## Overview

:::warning New DSN Connector Format (December 2024)

Snowflake has issued a [deprecation notice](https://www.snowflake.com/en/blog/blog-notice-snowflake-connector-python/) for the prior authentication method supported by Rill. We have migrated to using the newer Go driver format starting Rill version 0.55, which introduces [new Snowflake connection configuration options](../../reference/connectors/snowflake.md).

:::

[Snowflake](https://docs.snowflake.com/en/user-guide-intro) is a cloud-based data platform designed to facilitate data warehousing, data lakes, data engineering, data science, data application development, and data sharing. It separates compute and storage, enabling users to scale up or down instantly without downtime, providing a cost-effective solution for data management. With its unique architecture and support for multi-cloud environments, including AWS, Azure, and Google Cloud Platform, Snowflake offers seamless data integration, secure data sharing across organizations, and real-time access to data insights, making it a common choice to power many business intelligence applications and use cases. You can connect to and read from Snowflake data warehouses using the [Go Snowflake Driver](https://pkg.go.dev/github.com/snowflakedb/gosnowflake).

## Local Credentials

### Adding a Snowflake Connector

Create a new connector YAML file in the `connectors` directory of your Rill project (e.g. `connectors/snowflake.yaml`) and configure it with your Snowflake credentials:

```yaml
# Snowflake Connector YAML
type: connector
driver: snowflake

# DSN connection string (see Snowflake documentation)
dsn: "{{ .env.connector.snowflake.dsn }}"
```

### Storing Credentials in `.env`

Store your Snowflake credentials in your project's `.env` file:

```bash
connector.snowflake.dsn=username:password@account/database/schema?warehouse=warehouse_name
```

The DSN format follows the structure:
```
username:password@account/database/schema?warehouse=warehouse_name&role=role_name
```

For detailed information on DSN parameters and authentication methods, see the [Snowflake connector reference](../../reference/connectors/snowflake.md).

:::tip Key-Pair Authentication

For enhanced security in production, consider using [key-pair authentication](../../reference/connectors/snowflake.md#key-pair-authentication) instead of password-based authentication.

:::

## Configuring Snowflake as a Source

Once your Snowflake connector is configured, you can reference Snowflake tables in your models.

### Method 1 - Direct SQL Query in Model

Create a model SQL file that directly queries your Snowflake table:

```sql
-- models/my_snowflake_data.sql

-- Query Snowflake table directly using the connector
-- Replace 'snowflake' with your connector name if different
SELECT * FROM snowflake.database.schema.table_name
```

### Method 2 - Creating a Model YAML File

For more control over refresh schedules and configuration, create a model YAML file:

```yaml
# models/my_snowflake_data.yaml
type: model

# Reference your Snowflake connector
connector: snowflake

# SQL query to execute in Snowflake
sql: SELECT * FROM database.schema.table_name

# Optional: Set a refresh schedule
refresh:
  cron: "0 */6 * * *"  # Refresh every 6 hours
```

:::info

For the complete list of model properties, see the [model YAML reference](../../reference/project-files/models.md).

:::

## Deploying to Rill Cloud

When deploying to Rill Cloud, you'll need to configure credentials that Rill Cloud can use to connect to your Snowflake instance.

### Setting Credentials for Deployment

Run the following command to configure your Snowflake credentials for Rill Cloud:

```bash
rill env configure
```

You'll be prompted to enter your Snowflake DSN. For production deployments, we recommend using key-pair authentication in your DSN string.

### Deploy Your Project

After configuring credentials, deploy your project:

```bash
rill deploy
```

Follow the prompts to complete the deployment to Rill Cloud.

:::tip Best Practices

1. **Use Service Accounts**: Create dedicated Snowflake users for Rill with minimal required permissions
2. **Key-Pair Authentication**: Use key-pair authentication instead of passwords for better security
3. **Connection Pooling**: The connector handles connection pooling automatically
4. **Query Optimization**: Use appropriate filters and column selection to minimize data transfer

:::

## Additional Resources

- [Snowflake Connector Reference](../../reference/connectors/snowflake.md) - Complete connector configuration options
- [Snowflake Documentation](https://docs.snowflake.com/) - Official Snowflake documentation
- [Deploy with Credentials](../../deploy/credentials.md) - More about credential management in Rill Cloud
