---
title: BigQuery
description: Power Rill dashboards using BigQuery as an OLAP engine
sidebar_label: BigQuery
sidebar_position: 20
---

[Google BigQuery](https://cloud.google.com/bigquery/docs) is a fully managed, serverless data warehouse that enables scalable and cost-effective analysis of large datasets using SQL-like queries. It supports a highly scalable and flexible architecture, allowing users to analyze large amounts of data in real time, making it suitable for BI/ML applications.

Rill now supports using BigQuery as an OLAP engine via a "live connector". This allows you to connect to existing BigQuery tables and use them to power Rill dashboards without ingesting data into a separate database. This is particularly useful when:
- Your data already exists in BigQuery and is optimized for analytical queries
- You want to leverage BigQuery's scalability and performance for large datasets
- You prefer to keep your data in your existing BigQuery warehouse

## BigQuery as Data Source vs OLAP Engine

BigQuery can be used in two ways with Rill:

1. **As a data source**: Ingest data from BigQuery tables into Rill's default OLAP engine (DuckDB)
2. **As an OLAP engine**: Connect directly to BigQuery and query tables without data ingestion

This documentation covers using BigQuery as an OLAP engine. For information about using BigQuery as a data source, see our [BigQuery data source documentation](/connect/data-source/bigquery).

## Connect to BigQuery as OLAP Engine

To use BigQuery as your OLAP engine, you'll need to:

1. Create a BigQuery connector with appropriate credentials
2. Set BigQuery as your default OLAP connector (or specify it for individual models/metrics views)

### Step 1: Create BigQuery Connector

Create a connector configuration file in your `connectors` directory. We recommend using Service Account JSON for authentication:

```yaml
type: connector
driver: bigquery

google_application_credentials: "{{ .env.connector.bigquery.google_application_credentials }}"
project_id: "your-project-id"
```

For detailed authentication options and credential setup, see the [BigQuery data source documentation](/connect/data-source/bigquery#connect-to-bigquery).

### Step 2: Set as Default OLAP Engine

To use BigQuery as your default OLAP engine, update your `rill.yaml`:

```yaml
olap_connector: bigquery
```

Alternatively, you can specify BigQuery as the OLAP engine for specific resources using the `connector` or `output.connector` property.

## Using BigQuery with Partitioned Models

BigQuery's OLAP support enables powerful use cases with partitioned models. You can use BigQuery as both the source and output connector for processing partitioned data:

```yaml
type: model

partitions:
  connector: bigquery
  sql: |
    SELECT DISTINCT _PARTITIONTIME AS partition_time
    FROM `project.dataset.table`
    WHERE TIMESTAMP_TRUNC(_PARTITIONTIME, MONTH) = TIMESTAMP("2025-08-01")

connector: bigquery
sql: |
  SELECT * FROM `project.dataset.table`
  WHERE _PARTITIONTIME = '{{ .partition.partition_time }}'
  LIMIT 1000

output:
  connector: duckdb  # Or another OLAP engine for final storage
```

This pattern is particularly useful for:
- Processing large partitioned tables incrementally
- Leveraging BigQuery's partition pruning for efficient queries
- Transforming data before loading into another OLAP engine

## External Tables

When using BigQuery as an OLAP engine, you can create metrics views directly against existing BigQuery tables without creating models. This allows you to:

- Build dashboards on top of tables managed outside of Rill
- Avoid data duplication when tables are already optimized
- Use BigQuery's compute for all analytical queries

For more information on external OLAP tables, see our [OLAP engines documentation](/connect/olap#external-olap-tables).

## Performance Considerations

When using BigQuery as an OLAP engine:

- **Query costs**: BigQuery charges based on data scanned. Use partitioning and clustering to minimize costs
- **Partitioned tables**: Leverage BigQuery's native partitioning for better performance and lower costs
- **Materialized views**: Consider using BigQuery's materialized views for frequently accessed aggregations
- **Billing project**: Ensure your service account has access to a billing project for query execution

## Deploy to Rill Cloud

When deploying to Rill Cloud with BigQuery as an OLAP engine, ensure you provide valid service account credentials with:

- **BigQuery Data Viewer** role (or equivalent) for reading data
- **BigQuery Job User** role for executing queries
- Access to the appropriate BigQuery datasets

For detailed deployment instructions, see our [BigQuery deployment documentation](/connect/data-source/bigquery#deploy-to-rill-cloud).

:::tip Need help with BigQuery?

If you need assistance connecting Rill to BigQuery as an OLAP engine, please don't hesitate to [contact us](/contact). We'd love to help!

:::

## Additional Resources

- [BigQuery as a data source](/connect/data-source/bigquery)
- [Partitioned models](/build/models/partitioned-models)
- [External OLAP tables](/connect/olap#external-olap-tables)
- [Multiple OLAP engines](/connect/olap/multiple-olap)
