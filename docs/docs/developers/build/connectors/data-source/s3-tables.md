---
title: Amazon S3 Tables
description: Read Iceberg tables from Amazon S3 table buckets
sidebar_label: Amazon S3 Tables
sidebar_position: 28
---

## Overview

[Amazon S3 Tables](https://docs.aws.amazon.com/AmazonS3/latest/userguide/s3-tables.html) store tabular data in [Apache Iceberg](/developers/build/connectors/data-source/iceberg) format inside S3 table buckets, with a built-in Iceberg REST catalog. Rill reads S3 Tables through DuckDB's [Iceberg extension](https://duckdb.org/docs/extensions/iceberg/amazon_s3_tables.html) by attaching the table bucket as a catalog.

:::note Use the catalog, not `iceberg_scan()`
S3 table buckets do not write a `version-hint` file and do not allow listing objects, so `iceberg_scan()` only works when given the exact path of the current `metadata.json` file. That path changes on every write, and S3 Tables maintenance eventually deletes old metadata. Attaching the table bucket always reads the latest snapshot.
:::

## Prerequisites

- The ARN of your table bucket, e.g. `arn:aws:s3tables:us-east-1:123456789012:bucket/my-table-bucket`.
- AWS credentials with read access to the table bucket:

```json
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Action": [
      "s3tables:GetTableBucket",
      "s3tables:ListNamespaces",
      "s3tables:GetNamespace",
      "s3tables:ListTables",
      "s3tables:GetTable",
      "s3tables:GetTableMetadataLocation",
      "s3tables:GetTableData"
    ],
    "Resource": [
      "arn:aws:s3tables:us-east-1:123456789012:bucket/my-table-bucket",
      "arn:aws:s3tables:us-east-1:123456789012:bucket/my-table-bucket/table/*"
    ]
  }]
}
```

## Manual Configuration

Create an [S3 connector](/developers/build/connectors/data-source/s3) in `connectors/s3.yaml`, setting `region` to the table bucket's region:

```yaml
type: connector
driver: s3
aws_access_key_id: '{{ .env.aws_access_key_id }}'
aws_secret_access_key: '{{ .env.aws_secret_access_key }}'
region: us-east-1
```

Then create a model that attaches the table bucket in `pre_exec` and selects from a table using `<catalog>.<namespace>.<table>`. `create_secrets_from_connectors` creates a DuckDB secret from the S3 connector before `pre_exec` runs.

Create `models/s3tables_orders.yaml`:

```yaml
type: model
connector: duckdb
materialize: true
create_secrets_from_connectors: s3

pre_exec: |
  INSTALL aws; LOAD aws; INSTALL iceberg; LOAD iceberg;
  ATTACH OR REPLACE 'arn:aws:s3tables:us-east-1:123456789012:bucket/my-table-bucket' AS s3t (
    TYPE iceberg,
    ENDPOINT_TYPE s3_tables
  )

sql: SELECT * FROM s3t.my_namespace.orders
```

Finally, add the credentials to your project's `.env` file:

```bash
aws_access_key_id=...
aws_secret_access_key=...
```

## Deploy to Rill Cloud

Push the credentials in your `.env` file to your Rill Cloud project:

```bash
rill env push
```

## Limitations

- **DuckDB engine**: S3 Tables support is provided through DuckDB's Iceberg extension.
- **Read-only**: Rill reads from S3 Tables but does not write to them.
