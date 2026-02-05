---
title: Metrics SQL 
description: Query metrics views using SQL syntax
sidebar_label: Metrics SQL
---

You can write a SQL query referring to metrics definitions and dimensions defined in a metrics view.
It should have the following structure:
    
```yaml
type: api
metrics_sql: SELECT publisher, domain, total_records FROM ad_bids_metrics
```

## Querying Fundamentals

Metrics SQL transforms queries that reference `dimensions` and `measures` within a `metrics view` into their corresponding database columns or expressions. This transformation is based on the mappings defined in a metrics view YAML configuration, enabling reuse of dimension or measure definitions. Additionally, any security policies defined in the metrics view are also inherited.

## Example: Crafting a Metrics SQL Query

Consider a metrics view configured as follows:
```yaml
#metrics/ad_bids_metrics.yaml
type: metrics_view
title: Ad Bids
model: ad_bids
timeseries: timestamp
dimensions:
  - name: publisher
    expression: toUpper(publisher)
  - name: domain
    column: domain
measures:
  - name: total_records
    display_name: Total records
    expression: COUNT(*)
```

To query this view, a user might write a Metrics SQL query like:
```sql
SELECT publisher, domain, total_records FROM ad_bids_metrics
```

This Metrics SQL is internally translated to a standard SQL query as follows:
```sql
SELECT toUpper(publisher) AS publisher, domain AS domain, COUNT(*) AS total_records FROM ad_bids_metrics GROUP BY publisher, domain
```

## Security and Compliance

Queries executed via Metrics SQL are subject to the security policies and access controls defined in the metrics view YAML configuration, ensuring data security and compliance.

## Limitations

Metrics SQL is specifically designed for querying metrics views and may not support all features found in standard SQL. Its primary focus is on providing an efficient and easy way to extract data within the constraints of metrics view configurations.

## Supported SQL Features

- **SELECT** statements with plain `dimension` and `measure` references.
- A single **FROM** clause referencing a `metrics view`.
- **WHERE** clause that can reference selected `dimensions` only.
- Operators in **WHERE** and **HAVING** clauses include `=`, `!=`, `>`, `>=`, `<`, `<=`, IN, LIKE, AND, OR, and parentheses for structuring the expression.
- **HAVING** clause for filtering on aggregated results, referencing selected dimension and measure names. Supports the same expression capabilities as the WHERE clause.
- **ORDER BY** clause for sorting the results.
- **LIMIT** and **OFFSET** clauses for controlling the result set size and pagination.

:::warning
 The Metrics SQL feature is currently evolving. We are dedicated to enhancing the syntax by introducing additional SQL features, while striving to maintain support for existing syntax. However, please be advised that backward compatibility cannot be guaranteed at all times. Additionally, users should be aware that there may be untested edge cases in the current implementation. We appreciate your understanding as we work to refine and improve this feature.
:::

## SQL Templating

You can use templating to make your Metrics SQL query dynamic. We support:
 - Dynamic arguments that can be passed in as query parameters during the API call using `{{ .args.<param-name> }}`
 - User attributes like email, domain, and admin if available using `{{ .user.<attr> }}` (see integration docs [here](/developers/integrate/custom-api) for when user attributes are available)
 - Conditional statements
 - Optional parameters paired with conditional statements.

See integration docs [here](/developers/integrate/custom-api) to learn how these are passed in when calling the API.

### Conditional statements

Assume an API endpoint defined as `my-api.yaml`:
```yaml
type: api
metrics_sql: |
  SELECT publisher, total_records
    {{ if ( .user.admin ) }} ,domain  {{ end }} 
    FROM ad_bids_metrics WHERE timestamp::DATE = '{{ .args.date }}' 
    {{ if ( .user.admin ) }} GROUP BY publisher, domain {{ else }} GROUP BY publisher {{ end }}
```
If the user is an admin, the API will return the count of records by `publisher` and `domain` for the given date. If the user is not an admin, the API will return the total count of records by `publisher` for the given date.

### Optional parameters

Rill utilizes standard Go templating together with [Sprig](http://masterminds.github.io/sprig/), which adds a number of useful utility functions.  
One of those functions is `hasKey`, which in the example below enables optional parameters to be passed to the Custom API endpoint. This allows you to build API endpoints that can handle a wider range of parameters and logic, reducing the need to duplicate API endpoints.

Assume an API endpoint defined as `my-api.yaml`:
```yaml
type: api
metrics_sql: |
  SELECT
    publisher,
    total_records
  FROM ad_bids_metrics
  {{ if hasKey .args "publisher" }} WHERE publisher = '{{ .args.publisher }}' {{ end }} 
  GROUP BY publisher
```

HTTP GET `.../runtime/api/my-api` would return `total_records` for all `publisher`s.  
HTTP GET `.../runtime/api/my-api?publisher=Google` would return `total_records` for `Google`.

## Add an OpenAPI spec

You can optionally provide OpenAPI annotations for the request and response schema in your custom API definition. These will automatically be incorporated in the OpenAPI spec for your project (see [Custom API Integration](/developers/integrate/custom-api) for details).

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

## How to use Metrics SQL APIs

Refer to the integration docs [here](/developers/integrate/custom-api) to learn how to use Metrics SQL APIs in your application.

