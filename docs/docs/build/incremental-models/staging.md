---
title: Staging Models
description: C
sidebar_label: Staging Models
sidebar_position: 10
---
Staging models are required for situations where the input and output type are not supported, such as Snowflake to ClickHouse. 

:::note Supported 
Staging Models is in ongoing development, while we do have support for the following, please reach out to us if you have any specific requirements.

Snowflake --> S3 --> ClickHouse

:::
```mermaid
  sequenceDiagram
    Snowflake->>Staging (S3): write data to temporary Staging
    Staging (S3)->>ClickHouse: write temporary data to ClickHouse
    ClickHouse ->> Staging (S3): remove temporary data from Staging
```

In the above example, during the ingestion from Snowflake to Clickhouse, we use the temporary staging table in S3 to write from Snowflake to S3, then from S3 to ClickHouse. Once this procedure is complete, we clear the temporary data from S3.

### Example:

```yaml
# Use DuckDB to generate a range of days from 1st Jan to today
partitions:
  connector: duckdb
  sql: SELECT range as day FROM range(TIMESTAMPTZ '2024-01-01', now(), INTERVAL 1 DAY)

# Don't reload previously ingested partitions on every refresh
incremental: true

# Query Snowflake for all events belonging to the current partition
connector: snowflake
sql: SELECT * FROM events WHERE date_trunc('day', event_time) = '{{ .partition.day }}'

# Since ClickHouse can't ingest from Snowflake or vice versa, we use S3 as a temporary staging connector
stage:
  connector: s3
  path: s3://bucket/temp-data

# Produce the final output into ClickHouse
output:
  connector: clickhouse
```