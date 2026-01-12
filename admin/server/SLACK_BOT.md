# Slack Bot Integration

The Slack bot is integrated directly into the Rill admin server, allowing users to interact with Rill data through Slack.

## Features

- **Multi-Workspace Support**: Handles multiple Slack workspaces with isolated token storage
- **Multi-Customer Support**: Optional customer grouping via `customer_id` field
- **Threaded Conversations**: Maintains conversation context per Slack thread
- **Direct Messages**: Works in both DMs and channel threads
- **Secure Token Storage**: Tokens are encrypted at rest using the database encryption keyring
- **Streaming Responses**: Real-time streaming of AI responses as they're generated

## Setup

### 1. Create a Slack App

1. Go to [api.slack.com/apps](https://api.slack.com/apps)
2. Click "Create New App" → "From scratch"
3. Name your app (e.g., "Rill Bot") and select your workspace
4. Go to "OAuth & Permissions" in the sidebar
5. Add the following Bot Token Scopes:
   - `app_mentions:read` - To listen for mentions
   - `chat:write` - To send messages
   - `im:history` - To read DM history
   - `im:read` - To read DM metadata
   - `im:write` - To send DMs
   - `channels:history` - To read channel messages
   - `groups:history` - To read private channel messages
   - `commands` - For slash commands
6. Go to "Event Subscriptions" and enable it
7. Set Request URL to: `https://your-admin-server.com/slack/events`
8. Subscribe to bot events:
   - `app_mentions` - When the bot is mentioned
   - `message.im` - Direct messages
   - `message.channels` - Channel messages (for thread replies)
9. Go to "Slash Commands" and create a new command:
   - Command: `/set-token`
   - Request URL: `https://your-admin-server.com/slack/commands/set-token`
   - Short description: "Set your Rill access token"
10. Install the app to your workspace
11. Copy the following tokens:
    - Bot User OAuth Token (starts with `xoxb-`)
    - Signing Secret (from "Basic Information")

### 2. Configure the Admin Server

Add the following environment variables:

```bash
RILL_ADMIN_SLACK_BOT_TOKEN=xoxb-your-bot-token-here
RILL_ADMIN_SLACK_SIGNING_SECRET=your-signing-secret-here
# Optional: If not set, the bot will automatically use the first available org and project
RILL_ADMIN_SLACK_DEFAULT_ORG=your-org-name
RILL_ADMIN_SLACK_DEFAULT_PROJECT=your-project-name
```

### 3. Database Migration

The database migration will run automatically when the admin server starts. The migration creates:
- `slack_workspaces` table
- `slack_user_tokens` table (with encryption)
- `slack_conversations` table

## Usage

### Setting Your Token

Users can set their Rill personal access token in two ways:

1. **Via Slash Command**:
   ```
   /set-token rill_usr_your-token-here
   ```

2. **Via DM**: Simply send your token as a direct message to the bot:
   ```
   rill_usr_your-token-here
   ```

### Using the Bot

#### In Direct Messages

1. Open a DM with the bot
2. Send your question:
   ```
   What were our top products last month?
   ```
3. The bot will respond with streaming results from Rill

#### In Channels

1. Mention the bot in a channel:
   ```
   @Rill Bot what were our sales last week?
   ```
2. The bot will respond in a thread
3. Continue the conversation by replying in the thread

#### Threaded Conversations

- Each Slack thread maintains its own conversation context
- You can ask follow-up questions in the same thread
- Different threads are separate conversations

## Architecture

### Database Schema

- **`slack_workspaces`**: Stores Slack workspace/team information
- **`slack_user_tokens`**: Stores encrypted Rill PATs per user per workspace
- **`slack_conversations`**: Stores conversation state per workspace/thread

### API Endpoints

- **`POST /slack/events`**: Slack Events API webhook
- **`POST /slack/commands/set-token`**: Slash command handler

### Integration Points

- **Database**: Uses existing admin database with encryption
- **Runtime API**: Uses runtime proxy to access Rill chat API
- **Authentication**: Tokens are encrypted using database encryption keyring

## Multi-Workspace Support

The bot automatically:
- Registers workspaces when first encountered
- Isolates tokens per workspace (same user can have different tokens in different workspaces)
- Maintains conversation state per workspace/thread

## Multi-Customer Support

To group workspaces by customer, you can set the `customer_id` field when registering workspaces. This allows you to:
- Query all workspaces for a customer
- Manage workspaces by customer
- Apply customer-specific configurations

## Security

- **Token Encryption**: All Rill PATs are encrypted at rest using the database encryption keyring
- **Signature Verification**: All Slack requests are verified using the signing secret
- **Per-Workspace Isolation**: Tokens and conversations are isolated per workspace
- **Secure Storage**: Uses the same encryption infrastructure as other sensitive data

## Configuration

The bot uses the following configuration (via environment variables):

- `RILL_ADMIN_SLACK_BOT_TOKEN`: Slack bot token (required)
- `RILL_ADMIN_SLACK_SIGNING_SECRET`: Slack signing secret (required)
- `RILL_ADMIN_SLACK_DEFAULT_ORG`: Default organization for Rill queries (required)
- `RILL_ADMIN_SLACK_DEFAULT_PROJECT`: Default project for Rill queries (required)

## Troubleshooting

### Bot Not Responding

1. Check that the admin server is running
2. Verify your Slack tokens are correct in environment variables
3. Check server logs for errors
4. Ensure the bot is installed in your workspace
5. Verify the webhook URL is accessible from Slack

### Authentication Errors

1. Verify your Rill token is valid: `rill token list`
2. Check that your token hasn't expired
3. Ensure you have access to the specified org/project

### API Errors

1. Verify your `SLACK_DEFAULT_ORG` and `SLACK_DEFAULT_PROJECT` are correct
2. Check that the project has a deployment
3. Review the error message in Slack for details

## Future Enhancements

- Per-workspace org/project configuration
- Support for multiple projects per workspace
- Token rotation reminders
- Usage analytics per workspace/customer
