---
title: User Tokens
description: Create and manage personal access tokens for development and scripting
sidebar_label: User Tokens
sidebar_position: 26
---

User tokens (also called personal access tokens or PATs) provide programmatic access to Rill Cloud tied to your personal user account. They inherit your user permissions and are ideal for local development, scripting, and integrations like MCP (Model Context Protocol).

## Overview

User tokens are designed for:
- **Local development** - Testing and developing with Rill APIs from your machine
- **Personal scripts** - Automating personal workflows and data analysis
- **AI integrations** - Connecting AI assistants (Claude Desktop, ChatGPT) via MCP
- **Experimentation** - Trying out Rill APIs without production concerns
- **CLI authentication** - Authenticating Rill CLI commands


## Creating User Tokens

### Basic Creation

Create a user token with the CLI:

```bash
rill token issue
```

You'll be prompted to provide a display name and optionally set an expiration time.

### With Display Name

Give your token a descriptive name:

```bash
rill token issue --display-name "Local Development"
```

### With Expiration

Set an expiration time in minutes:

```bash
# Expires in 24 hours (1440 minutes)
rill token issue --display-name "MCP Token" --ttl-minutes 1440

# Expires in 7 days (10080 minutes)
rill token issue --display-name "Testing Token" --ttl-minutes 10080
```

:::warning Store tokens securely
User tokens provide access to your data with your permissions. Store them securely and never commit them to version control. Treat them like passwords.
:::

## Managing User Tokens

### Listing Your Tokens

View all your active user tokens:

```bash
rill token list
```

Output:
```
ID                                    DISPLAY NAME           CREATED              EXPIRES
rill_usr_abc123...                   Local Development       2024-01-15 10:30     Never
rill_usr_def456...                   MCP Token              2024-01-16 14:20     2024-01-17 14:20
rill_usr_ghi789...                   Testing Token          2024-01-14 09:00     2024-01-21 09:00
```

### Revoking Tokens

Revoke a token by its ID:

```bash
rill token revoke <token-id>
```

Or revoke by display name:

```bash
rill token revoke --display-name "Local Development"
```

:::tip Token rotation
For security, periodically rotate your tokens by creating new ones and revoking old ones. This is especially important for long-lived tokens.
:::

## Using User Tokens

### With Custom APIs

User tokens can be used to authenticate requests to Rill's custom APIs:

```bash
curl https://api.rilldata.com/v1/organizations/<org>/projects/<project>/runtime/api/<api-name> \
  -H "Authorization: Bearer <user-token>"
```

For more details, see [Custom API Integration](/integrate/custom-api).

## Token Permissions

Service tokens can be assigned roles at both the organization and project levels. For more details on using attributes with security policies, see the [security policies](/build/metrics-view/security) documentation.

## Use Cases

### Local Development

Create a long-lived token for local development:

```bash
# Create token without expiration
rill token issue --display-name "Dev Machine"

# Store in environment variable
echo 'export RILL_TOKEN="rill_usr_..."' >> ~/.bashrc
source ~/.bashrc

# Use in your development workflow
rill project list
```

### Personal Data Analysis Scripts

Automate data analysis workflows:

```bash
#!/bin/bash
# daily-report.sh

RILL_TOKEN="rill_usr_..."
ORG="my-org"
PROJECT="analytics"

# Fetch data from custom API
curl -s "https://api.rilldata.com/v1/organizations/$ORG/projects/$PROJECT/runtime/api/daily-metrics" \
  -H "Authorization: Bearer $RILL_TOKEN" \
  | jq '.results[] | {date: .date, revenue: .revenue, users: .users}'
```

### Temporary Access

Create short-lived tokens for temporary testing:

```bash
# 2-hour token for testing
rill token issue --display-name "Quick Test" --ttl-minutes 120
```