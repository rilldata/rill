---
title: Custom APIs
description: Expose your Rill data as HTTP API endpoints
sidebar_label: Custom APIs
sidebar_position: 10
---

Rill lets you create custom API endpoints that return data from your project as JSON over HTTP. Define a YAML file, write a SQL query, and you have an API — no backend code required.

Custom APIs are ideal for:
- **Powering internal tools** — feed Rill data into dashboards, Slack bots, or scripts
- **Building customer-facing integrations** — expose filtered data to external applications
- **Automating workflows** — pull data into CI/CD pipelines, scheduled jobs, or ETL processes
- **Multi-tenant data access** — serve different data to different customers using [custom attributes](/developers/build/custom-apis/security)

## API types

Rill supports two types of custom APIs:

| Type | Best for | Query target |
|------|----------|-------------|
| [**SQL API**](/developers/build/custom-apis/sql) | Querying models, tables, or external databases directly | Any model, table, or external connector (DuckDB, BigQuery, Snowflake, etc.) |
| [**Metrics SQL API**](/developers/build/custom-apis/metrics-sql) | Querying metrics views using dimension and measure names | Metrics views (inherits security policies automatically) |

## Your first custom API

### 1. Create an API file

Create a YAML file in your project's `apis/` directory. For example, `apis/top-publishers.yaml`:

```yaml
type: api
sql: |
  SELECT publisher, COUNT(*) as total_records
  FROM ad_bids
  GROUP BY publisher
  ORDER BY total_records DESC
  LIMIT 10
```

### 2. Test it locally

With Rill Developer running (`rill start`), call your API at:

```bash
curl "http://localhost:9009/v1/instances/default/api/top-publishers"
```

You'll get a JSON response:
```json
[
  {"publisher": "Facebook", "total_records": 15234},
  {"publisher": "Google", "total_records": 12876},
  {"publisher": "Microsoft", "total_records": 9541}
]
```

:::note
Local development does not require authentication. When deployed to Rill Cloud, all API calls require a bearer token.
:::

### 3. Deploy and call from Rill Cloud

After deploying your project, call the API with authentication:

```bash
curl "https://api.rilldata.com/v1/organizations/<org>/projects/<project>/runtime/api/top-publishers" \
  -H "Authorization: Bearer <token>"
```

See [Custom API Integration](/developers/integrate/custom-api) for full details on authentication and calling APIs.

## Make it dynamic

Add [templating](/developers/build/custom-apis/templating) to make your API accept parameters:

```yaml
type: api
sql: |
  SELECT publisher, COUNT(*) as total_records
  FROM ad_bids
  WHERE domain = '{{ .args.domain }}'
  GROUP BY publisher
  ORDER BY total_records DESC
  LIMIT {{ default 10 .args.limit }}
```

Call it with query parameters:

```bash
curl "http://localhost:9009/v1/instances/default/api/top-publishers?domain=google.com&limit=5"
```

## Add access control

Use [security rules](/developers/build/custom-apis/security) to control who can access your API and what data they see:

```yaml
type: api
sql: |
  SELECT publisher, domain, COUNT(*) as total_records
  FROM ad_bids
  WHERE customer_id = '{{ .user.customer_id }}'
  GROUP BY publisher, domain
security:
  access: true
```

Each customer sees only their own data, based on the [custom attributes](/developers/build/custom-apis/security#custom-attributes-on-service-tokens) in their token.

## Next steps

- [**Templating**](/developers/build/custom-apis/templating) — Dynamic arguments, user attributes, and conditional logic
- [**Security & Access Control**](/developers/build/custom-apis/security) — Custom attributes, multi-tenant patterns, and access rules
- [**OpenAPI Documentation**](/developers/build/custom-apis/openapi) — Document your APIs with OpenAPI specs
- [**Calling APIs**](/developers/integrate/custom-api) — HTTP endpoints, authentication, and client generation
