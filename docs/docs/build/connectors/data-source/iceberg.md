---
title: Apache Iceberg
description: Connect to Apache Iceberg tables
sidebar_label: Iceberg
sidebar_position: 26
---

<!-- WARNING: There are links to this page in source code. If you move it, find and replace the links and consider adding a redirect in docusaurus.config.js. -->

## Overview

[Apache Iceberg](https://iceberg.apache.org/) is an open table format for large analytic tables that brings reliability and simplicity to data lakes. Iceberg provides features like schema evolution, hidden partitioning, and time travel queries, making it easier to manage large datasets in object storage. It is designed to solve common data lake problems including schema evolution, partition management, and ACID transactions on large tables.

Iceberg works with Parquet, Avro, and ORC file formats and is compatible with multiple query engines including Spark, Trino, Flink, and others, making it ideal for large-scale data lakehouse architectures.

## Connect to Iceberg

To connect to Apache Iceberg tables, you'll need to configure:

- **Catalog Type**: Hive Metastore, AWS Glue, JDBC, or REST catalog
- **Storage Backend**: S3, GCS, Azure Blob Storage, or HDFS
- **Authentication**: Credentials for catalog and storage access
- **Table Location**: Path to Iceberg table metadata and data files


### Authentication

Authentication follows the same patterns as other object storage connectors. For S3-based storage, use Access Key/Secret Key or IAM Role Assumption. For GCS, use Service Account JSON credentials. See the [S3](/build/connectors/data-source/s3) and [GCS](/build/connectors/data-source/gcs) connector documentation for detailed authentication setup.

## Use Cases

Iceberg is ideal for:

- **Large Data Lakes**: Manage petabytes of data across multiple partitions
- **Frequent Updates**: Tables that require updates, deletes, and merges
- **Schema Evolution**: Tables that need to evolve their structure over time
- **Time Travel Queries**: Analyzing historical data states
- **Multi-Engine Support**: Tables accessed by multiple query engines

## Benefits

- **Reliability**: ACID transactions ensure data consistency
- **Performance**: Efficient metadata management and query planning
- **Flexibility**: Schema evolution without breaking queries
- **Compatibility**: Works with multiple query engines (Spark, Trino, Flink, etc.)

## Iceberg Table Features

- **Schema Evolution**: Safely evolve table schemas over time without breaking existing queries
- **Hidden Partitioning**: Automatic partition management without exposing partition details to queries
- **Time Travel**: Query data as it existed at a specific point in time using snapshots
- **Partition Evolution**: Change partitioning without rewriting data
- **Snapshot Management**: Track table changes over time with efficient metadata
- **Metadata Optimization**: Compact metadata files for faster queries
- **Concurrent Writes**: Safe concurrent writes from multiple processes

:::info Interested in Iceberg Support?

Apache Iceberg connector support is currently in development. If you're interested in using Iceberg tables with Rill, please [contact our team](/contact) to discuss your use case and requirements.

Our team can help you:
- Understand how Iceberg would integrate with Rill
- Plan your data lakehouse architecture with Iceberg
- Explore alternative solutions for your use case

:::

