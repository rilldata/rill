---
title: Delta Lake
description: Read Delta Lake tables from object storage
sidebar_label: Delta Lake
sidebar_position: 12
---

## Overview

[Delta Lake](https://delta.io/) is an open-source storage framework that brings ACID transactions to data lakes. Rill supports reading Delta tables directly from object storage through compatible query engine integrations. Today, this is powered by DuckDB's [Delta extension](https://duckdb.org/docs/stable/core_extensions/delta).

:::note Direct file access only
Rill reads Delta tables by scanning the table's transaction log and data files directly from object storage. Catalog-based access (e.g., through Unity Catalog) is not currently supported.
:::

## Storage Backends

Delta tables can be read from any of the following storage backends:

| Backend | URI format | Authentication |
|---|---|---|
| Amazon S3 | `s3://bucket/path/to/table` | Requires an [S3 connector](/developers/build/connectors/data-source/s3) |
| Azure Blob Storage | `azure://container/path/to/table` | Requires an [Azure connector](/developers/build/connectors/data-source/azure) |
| Local filesystem | `/path/to/table` | No authentication needed |

:::info GCS not yet supported
Google Cloud Storage is not currently supported for Delta tables. GCS support depends on the upstream DuckDB Delta extension adding it.
:::

For cloud storage backends, you must first configure the corresponding storage connector with valid credentials. Rill uses these credentials to authenticate when reading the Delta table files.

## Using the UI

1. Click **Add Data** in your Rill project
2. Select **Delta Lake** as the data source type
3. Choose your storage backend (S3, Azure, or Local)
4. Enter the path to your Delta table directory
5. Enter a model name and click **Create**

For cloud storage backends, the UI will prompt you to set up the corresponding storage connector if one doesn't already exist.

## Manual Configuration

Create a model that uses DuckDB's `delta_scan()` function to read the table.

### Reading from S3

Create `models/delta_data.yaml`:

```yaml
type: model
connector: duckdb
create_secrets_from_connectors: s3
materialize: true

sql: |
  SELECT *
  FROM delta_scan('s3://my-bucket/path/to/delta_table')
```

### Reading from Azure

```yaml
type: model
connector: duckdb
create_secrets_from_connectors: azure
materialize: true

sql: |
  SELECT *
  FROM delta_scan('azure://my-container/path/to/delta_table')
```

### Reading from local filesystem

```yaml
type: model
connector: duckdb
materialize: true

sql: |
  SELECT *
  FROM delta_scan('/path/to/delta_table')
```

## Deploy to Rill Cloud

Since Delta tables are read through DuckDB using your existing storage connector credentials, deploying to Rill Cloud follows the same process as the underlying storage connector:

- **S3**: Follow the [S3 deployment guide](/developers/build/connectors/data-source/s3#deploy-to-rill-cloud)
- **Azure**: Follow the [Azure deployment guide](/developers/build/connectors/data-source/azure#deploy-to-rill-cloud)

Ensure your storage connector credentials are configured in your Rill Cloud project before deploying.

## Limitations

- **Direct file access only**: Rill reads Delta transaction logs and data files directly from storage. Catalog integrations (e.g., Unity Catalog) are not supported.
- **DuckDB engine**: Delta support is currently provided through DuckDB's Delta extension. Additional engine support is planned.
- **No GCS support**: Google Cloud Storage is not yet supported by the Delta extension.
- **Read-only**: Rill reads from Delta tables but does not write to them.
- **Experimental**: The DuckDB Delta extension is currently marked as experimental.
