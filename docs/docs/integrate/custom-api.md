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

## OpenAPI schema

Rill automatically generates an OpenAPI spec that combines the built-in metrics APIs with your custom API definitions. You can use this OpenAPI spec to generate a typed client for accessing Rill from your programming language of choice. You can download the customized OpenAPI spec with:
```bash
curl https://api.rilldata.com/v1/organizations/<org-name>/projects/<project-name>/runtime/api/openapi \
  -H "Authorization: Bearer <token>" \
  -o openapi.json
```

## Access tokens

There are three types of access tokens you can use with custom APIs.

### User tokens

These tokens are tied to your personal user and access permissions. They are useful for local scripting and experimentation. Not recommended for production use. 

Use the `rill token issue` CLI command to obtain a personal access token. See the [CLI reference](../reference/cli/token) for details.

### Service tokens

These tokens are tied to your Rill organization and persist even if the creating user is removed. Service tokens can be assigned specific roles (organization-level or project-level) and custom attributes. They are recommended for use cases that integrate Rill into production systems (such as scheduled jobs or backend APIs).

Since service tokens can have broad permissions, they MUST NOT be embedded directly in your frontend or otherwise shared with end users. See "Ephemeral tokens" below for how to create safe, short-lived access tokens.

**Creating a service token:**
```bash
# Basic service token with organization role
rill service create my-service --org-role admin

# Service token with project role and custom attributes
rill service create my-service \
  --project my-project \
  --project-role viewer \
  --attributes '{"department":"engineering","region":"us-west"}'
```

**Using custom attributes with security policies:**

Custom attributes allow you to implement fine-grained access control. For example, if you create a service token with a `customer_id` attribute:

```bash
rill service create customer-api \
  --project-role viewer \
  --attributes '{"customer_id":"acme-corp"}'
```

You can reference these attributes in your security policies:

```yaml
# In your metrics view
security:
  access: true
  row_filter: customer_id = '{{ .user.customer_id }}'
```

For comprehensive documentation on service tokens, including roles, attributes, and management, see [Service Tokens](/manage/service-tokens). Also see the [CLI reference](../reference/cli/service) for command details.

### Ephemeral user tokens

You can use a service token to issue a short-lived, ephemeral access token with arbitrary user attributes. This enables you to create tokens that mimic an end user, even for users who are not signed up for Rill. Unlike service tokens, the access permissions of an ephemeral token are scoped to a specific project and user attributes.

The primary use case for these tokens is to have your backend issue a short-lived token that represents your current user, which your frontend can use to make direct calls to APIs in Rill. This is the same feature that powers Rill's embedded dashboards.

To get an ephemeral user token, you need to use a service token to perform a handshake with Rill's credentials API at `https://api.rilldata.com/v1/orgs/<org-name>/projects/<project-name>/credentials`. For example:
```bash
curl -X POST https://api.rilldata.com/v1/orgs/<org-name>/projects/<project-name>/credentials \
  -H "Authorization: Bearer <service-account-token>" \
  --data-raw '{ "user_email":"<user-email>" }'
``` 

The API accepts the following parameters:
- `user_email`: Optional user email that the token should represent. The user does not need to exist in Rill.
- `attributes`: Optional raw JSON payload of user attributes. This setting is not compatible with `user_email`. When using this, make sure to explicitly pass all the attributes used in your security policies, like `email` or `domain`.
- `ttl_seconds`: Optional time-to-live for the token. Defaults to 24 hours (86400 seconds).
