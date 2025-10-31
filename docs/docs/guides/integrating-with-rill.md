---
title: "Integrating with Rill's Admin API"
sidebar_label: "Integrating with Rill's Admin API"
sidebar_position: 30
hide_table_of_contents: false
tags:
  - Tutorial
  - Quickstart
  - Example Project
---

# Integrating with Rill's Admin API

Rill's Admin API allows you to programmatically manage your Rill organization, including workspaces, users, service accounts, and permissions. This guide will walk you through the steps to integrate with Rill using the Admin API. 

> In this scenario, we will generate an admin service account and use it to make requests to the Admin API to provision a new 'viewer' user with custom attributes. This can be useful for automating temporary viewer user provisioning in your organization.

## Prerequisites

This guide assumes you have completed the following prerequisites:

1. Installed the Rill developer CLI - see [Install Rill Developer](/get-started/install)
2. A Rill Cloud organization and workspace - see [Rill Quickstart Guide](https://docs.rilldata.com/quickstart)
3. Administrative access to your Rill organization

:::info
This guide can be completed with user tokens. These tokens are tied to your personal user and access permissions. They are useful for local scripting and experimentation. We don't recommend for production use.
:::

- [Rill Admin OpenAPI Specification](/api/admin)
- [Service Account CLI Reference](/reference/cli/service)
- [Custom API Integration Guide](/integrate/custom-api)

## Example: Managing Users with a Rill Service Account

This example demonstrates how to create a service account, use it to manage users with custom attributes, and then clean up all resources. This is useful for automating user provisioning and management in your Rill organization.

### Step 1: Create a Service Account

First, create a service account with organization admin permissions and custom attributes:

```bash
# Create a service account with custom attributes
rill service create user-management-service \
  --org-role admin \
  --attributes '{"department": "engineering", "environment": "production"}'

Service created: user-management-service
Token: rill_svc_A1B2C3D4E5F6G7H8I9J0K1L2M3N4O5P6Q7R8S9T0U1V2W3X4Y5Z6
```

### Step 2: Use Service Account to Add a User with Custom Attributes

Now use the service account token to add a user to your organization:

```bash
# Add a user to the organization using the service account
curl -X POST "https://admin.rilldata.com/v1/orgs/example/members" \
     -H "Authorization: Bearer rill_svc_A1B2C3D4E5F6G7H8I9J0K1L2M3N4O5P6Q7R8S9T0U1V2W3X4Y5Z6" \
     -H "Content-Type: application/json" \
     -d '{
           "email": "newuser@example.com",
           "role": "viewer",
         }'
```

**Expected Response:**
```json
{
  "pendingSignup": true
}
```

The user will receive an email invitation to join the Rill organization as a viewer.

### Step 4: Verify User Access and Attributes

Check that the user was added successfully:

```bash
# List organization members to verify the user was added
curl -X GET "https://admin.rilldata.com/v1/orgs/example/members" \
     -H "Authorization: Bearer rill_svc_A1B2C3D4E5F6G7H8I9J0K1L2M3N4O5P6Q7R8S9T0U1V2W3X4Y5Z6"
```

### Step 5: Clean Up Resources

Remove the user and service account when done:

```bash
# Remove the user from the organization
curl -X DELETE "https://admin.rilldata.com/v1/orgs/example/members/newuser@example.com" \
     -H "Authorization: Bearer rill_svc_A1B2C3D4E5F6G7H8I9J0K1L2M3N4O5P6Q7R8S9T0U1V2W3X4Y5Z6"

# Revoke the service account token
curl -X DELETE "https://admin.rilldata.com/v1/services/tokens/rill_svc_A1B2C3D4E5F6G7H8I9J0K1L2M3N4O5P6Q7R8S9T0U1V2W3X4Y5Z6" \
     -H "Authorization: Bearer rill_svc_A1B2C3D4E5F6G7H8I9J0K1L2M3N4O5P6Q7R8S9T0U1V2W3X4Y5Z6"

# Delete the service account
rill service remove user-management-service
```

## Service Account API Access

Service accounts have access to all Admin API endpoints based on their assigned roles and permissions. Key endpoints include:

- **Organization Management**: Add/remove users, manage roles and permissions
- **Project Management**: Create, update, and delete projects  
- **User Management**: Provision users, manage attributes, issue tokens
- **Service Management**: Create and manage other service accounts
- **Deployment Management**: Start, stop, and configure deployments
- **Token Management**: Issue ephemeral user tokens with custom attributes

## Service Account Best Practices

- _Secure Storage of Credentials_: Store service account credentials securely, using encrypted storage solutions and access controls to prevent unauthorized access.
- _Regular Rotation of Credentials_: Regularly update service account passwords and keys to reduce the risk of compromise.
- _Minimum Necessary Permissions_: Grant only the permissions necessary for the specific tasks the service account needs to perform, and review permissions regularly to adapt to changes in application functionality.
- _Monitoring and Logging_: Depending on your organization's security needs, you can also consider implementing monitoring and logging of all access and actions taken by service accounts to detect and respond to anomalous activities promptly.
- _Use Custom Attributes_: Leverage custom attributes on service accounts to implement fine-grained access controls and security policies.
- _Token Lifecycle Management_: Implement proper token lifecycle management, including regular rotation and revocation of unused tokens.
