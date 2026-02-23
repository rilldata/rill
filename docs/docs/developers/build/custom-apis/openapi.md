---
title: OpenAPI Documentation
description: Add OpenAPI specs to your custom APIs for documentation and client generation
sidebar_label: OpenAPI Documentation
sidebar_position: 60
---

Rill automatically generates an OpenAPI specification for your project that combines built-in APIs with your custom API definitions. You can add request and response schemas to your APIs for better documentation and typed client generation.

## Adding an OpenAPI spec to your API

Add an `openapi` block to your API definition with `request_schema` and `response_schema`:

```yaml
type: api

sql: |
  SELECT publisher, COUNT(*) as total_records
  FROM ad_bids
  WHERE domain = '{{ .args.domain }}'
  {{ if hasKey .args "publisher" }}
    AND publisher = '{{ .args.publisher }}'
  {{ end }}
  ORDER BY total_records DESC
  LIMIT {{ default 25 .args.limit }}
  OFFSET {{ default 0 .args.offset }}

openapi:
  summary: Get ad bid statistics by publisher, filtered by domain

  request_schema:
    type: object
    required:
      - domain
    properties:
      domain:
        type: string
        description: Domain to filter results by
      publisher:
        type: string
        description: Optional publisher filter
      limit:
        type: integer
        description: Number of results to return (default 25)
      offset:
        type: integer
        description: Offset for pagination (default 0)

  response_schema:
    type: object
    properties:
      publisher:
        type: string
        description: Publisher name
      total_records:
        type: integer
        description: Total number of ad bid records
```

### OpenAPI fields

| Field | Description |
|-------|-------------|
| `openapi.summary` | Short description of what the API does (appears in OpenAPI docs) |
| `openapi.request_schema` | JSON Schema describing the request parameters |
| `openapi.response_schema` | JSON Schema describing a single row in the response array |

Schemas follow the [JSON Schema](https://json-schema.org/) format. You can use `type`, `required`, `properties`, `description`, and other standard JSON Schema keywords.

## Downloading the OpenAPI spec

### Locally

```bash
curl http://localhost:9009/v1/instances/default/api/openapi -o openapi.json
```

### From Rill Cloud

```bash
curl "https://api.rilldata.com/v1/organizations/<org>/projects/<project>/runtime/api/openapi" \
  -H "Authorization: Bearer <token>" \
  -o openapi.json
```

The generated spec includes all your custom APIs with their schemas, plus Rill's built-in API endpoints.

## Generating typed clients

Use the downloaded OpenAPI spec with any code generation tool to create typed clients for your language:

**JavaScript/TypeScript** (using [openapi-typescript](https://github.com/openapi-ts/openapi-typescript)):
```bash
npx openapi-typescript openapi.json -o ./src/api-types.ts
```

**Python** (using [openapi-python-client](https://github.com/openapi-generators/openapi-python-client)):
```bash
openapi-python-client generate --path openapi.json
```

**Go** (using [oapi-codegen](https://github.com/oapi-codegen/oapi-codegen)):
```bash
oapi-codegen -package api openapi.json > api/client.go
```

## Full example

Here's a complete API with OpenAPI documentation, templating, and security:

```yaml
type: api

metrics_sql: |
  SELECT publisher, domain, total_records, avg_bid_price
  FROM ad_bids_metrics
  WHERE domain = '{{ .args.domain }}'
  {{ if hasKey .args "publisher" }}
    AND publisher = '{{ .args.publisher }}'
  {{ end }}
  ORDER BY total_records DESC
  LIMIT {{ default 25 .args.limit }}
  OFFSET {{ default 0 .args.offset }}

security:
  access: true

openapi:
  summary: Query ad bid metrics by domain with optional publisher filter

  request_schema:
    type: object
    required:
      - domain
    properties:
      domain:
        type: string
        description: Domain to filter metrics by (e.g., "google.com")
      publisher:
        type: string
        description: Optional publisher to filter by (e.g., "Facebook")
      limit:
        type: integer
        description: Max results to return (default 25, max 1000)
      offset:
        type: integer
        description: Pagination offset (default 0)

  response_schema:
    type: object
    properties:
      publisher:
        type: string
        description: Publisher name
      domain:
        type: string
        description: Domain name
      total_records:
        type: integer
        description: Total number of records
      avg_bid_price:
        type: number
        description: Average bid price in USD
```
