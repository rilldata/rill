---
title: Configure connector credentials
sidebar_label: Configure Credentials
sidebar_position: 00
---

Rill requires credentials to connect to remote data sources such as private buckets (S3, GCS, Azure), data warehouses (Snowflake, BigQuery), OLAP engines (ClickHouse, Apache Druid) or other DuckDB sources (MotherDuck). Please refer to the appropriate [connector](../../reference/connectors/connectors.md) and [OLAP engine](../../reference/olap-engines/olap-engines.md) page for instructions to configure credentials accordingly.

At a high level, configuring credentials and credentials management can be broken down into the three categories:
- Setting credentials for Rill Developer
- Setting credentials for a Rill Cloud project
- Pushing and pulling credentials to / from Rill Cloud

## Setting credentials for Rill Developer

When using a source (or different OLAP engine), 


## Setting credentials for a Rill Cloud project

## Pushing and pulling credentials to / from Rill Cloud



When running Rill locally, Rill attempts to find existing credentials configured on your computer. When deploying projects to Rill Cloud, you must explicitly provide service account credentials with correct access permissions.

For instructions on how to create a service account and set credentials in Rill Cloud, see our reference docs for:

- [Amazon S3](../../reference/connectors/s3.md) 
- [Google Cloud Storage (GCS)](../../reference/connectors/gcs.md)
- [Azure Blob Storage (Azure)](../../reference/connectors/azure.md)
- [Amazon Athena](../../reference/connectors/athena.md)
- [BigQuery](../../reference/connectors/bigquery.md)
- [ClickHouse](../../reference/olap-engines/clickhouse.md)
- [MotherDuck](../../reference/connectors/motherduck.md)
- [Postgres](../../reference/connectors/postgres.md)
- [Salesforce](../../reference/connectors/salesforce.md)
- [Snowflake](../../reference/connectors/snowflake.md)
- [Google Sheets](../../reference/connectors/googlesheets.md)



