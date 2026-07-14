---
title: Dynamic Queries with Templating
description: Use templating to make your custom APIs dynamic with arguments, user attributes, and conditional logic
sidebar_label: Templating
sidebar_position: 40
---

Rill's custom APIs support Go-style templating to make your SQL queries dynamic. You can accept parameters from API callers, reference user attributes from authentication tokens, and use conditional logic to build flexible endpoints.

Templating works with both [SQL APIs](/developers/build/custom-apis/sql) and [Metrics SQL APIs](/developers/build/custom-apis/metrics-sql).

## Template context reference

Every template has access to the following context:

| Variable | Type | Description |
|----------|------|-------------|
| `{{ .args.<name> }}` | `any` | Runtime arguments passed via query parameters or POST body |
| `{{ .user.email }}` | `string` | Authenticated user's email address |
| `{{ .user.domain }}` | `string` | Email domain of the authenticated user |
| `{{ .user.name }}` | `string` | Display name of the authenticated user |
| `{{ .user.admin }}` | `bool` | Whether the user has admin permissions on the project |
| `{{ .user.groups }}` | `[]string` | User groups the authenticated user belongs to |
| `{{ .user.<custom> }}` | `any` | Custom attributes from service tokens (e.g., `customer_id`, `region`) |
| `{{ .export }}` | `bool` | `true` when the API is being resolved for export (CSV, Excel, Parquet) |

:::note
When testing locally (`localhost:9009`), `.user` attributes are not available since no authentication is required. To test with user attributes, deploy to Rill Cloud and use a service token with [custom attributes](/developers/build/custom-apis/security#custom-attributes-on-service-tokens).
:::

## Dynamic arguments

Pass arguments to your API via query parameters (GET) or JSON body (POST). Reference them with `{{ .args.<name> }}`:

```yaml
type: api
sql: |
  SELECT publisher, domain, COUNT(*) as total
  FROM ad_bids
  WHERE domain = '{{ .args.domain }}'
  GROUP BY publisher, domain
```

**Calling the API:**
```bash
# Via query parameter
curl "http://localhost:9009/v1/instances/default/api/my-api?domain=google.com"

# Via POST body
curl -X POST http://localhost:9009/v1/instances/default/api/my-api \
  -H "Content-Type: application/json" \
  -d '{"domain": "google.com"}'
```

### Multiple arguments

```yaml
type: api
sql: |
  SELECT publisher, domain, COUNT(*) as total
  FROM ad_bids
  WHERE domain = '{{ .args.domain }}'
    AND publisher = '{{ .args.publisher }}'
  GROUP BY publisher, domain
```

```bash
curl "http://localhost:9009/v1/instances/default/api/my-api?domain=google.com&publisher=Facebook"
```

## User attributes

Reference the authenticated user's attributes with `{{ .user.<attr> }}`. These come from the user's identity (for user tokens) or from custom attributes (for service tokens).

### Built-in attributes

```yaml
type: api
sql: |
  SELECT *
  FROM reports
  WHERE owner_email = '{{ .user.email }}'
  LIMIT 50
```

### Custom attributes from service tokens

[Service tokens with custom attributes](/developers/build/custom-apis/security#custom-attributes-on-service-tokens) are available in templates as `{{ .user.<attribute> }}`:

```yaml
type: api
sql: |
  SELECT order_id, product, total
  FROM orders
  WHERE customer_id = '{{ .user.customer_id }}'
    AND region = '{{ .user.region }}'
```

See [Security & Access Control](/developers/build/custom-apis/security) for creating tokens and a full multi-tenant walkthrough.

## Conditional logic

Use Go template `{{ if }}` / `{{ else }}` / `{{ end }}` blocks to conditionally include SQL:

### Admin-only columns

```yaml
type: api
sql: |
  SELECT
    publisher,
    COUNT(*) as total_records
    {{ if .user.admin }}
      , SUM(revenue) as total_revenue
      , AVG(bid_price) as avg_bid
    {{ end }}
  FROM ad_bids
  GROUP BY publisher
  ORDER BY total_records DESC
```

Admins see revenue and bid data; non-admins see only publisher and record count.

### Admin-only filters

```yaml
type: api
sql: |
  SELECT publisher, domain, COUNT(*) as total
  FROM ad_bids
  WHERE timestamp >= '{{ .args.start_date }}'
  {{ if (not .user.admin) }}
    AND domain = '{{ .user.domain }}'
  {{ end }}
  GROUP BY publisher, domain
```

Non-admins see only data for their domain; admins see everything.

## Optional parameters

Use the `hasKey` function to check whether an argument was provided, making parameters optional:

```yaml
type: api
sql: |
  SELECT publisher, COUNT(*) as total_records
  FROM ad_bids
  WHERE 1=1
  {{ if hasKey .args "publisher" }}
    AND publisher = '{{ .args.publisher }}'
  {{ end }}
  {{ if hasKey .args "domain" }}
    AND domain = '{{ .args.domain }}'
  {{ end }}
  GROUP BY publisher
  ORDER BY total_records DESC
```

**Without parameters** — returns all publishers:
```bash
curl "http://localhost:9009/v1/instances/default/api/my-api"
```

**With one parameter** — filters by publisher:
```bash
curl "http://localhost:9009/v1/instances/default/api/my-api?publisher=Google"
```

**With both** — filters by publisher and domain:
```bash
curl "http://localhost:9009/v1/instances/default/api/my-api?publisher=Google&domain=news.google.com"
```

## Pagination pattern

Build paginated APIs using `LIMIT` and `OFFSET` with the `default` function for sensible defaults:

```yaml
type: api
sql: |
  SELECT publisher, domain, bid_price, timestamp
  FROM ad_bids
  ORDER BY timestamp DESC
  LIMIT {{ default 25 .args.limit }}
  OFFSET {{ default 0 .args.offset }}
```

**Page 1 (first 25 results):**
```bash
curl "http://localhost:9009/v1/instances/default/api/my-api"
```

**Page 2:**
```bash
curl "http://localhost:9009/v1/instances/default/api/my-api?offset=25"
```

**Custom page size:**
```bash
curl "http://localhost:9009/v1/instances/default/api/my-api?limit=10&offset=20"
```

## Sprig utility functions

Rill uses standard Go templating together with [Sprig](http://masterminds.github.io/sprig/), which provides many utility functions. Commonly used ones:

| Function | Example | Description |
|----------|---------|-------------|
| `default` | `{{ default 100 .args.limit }}` | Use a default value if the argument is empty |
| `hasKey` | `{{ if hasKey .args "name" }}` | Check if an argument was provided |
| `lower` | `{{ lower .args.status }}` | Convert to lowercase |
| `upper` | `{{ upper .args.code }}` | Convert to uppercase |
| `trim` | `{{ trim .args.name }}` | Trim whitespace |
| `ne` | `{{ if (ne .user.domain "") }}` | Not equal comparison |
| `eq` | `{{ if (eq .user.region "us") }}` | Equal comparison |

### Example: combining multiple patterns

Here's a real-world example combining optional params, defaults, user attributes, and conditionals:

```yaml
type: api
sql: |
  SELECT
    publisher,
    domain,
    COUNT(*) as impressions,
    AVG(bid_price) as avg_bid
    {{ if .user.admin }}
      , SUM(revenue) as total_revenue
    {{ end }}
  FROM ad_bids
  WHERE 1=1
  {{ if hasKey .args "publisher" }}
    AND publisher = '{{ .args.publisher }}'
  {{ end }}
  {{ if hasKey .args "start_date" }}
    AND timestamp >= '{{ .args.start_date }}'
  {{ end }}
  {{ if (not .user.admin) }}
    AND domain = '{{ .user.domain }}'
  {{ end }}
  GROUP BY publisher, domain
  ORDER BY impressions DESC
  LIMIT {{ default 50 .args.limit }}
  OFFSET {{ default 0 .args.offset }}
```

This single endpoint handles:
- Optional filtering by publisher and date range
- Admin users see revenue; non-admins don't
- Non-admin users are scoped to their domain
- Pagination with sensible defaults
