---
title: User Tokens
description: Create and manage personal access tokens for development and scripting
sidebar_label: User Tokens
sidebar_position: 24
---

User tokens (also called personal access tokens or PATs) provide programmatic access to Rill Cloud tied to your personal user account. They inherit your user permissions and are ideal for local development, scripting, and integrations like MCP (Model Context Protocol).

## Overview

User tokens are designed for:
- **Local development** - Testing and developing with Rill APIs from your machine
- **Personal scripts** - Automating personal workflows and data analysis
- **AI integrations** - Connecting AI assistants (Claude Desktop, ChatGPT) via MCP
- **Experimentation** - Trying out Rill APIs without production concerns
- **CLI authentication** - Authenticating Rill CLI commands

### User Tokens vs Service Tokens

| Feature | User Tokens | Service Tokens |
|---------|-------------|---------------|
| **Tied To** | Your user account | Organization |
| **Permissions** | Inherits your user permissions | Assigned roles (org/project) |
| **Persistence** | Revoked if user is removed | Persist after creator is removed |
| **Best For** | Development & personal use | Production systems |
| **Custom Attributes** | User profile attributes | Configurable custom attributes |
| **Recommended Use** | Local scripting, MCP, testing | Backend APIs, scheduled jobs |

:::tip When to use what
- **Development/Testing** → User tokens
- **Production systems** → [Service tokens](/manage/service-tokens)
- **AI assistants (MCP)** → User tokens
- **Embedded dashboards** → Service tokens (to issue ephemeral tokens)
:::

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

### Via the UI

You can also create user tokens through the Rill Cloud UI:

1. Navigate to your project's **AI tab**
2. Click **"Create Token"** or **"Copy MCP Config"**
3. The token will be automatically created and included in the configuration

<img src='/img/explore/mcp/project-ai.png' class='rounded-gif'/>

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

### With MCP (AI Assistants)

User tokens are the primary authentication method for connecting AI assistants to Rill via MCP:

**Automatic setup via UI:**
1. Go to your project's AI tab in Rill Cloud
2. Click "Copy MCP Config" - this creates a token automatically
3. Paste the config into your AI assistant (Claude Desktop, ChatGPT, etc.)

**Manual setup:**
```bash
# Create a token
rill token issue --display-name "Claude Desktop MCP"

# Configure your MCP client with the token
# (see MCP documentation for client-specific setup)
```

For comprehensive MCP setup, see:
- [MCP Server Documentation](/explore/mcp)
- [MCP Setup Guide](/guides/setting-up-mcp)

### With Rill CLI

User tokens can authenticate CLI commands:

```bash
# Set token as environment variable
export RILL_TOKEN="rill_usr_your_token_here"

# Or use the --api-token flag
rill project list --api-token "rill_usr_your_token_here"
```

### With Custom Scripts

Use user tokens in scripts to automate workflows:

```python
import requests

RILL_TOKEN = "rill_usr_your_token_here"
ORG = "my-org"
PROJECT = "my-project"

headers = {
    "Authorization": f"Bearer {RILL_TOKEN}",
    "Content-Type": "application/json"
}

# Query a custom API
response = requests.get(
    f"https://api.rilldata.com/v1/organizations/{ORG}/projects/{PROJECT}/runtime/api/my-api",
    headers=headers
)

data = response.json()
print(data)
```

## Token Permissions

User tokens inherit your personal permissions in Rill Cloud:

### Organization Permissions

Your token has the same organization-level permissions as your user account:
- **Admin** - Full access to organization and all projects
- **Editor** - Can create projects and manage members
- **Viewer** - Read-only access
- **Guest** - Limited access, requires explicit project permissions

### Project Permissions

Your token has the same project-level permissions as your user account:
- **Admin** - Full project access
- **Editor** - Can edit resources and create reports
- **Viewer** - Read-only dashboard access

### Security Policies

When using tokens to access data, security policies are evaluated using your user attributes:
- `{{ .user.email }}` - Your email address
- `{{ .user.domain }}` - Your email domain
- `{{ .user.admin }}` - Whether you're an admin
- `{{ .user.groups }}` - Your user groups
- Custom attributes set by your organization

Example security policy:
```yaml
# In metrics view
security:
  access: "'{{ .user.domain }}' == 'example.com'"
  row_filter: region = '{{ .user.region }}'
```

For more on security policies, see [Data Access Control](/build/metrics-view/security).

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

### AI Assistant Integration (MCP)

Set up AI assistants to query your Rill metrics:

**Via UI (recommended):**
1. Open your project in Rill Cloud
2. Navigate to the AI tab
3. Click "Copy MCP Config"
4. Paste into Claude Desktop, ChatGPT, or other MCP clients

**Via CLI:**
```bash
# Create token for AI assistant
rill token issue --display-name "Claude Desktop" --ttl-minutes 10080

# Configure your MCP client with the token
```

Then ask your AI assistant questions like:
- "What were my top products by revenue last month?"
- "Show me week-over-week growth in user signups"
- "Are there any anomalies in website traffic this week?"

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

## Best Practices

### Security

1. **Use short expirations for testing**
   ```bash
   # Good for temporary testing
   rill token issue --display-name "Test" --ttl-minutes 60
   ```

2. **Never commit tokens to version control**
   ```bash
   # Add to .gitignore
   echo ".env" >> .gitignore
   echo "*.token" >> .gitignore

   # Store in environment variables
   export RILL_TOKEN="rill_usr_..."
   ```

3. **Use descriptive display names**
   ```bash
   # Good: Describes purpose and location
   rill token issue --display-name "MBP - Local Dev"
   rill token issue --display-name "Claude Desktop - Personal"

   # Avoid: Generic names
   rill token issue --display-name "token1"
   ```

4. **Revoke unused tokens regularly**
   ```bash
   # List and review tokens
   rill token list

   # Revoke old tokens
   rill token revoke <token-id>
   ```

5. **Use service tokens for production**
   - Don't use user tokens in production systems
   - User tokens are revoked when user leaves organization
   - Use [service tokens](/manage/service-tokens) instead

### Token Lifecycle Management

**For development:**
- Create one long-lived token per development machine
- Revoke tokens when you stop using a machine
- Rotate tokens every 3-6 months

**For MCP/AI assistants:**
- Create one token per AI assistant
- Set reasonable expiration (7-30 days)
- Regenerate when prompted

**For scripts:**
- Use tokens stored in environment variables or secrets manager
- Set expirations appropriate to script usage frequency
- Document which scripts use which tokens

### Naming Conventions

Use clear, consistent naming:

```bash
# Format: [Device/Tool] - [Purpose]
rill token issue --display-name "MacBook Pro - Local Dev"
rill token issue --display-name "Claude Desktop - MCP"
rill token issue --display-name "Linux Server - Automation Scripts"
rill token issue --display-name "iPad - Mobile Testing"
```

## Troubleshooting

### Token Not Working

If your token isn't working:

1. **Check if token exists**
   ```bash
   rill token list
   ```

2. **Verify token format**
   - User tokens start with `rill_usr_`
   - Should be used in `Authorization: Bearer <token>` header

3. **Check expiration**
   - Look at the "Expires" column in `rill token list`
   - Create a new token if expired

4. **Verify permissions**
   - Check your user role in organization and project
   - Ensure you have necessary permissions for the operation

### Token Expired

If your token has expired:

```bash
# Check expiration
rill token list

# Create new token
rill token issue --display-name "Replacement Token"

# Revoke old token
rill token revoke <old-token-id>
```

### Permission Denied

If you get permission denied errors:

1. **Check your user permissions**
   - Verify your organization role
   - Verify your project role
   - Contact your organization admin if you need elevated access

2. **Check security policies**
   - Security policies may be restricting access
   - Review security policies in metrics views
   - Verify your user attributes match policy requirements

3. **Check project access**
   - Ensure you're a member of the project
   - Guest users need explicit project membership

### MCP Not Connecting

If MCP isn't connecting:

1. **Verify token in config**
   ```json
   {
     "mcpServers": {
       "rill": {
         "command": "npx",
         "args": [
           "-y",
           "@rilldata/rill-mcp-server",
           "--project", "your-project",
           "--token", "rill_usr_your_token_here"
         ]
       }
     }
   }
   ```

2. **Check token hasn't expired**
   ```bash
   rill token list
   ```

3. **Regenerate token if needed**
   - Go to AI tab in Rill Cloud
   - Click "Copy MCP Config" to get fresh config

4. **Restart your MCP client**
   - Close and reopen Claude Desktop/ChatGPT
   - Some clients require restart to pick up config changes

## Comparison: When to Use User vs Service Tokens

### Use User Tokens When:

✅ Developing locally on your machine
✅ Running personal analysis scripts
✅ Connecting AI assistants (Claude, ChatGPT) via MCP
✅ Testing and experimenting with APIs
✅ You want permissions tied to your user account

### Use Service Tokens When:

✅ Building production integrations
✅ Creating scheduled jobs or automation
✅ Building backend APIs that serve multiple users
✅ Embedding dashboards in your application
✅ You need tokens that persist after user changes
✅ You need fine-grained custom attributes for security

**Example scenarios:**

| Scenario | Token Type | Why |
|----------|-----------|-----|
| Local Rill CLI usage | User token | Personal, inherits your permissions |
| Claude Desktop MCP | User token | Personal AI assistant |
| Embedded customer dashboard | Service token | Production system, needs to issue ephemeral tokens |
| Nightly data sync job | Service token | Automated system, should persist |
| Personal Python analysis script | User token | Personal use, testing |
| Backend API serving users | Service token | Production, needs specific permissions |

## Related Topics

- [Service Tokens](/manage/service-tokens) - For production systems and automation
- [Custom API Integration](/integrate/custom-api) - Using tokens with Rill APIs
- [MCP Server](/explore/mcp) - Connecting AI assistants to Rill
- [Data Access Control](/build/metrics-view/security) - How security policies work
- [Roles and Permissions](/manage/roles-permissions) - Understanding permission levels
- [CLI Reference - Token Commands](/reference/cli/token) - Complete CLI command reference
