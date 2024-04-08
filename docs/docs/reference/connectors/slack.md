---
title: Slack
description: Deliver notifications to Slack
sidebar_label: Slack
sidebar_position: 999
---


## Overview

[Slack](https://slack.com/) is a popular messaging platform that allows teams to communicate and collaborate in real-time. 
Rill supports sending notifications to Slack channels using the [Slack API](https://api.slack.com/). 
This can be useful for sending alerts and reports to your team members.

Although Rill does not support reading data from Slack, Slack support is implemented as a connector 
with a corresponding bonus like passing a token as an environment variable.

Slack connector can send notifications to channels, direct messages, and via webhooks. 
All of these methods require a Slack application to be created and configured with the necessary scopes.
Follow the [Slack documentation](https://api.slack.com/start/quickstart) to create a Slack application and configure it 
depending on notification type you want to use: channels, direct messages, or webhooks.

### Slack channels

Sending notifications to a channel requires [`chat:write`](https://api.slack.com/scopes/chat:write) scope. The application needs to be added to the channel.

### Direct messages

Sending notifications to a direct message requires [`chat:write`](https://api.slack.com/scopes/chat:write), [`users:read`](https://api.slack.com/scopes/users:read), and [`users:read.email`](https://api.slack.com/scopes/users:read.email) scopes. 
The last two scopes are required to find the user's ID by email.

### Webhooks

Sending notifications via webhooks requires [`incoming-webhook`](https://api.slack.com/scopes/incoming-webhook) scope only and no bot token is required.  
The application needs to be added to the channel.

## Local credentials
    
When using Rill Developer on your local machine (i.e. `rill start`), Rill expects a [bot token](https://api.slack.com/authentication/token-types#bot) to be passed as follows:
    
```bash
start devproject --env connector.slack.bot_token=xoxb-...
```

## Cloud deployment

Once a project that requires Slack notifications has been deployed using `rill deploy`, Rill Cloud will need to be able to have access to the [bot token](https://api.slack.com/authentication/token-types#bot). This can be done as follows:
    
```bash
rill env configure
```

## Example usage

### Reports
```yaml
...
notify:
  email:
    recipients:
      - recipient@example.com
      ...
  slack:
    users:
      - recipient@example.com
      ...
    channels:
      - alerts
      ...
    webhooks:
      - https://hooks.slack.com/services/...
      ...
```
### Alerts
```yaml
...
notify:
  on_recover: true
  renotify: true
  renotify_after: 24h
  email:
    recipients:
      - recipient@example.com
      ...
  slack:
    users:
      - recipient@example.com
      ...
    channels:
      - alerts
      ...
    webhooks:
      - https://hooks.slack.com/services/...
      ...
```

