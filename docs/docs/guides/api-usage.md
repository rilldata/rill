---
title: "API Usage Guide"
description: "Guide to using Rill APIs with tokens, admin service endpoints, and OpenAPI integration"
sidebar_label: "API Usage"
sidebar_position: 10
---

# API Usage Guide

This guide covers how to authenticate with and use Rill's APIs, including user and service tokens, admin service endpoints, and OpenAPI integration.

## Authentication Tokens

Rill provides two types of authentication tokens for API access: user tokens and service tokens.

### User Tokens

User tokens are personal access tokens tied to your individual user account. They inherit your permissions and are useful for local scripting, experimentation, and development work.

**Creating a user token:**
```bash
rill token issue
? Please enter a display name for the token: "my-token"
Token: rill_usr_******************************************************
```

This command will prompt you to:
- Choose a display name for the token
- Generate the token

**Managing user tokens:**
```bash
# List all your tokens
rill token list
  ID (9)                                 DESCRIPTION       PREFIX                CLIENT             CREATED               EXPIRES   LAST USED
 -------------------------------------- ----------------- --------------------- ------------------ --------------------- --------- ---------------------
  4cc7307d-75a8-42ef-9e05-f0e6599849c7   test token        rill_usr_**********   Created manually   2025-09-16 14:58:49             2025-09-16 14:58:49

# Revoke a specific token
rill token revoke 4cc7307d-75a8-42ef-9e05-f0e6599849c7
```

> User tokens are not recommended for production use as they are tied to individual users and may expire when users leave your organization.

### Service Tokens

Service tokens are organization-level tokens that persist independently of individual users. They currently have admin access to all projects in the organization and are recommended for production integrations.

**Creating a service account and token:**
```bash
# Create a service account
rill service create <service-account-name>

# Generate a token for the service
rill service token issue <service-account-name>
```

**Managing service accounts:**
```bash
# List service accounts
rill service list

# Show service details
rill service show <service-name>

# Set role for a service
rill service set-role <service-name> --role <role>

# Delete a service account
rill service delete <service-name>
```

:::note
Service tokens have admin permissions across the organization. They should never be embedded directly in frontend applications or shared with end users.
:::

## Admin Service REST Endpoints

The Rill admin service (`api.rilldata.com`) provides REST endpoints for managing organizations, projects, deployments, billing, and more. Below are key endpoints with examples.

### Authentication

All admin API requests require authentication using a Bearer token:

```bash
curl -H "Authorization: Bearer <your-token>" \
     https://api.rilldata.com/v1/...
```

### Organizations

**List organizations:**
```bash
GET /v1/organizations
```

**Get organization details:**
```bash
GET /v1/organizations/{org-name}
```

**Update organization:**
```bash
PUT /v1/organizations/{org-name}
Content-Type: application/json

{
  "description": "Updated description",
  "name": "new-org-name"
}
```

### Projects

**List projects in an organization:**
```bash
GET /v1/organizations/{org-name}/projects
```

**Get project details:**
```bash
GET /v1/organizations/{org-name}/projects/{project-name}
```

**Create a project:**
```bash
POST /v1/organizations/{org-name}/projects
Content-Type: application/json

{
  "name": "my-project",
  "description": "Project description",
  "public": false,
  "githubUrl": "https://github.com/org/repo"
}
```

**Update project:**
```bash
PUT /v1/organizations/{org-name}/projects/{project-name}
Content-Type: application/json

{
  "description": "Updated description",
  "public": true
}
```

**Delete project:**
```bash
DELETE /v1/organizations/{org-name}/projects/{project-name}
```

### Deployments

**List deployments for a project:**
```bash
GET /v1/organizations/{org-name}/projects/{project-name}/deployments
```

**Get deployment details:**
```bash
GET /v1/deployments/{deployment-id}
```

**Create deployment:**
```bash
POST /v1/organizations/{org-name}/projects/{project-name}/deployments
Content-Type: application/json

{
  "branch": "main",
  "runtimeHost": "https://runtime.rilldata.com",
  "runtimeAudience": "https://runtime.rilldata.com"
}
```

**Delete deployment:**
```bash
DELETE /v1/deployments/{deployment-id}
```

### Billing

**List public billing plans:**
```bash
GET /v1/billing/plans
```

**Get billing project credentials:**
```bash
POST /v1/billing/metrics-project-credentials
Content-Type: application/json

{
  "orgName": "my-org"
}
```

## OpenAPI Integration with Orval

Rill generates OpenAPI specifications for both the admin service and runtime APIs. You can use these specs with tools like Orval to generate type-safe client libraries.

### Admin Service OpenAPI

The admin service OpenAPI spec is available at build time and used internally by Rill's web admin interface. The spec includes endpoints for:

- Organization management
- Project management
- Deployment management
- Billing operations
- User management
- Service account management

### Runtime OpenAPI

Each Rill project exposes a runtime-specific OpenAPI spec that includes:

- Query APIs for metrics views
- Custom APIs defined in your project
- Connector operations
- Export functionality

**Download runtime OpenAPI spec:**
```bash
curl https://api.rilldata.com/v1/organizations/{org-name}/projects/{project-name}/runtime/api/openapi \
  -H "Authorization: Bearer <token>" \
  -o openapi.json
```

### Using Orval for TypeScript Clients

Rill uses Orval internally to generate TypeScript clients from OpenAPI specs. Here's how you can set it up for your own projects:

**Install Orval:**
```bash
npm install --save-dev orval
```

**Create Orval configuration:**
```typescript
// orval.config.ts
import { defineConfig } from "orval";

export default defineConfig({
  runtime: {
    input: {
      target: "https://api.rilldata.com/v1/organizations/{org}/projects/{project}/runtime/api/openapi",
      headers: {
        Authorization: "Bearer YOUR_SERVICE_TOKEN"
      }
    },
    output: {
      target: "./src/client/runtime.ts",
      client: "fetch",
      mode: "tags-split",
      prettier: true,
      override: {
        mutator: {
          path: "./src/client/http-client.ts",
          name: "customInstance"
        }
      }
    }
  }
});
```

**Generate the client:**
```bash
npx orval
```

**Use the generated client:**
```typescript
import { runtimeApi } from "./client/runtime";

// Query a metrics view
const result = await runtimeApi.metricsViews.query({
  metricsView: "my_metrics",
  measures: [{ name: "total_revenue" }],
  dimensions: [{ name: "category" }]
});
```

### Custom HTTP Client

For proper authentication, create a custom HTTP client:

```typescript
// src/client/http-client.ts
import { getToken } from "../auth";

export const customInstance = async (url: string, options: RequestInit) => {
  const token = await getToken();
  
  return fetch(url, {
    ...options,
    headers: {
      ...options.headers,
      Authorization: `Bearer ${token}`,
    },
  });
};
```

This setup provides type-safe access to Rill's APIs with automatic authentication handling.
