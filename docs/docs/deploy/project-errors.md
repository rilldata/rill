---
title: Project Errors
description: Explore and configure alerts for project errors
sidebar_label: Project Errors
sidebar_position: 25
---

## Overview

Rill projects can go into an error state for many reasons, such as a malformed YAML file, missing credentials for a connector, or a breaking change in a data type.
Regardless of the error, Rill Cloud takes several steps to manage and contain errors:

- **Isolation:** It handles errors at the individual resource level. For example, if a dashboard breaks, all other dashboards stay online.
- **Fallback:** It strives to fall back to the most recent valid state when possible. For example, if the model underneath one of your dashboards breaks, the dashboard will keep serving from the most recent valid state.
- **Overview:** You can always view project status at the individual resource level using the "Status" tab in Rill Cloud or using the `rill project status` CLI command.

## Receive alerts for project errors

To help you quickly identify and fix errors, you can configure a Rill alert that will trigger when one or more resources in your project enter an error state. The alert must be configured using a YAML file committed to your Rill project repository (configuration through the UI is not yet possible).

### Configure an email alert

To configure an email alert for project errors, add a file named `project_errors.yaml` to your repository with the contents below. Remember to update the `recipients` field to your desired alert recipients.

```yaml
type: alert

# Check the alert every 10 minutes.
refresh:
  cron: "*/10 * * * *"

# Query for all resources with a reconcile error.
# The alert will trigger when the query result is not empty.
data:
  resource_status:
    where_error: true

# Send notifications by email
notify:
  email:
    recipients: [john@example.com]
```

### Configure a Slack alert

To configure a Slack alert for project errors, first follow the Slack configuration steps described on [Configuring Slack integration](../explore/alerts/slack.md). Then add a file named `project_errors.yaml` to your repository with the contents below. Remember to update the `channels` field to your desired destination channel.

```yaml
type: alert

# Check the alert every 10 minutes.
refresh:
  cron: "*/10 * * * *"

# Query for all resources with a reconcile error.
# The alert will trigger when the query result is not empty.
data:
  resource_status:
    where_error: true

# Send notifications in Slack.
# Follow these steps to configure a Slack token: https://docs.rilldata.com/explore/alerts/slack.
notify:
  slack:
    channels: [rill-alerts]
```
