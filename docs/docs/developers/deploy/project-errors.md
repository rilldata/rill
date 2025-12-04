---
title: Managing Project Errors
description: Troubleshoot and configure alerts for project errors
sidebar_label: Managing Project Errors
sidebar_position: 25
---

When you deploy to Rill Cloud, projects can encounter errors—from missing credentials to data type mismatches. This guide will help you:

- **Troubleshoot common errors** in models and dashboards
- **Understand error messages** and their solutions
- **Set up automated alerts** to catch issues early

## How Rill Handles Errors

Before diving into troubleshooting, it's helpful to understand Rill's error management approach:

- **Visibility:** View project status at the resource level via the `Status` tab or [`rill project status`](/reference/cli/project/status) CLI command
- **Isolation:** Errors are contained to individual resource trees—if one dashboard fails, others remain available
- **Fallback:** Rill attempts to serve from the most recent valid state when possible

:::tip Check upstream dependencies
The surfaced error might not be the root cause. A dashboard error could stem from an underlying model timeout. Always check the [project status page](/manage/project-management#checking-deployment-status) to trace errors to their source.
:::

## Model Errors

Model errors typically occur when there are issues with [production credentials](/build/connectors/templating), data processing, SQL syntax, or data type mismatches. Most of these errors will have surfaced in Rill Developer but a few possible issues after deploying to Rill Cloud are:

1. **Production configuration missing** - Your YAML files reference `prod:` parameters that have been defined incorrectly. Verify your [dev/prod setup](/build/connectors/templating).
2. **Timeouts, OOM** - Assuming the data in Rill Cloud is a lot larger, timeouts and OOM issues can occur. [Contact us](/contact) if you see any related error messages.

To troubleshoot model errors:

1. **Check the model's status** in the [project status page](/manage/project-management#checking-deployment-status)
2. **Review the error message** for specific details about what failed. Common error messages and their solutions:
   - **`Failed to connect to ...`**: Issue with your connector. If it worked locally, check your production credentials and [firewall settings](/build/connectors/data-source#externally-hosted-services)
   - **`Table with name ... does not exist!`**: Verify the table exists by running `rill query --sql "select * from {table_name} limit 1" --project {project_name}` or `rill project tables --project {project_name}`
   - **`IO Error: No files found that match the pattern...`**: Check that your production cloud storage folder path is correct and files exist
   - **`some partitions have errors`**: Run `rill project refresh --model {model_name} --errored-partitions` from your authenticated CLI
   - **`Out of Memory Error: ...`**: [Contact support](/contact) for assistance with memory issues 

If after going through the above steps, you are still unable to resolve the issue, [contact us!](/contact)

## Metrics View / Dashboard Errors

Metrics view and dashboard errors often stem from issues with the underlying models or configuration problems. Common issues include:

- **Model Dependencies:** Dashboards failing because their underlying models have errors
- **Missing Dimensions/Measures:** References to fields that don't exist in the underlying model

To resolve metrics view and dashboard errors:

1. Verify that all referenced models are building successfully
2. Check the measures and dimensions in your metrics YAML in GitHub or Rill Developer matches an existing column in your data
3. Check that any changes to the metrics dimensions and measures are reflected in the explore YAML.

## Setting Up Error Alerts

You can configure alerts to automatically notify you when project errors occur. Once set up, you'll receive notifications (via email or Slack) whenever any resource in your project enters an error state.

Besides alerting on project errors, it is possible to configure generic alerts in your dashboards based on specific thresholds or conditions being met. For more details, check out our [alerts documentation](/users/explore/alerts)!

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

After making these changes, you should commit and [push these changes](/developers/deploy/deploy-dashboard/github-101#pushing-changes) to your git repository.

### Configure a Slack alert

<<<<<<< HEAD:docs/docs/developers/deploy/project-errors.md
To configure a Slack alert for project errors, first follow the Slack configuration steps described on [Configuring Slack integration](/users/explore/alerts/slack). Next, add a file named `project_errors.yaml` to your Rill project with the contents below. Remember to update the `channels` field to your desired destination channel.
=======
To configure a Slack alert for project errors, first follow the Slack configuration steps described on [Configuring Slack integration](/build/connectors/data-source/slack#setting-up-the-slack-integration). Next, add a file named `project_errors.yaml` to your Rill project with the contents below. Remember to update the `channels` field to your desired destination channel.
>>>>>>> main:docs/docs/deploy/project-errors.md

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
# Follow these steps to configure a Slack token: https://docs.rilldata.com/users/explore/alerts/slack.
notify:
  slack:
    channels: [rill-alerts]
```

<<<<<<< HEAD:docs/docs/developers/deploy/project-errors.md
:::info Don't forget to commit your changes!

After making these changes, you should commit and [push these changes](/developers/deploy/deploy-dashboard/github-101#pushing-changes) to your git repository.

:::
=======
After making these changes, you should commit and [push these changes](/deploy/deploy-dashboard/github-101#pushing-changes) to your git repository or update your Rill project via the Deploy button.
>>>>>>> main:docs/docs/deploy/project-errors.md
