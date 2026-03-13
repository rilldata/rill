---
title: Apache Iceberg
description: Read Iceberg tables from object storage via ClickHouse
sidebar_label: Apache Iceberg
sidebar_position: 20
---

import WrongOLAP from '@site/src/components/WrongOLAP';
import ClickHousePrereq from '@site/src/components/ClickHousePrereq';

<WrongOLAP engine="clickhouse" />

## Overview

[Apache Iceberg](https://iceberg.apache.org/) is an open table format for large analytical datasets. ClickHouse can read Iceberg tables directly from object storage using its [`icebergS3()` table function](https://clickhouse.com/docs/en/sql-reference/table-functions/iceberg), with support for schema evolution and partition pruning.

<ClickHousePrereq />

## Credentials

Add your storage credentials to your project's `.env` file.

**For S3:**

```bash
AWS_ACCESS_KEY_ID=your_access_key_id
AWS_SECRET_ACCESS_KEY=your_secret_access_key
```

**For Azure:**

```bash
AZURE_STORAGE_CONNECTION_STRING=DefaultEndpointsProtocol=https;AccountName=myaccount;AccountKey=mykey;EndpointSuffix=core.windows.net
```

For details on managing credentials, see [Configure Local Credentials](/developers/build/connectors/credentials).

## Model Configuration

### Reading from S3

```yaml
type: model
connector: my_clickhouse

sql: |
  SELECT *
  FROM icebergS3(
    'https://my-bucket.s3.amazonaws.com/path/to/iceberg/table',
    '{{ .env.AWS_ACCESS_KEY_ID }}',
    '{{ .env.AWS_SECRET_ACCESS_KEY }}'
  )
```

### Reading from Azure

```yaml
type: model
connector: my_clickhouse

sql: |
  SELECT *
  FROM icebergAzure(
    '{{ .env.AZURE_STORAGE_CONNECTION_STRING }}',
    'my-container',
    'path/to/iceberg/table'
  )
```

:::info
The Iceberg table functions point to the table's root directory (containing the `metadata/` folder). ClickHouse reads the latest snapshot by default.
:::

## Reference

For general Iceberg configuration and storage backend setup, see the [DuckDB Iceberg connector guide](/developers/build/connectors/data-source/duckdb/iceberg). See also the [ClickHouse Iceberg documentation](https://clickhouse.com/docs/en/sql-reference/table-functions/iceberg) for full syntax details.
