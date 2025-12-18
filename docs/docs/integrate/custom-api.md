---
title: "Custom API Integration"
description: How to integrate custom APIs with your application
sidebar_label: "Custom API Integration"
sidebar_position: 20
---

Rill exposes [custom APIs](/build/custom-apis) you have created with `type: api` as HTTP endpoints 
at `https://api.rilldata.com/v1/organizations/<org-name>/projects/<project-name>/runtime/api/<name of api>`.

## Accessing custom APIs

Custom APIs accept both GET and POST requests to the API endpoint with a bearer token in the `Authorization` header. Parameters can always be passed using query arguments in the URL. For example:

```bash
curl https://api.rilldata.com/v1/organizations/<org-name>/projects/<project-name>/runtime/api/<name of api>[?query-args] \
  -H "Authorization: Bearer <token>"
```

For POST requests, if you send the `Content-Type: application/json` header, you can optionally also pass arguments as a JSON object in the request body.

```bash
curl -X POST https://api.rilldata.com/v1/organizations/<org-name>/projects/<project-name>/runtime/api/<name of api> \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"param1": "value1", "param2": "value2"}'
```

## Testing custom APIs locally

When developing and testing custom APIs with Rill Developer on localhost, you can access your APIs without authentication at:

```
http://localhost:9009/v1/instances/default/api/<filename>
```

Where `<filename>` is the name of your API file (without the `.yaml` extension).

### Local API examples

For a custom API defined in `my-api.yaml`:

**GET request:**
```bash
curl "http://localhost:9009/v1/instances/default/api/my-api?param1=value1&param2=value2"
```

**POST request:**
```bash
curl -X POST http://localhost:9009/v1/instances/default/api/my-api \
  -H "Content-Type: application/json" \
  -d '{"param1": "value1", "param2": "value2"}'
```

### Local OpenAPI schema

You can also access the OpenAPI spec locally without authentication:
```bash
curl http://localhost:9009/v1/instances/default/api/openapi -o openapi.json
```

:::note
Local development URLs do not require authentication tokens. This makes it easy to test your APIs during development, but remember to implement proper authentication when deploying to production.
:::

## OpenAPI schema

Rill automatically generates an OpenAPI spec that combines the built-in metrics APIs with your custom API definitions. You can use this OpenAPI spec to generate a typed client for accessing Rill from your programming language of choice. You can download the customized OpenAPI spec with:
```bash
curl https://api.rilldata.com/v1/organizations/<org-name>/projects/<project-name>/runtime/api/openapi \
  -H "Authorization: Bearer <token>" \
  -o openapi.json
```

## Authentication

Rill APIs require authentication tokens. Choose the appropriate token type for your use case:

### Quick Start

**For local testing (No authentication required when running locally):**  
```bash
# Test your API endpoint at http://localhost:9009/v1/instances/default/api/<filename>-

# Test your API endpoint locally (no auth required)
curl http://localhost:9009/v1/instances/default/api/<filename>
```

**For Rill Cloud testing:**
```bash
rill token issue --display-name "API Testing"
# Returns: rill_usr_...

curl https://api.rilldata.com/v1/organizations/<org>/projects/<project>/runtime/api/<api-name> \
  -H "Authorization: Bearer rill_usr_..."
```

**For production systems:**
```bash
rill service create my-api \
  --project-role viewer \
  --attributes '{"customer_id":"acme-corp"}'
# Returns: rill_svc_...
```

:::tip Token Documentation
For comprehensive guidance on token types, roles, custom attributes, and management:
- **[User Tokens](/manage/user-tokens)** - Personal access tokens for development
- **[Service Tokens](/manage/service-tokens)** - Long-lived tokens for production systems
- **[Roles and Permissions](/manage/roles-permissions)** - Understand access levels

:::

### Using custom attributes with security policies

Service tokens can include custom attributes for fine-grained access control. Reference these attributes in your [security policies](/build/metrics-view/security#advanced-example-custom-attributes-embed-dashboards):

```yaml
# In your metrics view
security:
  access: true
  row_filter: customer_id = '{{ .user.customer_id }}'
```
