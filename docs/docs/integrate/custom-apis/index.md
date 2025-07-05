---
title: "Custom API"
description: Create your own API to pull data out in flexible manner 
sidebar_label: "Custom API"
sidebar_position: 30
---

Rill allows you to create custom APIs to pull data out in a flexible manner. You can write custom SQL queries and expose them as an API endpoint.

## Create a custom API

To create a custom API, create a new yaml file under `apis` directory in your Rill project. Currently, we support two types of custom APIs:

### SQL API

You can write a SQL query and expose it as an API endpoint. This is useful when you want to directly write queries against a [model](/transform/models) that you have created. It should have the following structure:
    
```yaml
type: api
sql: SELECT abc FROM my_model
```

Read more details about [SQL apis](./sql-api.md).

### Metrics SQL API

You can write a SQL query referring to metrics definition and dimensions defined in a [metrics view](/define/metrics-view/metrics-view.md). 
It should have the following structure:
    
```yaml
type: api
metrics_sql: SELECT dimension, measure FROM my_metrics
```

Read more details about [Metrics SQL API](./metrics-sql-api.md).

## Add an OpenAPI spec

You can optionally provide OpenAPI annotations for the request and response schema in your custom API definition. These will automatically be incorporated in the OpenAPI spec for your project (see [Custom API Integration](/integrate/custom-api.md) for details).

Example custom API with request and response schema:

```yaml
type: api

metrics_sql: >
  SELECT product, total_sales
  FROM sales_metrics
  WHERE country = '{{ .args.country }}'
  {{ if hasKey .args "limit" }} LIMIT {{ .args.limit }} {{ end }}
  {{ if hasKey .args "offset" }} OFFSET {{ .args.offset }} {{ end }}

openapi:
  request_schema:
    type: object
    required:
      - country
    properties:
      country:
        type: string
        description: Country to filter sales by
      limit:
        type: integer
        description: Optional limit for pagination
      offset:
        type: integer
        description: Optional offset for pagination
  
  response_schema:
    type: object
    properties:
      product:
        type: string
        description: Product name
      total_sales:
        type: number
        description: Total sales for the product
```

## How to use custom APIs

Refer to the integration docs [here](/integrate/custom-api.md) to learn how to use custom APIs in your application.
