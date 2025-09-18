---
title: Create YAML Alerts
description: Create custom YAML alerts 
sidebar_label: Alerts
---

Along with alerts in [Rill Cloud](/explore/alerts) that can alert you to certain data level conditions, it is also possible to create YAML alerts that allow for a bit more flexibility than what is provided in the UI. Rill [alerts](/reference/project-files/alerts) allow for resource checking, model checks and running custom APIs and more. When creating an alert via a YAML file, you'll see this denoted in the UI as `Created through code` and cannot be modified via the Rill Cloud UI.


## Project Status Alerts

To help you quickly identify and fix errors, you can configure a Rill alert that will trigger when one or more resources in your project enter an error state. The alert must be configured using a YAML file committed to your Rill project repository (configuration through the UI is not yet possible).

### Configure an email alert
To configure an email alert for project errors, add a file named project_errors.yaml to your Rill project with the contents below. Remember to update the recipients' field to your desired alert recipients.

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
To configure a Slack alert for project errors, first follow the Slack configuration steps described on Configuring Slack integration. Next, add a file named project_errors.yaml to your Rill project with the contents below. Remember to update the channels field to your desired destination channel.
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
## SQL / Metrics SQL Queries

Alerts using `sql` or `metrics_sql` on an underlying model or table can be used to check for data accuracy, anomalies, and business logic validation. These alerts allow you to run custom SQL queries against your data and trigger notifications when a non-empty result is returned.

### Example: Data Quality Alert

```yaml
type: alert

# Check every hour
refresh:
  cron: "0 * * * *"

# SQL query to check for data quality issues
data:
  sql: |
    SELECT 
      COUNT(*) as missing_records
    FROM your_table 
    WHERE column IS NULL

# Alert triggers if any records are found (non-zero result)
# Send email notification
notify:
  email:
    recipients: [data-team@example.com]

```

### Example: Metrics SQL Alert

```yaml
type: alert

# Check every 15 minutes
refresh:
  cron: "*/15 * * * *"

# Use metrics SQL to check for anomalies in key metrics
data:
  metrics_sql: |
    SELECT 
      SUM(revenue) as total_revenue,
      COUNT(DISTINCT customer_id) as unique_customers
    FROM your_metrics_view
    WHERE total_revenue < 10000

notify:
  slack:
    channels: [alerts]
```

## Custom API

Custom API alerts allow you to integrate with external systems and trigger alerts based on responses from third-party APIs. This is useful for monitoring external dependencies, checking service health, or validating data against external sources.

### Example: Trigger a custom API

```yaml
type: alert

# Check every 5 minutes
refresh:
  cron: "*/5 * * * *"

# Call external API to check service health
data:
  api: custom_api
  args:
    - startDate: '2025-01-01'


notify:
  email:
    recipients: [ops-team@example.com]
```

For more information, see our [Alerts YAML reference page](/reference/project-files/alerts)!