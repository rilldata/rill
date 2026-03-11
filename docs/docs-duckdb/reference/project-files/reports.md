---
note: GENERATED. DO NOT EDIT.
title: Report YAML
sidebar_position: 38
---

Reports allow you to schedule and deliver data exports or AI-powered insights to recipients via email or Slack.


## Properties

### `type`

_[string]_ - Refers to the resource type and must be `report` _(required)_

### `display_name`

_[string]_ - Display name for the report shown in notifications and UI 

### `title`

_[string]_ - Deprecated: use display_name instead 

### `refresh`

_[object]_ - Refresh schedule for the report
```yaml
refresh:
  cron: "0 9 * * *"
```
 

  - **`cron`** - _[string]_ - A cron expression that defines the execution schedule 

  - **`time_zone`** - _[string]_ - Time zone to interpret the schedule in (e.g., 'UTC', 'America/Los_Angeles'). 

  - **`disable`** - _[boolean]_ - If true, disables the resource without deleting it. 

  - **`ref_update`** - _[boolean]_ - If true, allows the resource to run when a dependency updates. 

  - **`run_in_dev`** - _[boolean]_ - If true, allows the schedule to run in development mode. 

### `watermark`

_[string]_ - Specifies how the watermark is determined for incremental processing. Use 'trigger_time' to set it at runtime or 'inherit' to use the upstream model's watermark. 

### `intervals`

_[object]_ - Define the interval of the report to check 

  - **`duration`** - _[string]_ - A valid ISO8601 duration to define the interval duration 

  - **`limit`** - _[integer]_ - Maximum number of intervals to check for on invocation 

  - **`check_unclosed`** - _[boolean]_ - Whether unclosed intervals should be checked 

### `timeout`

_[string]_ - Define the timeout for the report execution (e.g., '5m', '1h') 

### `data`

_[oneOf]_ - Data source for the report using the generic resolver pattern.
Supports ai resolvers only as of now.
 

  - **option 1** - _[object]_ - Executes a raw SQL query against the project's data models.

    - **`sql`** - _[string]_ - Raw SQL query to run against existing models in the project. _(required)_

    - **`connector`** - _[string]_ - specifies the connector to use when running SQL or glob queries. 

  - **option 2** - _[object]_ - Executes a SQL query that targets a defined metrics view.

    - **`metrics_sql`** - _[string]_ - SQL query that targets a metrics view in the project _(required)_

  - **option 3** - _[object]_ - Calls a custom API defined in the project to compute data.

    - **`api`** - _[string]_ - Name of a custom API defined in the project. _(required)_

    - **`args`** - _[object]_ - Arguments to pass to the custom API. 

  - **option 4** - _[object]_ - Uses a file-matching pattern (glob) to query data from a connector.

    - **`glob`** - _[anyOf]_ - Defines the file path or pattern to query from the specified connector. _(required)_

      - **option 1** - _[string]_ - A simple file path/glob pattern as a string.

      - **option 2** - _[object]_ - An object-based configuration for specifying a file path/glob pattern with advanced options.

    - **`connector`** - _[string]_ - Specifies the connector to use with the glob input. 

  - **option 5** - _[object]_ - Uses the status of a resource as data.

    - **`resource_status`** - _[object]_ - Based on resource status _(required)_

      - **`where_error`** - _[boolean]_ - Indicates whether the condition should trigger when the resource is in an error state. 

  - **option 6** - _[object]_ - Uses AI to generate insights and analysis from metrics data. Only available for reports.

    - **`ai`** - _[object]_ - AI resolver configuration for generating automated insights _(required)_

      - **`prompt`** - _[string]_ - Custom prompt to guide the AI analysis. If not provided, a default analysis prompt is used. 

      - **`time_range`** - _[object]_ - Time range for the analysis period 

        - **`iso_duration`** - _[string]_ - ISO 8601 duration (e.g., P7D for 7 days, P1M for 1 month) 

        - **`iso_offset`** - _[string]_ - ISO 8601 offset from current time (e.g., P1D to start from yesterday) 

        - **`start`** - _[string]_ - Start timestamp in ISO 8601 format 

        - **`end`** - _[string]_ - End timestamp in ISO 8601 format 

        - **`expression`** - _[string]_ - Rill time expression (e.g., 'last 7 days', 'this month') 

      - **`comparison_time_range`** - _[object]_ - Optional comparison time range for period-over-period analysis 

        - **`iso_duration`** - _[string]_ - ISO 8601 duration for comparison period 

        - **`iso_offset`** - _[string]_ - ISO 8601 offset for comparison period (e.g., P7D to compare with previous week) 

        - **`start`** - _[string]_ - Start timestamp in ISO 8601 format 

        - **`end`** - _[string]_ - End timestamp in ISO 8601 format 

        - **`expression`** - _[string]_ - Rill time expression for comparison period 

      - **`context`** - _[object]_ - Context to constrain the AI analysis 

        - **`explore`** - _[string]_ - Name of the explore dashboard to analyze 

        - **`dimensions`** - _[array of string]_ - List of dimensions to include in analysis 

        - **`measures`** - _[array of string]_ - List of measures to include in analysis 

### `query`

_[object]_ - Legacy query-based report configuration 

  - **`name`** - _[string]_ - Name of the query to execute (e.g., MetricsViewAggregation) 

  - **`args`** - _[object]_ - Arguments to pass to the query 

  - **`args_json`** - _[string]_ - Query arguments as a JSON string (alternative to args) 

### `export`

_[object]_ - Export configuration for query-based reports 

  - **`format`** - _[string]_ - Export file format 

  - **`include_header`** - _[boolean]_ - Include column headers in the export 

  - **`limit`** - _[integer]_ - Maximum number of rows to export 

### `notify`

_[object]_ - Notification configuration for email and Slack delivery 

  - **`email`** - _[object]_ - Send notifications via email. 

    - **`recipients`** - _[array of string]_ - An array of email addresses to notify. _(required)_

  - **`slack`** - _[object]_ - Send notifications via Slack. 

    - **`users`** - _[array of string]_ - An array of Slack user IDs to notify. 

    - **`channels`** - _[array of string]_ - An array of Slack channel names to notify. 

    - **`webhooks`** - _[array of string]_ - An array of Slack webhook URLs to send notifications to. 

### `annotations`

_[object]_ - Key-value pairs for report metadata (e.g., admin_owner_user_id for AI reports) 

## Common Properties

### `name`

_[string]_ - Name is usually inferred from the filename, but can be specified manually. 

### `refs`

_[array of string]_ - List of resource references 

### `dev`

_[object]_ - Overrides any properties in development environment. 

### `prod`

_[object]_ - Overrides any properties in production environment. 

## Examples

```yaml
# Example: query-based report with CSV export
type: report
display_name: Weekly Sales Report
refresh:
    cron: "0 9 * * 1"
data:
    metrics:
        metrics_view: sales_metrics
        dimensions:
            - name: region
        measures:
            - name: total_sales
        time_range:
            expression: "7D as of latest/D"
export:
    format: csv
    limit: 1000
notify:
    email:
        recipients:
            - sales@example.com
annotations:
    admin_owner_user_id: user-123
    web_open_mode: recipient # report will use recipient's permission to run the query
```

```yaml
# Example: AI-powered insight report
type: report
display_name: Daily AI Insights
refresh:
    cron: "0 8 * * *"
data:
    ai:
        prompt: "Analyze key metrics and identify significant changes"
        time_range:
            expression: "1D as of latest/D"
        comparison_time_range:
            expression: "1D as of latest/D offset -1D"
        context:
            explore: my_explore
notify:
    email:
        recipients:
            - team@example.com
annotations:
    admin_owner_user_id: user-123 # report will be run with this user permission
```