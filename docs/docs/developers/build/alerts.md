---
title: Alerts
description: Define alerts as code for automated monitoring and notifications
sidebar_label: Alerts
sidebar_position: 42
---

## Overview

Alerts in Rill allow you to monitor your data and receive notifications when specific conditions are met. While alerts can be created through the UI, defining them as code in YAML files provides version control, reproducibility, and the ability to manage complex alerting logic programmatically.

When you create an alert via a YAML file, it appears in the UI marked as `Created through code`.

:::tip Using live connectors?

If you're using [live connectors](/developers/build/connectors/olap) (ClickHouse, Druid, Pinot, StarRocks, etc.), **alerts are your primary tool for data quality monitoring**. Since live connectors don't create local models, [data quality tests](/developers/build/models/data-quality-tests) won't run. Use alerts instead to validate your data on a schedule.

:::

## Alert Structure

An alert YAML file has the following core components:

```yaml
type: alert
display_name: My Alert Name
description: A brief description of what this alert monitors

# When to check the alert
refresh:
  cron: "0 * * * *"  # Every hour

# What data to check
data:
  sql: SELECT * FROM my_model WHERE condition_is_bad

# Where to send notifications
notify:
  email:
    recipients:
      - team@example.com
```

## Scheduling Alerts

The [`refresh`](/reference/project-files/alerts#refresh) property defines when and how often the alert runs.

### Cron Schedule

Use standard `cron` expressions to define the schedule:

```yaml
refresh:
  cron: "0 * * * *"      # Every hour
  time_zone: "America/New_York"  # Optional timezone
```

### Interval-Based Monitoring

Use `intervals` when you need to check data across multiple time windows, such as validating metrics for each hour or day. This is useful for time-series monitoring where you want to ensure data quality across a rolling window of time periods. Interval-based monitoring is more flexible than simple cron schedules when you need to check multiple historical periods on each evaluation.

```yaml
refresh:
  cron: "5 * * * *"  # 5 minutes past each hour

intervals:
  duration: PT1H      # 1 hour intervals
  limit: 24           # Check last 24 intervals
  check_unclosed: false
```

## Data Sources

Alerts support multiple data source types to query your data.

### SQL Query

Execute raw SQL against your models:

```yaml
data:
  sql: |
    SELECT *
    FROM orders
    WHERE created_at < NOW() - INTERVAL '24 hours'
      AND status = 'pending'
```

The alert triggers when the query returns **any rows**.

### Metrics SQL

Use `metrics_sql` when you want to query a [metrics view](/developers/build/metrics-view) using its defined dimensions and measures, rather than writing raw SQL against the underlying model. This approach leverages the metrics view's security policies and allows you to reference measures and dimensions by name. For details on the `metrics_sql` syntax, see [Custom APIs](/developers/build/custom-apis#metrics-sql-api).

```yaml
data:
  metrics_sql: |
    SELECT *
    FROM sales_metrics
    WHERE total_revenue < 1000
```

### Custom API

Use a custom API when you want to reuse complex query logic that's already defined as a [Custom API](/developers/build/custom-apis) in your project. This approach is useful for sharing validation logic between alerts and other integrations, or when you need to pass dynamic arguments to your alert queries.

```yaml
data:
  api: my_custom_validation_api
  args:
    threshold: 100
    date_range: "7d"
```

### Resource Status

Monitor the health of your Rill resources to catch pipeline failures and reconciliation errors. This is useful for monitoring pipeline health and catching reconciliation failures before they impact downstream processes.

```yaml
data:
  resource_status:
    where_error: true
```

This triggers when any resource in your project has a reconciliation error.

## Notification Configuration

Configure where and how you receive notifications when alerts trigger. You can send notifications via email, Slack, or both. Notifications are sent when the alert condition is met (when the data query returns rows), and optionally when the alert recovers or encounters evaluation errors.

### Email Notifications

```yaml
notify:
  email:
    recipients:
      - alice@example.com
      - bob@example.com
      - data-team@example.com
```

### Slack Notifications

Before using Slack notifications, you must [configure the Slack integration](/developers/build/connectors/services/slack) for your project.

```yaml
notify:
  slack:
    channels:
      - "#data-alerts"
      - "#engineering"
    users:
      - "U1234567890"  # Slack user IDs
    webhooks:
      - "https://hooks.slack.com/services/..."
```

### Combined Notifications

Send to multiple destinations:

```yaml
notify:
  email:
    recipients:
      - team@example.com
  slack:
    channels:
      - "#alerts"
```

## Alert Behavior

### Recovery Notifications

Control when you receive notifications about alert state changes. Use `on_recover` to confirm issues are resolved and get peace of mind that problems have been fixed. Use `on_error` to catch alert evaluation failures (e.g., query syntax errors) that prevent the alert from running properly.

```yaml
on_recover: true   # Notify when alert recovers
on_fail: true      # Notify when alert triggers (default)
on_error: false    # Notify on evaluation errors
```

### Re-notification (Snooze)

Control how often you're notified for ongoing issues. This prevents alert fatigue while ensuring ongoing issues aren't forgotten. Instead of receiving notifications on every evaluation cycle, you'll only be re-notified after the specified duration if the alert is still failing.

```yaml
renotify: true
renotify_after: "24h"  # Re-notify every 24 hours if still failing
```

## Working Examples

### Data Freshness Alert

This example demonstrates a data freshness check that queries the maximum timestamp from an events model and triggers when data is older than 24 hours. It uses both email and Slack notifications, includes recovery notifications to confirm when data freshness is restored, and implements re-notification every 6 hours to prevent alert fatigue while ensuring ongoing issues are tracked.

```yaml
# alerts/data_freshness.yaml
type: alert
display_name: Data Freshness Check
description: Alert when event data is stale

refresh:
  cron: "0 * * * *"  # Check every hour

data:
  sql: |
    SELECT 'Data is stale' AS error_message
    FROM (
      SELECT MAX(event_timestamp) AS latest_event
      FROM events_model
    )
    WHERE latest_event < NOW() - INTERVAL '24 hours'

notify:
  email:
    recipients:
      - data-ops@example.com
  slack:
    channels:
      - "#data-alerts"

on_recover: true
renotify: true
renotify_after: "6h"
```

### Project Health Monitor

This example monitors the overall health of your Rill project by checking for any resource reconciliation errors. It runs every 10 minutes for rapid detection of pipeline failures, uses the `resource_status` data source to automatically detect errors across all resources, and sends notifications to both Slack and email channels. Recovery notifications ensure you're alerted when issues are resolved.

```yaml
# alerts/project_health.yaml
type: alert
display_name: Project Health Monitor
description: Alert when any resource has a reconciliation error

refresh:
  cron: "*/10 * * * *"  # Every 10 minutes

data:
  resource_status:
    where_error: true

notify:
  slack:
    channels:
      - "#rill-alerts"
  email:
    recipients:
      - platform-team@example.com

on_recover: true
```

### Interval-Based Monitoring Example

This example shows how to use interval-based monitoring to validate metrics across multiple time periods. It checks hourly aggregates for the last 24 hours, looking for any hours with zero event counts. The alert runs 5 minutes past each hour to ensure the previous hour's data is complete, and uses the `intervals` configuration to systematically check each hour in the rolling window. This pattern is ideal for time-series data quality monitoring where you need to validate multiple periods on each evaluation.

```yaml
# alerts/hourly_metrics.yaml
type: alert
display_name: Hourly Metrics Check
description: Validate metrics for each hour

refresh:
  cron: "5 * * * *"  # 5 minutes past each hour

intervals:
  duration: PT1H      # 1 hour intervals
  limit: 24           # Check last 24 intervals
  check_unclosed: false

data:
  sql: |
    SELECT *
    FROM hourly_aggregates
    WHERE hour_start = DATE_TRUNC('hour', NOW() - INTERVAL '1 hour')
      AND event_count = 0

notify:
  slack:
    channels:
      - "#monitoring"
```

## Reference

For the complete specification of all available properties, see the [Alert YAML Reference](/reference/project-files/alerts).

:::note Advanced Properties

For advanced properties like `glob`, `for`, `watermark`, and `timeout`, see the [Alert YAML Reference](/reference/project-files/alerts).

:::

