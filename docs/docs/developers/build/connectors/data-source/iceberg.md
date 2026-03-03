---
title: Apache Iceberg
description: Read Iceberg tables from object storage
sidebar_label: Apache Iceberg
sidebar_position: 27
---

## Overview

[Apache Iceberg](https://iceberg.apache.org/) is an open table format for large analytic datasets. Rill supports reading Iceberg tables directly from object storage through compatible query engine integrations. Today, this is powered by DuckDB's native [Iceberg extension](https://duckdb.org/docs/extensions/iceberg/overview.html).

:::note Direct file access only
Rill reads Iceberg tables by scanning the table's metadata and data files directly from object storage. Catalog-based access (e.g., through a Hive Metastore, AWS Glue, or REST catalog) is not currently supported.
:::

## Storage Backends

Iceberg tables can be read from any of the following storage backends:

| Backend | URI format | Authentication |
|---|---|---|
| Amazon S3 | `s3://bucket/path/to/table` | Requires an [S3 connector](/developers/build/connectors/data-source/s3) |
| Google Cloud Storage | `gs://bucket/path/to/table` | Requires a [GCS connector](/developers/build/connectors/data-source/gcs) |
| Azure Blob Storage | `azure://container/path/to/table` | Requires an [Azure connector](/developers/build/connectors/data-source/azure) |
| Local filesystem | `/path/to/table` | No authentication needed |

For cloud storage backends, you must first configure the corresponding storage connector with valid credentials. Rill uses these credentials to authenticate when reading the Iceberg table files.

## Using the UI

1. Click **Add Data** in your Rill project
2. Select **Apache Iceberg** as the data source type
3. Choose your storage backend (S3, GCS, Azure, or Local)
4. Enter the path to your Iceberg table directory
5. Optionally configure advanced parameters (allow moved paths, snapshot version)
6. Enter a model name and click **Create**

For cloud storage backends, the UI will prompt you to set up the corresponding storage connector if one doesn't already exist.

## Manual Configuration

Create a model that uses DuckDB's `iceberg_scan()` function to read the table.

### Reading from S3

Create `models/iceberg_data.yaml`:

```yaml
type: model
connector: duckdb
create_secrets_from_connectors: s3
materialize: true

sql: |
  SELECT *
  FROM iceberg_scan('s3://my-bucket/path/to/iceberg_table')
```

### Reading from GCS

```yaml
type: model
connector: duckdb
create_secrets_from_connectors: gcs
materialize: true

sql: |
  SELECT *
  FROM iceberg_scan('gs://my-bucket/path/to/iceberg_table')
```

### Reading from Azure

```yaml
type: model
connector: duckdb
create_secrets_from_connectors: azure
materialize: true

sql: |
  SELECT *
  FROM iceberg_scan('azure://my-container/path/to/iceberg_table')
```

### Reading from local filesystem

```yaml
type: model
connector: duckdb
materialize: true

sql: |
  SELECT *
  FROM iceberg_scan('/path/to/iceberg_table')
```

## Optional Parameters

The `iceberg_scan()` function accepts additional parameters:

| Parameter | Type | Description |
|---|---|---|
| `allow_moved_paths` | boolean | Allow reading tables where data files have been moved from their original location. Defaults to `true` in the UI. |
| `version` | string | Read a specific Iceberg snapshot version instead of the latest. |

Example with optional parameters:

```sql
SELECT *
FROM iceberg_scan('s3://my-bucket/path/to/iceberg_table',
  allow_moved_paths = true,
  version = '2')
```

## Deploy to Rill Cloud

Since Iceberg tables are read through DuckDB using your existing storage connector credentials, deploying to Rill Cloud follows the same process as the underlying storage connector:

- **S3**: Follow the [S3 deployment guide](/developers/build/connectors/data-source/s3#deploy-to-rill-cloud)
- **GCS**: Follow the [GCS deployment guide](/developers/build/connectors/data-source/gcs#deploy-to-rill-cloud)
- **Azure**: Follow the [Azure deployment guide](/developers/build/connectors/data-source/azure#deploy-to-rill-cloud)

Ensure your storage connector credentials are configured in your Rill Cloud project before deploying.

## Limitations

- **Direct file access only**: Rill reads Iceberg metadata and data files directly from storage. Catalog integrations (Hive Metastore, AWS Glue, REST catalog) are not supported.
- **DuckDB engine**: Iceberg support is currently provided through DuckDB's Iceberg extension. Additional engine support (e.g., ClickHouse) is planned.
- **Read-only**: Rill reads from Iceberg tables but does not write to them.
