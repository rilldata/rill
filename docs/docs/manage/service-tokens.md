---
title: Service Tokens
description: Create and manage service tokens for programmatic access to Rill
sidebar_label: Service Tokens
sidebar_position: 25
---

Service tokens (also called service accounts) provide programmatic access to Rill Cloud for production systems, scheduled jobs, backend APIs, and other automated workflows. Unlike user tokens, service tokens persist even if the creating user is removed from the organization.

## Overview

Service tokens are designed for:
- **Production integrations** - Backend services that need to access Rill APIs
- **Scheduled jobs** - Automated reports, data syncs, or ETL processes
- **CI/CD pipelines** - Automated testing and deployment workflows
- **Custom applications** - Applications that integrate with Rill APIs
- **Embedded analytics** - Issuing ephemeral tokens for end users

### Service Tokens vs User Tokens

| Feature | Service Tokens | User Tokens |
|---------|---------------|-------------|
| **Persistence** | Persist after user removal | Tied to user account |
| **Recommended Use** | Production systems | Local development |
| **Permissions** | Org and/or project roles | Inherits user permissions |
| **Custom Attributes** | Supported | N/A |
| **Security Policies** | Evaluated with attributes | Evaluated with user profile |

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

Service tokens can be assigned roles at both the organization and project levels.

### Organization Roles

Organization-level roles grant permissions across the entire organization:

- **admin** - Full access to all organization resources and projects
- **editor** - Can create projects and manage organization members
- **viewer** - Read-only access to organization resources
- **guest** - Limited access, requires explicit project permissions

```bash
rill service create global-admin \
  --org-role admin
```

### Project Roles

Project-level roles grant permissions to specific projects:

- **admin** - Full access to the project
- **editor** - Can edit project resources and create reports
- **viewer** - Read-only access to project dashboards

```bash
rill service create dashboard-viewer \
  --project sales-analytics \
  --project-role viewer
```

### Combined Roles

You can assign both organization and project roles:

```bash
rill service create hybrid-service \
  --org-role viewer \
  --project analytics-dashboard \
  --project-role admin \
  --attributes '{"scope":"analytics-only"}'
```

For more details on permissions, see [Roles and Permissions](/manage/roles-permissions).

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

:::tip Token rotation
For security, periodically rotate service tokens by creating a new token, updating your applications, and deleting the old token.
:::

## Using Custom Attributes with Security Policies

Custom attributes are most powerful when combined with security policies to implement fine-grained access control.

### Example: Multi-Tenant Access

**Service token with customer attribute:**
```bash
rill service create customer-api \
  --project-role viewer \
  --attributes '{"customer_id":"acme-corp"}'
```

**Security policy in metrics view:**
```yaml
# metrics_view.yaml
security:
  access: true  # Allow authenticated access
  row_filter: customer_id = '{{ .user.customer_id }}'
```

This ensures the service token can only access data for `customer_id = 'acme-corp'`.

### Example: Regional Access

**Service token with region attribute:**
```bash
rill service create regional-service \
  --project-role viewer \
  --attributes '{"region":"us-west","allowed_regions":"us-west,us-east"}'
```

**Security policy:**
```yaml
security:
  row_filter: region IN ({{ .user.allowed_regions }})
```

### Example: Department-Based Access

**Service token with department attribute:**
```bash
rill service create dept-service \
  --project-role viewer \
  --attributes '{"department":"engineering"}'
```

**Security policy:**
```yaml
security:
  row_filter: department = '{{ .user.department }}'

  field_access:
    - if: '{{ eq .user.department "engineering" }}'
      names: [salary, bonus]
      allow: true
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

For more details, see [Custom API Integration](/integrate/custom-api#ephemeral-user-tokens).

## Best Practices

### Security

1. **Use least privilege** - Grant only the minimum permissions needed
   ```bash
   # Good: Project-specific viewer role
   rill service create reader --project my-project --project-role viewer

   # Avoid: Org-wide admin unless truly needed
   rill service create admin --org-role admin
   ```

2. **Store tokens securely** - Never commit tokens to version control
   - Use environment variables: `RILL_TOKEN=rill_svc_...`
   - Use secrets managers: AWS Secrets Manager, Vault, etc.
   - Rotate tokens regularly

3. **Use custom attributes** - Leverage attributes for fine-grained access control
   ```bash
   rill service create api-service \
     --project-role viewer \
     --attributes '{"environment":"production","customer_id":"${CUSTOMER_ID}"}'
   ```

4. **Monitor token usage** - Track which services are accessing your data
   - Review service token list regularly
   - Remove unused tokens
   - Update attributes as requirements change

### Naming Conventions

Use descriptive names that indicate purpose:

```bash
# Good names
rill service create prod-reporting-api
rill service create staging-etl-pipeline
rill service create customer-dashboard-embedder

# Avoid generic names
rill service create service1
rill service create token
```

### Attribute Design

Design attributes to match your security requirements:

```json
{
  "environment": "production",
  "purpose": "reporting",
  "customer_id": "acme-corp",
  "region": "us-west",
  "tier": "premium",
  "department": "engineering",
  "created_by": "devops-team",
  "expires_on": "2024-12-31"
}
```

Keep attribute names consistent across your organization for easier security policy management.

## Troubleshooting

### Service Token Has No Access

If a service token can't access resources:

1. **Check roles**: Ensure the service has appropriate org or project roles
   ```bash
   rill service show my-service
   ```

2. **Verify security policies**: Check if security policies are blocking access
   - Review `security:` section in metrics views
   - Ensure required attributes are present
   - Test with `{{ .user.admin }}` to check if it's a permission issue

3. **Check token validity**: Ensure the token hasn't been revoked
   ```bash
   rill service list
   ```

### Attributes Not Working in Security Policies

If custom attributes aren't being evaluated:

1. **Verify attribute syntax**: Ensure attributes are valid JSON
   ```bash
   rill service show my-service
   ```

2. **Check security policy syntax**: Use correct templating
   ```yaml
   # Correct
   row_filter: customer_id = '{{ .user.customer_id }}'

   # Incorrect (missing quotes)
   row_filter: customer_id = {{ .user.customer_id }}
   ```

3. **Provide fallback values**: Handle missing attributes gracefully
   ```yaml
   row_filter: customer_id = '{{ .user.customer_id | default "none" }}'
   ```

## Related Topics

- [Roles and Permissions](/manage/roles-permissions) - Understanding role-based access control
- [Security Policies](/build/metrics-view/security) - Implementing fine-grained access control
- [Custom API Integration](/integrate/custom-api) - Using service tokens in your applications
- [CLI Reference - Service Commands](/reference/cli/service) - Complete CLI command reference
