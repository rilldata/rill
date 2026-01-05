---
title: Alerts
description: Define alerts as code for automated monitoring and notifications
sidebar_label: Code Alerts
sidebar_position: 0
---

## Overview

Alerts in Rill allow you to monitor your data and receive notifications when specific conditions are met. While alerts can be created through the UI, defining them as code in YAML files provides version control, reproducibility, and the ability to manage complex alerting logic programmatically.

When you create an alert via a YAML file, it appears in the UI marked as `Created through code`.

:::tip Using live connectors?

If you're using [live connectors](/build/connectors/olap) (ClickHouse, Druid, Pinot, StarRocks, etc.), **alerts are your primary tool for data quality monitoring**. Since live connectors don't create local models, [data quality tests](/build/models/data-quality-tests) won't run. Use alerts instead to validate your data on a schedule.

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

Query a metrics view directly:

```yaml
data:
  metrics_sql: |
    SELECT *
    FROM sales_metrics
    WHERE total_revenue < 1000
```

### Custom API

Call a custom API defined in your project:

```yaml
data:
  api: my_custom_validation_api
  args:
    threshold: 100
    date_range: "7d"
```

### Resource Status

Monitor the health of your Rill resources:

```yaml
data:
  resource_status:
    where_error: true
```

This triggers when any resource in your project has a reconciliation error.

## Notification Configuration

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

Before using Slack notifications, you must [configure the Slack integration](/build/connectors/data-source/slack) for your project.

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

Get notified when an alert condition resolves:

```yaml
on_recover: true   # Notify when alert recovers
on_fail: true      # Notify when alert triggers (default)
on_error: false    # Notify on evaluation errors
```

### Re-notification (Snooze)

Control how often you're notified for ongoing issues:

```yaml
renotify: true
renotify_after: "24h"  # Re-notify every 24 hours if still failing
```

## Working Examples

### Data Freshness Alert

Alert when data hasn't been updated in over 24 hours:

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

Alert on any resource errors in your project:

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

### Interval-Based Monitoring

Check data across time intervals:

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

