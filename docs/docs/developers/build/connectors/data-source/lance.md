---
title: Lance
description: Read Lance datasets from local or cloud storage
sidebar_label: Lance
sidebar_position: 28
---

## Overview

[Lance](https://lancedb.github.io/lance/) is a modern columnar data format optimized for ML and AI workloads, with native support for cloud storage. Rill supports reading Lance datasets directly from local or cloud storage through DuckDB's [Lance extension](https://duckdb.org/docs/stable/core_extensions/lance).

## Storage Backends

Lance datasets can be read from any of the following storage backends:

| Backend | URI format | Authentication |
|---|---|---|
| Amazon S3 | `s3://bucket/path/to/dataset.lance` | Requires an [S3 connector](/developers/build/connectors/data-source/s3) |
| Google Cloud Storage | `gs://bucket/path/to/dataset.lance` | Requires a [GCS connector](/developers/build/connectors/data-source/gcs) |
| Azure Blob Storage | `az://container/path/to/dataset.lance` | Requires an [Azure connector](/developers/build/connectors/data-source/azure) |
| Local filesystem | `/path/to/dataset.lance` | No authentication needed |

For cloud storage backends, you must first configure the corresponding storage connector with valid credentials. Rill uses these credentials to authenticate when reading the Lance dataset files.

## Using the UI

1. Click **Add Data** in your Rill project
2. Select **Lance** as the data source type
3. Choose your storage backend (S3, GCS, Azure, or Local)
4. Enter the path to your Lance dataset
5. Enter a model name and click **Create**

For cloud storage backends, the UI will prompt you to set up the corresponding storage connector if one doesn't already exist.

## Manual Configuration

Create a model that reads from a Lance dataset path. DuckDB's Lance extension recognizes `.lance` paths automatically.

### Reading from S3

Create `models/lance_data.yaml`:

```yaml
type: model
connector: duckdb
create_secrets_from_connectors: s3
materialize: true

sql: |
  SELECT *
  FROM 's3://my-bucket/path/to/dataset.lance'
```

### Reading from GCS

```yaml
type: model
connector: duckdb
create_secrets_from_connectors: gcs
materialize: true

sql: |
  SELECT *
  FROM 'gs://my-bucket/path/to/dataset.lance'
```

### Reading from Azure

```yaml
type: model
connector: duckdb
create_secrets_from_connectors: azure
materialize: true

sql: |
  SELECT *
  FROM 'az://my-container/path/to/dataset.lance'
```

### Reading from local filesystem

```yaml
type: model
connector: duckdb
materialize: true

sql: |
  SELECT *
  FROM '/path/to/dataset.lance'
```

## Deploy to Rill Cloud

Since Lance datasets are read through DuckDB using your existing storage connector credentials, deploying to Rill Cloud follows the same process as the underlying storage connector:

- **S3**: Follow the [S3 deployment guide](/developers/build/connectors/data-source/s3#deploy-to-rill-cloud)
- **GCS**: Follow the [GCS deployment guide](/developers/build/connectors/data-source/gcs#deploy-to-rill-cloud)
- **Azure**: Follow the [Azure deployment guide](/developers/build/connectors/data-source/azure#deploy-to-rill-cloud)

Ensure your storage connector credentials are configured in your Rill Cloud project before deploying.

## Limitations

- **Direct file access only**: Rill reads Lance datasets directly from storage. There is no catalog integration.
- **DuckDB engine**: Lance support is provided through DuckDB's Lance extension.
- **Read-only**: Rill reads from Lance datasets but does not write to them.
