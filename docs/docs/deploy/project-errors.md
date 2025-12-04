---
title: Managing Project Errors
description: Configure alerts and manage errors for deployed projects in Rill Cloud
sidebar_label: Managing Project Errors
sidebar_position: 25
---

When you deploy to Rill Cloud, projects can encounter errors—from missing credentials to data type mismatches. This guide focuses on managing errors in deployed projects and setting up automated alerts.

:::info General Troubleshooting
For general troubleshooting guidance, error message explanations, and debugging techniques, see the [Debugging Rill Projects](/build/debugging) documentation.
:::

## How Rill Handles Errors

Rill's error management approach ensures visibility and isolation:

- **Visibility:** View project status at the resource level via the `Status` tab or [`rill project status`](/reference/cli/project/status) CLI command
- **Isolation:** Errors are contained to individual resource trees—if one dashboard fails, others remain available
- **Fallback:** Rill attempts to serve from the most recent valid state when possible

:::tip Check upstream dependencies
The surfaced error might not be the root cause. A dashboard error could stem from an underlying model timeout. Always check the [project status page](/manage/project-management#checking-deployment-status) to trace errors to their source.
:::

## Deployment-Specific Error Scenarios

Most errors will surface during local development in Rill Developer. However, after deploying to Rill Cloud, you may encounter additional issues:

1. **Production configuration missing** - Your YAML files reference `prod:` parameters that have been defined incorrectly. Verify your [dev/prod setup](/build/connectors/templating).
2. **Timeouts, OOM** - Production data volumes may be larger than local development data, leading to timeouts and out-of-memory issues. [Contact us](/contact) if you see any related error messages.
3. **Production credentials** - Connector credentials configured for production may differ from local development. Verify your [production credentials](/build/connectors/templating).

To troubleshoot deployment errors:

1. **Check the resource status** in the [project status page](/manage/project-management#checking-deployment-status)
2. **Review project logs** using `rill project logs` or the Rill Cloud UI
3. **Compare with local behavior** - If it worked locally, check production-specific configuration differences

For detailed troubleshooting steps and common error solutions, see the [Debugging Rill Projects](/build/debugging#troubleshooting-common-errors) guide.

## Setting Up Error Alerts

You can configure alerts to automatically notify you when project errors occur. Once set up, you'll receive notifications (via email or Slack) whenever any resource in your project enters an error state.

Besides alerting on project errors, it is possible to configure generic alerts in your dashboards based on specific thresholds or conditions being met. For more details, check out our [alerts documentation](/explore/alerts)!

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

This will give you a good idea of what object has an issue, and you can browse the [status page](/manage/project-management#checking-deployment-status) for more information.

After making these changes, you should commit and [push these changes](/deploy/deploy-dashboard/github-101#pushing-changes) to your git repository.

### Configure a Slack alert

To configure a Slack alert for project errors, first follow the Slack configuration steps described on [Configuring Slack integration](/build/connectors/data-source/slack#setting-up-the-slack-integration). Next, add a file named `project_errors.yaml` to your Rill project with the contents below. Remember to update the `channels` field to your desired destination channel.

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

After making these changes, you should commit and [push these changes](/deploy/deploy-dashboard/github-101#pushing-changes) to your git repository or update your Rill project via the Deploy button.
