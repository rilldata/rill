---
title: Security & Access Control
description: Control who can access your APIs and what data they see using security rules and custom attributes
sidebar_label: Security & Access Control
sidebar_position: 50
---

Rill's custom APIs support fine-grained access control through security rules and custom attributes on tokens. You can restrict who can call an API, and filter the data each caller sees — all without writing backend code.

## API access rules

Control who can access an API using the `security` block:

### Allow all authenticated users

```yaml
type: api
sql: SELECT publisher, COUNT(*) as total FROM ad_bids GROUP BY publisher
security:
  access: true
```

### Restrict to admins only

```yaml
type: api
sql: |
  SELECT publisher, SUM(revenue) as total_revenue
  FROM ad_bids
  GROUP BY publisher
security:
  access: "{{ .user.admin }}"
```

Only users with admin permissions on the project can call this API. Non-admins receive a 403 Forbidden response.

### Restrict by custom attribute

```yaml
type: api
sql: SELECT * FROM internal_reports
security:
  access: "{{ eq .user.tier \"enterprise\" }}"
```

Only users whose token has `tier: "enterprise"` can access this endpoint.

## Custom attributes on service tokens

Custom attributes are key-value pairs you attach to [service tokens](/guide/administration/access-tokens/service-tokens). When a service token is used to call an API, its attributes are available in templates as `{{ .user.<attribute> }}`.

### Creating a service token with attributes

```bash
rill service create acme-api \
  --project my-project \
  --project-role viewer \
  --attributes '{"customer_id": "acme-corp", "region": "us-west", "tier": "premium"}'
```

This creates a token with three custom attributes: `customer_id`, `region`, and `tier`.

### Updating attributes on an existing service

```bash
rill service edit acme-api \
  --attributes '{"customer_id": "acme-corp", "region": "eu-central", "tier": "enterprise"}'
```

### Common attribute patterns

| Attribute | Use case |
|-----------|----------|
| `customer_id` | Multi-tenant data isolation |
| `region` | Geographic data filtering |
| `department` | Departmental access control |
| `tier` | Feature gating (free, premium, enterprise) |
| `environment` | Environment-specific data (production, staging) |

## How attributes flow through the system

When an API is called with a service token, here's what happens:

```
1. Service token created with attributes: {"customer_id": "acme"}
                    ↓
2. API call with bearer token
                    ↓
3. Rill extracts attributes from the token into JWT claims
                    ↓
4. Template engine makes attributes available as {{ .user.customer_id }}
                    ↓
5. SQL query is rendered with the actual values
                    ↓
6. Query executes and returns filtered results
```

## End-to-end example: multi-tenant API

This walkthrough shows how to build an API that serves different data to different customers.

### Step 1: Create the API

Create `apis/customer-orders.yaml`:

```yaml
type: api
sql: |
  SELECT
    order_id,
    product_name,
    quantity,
    total_price,
    order_date
  FROM orders
  WHERE customer_id = '{{ .user.customer_id }}'
  ORDER BY order_date DESC
  LIMIT {{ default 50 .args.limit }}
  OFFSET {{ default 0 .args.offset }}
security:
  access: true
```

### Step 2: Create service tokens for each customer

```bash
# Token for Acme Corp
rill service create acme-api \
  --project my-project \
  --project-role viewer \
  --attributes '{"customer_id": "acme-corp"}'
# Returns: rill_svc_abc123...

# Token for Globex Inc
rill service create globex-api \
  --project my-project \
  --project-role viewer \
  --attributes '{"customer_id": "globex-inc"}'
# Returns: rill_svc_def456...
```

### Step 3: Call the API

**Acme sees only their orders:**
```bash
curl "https://api.rilldata.com/v1/organizations/my-org/projects/my-project/runtime/api/customer-orders" \
  -H "Authorization: Bearer rill_svc_abc123..."
```

```json
[
  {"order_id": "A-1001", "product_name": "Widget Pro", "quantity": 50, "total_price": 2500, "order_date": "2025-01-15"},
  {"order_id": "A-1002", "product_name": "Gadget Plus", "quantity": 25, "total_price": 1250, "order_date": "2025-01-14"}
]
```

**Globex sees only their orders:**
```bash
curl "https://api.rilldata.com/v1/organizations/my-org/projects/my-project/runtime/api/customer-orders" \
  -H "Authorization: Bearer rill_svc_def456..."
```

```json
[
  {"order_id": "G-2001", "product_name": "Sprocket X", "quantity": 100, "total_price": 5000, "order_date": "2025-01-16"},
  {"order_id": "G-2002", "product_name": "Bolt Kit", "quantity": 200, "total_price": 800, "order_date": "2025-01-13"}
]
```

Same API, same endpoint — different data based on the token's `customer_id` attribute.

## Admin vs non-admin patterns

Use `{{ .user.admin }}` to expose different data or behavior based on the user's role:

### Show extra columns for admins

```yaml
type: api
sql: |
  SELECT
    publisher,
    domain,
    COUNT(*) as impressions
    {{ if .user.admin }}
      , SUM(revenue) as total_revenue
      , AVG(cost_per_click) as avg_cpc
    {{ end }}
  FROM ad_bids
  GROUP BY publisher, domain
  ORDER BY impressions DESC
  LIMIT 50
```

### Remove filters for admins

```yaml
type: api
sql: |
  SELECT publisher, domain, COUNT(*) as total
  FROM ad_bids
  WHERE 1=1
  {{ if (not .user.admin) }}
    AND customer_id = '{{ .user.customer_id }}'
  {{ end }}
  GROUP BY publisher, domain
```

Admins see all data across all customers; non-admins see only their customer's data.

## Metrics SQL security inheritance

When using [Metrics SQL APIs](/developers/build/custom-apis/metrics-sql), security policies defined on the metrics view are automatically enforced. You don't need to add a `security` block to the API — it inherits the metrics view's policies:

```yaml
# metrics/ad_bids_metrics.yaml
type: metrics_view
model: ad_bids
security:
  access: true
  row_filter: "customer_id = '{{ .user.customer_id }}'"
```

```yaml
# apis/customer-metrics.yaml
type: api
metrics_sql: |
  SELECT publisher, total_records
  FROM ad_bids_metrics
  ORDER BY total_records DESC
```

The `row_filter` from the metrics view is automatically applied — each customer only sees their own data, even though the API definition doesn't mention security at all.

## Skipping nested security

By default, when an API queries a metrics view, Rill enforces the security policies on both the API itself and the underlying metrics view. In some cases, you may want the API to handle all access control itself and skip checks on nested resources:

```yaml
type: api
sql: |
  SELECT * FROM sensitive_model
  WHERE access_level <= {{ .user.access_level }}
security:
  access: true
skip_nested_security: true
```

Use `skip_nested_security: true` when your API already handles all necessary access control in its own query logic.

## Issuing ephemeral tokens

For applications that need to issue short-lived tokens to end users (e.g., for embedded dashboards or temporary API access), service tokens can issue ephemeral tokens with custom user attributes:

```bash
curl -X POST "https://api.rilldata.com/v1/orgs/<org>/projects/<project>/credentials" \
  -H "Authorization: Bearer <service-token>" \
  -H "Content-Type: application/json" \
  -d '{
    "attributes": {
      "email": "user@acme.com",
      "customer_id": "acme-corp",
      "department": "engineering"
    },
    "ttl_seconds": 3600
  }'
```

The response contains a short-lived JWT that can be used to call APIs with those attributes. See [Service Tokens](/guide/administration/access-tokens/service-tokens#issuing-ephemeral-tokens) for details.
