---
note: GENERATED. DO NOT EDIT.
title: Alert YAML
sidebar_position: 38
---

Along with alertings at the dashboard level and can be created via the UI, there might be more extensive alerting that you might want to develop and can be done so the an alert.yaml. When creating an alert via a YAML file, you'll see this denoted in the UI as `Created through code`.

## Properties

### `type`

_[string]_ - Refers to the resource type and must be `alert` _(required)_

### `refresh`

_[object]_ - Refresh schedule for the alert
  ```yaml
  refresh:
    cron: "* * * * *"
    #every: "24h"
  ```
 _(required)_

  - **`cron`** - _[string]_ - A cron expression that defines the execution schedule 

  - **`time_zone`** - _[string]_ - Time zone to interpret the schedule in (e.g., 'UTC', 'America/Los_Angeles'). 

  - **`disable`** - _[boolean]_ - If true, disables the resource without deleting it. 

  - **`ref_update`** - _[boolean]_ - If true, allows the resource to run when a dependency updates. 

  - **`run_in_dev`** - _[boolean]_ - If true, allows the schedule to run in development mode. 

### `display_name`

_[string]_ - Display name for the alert 

### `description`

_[string]_ - Description for the alert 

### `data`

_[oneOf]_ - Data source for the alert _(required)_

    - **`sql`** - _[string]_ - Raw SQL query to run against existing models in the project. _(required)_

    - **`connector`** - _[string]_ - specifies the connector to use when running SQL or glob queries. 

    - **`metrics_sql`** - _[string]_ - SQL query that targets a metrics view in the project _(required)_

    - **`api`** - _[string]_ - Name of a custom API defined in the project. _(required)_

    - **`args`** - _[object]_ - Arguments to pass to the custom API. 

    - **`glob`** - _[anyOf]_ - Defines the file path or pattern to query from the specified connector. _(required)_

      - **option 1** - _[string]_ - A simple file path/glob pattern as a string.

      - **option 2** - _[object]_ - An object-based configuration for specifying a file path/glob pattern with advanced options.

    - **`connector`** - _[string]_ - Specifies the connector to use with the glob input. 

    - **`resource_status`** - _[object]_ - Based on resource status _(required)_

      - **`where_error`** - _[boolean]_ - Indicates whether the condition should trigger when the resource is in an error state. 

```yaml
resource_status:
  where_error: true
```


### `intervals`

_[object]_ - define the interval of the alert to check 

  - **`duration`** - _[string]_ - a valid ISO8601 duration to define the interval duration 

  - **`limit`** - _[integer]_ - maximum number of intervals to check for on invocation 

  - **`check_unclosed`** - _[boolean]_ - boolean, whether unclosed intervals should be checked 

### `watermark`

_[string]_ - Specifies how the watermark is determined for incremental processing. Use 'trigger_time' to set it at runtime or 'inherit' to use the upstream model's watermark. 

### `timeout`

_[string]_ - define the timeout of the alert in seconds (optional). 

### `for`

_[oneOf]_ - Specifies how user identity or attributes should be evaluated for security policy enforcement. 

    - **`user_id`** - _[string]_ - The unique user ID used to evaluate security policies. _(required)_

    - **`user_email`** - _[string]_ - The user's email address used to evaluate security policies. _(required)_

    - **`attributes`** - _[object]_ - A dictionary of user attributes used to evaluate security policies. _(required)_

### `on_recover`

_[boolean]_ - Send an alert when a previously failing alert recovers. Defaults to false. 

### `on_fail`

_[boolean]_ - Send an alert when a failure occurs. Defaults to true. 

### `on_error`

_[boolean]_ - Send an alert when an error occurs during evaluation. Defaults to false. 

### `renotify`

_[boolean]_ - Enable repeated notifications for unresolved alerts. Defaults to false. 

### `renotify_after`

_[string]_ - Defines the re-notification interval for the alert (e.g., '10m','24h'), equivalent to snooze duration in UI, defaults to 'Off' 

### `condition`

_[object]_ - Condition that triggers the alert 

  - **`operator`** - _[string]_ - Comparison operator (gt, lt, eq, etc.) 

  - **`threshold`** - _[no type]_ - Threshold value for the condition 

  - **`measure`** - _[string]_ - Measure to compare against the threshold 

### `notify`

_[object]_ - Notification configuration _(required)_

  - **`email`** - _[object]_ - Send notifications via email. 

    - **`recipients`** - _[array of string]_ - An array of email addresses to notify. _(required)_

  - **`slack`** - _[object]_ - Send notifications via Slack. 

    - **`users`** - _[array of string]_ - An array of Slack user IDs to notify. 

    - **`channels`** - _[array of string]_ - An array of Slack channel IDs to notify. 

    - **`webhooks`** - _[array of string]_ - An array of Slack webhook URLs to send notifications to. 

### `annotations`

_[object]_ - Key value pair used for annotations 

## Common Properties

### `name`

_[string]_ - Name is usually inferred from the filename, but can be specified manually. 

### `refs`

_[array of string]_ - List of resource references 

### `dev`

_[object]_ - Overrides any properties in development environment. 

### `prod`

_[object]_ - Overrides any properties in production environment. 