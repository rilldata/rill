---
title: Slack
description: Connect to Slack for data extraction and analytics
sidebar_label: Slack
sidebar_position: 70
---

<!-- WARNING: There are links to this page in source code. If you move it, find and replace the links and consider adding a redirect in docusaurus.config.js. -->

## Overview

[Slack](https://slack.com/) is a popular messaging platform that allows teams to communicate and collaborate in real-time. 
Rill supports sending notifications to Slack channels using the [Slack API](https://api.slack.com/). 
This can be useful for sending alerts and reports to your team members. 

## Setting up the Slack integration

Rill Cloud can send alert notifications to channels and/or as direct messages. This will require a Slack application to first be created and configured in your workspace with the necessary [permission scopes](https://api.slack.com/scopes). To set up your Slack application, follow the steps provided in the [Slack documentation](https://api.slack.com/start/quickstart) and configure your app within the appropriate permissions depending on the notification type that you wish to use (see below).




### Slack channels

Sending notifications to a specific channel (public or private) requires the [`chat:write`](https://api.slack.com/scopes/chat:write) scope.

:::info
The application will also need to be added to the channel for the notification to be sent.
:::

### Direct messages

Sending notifications via a direct message requires the [`chat:write`](https://api.slack.com/scopes/chat:write), [`users:read`](https://api.slack.com/scopes/users:read), and [`users:read.email`](https://api.slack.com/scopes/users:read.email) scopes. 

:::tip
The last two scopes are required to find the user's ID by email.
:::

## Enabling the Slack integration in your project

Once the Slack integration has been set up, the Slack destination will need to be enabled on a per project basis (note - alerts can only be configured on projects deployed to Rill Cloud). This requires the `connector.slack.bot_token` connector variable to be set, which can be configured in Rill in a manner very similar to [setting credentials](/deploy/deploy-credentials#configure-environmental-variables-and-credentials-for-rill-cloud) for other connectors. Please use one of the available options below.

### Creating a Slack.yaml connector

Please refer to our [connector YAML reference documentation](/reference/project-files/connectors#slack) for more details. 

### Updating the `.env` file directly

Within your project's `.env` file (i.e. `<RILL_PROJECT_HOME>/.env`), you can set this connector variable with the Slack Bot User OAuth Token:

```shell
connector.slack.bot_token=<BOT_USER_OAUTH_TOKEN>
```


Afterwards, if the project has already been deployed to Rill Cloud, you can `rill env push` to update your cloud deployment accordingly.

### Using the `rill env set` command

Another option to set this connector variable within your project is to use the `rill env set` command, i.e.:

```shell
rill env set connector.slack.bot_token <BOT_USER_OAUTH_TOKEN>
```

Afterwards, if the project has already been deployed to Rill Cloud, you can `rill env push` to update your cloud deployment accordingly.

### Enabling the Slack connector through `rill.yaml`

You can enable the Slack "connector" within your project by updating your project's `rill.yaml` file with the following configuration:

```yaml
# Rest of your rill.yaml contents
connectors:
- name: slack
  type: slack
```

Afterwards, when you next deploy the project, you will be prompted to set your Slack Bot User OAuth Token via `rill env configure`.

