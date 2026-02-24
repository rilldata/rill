---
title: SQL APIs
description: Create custom APIs with SQL queries against models, tables, and external databases
sidebar_label: SQL APIs
sidebar_position: 20
---

SQL APIs let you write SQL queries and expose them as HTTP endpoints. By default, queries execute against your project's default [OLAP connector](/developers/build/connectors/olap) (e.g., DuckDB, ClickHouse). You can also query external databases like BigQuery, Snowflake, and Postgres by specifying a different `connector`.

## Basic syntax

Create a YAML file in your project's `apis/` directory:

```yaml
type: api
sql: SELECT publisher, domain, bid_price FROM ad_bids LIMIT 100
```

This queries your default OLAP connector and returns the results as JSON.

### Multi-line queries

For longer queries, use YAML multi-line syntax:

```yaml
type: api
sql: |
  SELECT
    publisher,
    domain,
    COUNT(*) as impressions,
    AVG(bid_price) as avg_bid
  FROM ad_bids
  WHERE timestamp >= '2024-01-01'
  GROUP BY publisher, domain
  ORDER BY impressions DESC
  LIMIT 50
```

## Querying models and tables

By default, SQL APIs execute against your project's default [OLAP connector](/developers/build/connectors/olap). You can query any [model](/developers/build/models) or source table:

```yaml
# Query a model
type: api
sql: SELECT * FROM my_model WHERE status = 'active' LIMIT 100
```

```yaml
# Query with joins across models
type: api
sql: |
  SELECT
    o.order_id,
    o.total,
    c.name as customer_name
  FROM orders o
  JOIN customers c ON o.customer_id = c.id
  ORDER BY o.total DESC
  LIMIT 25
```

## Querying external databases

You can query external databases directly by specifying a `connector`. This lets you access data in real-time without ingesting it into Rill.

### Data warehouses

**Athena:**
```yaml
type: api
connector: athena
sql: SELECT * FROM s3_data_table WHERE event_date >= '2024-01-01' LIMIT 100
```

**BigQuery:**
```yaml
type: api
connector: bigquery
sql: SELECT * FROM `my-project.my_dataset.my_table` WHERE region = 'us' LIMIT 100
```

**Redshift:**
```yaml
type: api
connector: redshift
sql: SELECT * FROM transactions WHERE transaction_date >= '2024-01-01' LIMIT 100
```

**Snowflake:**
```yaml
type: api
connector: snowflake
sql: SELECT * FROM my_database.my_schema.events WHERE created_at >= '2024-01-01' LIMIT 100
```

### Databases

**MySQL:**
```yaml
type: api
connector: mysql
sql: SELECT * FROM orders WHERE order_date >= '2025-01-01' LIMIT 100
```

**Postgres:**
```yaml
type: api
connector: postgres
sql: SELECT * FROM events WHERE created_at >= '2025-01-01' LIMIT 100
```

:::warning External database costs
Queries to external connectors execute directly on your data source and incur costs based on your provider's billing model:
- **Athena / BigQuery** — charged per TB of data scanned
- **Redshift / Snowflake** — charged for compute time
- **MySQL / Postgres** — may incur costs based on instance compute and IOPS

To minimize costs: use `LIMIT` clauses, apply filters to reduce data scanned, and consider materializing frequently accessed queries as [models](/developers/build/models) in DuckDB.
:::

## When to use external connectors vs your OLAP engine

| Factor | OLAP engine (default) | External connector |
|--------|-------------------|--------------------|
| **Query speed** | Fast — data is local | Depends on source (network + query time) |
| **Data freshness** | As of last refresh | Real-time from the source |
| **Cost** | No additional cost | Per-query costs from your provider |
| **Best for** | Low-latency APIs, pre-modeled data | Real-time access, ad-hoc queries |

**Use your OLAP engine** when you need fast, low-cost queries against data already modeled in Rill. Your data refreshes on a schedule and is optimized for analytical queries.

**Use external connectors** when you need real-time access to the latest data, your data lives in the source database, or you're building internal tools where query costs are acceptable.

## Adding dynamic behavior

SQL APIs support [templating](/developers/build/custom-apis/templating) for dynamic arguments, user attributes, and conditional logic:

```yaml
type: api
sql: |
  SELECT publisher, COUNT(*) as total
  FROM ad_bids
  WHERE domain = '{{ .args.domain }}'
  {{ if .user.admin }}
    AND internal_flag IS NOT NULL
  {{ end }}
  GROUP BY publisher
  LIMIT {{ default 100 .args.limit }}
```

See [Dynamic Queries with Templating](/developers/build/custom-apis/templating) for the full guide.
