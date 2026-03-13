---
title: Apache Iceberg
description: Read Iceberg tables from object storage via ClickHouse
sidebar_label: Apache Iceberg
sidebar_position: 27
---

import WrongOLAP from '@site/src/components/WrongOLAP';

<WrongOLAP engine="clickhouse" />

## Overview

Connect your ClickHouse OLAP engine to Apache Iceberg tables using ClickHouse's [`icebergS3()` table function](https://clickhouse.com/docs/en/sql-reference/table-functions/iceberg).

## Model Configuration

### Reading from S3

Create `models/iceberg_data.yaml`:

```yaml
type: model
connector: my_clickhouse

sql: |
  SELECT *
  FROM icebergS3(
    'https://my-bucket.s3.amazonaws.com/path/to/iceberg/table',
    '{{ .env.connector.s3.aws_access_key_id }}',
    '{{ .env.connector.s3.aws_secret_access_key }}'
  )
```

### Reading from Azure

```yaml
type: model
connector: my_clickhouse

sql: |
  SELECT *
  FROM icebergAzure(
    '{{ .env.connector.azure.azure_storage_connection_string }}',
    'my-container',
    'path/to/iceberg/table'
  )
```

## Reference

For general Iceberg configuration and storage backend setup, see the [DuckDB Iceberg connector guide](/developers/build/connectors/data-source/duckdb/iceberg). See also the [ClickHouse Iceberg documentation](https://clickhouse.com/docs/en/sql-reference/table-functions/iceberg) for full syntax details.
