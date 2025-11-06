---
title: Managing Project Errors
description: Explore and configure alerts for project errors
sidebar_label: Managing Project Errors
sidebar_position: 25
---

## Overview

Rill projects can go into an error state for many reasons, such as a malformed YAML file, missing credentials for a connector, or a breaking change in a data type.
Regardless of the error, Rill Cloud takes various steps to surface, manage, and contain errors:

- **Visibility:** Admins will always be able to view the project status at the individual resource level within Rill Cloud using the `Status` tab or by using the [rill project status](/reference/cli/project/status) CLI command.
- **Isolation:** Rill Cloud will handle errors at the individual resource level. For example, if a dashboard falls into an error state or fails to reconcile, all other dashboards should remain available. 
- **Fallback:** Rill Cloud will attempt to fall back to the most recent valid state when possible. For example, if the underlying model for a dashboard fails to build, the dashboard will keep serving from the most recent valid state.

## Receive alerts for project errors

To help you quickly identify and fix errors, you can configure a Rill alert that will trigger when one or more resources in your project enter an error state. The alert must be configured using a YAML file committed to your Rill project repository (configuration through the UI is not yet possible).

:::tip Want to learn more about alerts?

Besides alerting on project errors, it is possible to configure generic alerts in your dashboards based on specific thresholds or conditions being met. For more details, check out our [alerts documentation](/explore/alerts)!

:::

### Configure an email alert

To configure an email alert for project errors, add a file named `project_errors.yaml` to your Rill project with the contents below. Remember to update the `recipients` field to your desired alert recipients.

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

:::info Don't forget to commit your changes!

After making these changes, you should commit and [push these changes](/deploy/deploy-dashboard/github-101#pushing-changes) to your git repository.

:::

### Configure a Slack alert

To configure a Slack alert for project errors, first follow the Slack configuration steps described on [Configuring Slack integration](../explore/alerts/slack). Next, add a file named `project_errors.yaml` to your Rill project with the contents below. Remember to update the `channels` field to your desired destination channel.

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

:::info Don't forget to commit your changes!

After making these changes, you should commit and [push these changes](/deploy/deploy-dashboard/github-101#pushing-changes) to your git repository.

:::