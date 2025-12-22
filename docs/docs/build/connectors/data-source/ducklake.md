---
title: DuckLake
description: Connect to DuckLake data lakehouse
sidebar_label: DuckLake
sidebar_position: 12
---

<!-- WARNING: There are links to this page in source code. If you move it, find and replace the links and consider adding a redirect in docusaurus.config.js. -->

## Overview

[DuckLake](https://duckdb.org/docs/guides/data_lakes/overview) is a data lakehouse solution built on DuckDB that provides a unified interface for querying data across multiple storage formats and locations. DuckLake enables you to query data directly from object storage (S3, GCS, Azure Blob Storage) without needing to load it into a separate database. It combines the performance of DuckDB's analytical engine with the flexibility of data lake storage, allowing you to query data directly from object storage without data movement.

DuckLake supports multiple file formats including Parquet, CSV, JSON, Iceberg, and Delta Lake, making it ideal for data lake analytics where you need to query large datasets stored in object storage without ETL processes.

## Connect to DuckLake

To connect to DuckLake, you'll need to configure:

- **Storage Backend**: S3, GCS, Azure Blob Storage, or local filesystem
- **Authentication**: Credentials for accessing your object storage
- **Catalog Configuration**: Metadata catalog for table discovery
- **Format Support**: Parquet, CSV, JSON, Iceberg, Delta Lake formats

### Storage Backend Configuration

DuckLake supports multiple storage backends. Depending which storage backend you decide to use, refer to the explicit connector page.

- [s3](/build/connectors/data-source/s3)
- [gcs](/build/connectors/data-source/gcs)


### Authentication

Authentication follows the same patterns as other object storage connectors. For S3, you can use Access Key/Secret Key or IAM Role Assumption. For GCS, use Service Account JSON credentials. See the [S3](/build/connectors/data-source/s3) and [GCS](/build/connectors/data-source/gcs) connector documentation for detailed authentication setup.

## Use Cases

DuckLake is ideal for:

- **Data Lake Analytics**: Query large datasets stored in object storage without ETL
- **Multi-Format Support**: Work with data in various formats (Parquet, CSV, JSON, Iceberg, Delta Lake)
- **Cost-Effective Queries**: Analyze data without moving it to a separate warehouse
- **Unified Data Access**: Query data across multiple storage locations and formats

## Benefits

- **No Data Movement**: Query data directly from object storage without copying or loading
- **High Performance**: Leverage DuckDB's optimized columnar query engine
- **Format Flexibility**: Support for multiple data formats in a single query
- **Cost Effective**: Avoid data duplication and warehouse storage costs

:::info Interested in DuckLake Support?

DuckLake connector support is currently in development. If you're interested in using DuckLake with Rill, please [contact our team](/contact) to discuss your use case and requirements.

Our team can help you:
- Understand how DuckLake would integrate with Rill
- Plan your data lakehouse architecture
- Explore alternative solutions for your use case

:::

