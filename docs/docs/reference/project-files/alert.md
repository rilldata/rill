---
note: GENERATED. DO NOT EDIT.
title: Alert YAML
sidebar_position: 1
---

Along with alertings at the dashboard level and can be created via the UI, there might be more extensive alerting that you might want to develop and can be done so the an alert.yaml. When creating an alert via a YAML file, you'll see this denoted in the UI as `Created through code`.

## Properties


**`type`**  - _[string]_ - Refers to the resource type and must be `alert`  _(required)_

**`name`**  - _[string]_ - Name is usually inferred from the filename, but can be specified manually. 

**`refs`**  - _[array of oneOf]_ - List of resource references, each as a string or map. 

  *option 1* - _[object]_ - An object reference with at least a `<name>` and `<type>`.

  - **`type`**  - _[string]_ - type of resource 

  - **`name`**  - _[string]_ - name of resource  _(required)_

  *option 2* - _[string]_ - A string reference like `<resource-name>` or `<type/resource-name>`.

**`dev`**  - _[object]_ - Overrides properties in development 

**`prod`**  - _[object]_ - Overrides properties in production 

**`display_name`**  - _[string]_ - Refers to the display name for the alert 

**`refresh`**  - _[object]_ - Specifies the refresh schedule that Rill should follow to re-ingest and update the underlying data  _(required)_

  - **`ref_update`**  - _[boolean]_ - If true, allows the resource to run when a dependency updates. 

  - **`cron`**  - _[string]_ - A cron expression that defines the execution schedule 

  - **`every`**  - _[string]_ - Run at a fixed interval using a Go duration string (e.g., '1h', '30m', '24h'). See: https://pkg.go.dev/time#ParseDuration 

  - **`time_zone`**  - _[string]_ - Time zone to interpret the schedule in (e.g., 'UTC', 'America/Los_Angeles'). 

  - **`disable`**  - _[boolean]_ - If true, disables the resource without deleting it. 

  - **`run_in_dev`**  - _[boolean]_ - If true, allows the schedule to run in development mode. 

**`watermark`**  - _[string]_ - Specifies how the watermark is determined for incremental processing.
Use 'trigger_time' to set it at runtime or 'inherit' to use the upstream model's watermark. 

**`intervals`**  - _[object]_ - define the interval of the alert to check 

  - **`duration`**  - _[string]_ - a valid ISO8601 duration to define the interval duration 

  - **`limit`**  - _[integer]_ - maximum number of intervals to check for on invocation 

  - **`check_unclosed`**  - _[boolean]_ - boolean, whether unclosed intervals should be checked 

**`timeout`**  - _[string]_ - define the timeout of the alert in seconds (optional). 

**`data`**  - _[oneOf]_ - Specifies one of the options to retrieve or compute the data used by alert  _(required)_

  *option 1* - _[object]_ 

  - **`sql`**  - _[string]_ - Raw SQL query to run against existing models in the project.  _(required)_

  - **`connector`**  - _[string]_ - specifies the connector to use when running SQL or glob queries. 

  *option 2* - _[object]_ 

  - **`metrics_sql`**  - _[string]_ - SQL query that targets a metrics view in the project  _(required)_

  *option 3* - _[object]_ 

  - **`api`**  - _[string]_ - Name of a custom API defined in the project.  _(required)_

  - **`args`**  - _[object]_ - Arguments to pass to the custom API. 

  *option 4* - _[object]_ 

  - **`glob`**  - _[anyOf]_ - Defines the file path or pattern to query from the specified connector.  _(required)_

    *option 1* - _[string]_ 

    *option 2* - _[object]_ 

  - **`connector`**  - _[string]_ - Specifies the connector to use with the glob input. 

  *option 5* - _[object]_ 

  - **`resource_status`**  - _[object]_ - Based on resource status  _(required)_

    - **`where_error`**  - _[boolean]_ - Indicates whether the condition should trigger when the resource is in an error state. 

**`for`**  - _[oneOf]_  

  *option 1* - _[object]_ 

  - **`user_id`**  - _[string]_   _(required)_

  *option 2* - _[object]_ 

  - **`user_email`**  - _[string]_   _(required)_

  *option 3* - _[object]_ 

  - **`attributes`**  - _[object]_   _(required)_

**`on_recover`**  - _[boolean]_ - Send an alert when a previously failing alert recovers. Defaults to false. 

**`on_fail`**  - _[boolean]_ - Send an alert when a failure occurs. Defaults to true. 

**`on_error`**  - _[boolean]_ - Send an alert when an error occurs during evaluation. Defaults to false. 

**`renotify`**  - _[boolean]_ - Enable repeated notifications for unresolved alerts. Defaults to false. 

**`renotify_after`**  - _[string]_ - Defines the re-notification interval for the alert (e.g., '10m', '24h'), equivalent to snooze duration in UI, defaults to 'Off' 

**`notify`**  - _[anyOf]_ - Defines how and where to send notifications. At least one method (email or Slack) is required.  _(required)_

  *option 1* - _[object]_ 

  - **`email`**  - _[object]_ - Send notifications via email.  _(required)_

    - **`recipients`**  - _[array of string]_ - An array of email addresses to notify.  _(required)_

  *option 2* - _[object]_ 

  - **`slack`**  - _[object]_ - Send notifications via Slack.  _(required)_

    - **`users`**  - _[array of string]_ - An array of Slack user IDs to notify. 

    - **`channels`**  - _[array of string]_ - An array of Slack channel IDs to notify. 

    - **`webhooks`**  - _[array of string]_ - An array of Slack webhook URLs to send notifications to. 

**`annotations`**  - _[object]_  