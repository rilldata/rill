---
title: Create Custom APIs with SQL
description: Create custom APIs to retrieve aggregated data from Rill
sidebar_label: Custom APIs
---


Rill allows you to create custom APIs to pull data out in a flexible manner. You can write custom SQL queries and expose them as API endpoints.

To create a custom API, create a new YAML file under the `apis` directory in your Rill project. Currently, we support two types of custom APIs: Metrics SQL and SQL.

## Metrics SQL API

You can write a SQL query referring to metrics definitions and dimensions defined in a [metrics view](/build/metrics-view). 
It should have the following structure:
    
```yaml
type: api
metrics_sql: SELECT publisher, domain, total_records FROM ad_bids_metrics
```

For complete documentation on Metrics SQL API, including querying fundamentals, examples, supported SQL features, templating, and OpenAPI specs, see the [Metrics SQL API documentation](/build/metrics-view/metrics-sql).



## SQL API

You can write a SQL query and expose it as an API endpoint. This is useful when you want to directly write queries against a [model](/developer/build/models) that you have created. It should have the following structure:

```yaml
type: api
sql: SELECT publisher, domain, timestamp FROM ad_bids
```

### Querying External Databases

By default, SQL APIs execute queries against your default OLAP engine (typically DuckDB). However, you can specify a different OLAP engine using the `connector` parameter. This allows you to query data directly from **Athena**, **BigQuery**, **MySQL**, **Postgres**, **Redshift**, or **Snowflake** without ingesting them into Rill.

**Data Warehouses**

Athena:

```yaml
type: api
connector: athena
sql: SELECT * FROM s3_data_table LIMIT 100
```

BigQuery:

```yaml
type: api
connector: bigquery
sql: SELECT * FROM `rilldata.pricing.cloud_pricing_export` LIMIT 100
```

Redshift:

```yaml
type: api
connector: redshift
sql: SELECT * FROM transactions WHERE transaction_date >= '2024-01-01' LIMIT 100
```

Snowflake:

```yaml
type: api
connector: snowflake
sql: SELECT * FROM database.schema.table LIMIT 100
```

**OLTP Databases**

MySQL:

```yaml
type: api
connector: mysql
sql: SELECT * FROM orders WHERE order_date >= '2025-01-01' LIMIT 100
```

Postgres:

```yaml
type: api
connector: postgres
sql: SELECT * FROM events WHERE created_at >= '2025-01-01' LIMIT 100
```

:::warning Data Warehouse and Database Costs

When using alternative connectors (Athena, BigQuery, MySQL, Postgres, Redshift, Snowflake), queries execute directly on your data source and will **incur costs based on your provider's billing model**:
- **Athena**: Charges based on data scanned (per TB)
- **BigQuery**: Charges based on data scanned (per TB)
- **MySQL/Postgres**: May incur costs based on instance compute time and IOPS
- **Redshift**: Charges based on cluster compute time
- **Snowflake**: Charges based on warehouse compute time

To minimize costs:
- Use `LIMIT` clauses to restrict result set sizes
- Apply filters to reduce data scanned
- Consider materializing frequently accessed queries as models in DuckDB
- Monitor your warehouse's query costs and usage patterns

:::

**When to use alternative connectors for APIs:**
- Your data is already in the source database/warehouse
- You want real-time access to the latest data from the source
- You're building internal tools where query costs are acceptable
- [Querying partitions](/developer/build/models/partitioned-models) from underlying data source to ingest data in partitions in Rill

**When to use DuckDB (default):**
- You need fast, low-cost queries for end-user facing APIs
- Your data is already in Rill models
- You want predictable performance and costs
- You're serving external customers or high-volume requests
- You need the fastest possible query response times


## SQL Templating

You can use templating to make your SQL query dynamic. We support:
 - Dynamic arguments that can be passed in as query parameters during the API call using `{{ .args.<param-name> }}`
 - User attributes like email, domain, and admin if available using `{{ .user.<attr> }}` (see integration docs [here](/developer/integrate/custom-api) for when user attributes are available)
 - Conditional statements
 - Optional parameters paired with conditional statements.

See integration docs [here](/developer/integrate/custom-api) to learn how these are passed in when calling the API.

### Conditional statements

Assume an API endpoint defined as `my-api.yaml`:
```yaml
type: api
sql: |
  SELECT count(*)
    {{ if ( .user.admin ) }} ,publisher  {{ end }} 
    FROM ad_bids WHERE timestamp::DATE = '{{ .args.date }}' 
    {{ if ( .user.admin ) }} GROUP BY 2 {{ end }}
```

will expose an API endpoint like `https://api.rilldata.com/v1/organizations/<org-name>/projects/<project-name>/runtime/api/my-api?date=2021-01-01`.
If the user is an admin, the API will return the count of records by `publisher` for the given date. If the user is not an admin, the API will return the total count of records for the given date.


### Optional parameters

Rill utilizes standard Go templating together with [Sprig](http://masterminds.github.io/sprig/), which adds a number of useful utility functions.  
One of those functions is `hasKey`, which in the example below enables optional parameters to be passed to the Custom API endpoint. This allows you to build API endpoints that can handle a wider range of parameters and logic, reducing the need to duplicate API endpoints.

Assume an API endpoint defined as `my-api.yaml`:
```yaml
type: api
sql: |
  SELECT
    publisher,
    COUNT(*) as total_records
  FROM ad_bids
  {{ if hasKey .args "publisher" }} WHERE publisher = '{{ .args.publisher }}' {{ end }} 
  GROUP BY publisher
```

HTTP GET `.../runtime/api/my-api` would return `total_records` for all `publisher`s.  
HTTP GET `.../runtime/api/my-api?publisher=Google` would return `total_records` for `Google`.

## Add an OpenAPI spec

You can optionally provide OpenAPI annotations for the request and response schema in your custom API definition. These will automatically be incorporated in the OpenAPI spec for your project (see [Custom API Integration](/developer/integrate/custom-api) for details).

Example custom API with request and response schema:

```yaml
type: api

metrics_sql: >
  SELECT publisher, total_records
  FROM ad_bids_metrics
  WHERE domain = '{{ .args.domain }}'
  {{ if hasKey .args "limit" }} LIMIT {{ .args.limit }} {{ end }}
  {{ if hasKey .args "offset" }} OFFSET {{ .args.offset }} {{ end }}

openapi:
  request_schema:
    type: object
    required:
      - domain
    properties:
      domain:
        type: string
        description: Domain to filter sales by
      limit:
        type: integer
        description: Optional limit for pagination
      offset:
        type: integer
        description: Optional offset for pagination
  
  response_schema:
    type: object
    properties:
      publisher:
        type: string
        description: Publisher name
      total_records:
        type: number
        description: Total records for the publisher
```

## How to use custom APIs

Refer to the integration docs [here](/developer/integrate/custom-api) to learn how to use custom APIs in your application.
