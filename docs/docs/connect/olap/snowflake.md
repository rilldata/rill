---
title: Snowflake
description: Power Rill dashboards using Snowflake as an OLAP engine
sidebar_label: Snowflake
sidebar_position: 30
---

[Snowflake](https://docs.snowflake.com/en/user-guide-intro) is a cloud-based data platform designed to facilitate data warehousing, data lakes, data engineering, data science, and data application development. It separates compute and storage, enabling users to scale up or down instantly without downtime, providing a cost-effective solution for data management. With its unique architecture and support for multi-cloud environments (AWS, Azure, and Google Cloud), Snowflake offers seamless data integration and real-time access to data insights.

Rill now supports using Snowflake as an OLAP engine via a "live connector". This allows you to connect to existing Snowflake tables and use them to power Rill dashboards without ingesting data into a separate database. This is particularly useful when:
- Your data already exists in Snowflake and is optimized for analytical queries
- You want to leverage Snowflake's elastic compute and storage separation
- You prefer to keep your data in your existing Snowflake warehouse

:::info Private Key Authentication Recommended

Snowflake has issued a [deprecation notice](https://www.snowflake.com/en/blog/blocking-single-factor-password-authentification/) for single-factor password authentication. Rill supports and recommends using private key authentication to avoid any disruption of your service.

:::

## Snowflake as Data Source vs OLAP Engine

Snowflake can be used in two ways with Rill:

1. **As a data source**: Ingest data from Snowflake tables into Rill's default OLAP engine (DuckDB)
2. **As an OLAP engine**: Connect directly to Snowflake and query tables without data ingestion

This documentation covers using Snowflake as an OLAP engine. For information about using Snowflake as a data source, see our [Snowflake data source documentation](/connect/data-source/snowflake).

## Connect to Snowflake as OLAP Engine

To use Snowflake as your OLAP engine, you'll need to:

1. Create a Snowflake connector with appropriate credentials
2. Set Snowflake as your default OLAP connector (or specify it for individual models/metrics views)

### Step 1: Create Snowflake Connector

Create a connector configuration file in your `connectors` directory:

```yaml
type: connector
driver: snowflake

dsn: "{{ .env.connector.snowflake.dsn }}"
```

#### Using Private Key Authentication

We recommend using private key authentication. Your connection string should follow this format:

```
<username>@<account_identifier>/<database>/<schema>?warehouse=<warehouse>&role=<role>&authenticator=SNOWFLAKE_JWT&privateKey=<privateKey_url_safe>
```

For detailed instructions on generating and configuring private keys, see the [Snowflake data source documentation](/connect/data-source/snowflake#using-keypair-authentication).

:::info Finding Your Account Identifier

Your [Snowflake account identifier](https://docs.snowflake.com/en/user-guide/admin-account-identifier) is everything before `.snowflakecomputing.com` in your Snowflake account URL.

:::

### Step 2: Set as Default OLAP Engine

To use Snowflake as your default OLAP engine, update your `rill.yaml`:

```yaml
olap_connector: snowflake
```

Alternatively, you can specify Snowflake as the OLAP engine for specific resources using the `connector` or `output.connector` property.

## Using Snowflake with Partitioned Models

Snowflake's OLAP support enables powerful use cases with partitioned models. You can use Snowflake as both the source and output connector for processing partitioned data:

```yaml
type: model

partitions:
  connector: snowflake
  sql: |
    SELECT DISTINCT DATE_TRUNC('MONTH', event_time) AS partition_month
    FROM database.schema.events_table
    WHERE event_time >= '2025-01-01'

connector: snowflake
sql: |
  SELECT *
  FROM database.schema.events_table
  WHERE DATE_TRUNC('MONTH', event_time) = '{{ .partition.partition_month }}'
  LIMIT 1000

output:
  connector: duckdb  # Or another OLAP engine for final storage
```

This pattern is particularly useful for:
- Processing large tables incrementally by time periods
- Leveraging Snowflake's clustering keys for efficient queries
- Transforming data before loading into another OLAP engine

## External Tables

When using Snowflake as an OLAP engine, you can create metrics views directly against existing Snowflake tables without creating models. This allows you to:

- Build dashboards on top of tables managed outside of Rill
- Avoid data duplication when tables are already optimized
- Use Snowflake's compute for all analytical queries

For more information on external OLAP tables, see our [OLAP engines documentation](/connect/olap#external-olap-tables).

## Performance Considerations

When using Snowflake as an OLAP engine:

- **Warehouse sizing**: Choose appropriate virtual warehouse sizes for your query workload
- **Auto-suspend**: Configure auto-suspend to minimize costs during idle periods
- **Clustering keys**: Use Snowflake's clustering keys on large tables for better query performance
- **Result caching**: Snowflake automatically caches query results for 24 hours
- **Time travel**: Be aware that queries access current data unless you specify a historical timestamp

## Cost Optimization

Snowflake charges for:
- **Compute**: Virtual warehouse usage (billed per-second with a 60-second minimum)
- **Storage**: Data stored in Snowflake tables
- **Data transfer**: Moving data between regions

To optimize costs when using Snowflake as an OLAP engine:
- Use appropriately sized warehouses
- Enable auto-suspend on warehouses
- Consider using multi-cluster warehouses for variable workloads
- Leverage Snowflake's query result caching

## Deploy to Rill Cloud

When deploying to Rill Cloud with Snowflake as an OLAP engine, ensure your connection string includes:

- Valid authentication credentials (private key recommended)
- Appropriate warehouse and role permissions
- Network access if using private connectivity

For detailed deployment instructions, see our [Snowflake deployment documentation](/connect/data-source/snowflake#deploy-to-rill-cloud).

:::tip Need help with Snowflake?

If you need assistance connecting Rill to Snowflake as an OLAP engine, please don't hesitate to [contact us](/contact). We'd love to help!

:::

## Additional Resources

- [Snowflake as a data source](/connect/data-source/snowflake)
- [Partitioned models](/build/models/partitioned-models)
- [External OLAP tables](/connect/olap#external-olap-tables)
- [Multiple OLAP engines](/connect/olap/multiple-olap)
- [Snowflake key-pair authentication](https://docs.snowflake.com/en/user-guide/key-pair-auth)
