---
title: Calling Custom APIs
description: How to call and consume your custom APIs from external applications
sidebar_label: Calling APIs
sidebar_position: 70
---

Custom APIs are exposed as HTTP endpoints that return JSON. You can call them from any language or tool that supports HTTP requests.

## API endpoints

### Rill Cloud

```
https://api.rilldata.com/v1/organizations/<org-name>/projects/<project-name>/runtime/api/<api-name>
```

### Local development

```
http://localhost:9009/v1/instances/default/api/<api-name>
```

Where `<api-name>` is the name of your API file without the `.yaml` extension (e.g., `my-api.yaml` → `my-api`).

## Making requests

Custom APIs accept both GET and POST requests.

### GET with query parameters

```bash
curl "https://api.rilldata.com/v1/organizations/<org>/projects/<project>/runtime/api/my-api?domain=google.com&limit=10" \
  -H "Authorization: Bearer <token>"
```

### POST with JSON body

```bash
curl -X POST "https://api.rilldata.com/v1/organizations/<org>/projects/<project>/runtime/api/my-api" \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"domain": "google.com", "limit": 10}'
```

Both methods produce the same result. If you provide both query parameters and a JSON body, query parameters take precedence.

### Response format

APIs return a JSON array of objects:

```json
[
  {"publisher": "Facebook", "domain": "google.com", "total": 15234},
  {"publisher": "Google", "domain": "google.com", "total": 12876}
]
```

## Testing locally

Local development does not require authentication:

```bash
# GET request
curl "http://localhost:9009/v1/instances/default/api/my-api?domain=google.com"

# POST request
curl -X POST http://localhost:9009/v1/instances/default/api/my-api \
  -H "Content-Type: application/json" \
  -d '{"domain": "google.com"}'
```

:::note
User attributes (`{{ .user.* }}`) are not available during local testing since no authentication token is provided. To test APIs that depend on user attributes, deploy to Rill Cloud and use a service token with [custom attributes](/developers/build/custom-apis/security#custom-attributes-on-service-tokens).
:::

## Authentication

Rill Cloud APIs require a bearer token in the `Authorization` header.

### For development and testing

Create a [user token](/guide/administration/access-tokens/user-tokens) (inherits your personal permissions):

```bash
rill token issue --display-name "API Testing"
# Returns: rill_usr_...

curl "https://api.rilldata.com/v1/organizations/<org>/projects/<project>/runtime/api/my-api" \
  -H "Authorization: Bearer rill_usr_..."
```

### For production systems

Create a [service token](/guide/administration/access-tokens/service-tokens) with optional custom attributes:

```bash
rill service create my-api-service \
  --project my-project \
  --project-role viewer \
  --attributes '{"customer_id": "acme-corp"}'
# Returns: rill_svc_...
```

Custom attributes on the token are available in your API templates as `{{ .user.customer_id }}`. See [Security & Access Control](/developers/build/custom-apis/security) for details on how to use custom attributes to build multi-tenant APIs.

:::tip Token Documentation
For full guidance on token types, roles, and management:
- **[User Tokens](/guide/administration/access-tokens/user-tokens)** — Personal access tokens for development
- **[Service Tokens](/guide/administration/access-tokens/service-tokens)** — Long-lived tokens for production systems
- **[Roles and Permissions](/guide/administration/users-and-access/roles-permissions)** — Understand access levels
:::

## OpenAPI schema

Rill automatically generates an OpenAPI spec for your project. Download it to generate typed clients:

```bash
# From Rill Cloud
curl "https://api.rilldata.com/v1/organizations/<org>/projects/<project>/runtime/api/openapi" \
  -H "Authorization: Bearer <token>" \
  -o openapi.json

# Locally
curl http://localhost:9009/v1/instances/default/api/openapi -o openapi.json
```

See [OpenAPI Documentation](/developers/build/custom-apis/openapi) for how to add request and response schemas to your API definitions.
