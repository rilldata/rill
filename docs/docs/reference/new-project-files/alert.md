---
note: GENERATED. DO NOT EDIT.
title: Alert YAML
sidebar_position: 1
---

Along with alertings at the dashboard level and can be created via the UI, there might be more extensive alerting that you might want to develop and can be done so the an alert.yaml. When creating an alert via a YAML file, you'll see this denoted in the UI as `Created through code`.

## Properties


**`type`**  - _[string]_ - Refers to the resource type and must be `alert`  _(required)_

**`name`**  - _[string]_ - Name is usually inferred from the filename, but can be specified manually. 

**`namespace`**  - _[string]_ - Optional value to group resources by. Prepended to the resource name as `<namespace>/<name>`. 

**`refs`**  - _[array]_ - List of resource references, each as a string or map. 

     *option 1* - _[string]_ - A string reference like 'resource-name' or 'Kind/resource-name'.

     *option 2* - _[object]_ - An object reference with at least a 'name' and 'type'.

    - **`name`**  - _[string]_ -   _(required)_

    - **`type`**  - _[string]_ -  

**`version`**  - _[integer]_ - Version of the parser to use for this file. Enables backwards compatibility for breaking changes. 

**`renotify_after`**  - _[string]_ - Defines the re-notification interval for the alert (e.g., '10m', '24h'), equivalent to snooze duration in UI, defaults to 'Off' 

**`watermark`**  - _[string]_ - Specifies how the watermark is determined for incremental processing.
Use 'trigger_time' to set it at runtime or 'inherit' to use the upstream model's watermark. 

**`annotations`**  - _[object]_ -  

**`data`**  - _[object]_ -   _(required)_

   *option 1* - 

   *option 2* - 

   *option 3* - 

   *option 4* - 

   *option 5* - 

**`for`**  - _[object]_ -  

   *option 1* - 

   *option 2* - 

   *option 3* - 

**`intervals`**  - _[object]_ - define the interval of the alert to check 

  - **`duration`**  - _[string]_ - a valid ISO8601 duration to define the interval duration 

  - **`limit`**  - _[integer]_ - maximum number of intervals to check for on invocation 

  - **`check_unclosed`**  - _[boolean]_ - boolean, whether unclosed intervals should be checked 

**`on_error`**  - _[boolean]_ - Send an alert when an error occurs during evaluation. Defaults to false. 

**`on_recover`**  - _[boolean]_ - Send an alert when a previously failing alert recovers. Defaults to false. 

**`timeout`**  - _[string]_ - define the timeout of the alert in seconds (optional). 

**`display_name`**  - _[string]_ - Refers to the display name for the alert 

**`notify`**  - _[object]_ - Defines how and where to send notifications. At least one method (email or Slack) is required.  _(required)_

**`on_fail`**  - _[boolean]_ - Send an alert when a failure occurs. Defaults to true. 

**`refresh`**  - _[object]_ - Specifies the refresh schedule that Rill should follow to re-ingest and update the underlying data  _(required)_

  - **`run_in_dev`**  - _[boolean]_ - If true, allows the schedule to run in development mode. 

  - **`time_zone`**  - _[string]_ - Time zone to interpret the schedule in (e.g., 'UTC', 'America/Los_Angeles'). 

  - **`cron`**  - _[string]_ - A cron expression that defines the execution schedule 

  - **`disable`**  - _[boolean]_ - If true, disables the resource without deleting it. 

  - **`every`**  - _[string]_ - Run at a fixed interval using a Go duration string (e.g., '1h', '30m', '24h'). See: https://pkg.go.dev/time#ParseDuration 

  - **`ref_update`**  - _[boolean]_ - If true, allows the resource to run when a dependency updates. 

**`renotify`**  - _[boolean]_ - Enable repeated notifications for unresolved alerts. Defaults to false. 