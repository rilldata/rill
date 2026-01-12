# Testing the Slack Bot Locally

This guide covers several approaches to test the Slack bot while running Rill Cloud locally.

## Prerequisites

1. Start your local Rill Cloud environment:
   ```bash
   rill devtool start cloud
   ```

2. Ensure the admin server is running (if not started by devtool):
   ```bash
   go run ./cli admin start
   ```

   The admin server should be available at `http://localhost:8080`

3. Set up your Slack app (see `SLACK_BOT.md` for full setup instructions)

## Approach 1: ngrok (Recommended for Quick Testing)

**Best for**: Quick local testing with minimal setup

### Setup

1. Install ngrok:
   ```bash
   # macOS
   brew install ngrok
   
   # Or download from https://ngrok.com/download
   ```

2. Start ngrok tunnel:
   ```bash
   ngrok http 8080
   ```

3. Copy the HTTPS URL (e.g., `https://abc123.ngrok.io`)

4. Update your Slack app configuration:
   - **Event Subscriptions Request URL**: `https://abc123.ngrok.io/slack/events`
   - **Slash Command Request URL**: `https://abc123.ngrok.io/slack/commands/set-token`

5. Set environment variables for the admin server:
   ```bash
   export RILL_ADMIN_SLACK_BOT_TOKEN="xoxb-your-bot-token"
   export RILL_ADMIN_SLACK_SIGNING_SECRET="your-signing-secret"
   export RILL_ADMIN_SLACK_DEFAULT_ORG="your-org"
   export RILL_ADMIN_SLACK_DEFAULT_PROJECT="your-project"
   ```

6. Restart the admin server with these environment variables

### Testing

1. In Slack, send a DM to your bot or mention it in a channel
2. The bot should respond (you may need to set a token first with `/set-token`)

### Pros
- ✅ Quick setup
- ✅ Works immediately
- ✅ Free tier available

### Cons
- ❌ URL changes on each restart (unless using paid plan)
- ❌ Requires internet connection
- ❌ Less secure (public URL)

---

## Approach 2: Cloudflare Tunnel (cloudflared)

**Best for**: More stable URLs, better for longer testing sessions

### Setup

1. Install cloudflared:
   ```bash
   # macOS
   brew install cloudflared
   
   # Or download from https://developers.cloudflare.com/cloudflare-one/connections/connect-apps/install-and-setup/installation/
   ```

2. Start tunnel:
   ```bash
   cloudflared tunnel --url http://localhost:8080
   ```

3. Use the provided HTTPS URL in your Slack app configuration (same as ngrok)

### Pros
- ✅ More stable URLs
- ✅ Free
- ✅ Better performance

### Cons
- ❌ Still requires internet
- ❌ URL changes on restart

---

## Approach 3: Slack Socket Mode (No Public URL Required)

**Best for**: Development without exposing localhost, more secure

This approach uses Slack's Socket Mode, which establishes a WebSocket connection instead of requiring webhooks.

### Implementation Changes Needed

You would need to modify the bot to support Socket Mode. Here's a high-level approach:

1. Add Socket Mode support to `slack.go`:
   ```go
   import "github.com/slack-go/slack/socketmode"
   
   // In registerSlackEndpoints or a new function
   socketClient := socketmode.New(
       slack.New(s.opts.SlackBotToken),
       socketmode.OptionLog(log.New(os.Stdout, "socketmode: ", log.Lshortfile|log.LstdFlags)),
   )
   
   // Handle events via socket instead of webhook
   go func() {
       for evt := range socketClient.Events {
           switch evt.Type {
           case socketmode.EventTypeEventsAPI:
               // Handle like webhook events
           case socketmode.EventTypeSlashCommand:
               // Handle slash commands
           }
       }
   }()
   
   socketClient.Run()
   ```

2. Enable Socket Mode in your Slack app:
   - Go to "Socket Mode" in Slack app settings
   - Enable Socket Mode
   - Create an app-level token with `connections:write` scope
   - Use this token instead of webhook URLs

### Pros
- ✅ No public URL needed
- ✅ More secure
- ✅ Works behind firewalls
- ✅ No tunnel setup required

### Cons
- ❌ Requires code changes
- ❌ Different from production webhook approach
- ❌ WebSocket connection management

---

## Approach 4: Local Testing Script

**Best for**: Quick manual testing, debugging

Create a simple script to send test events to your local server:

### Create `test-slack-bot.sh`

```bash
#!/bin/bash

# Test Slack bot locally by sending mock events
# Usage: ./test-slack-bot.sh

ADMIN_URL="http://localhost:8080"
SIGNING_SECRET="your-signing-secret"

# Generate timestamp
TIMESTAMP=$(date +%s)

# Test event payload (app_mention)
EVENT_PAYLOAD='{
  "token": "test-token",
  "team_id": "T123456",
  "api_app_id": "A123456",
  "event": {
    "type": "app_mention",
    "user": "U123456",
    "text": "<@U123456> what is the total revenue?",
    "ts": "1234567890.123456",
    "channel": "C123456",
    "event_ts": "1234567890.123456"
  },
  "type": "event_callback",
  "event_id": "Ev123456",
  "event_time": 1234567890
}'

# Note: You'll need to generate a valid Slack signature
# This is a simplified example - in production, Slack signs requests
curl -X POST "$ADMIN_URL/slack/events" \
  -H "Content-Type: application/json" \
  -H "X-Slack-Signature: v0=..." \
  -H "X-Slack-Request-Timestamp: $TIMESTAMP" \
  -d "$EVENT_PAYLOAD"
```

**Note**: This approach is limited because Slack's signature verification requires proper signing, which is complex to mock.

---

## Approach 5: Integration Tests

**Best for**: Automated testing, CI/CD

Create integration tests that mock Slack events:

### Example Test File: `admin/server/slack_test.go`

```go
package server

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    
    "github.com/slack-go/slack/slackevents"
    "github.com/stretchr/testify/require"
)

func TestSlackWebhook(t *testing.T) {
    // Setup test server
    s := setupTestServer(t)
    
    // Create test event
    event := slackevents.EventsAPIEvent{
        Type: "event_callback",
        InnerEvent: &slackevents.EventsAPIInnerEvent{
            Type: "app_mention",
            Data: &slackevents.AppMentionEvent{
                Type:      "app_mention",
                User:      "U123456",
                Text:      "<@U123456> test message",
                Channel:   "C123456",
                TimeStamp: "1234567890.123456",
            },
        },
    }
    
    body, _ := json.Marshal(event)
    req := httptest.NewRequest(http.MethodPost, "/slack/events", bytes.NewReader(body))
    // Add proper Slack signature headers
    // ...
    
    w := httptest.NewRecorder()
    s.slackWebhook(w, req)
    
    require.Equal(t, http.StatusOK, w.Code)
}
```

---

## Recommended Workflow

For local development, I recommend:

1. **Start with ngrok** for quick testing:
   ```bash
   # Terminal 1: Start Rill Cloud
   rill devtool start cloud
   
   # Terminal 2: Start ngrok
   ngrok http 8080
   # Copy the HTTPS URL (e.g., https://abc123.ngrok.io)
   
   # Terminal 3: Start admin server with Slack config
   export RILL_ADMIN_SLACK_BOT_TOKEN="xoxb-..."
   export RILL_ADMIN_SLACK_SIGNING_SECRET="..."
   # Optional: If not set, bot will use first available org/project
   export RILL_ADMIN_SLACK_DEFAULT_ORG="your-org"
   export RILL_ADMIN_SLACK_DEFAULT_PROJECT="your-project"
   # Optional: Set external URL to ngrok URL (for any URLs the bot might generate)
   export RILL_ADMIN_EXTERNAL_URL="https://abc123.ngrok.io"
   go run ./cli admin start
   ```

2. **Update Slack app URLs** with the ngrok URL:
   - Event Subscriptions: `https://abc123.ngrok.io/slack/events`
   - Slash Command: `https://abc123.ngrok.io/slack/commands/set-token`

3. **Test the bot** in Slack:
   - Send a DM to the bot or mention it in a channel
   - Use `/set-token <your-rill-token>` to set your token
   - Ask questions about your Rill data

4. **For production-like testing**, consider implementing Socket Mode support

### Quick Start Script

Create a file `test-slack-bot.sh`:

```bash
#!/bin/bash
set -e

# Configuration
NGROK_PORT=8080
SLACK_BOT_TOKEN="${RILL_ADMIN_SLACK_BOT_TOKEN:-}"
SLACK_SIGNING_SECRET="${RILL_ADMIN_SLACK_SIGNING_SECRET:-}"
SLACK_DEFAULT_ORG="${RILL_ADMIN_SLACK_DEFAULT_ORG:-}"
SLACK_DEFAULT_PROJECT="${RILL_ADMIN_SLACK_DEFAULT_PROJECT:-}"

if [ -z "$SLACK_BOT_TOKEN" ] || [ -z "$SLACK_SIGNING_SECRET" ]; then
    echo "Error: RILL_ADMIN_SLACK_BOT_TOKEN and RILL_ADMIN_SLACK_SIGNING_SECRET must be set"
    exit 1
fi

echo "Starting ngrok on port $NGROK_PORT..."
ngrok http $NGROK_PORT > /tmp/ngrok.log 2>&1 &
NGROK_PID=$!

# Wait for ngrok to start
sleep 3

# Get ngrok URL (this is a simplified approach - you may need to adjust)
NGROK_URL=$(curl -s http://localhost:4040/api/tunnels | grep -o 'https://[^"]*\.ngrok\.io' | head -1)

if [ -z "$NGROK_URL" ]; then
    echo "Error: Could not get ngrok URL. Check ngrok is running."
    kill $NGROK_PID 2>/dev/null
    exit 1
fi

echo "✓ Ngrok URL: $NGROK_URL"
echo ""
echo "Update your Slack app with these URLs:"
echo "  Event Subscriptions: $NGROK_URL/slack/events"
echo "  Slash Command: $NGROK_URL/slack/commands/set-token"
echo ""
echo "Press Enter when ready to start the admin server..."
read

# Start admin server
export RILL_ADMIN_SLACK_BOT_TOKEN="$SLACK_BOT_TOKEN"
export RILL_ADMIN_SLACK_SIGNING_SECRET="$SLACK_SIGNING_SECRET"
export RILL_ADMIN_SLACK_DEFAULT_ORG="$SLACK_DEFAULT_ORG"
export RILL_ADMIN_SLACK_DEFAULT_PROJECT="$SLACK_DEFAULT_PROJECT"
export RILL_ADMIN_EXTERNAL_URL="$NGROK_URL"

echo "Starting admin server..."
go run ./cli admin start

# Cleanup
echo "Stopping ngrok..."
kill $NGROK_PID 2>/dev/null
```

Make it executable and run:
```bash
chmod +x test-slack-bot.sh
./test-slack-bot.sh
```

---

## Troubleshooting

### Bot not responding

1. Check admin server logs for errors
2. Verify environment variables are set correctly
3. Check Slack app event subscriptions are enabled
4. Verify the ngrok/tunnel URL is accessible

### Signature verification errors

- Ensure `RILL_ADMIN_SLACK_SIGNING_SECRET` matches your Slack app's signing secret
- Check that the request is coming from Slack (not a direct curl)

### Database errors

- Ensure PostgreSQL is running: `docker ps | grep postgres`
- Check database migrations are applied
- Verify database connection string in your `.env`

### Rate limiting

- The bot implements rate limiting (10 requests/minute per user)
- If testing frequently, you may hit rate limits
- Check logs for rate limit messages

---

## Environment Variables Reference

```bash
# Required for Slack bot
RILL_ADMIN_SLACK_BOT_TOKEN=xoxb-your-bot-token
RILL_ADMIN_SLACK_SIGNING_SECRET=your-signing-secret

# Optional: If not set, bot will automatically use the first available org and project
RILL_ADMIN_SLACK_DEFAULT_ORG=your-org-name
RILL_ADMIN_SLACK_DEFAULT_PROJECT=your-project-name

# Optional: Override default URLs (if using tunnel)
RILL_ADMIN_EXTERNAL_URL=https://your-ngrok-url.ngrok.io
```

---

## Next Steps

- See `SLACK_BOT.md` for production deployment
- Check `admin/server/slack.go` for implementation details
- Review Slack API documentation: https://api.slack.com/apis
