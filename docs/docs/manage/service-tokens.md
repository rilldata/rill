---
title: Service Tokens
description: Create and manage service tokens for programmatic access to Rill
sidebar_label: Service Tokens
sidebar_position: 27
---

Service tokens (also called service accounts) provide programmatic access to Rill Cloud for production systems, scheduled jobs, backend APIs, and other automated workflows. Unlike [user tokens](/manage/user-tokens), service tokens persist even if the creating user is removed from the organization.

## Overview

Service tokens are designed for:
- **Production integrations** - Backend services that need to access Rill APIs
- **Scheduled jobs** - Automated reports, data syncs, or ETL processes
- **CI/CD pipelines** - Automated testing and deployment workflows
- **Custom applications** - Applications that integrate with Rill APIs


## Creating Service Tokens

### Basic Creation

Create a service token with an organization-level role:

```bash
rill service create my-service --org-role admin
```

Or with a project-level role:

```bash
rill service create my-service --project my-project --project-role viewer
```

### With Custom Attributes

Custom attributes allow you to pass metadata that can be used in [security policies](/build/metrics-view/security). This is particularly useful for multi-tenant applications or when you need fine-grained access control.

```bash
rill service create my-service \
  --org-role admin \
  --attributes '{"department":"engineering","region":"us-west","tier":"premium"}'
```

**Example attributes:**
- `department` - Organizational department (engineering, sales, finance)
- `region` - Geographic region (us-west, eu-central, ap-south)
- `customer_id` - Customer identifier for multi-tenant systems
- `tier` - Service tier (free, premium, enterprise)
- `environment` - Deployment environment (production, staging)

### Complete Example

```bash
# Create a service for automated reporting
rill service create reporting-service \
  --org my-org \
  --project analytics-dashboard \
  --project-role viewer \
  --attributes '{"purpose":"reporting","schedule":"daily","region":"us-east"}'
```

The command will output:
```
Created service "reporting-service" in org "my-org".
Access token: rill_svc_[TOKEN_HERE]
```

:::warning Store tokens securely
Service tokens have powerful permissions. Store them securely in a secrets manager (AWS Secrets Manager, HashiCorp Vault, etc.) and never commit them to version control.
:::

## Roles and Permissions

Service tokens can be assigned roles at both the organization and project levels. For more details on using attributes with security policies, see the [security policies](/build/metrics-view/security) documentation.


## Managing Service Tokens

### Listing Service Tokens

View all service tokens in your organization:

```bash
rill service list --org my-org
```

Output:
```
NAME                 ORG ROLE    PROJECT ROLES           ATTRIBUTES
reporting-service    -           analytics-dashboard     {"purpose":"reporting"}
global-admin         admin       -                       {}
dashboard-viewer     -           sales-analytics         {}
```

### Editing Service Tokens

Update a service token's name or attributes:

```bash
# Change the name
rill service edit my-service --new-name renamed-service

# Update attributes
rill service edit my-service \
  --attributes '{"department":"finance","region":"eu-west"}'

# Clear attributes
rill service edit my-service --attributes '{}'
```

### Managing Roles

Update organization role:

```bash
rill service set-role my-service --org-role editor
```

Add or update project role:

```bash
rill service set-role my-service \
  --project analytics-dashboard \
  --project-role admin
```

Remove project role:

```bash
rill service set-role my-service \
  --project analytics-dashboard \
  --remove
```

### Viewing Service Details

Get detailed information about a service:

```bash
rill service show my-service --org my-org
```

### Revoking Access

Delete a service token to immediately revoke its access:

```bash
rill service delete my-service --org my-org
```

## Issuing Ephemeral Tokens

Service tokens can issue short-lived ephemeral tokens for end users. This is useful for:
- Embedded dashboards
- Temporary API access
- User-specific data access

### With Custom User Attributes

```bash
curl -X POST https://api.rilldata.com/v1/orgs/<org-name>/projects/<project-name>/credentials \
  -H "Authorization: Bearer <service-token>" \
  -H "Content-Type: application/json" \
  -d '{
    "attributes": {
      "email": "user@example.com",
      "department": "sales",
      "region": "us-west",
      "customer_id": "acme-corp"
    },
    "ttl_seconds": 3600
  }'
```

### With User Email

For simpler cases, you can just provide the user's email:

```bash
curl -X POST https://api.rilldata.com/v1/orgs/<org-name>/projects/<project-name>/credentials \
  -H "Authorization: Bearer <service-token>" \
  -H "Content-Type: application/json" \
  -d '{
    "user_email": "user@example.com",
    "ttl_seconds": 3600
  }'
```

The response contains a short-lived JWT token that can be used to access Rill APIs on behalf of the user.
