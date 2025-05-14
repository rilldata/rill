---
note: GENERATED. DO NOT EDIT.
title: Report YAML
sidebar_position: 39
---



## Properties

### `type`

_[string]_ - Refers to the resource type and must be `report`  _(required)_

### `name`

_[string]_ - Name is usually inferred from the filename, but can be specified manually. 

### `refs`

_[array of string]_ - List of resource references, each as a string or map. 

### `dev`

_[object]_ - Overrides properties in development 

### `prod`

_[object]_ - Overrides properties in production 

### `display_name`

_[string]_ - the display name of your report. 

### `refresh`

_[object]_ - Specifies the refresh schedule that Rill should follow to re-ingest and update the underlying data 

  - **`cron`** - _[string]_ - A cron expression that defines the execution schedule 

  - **`time_zone`** - _[string]_ - Time zone to interpret the schedule in (e.g., 'UTC', 'America/Los_Angeles'). 

  - **`disable`** - _[boolean]_ - If true, disables the resource without deleting it. 

  - **`ref_update`** - _[boolean]_ - If true, allows the resource to run when a dependency updates. 

  - **`run_in_dev`** - _[boolean]_ - If true, allows the schedule to run in development mode. 

### `watermark`

_[string]_ - Specifies how the watermark is determined for incremental processing.
Use 'trigger_time' to set it at runtime or 'inherit' to use the upstream model's watermark. 

### `intervals`

_[object]_ - define the interval of the report to check 

  - **`duration`** - _[string]_ - a valid ISO8601 duration to define the interval duration 

  - **`limit`** - _[integer]_ - maximum number of intervals to check for on invocation 

  - **`check_unclosed`** - _[boolean]_ - boolean, whether unclosed intervals should be checked 

### `timeout`

_[string]_ - define the timeout of the report in seconds (optional). 

### `query`

_[object]_   _(required)_

  - **`name`** - _[string]_   _(required)_

  - **`args`** - _[object]_  

  - **`args_json`** - _[string]_  

### `export`

_[object]_ - to define the export properties  _(required)_

  - **`format`** - _[string]_ - Format for exported report: can be 'csv', 'xlsx', or 'parquet'.  _(required)_

  - **`limit`** - _[integer]_  

### `notify`

_[anyOf]_ - Defines how and where to send notifications. At least one method (email or Slack) is required  _(required)_

  **&nbsp;&nbsp;&nbsp;&nbsp;option 1** - _[object]_ 

  - **`email`** - _[object]_ - Send notifications via email.  _(required)_

    - **`recipients`** - _[array of string]_ - An array of email addresses to notify.  _(required)_

  **&nbsp;&nbsp;&nbsp;&nbsp;option 2** - _[object]_ 

  - **`slack`** - _[object]_ - Send notifications via Slack.  _(required)_

    - **`users`** - _[array of string]_ - An array of Slack user IDs to notify. 

    - **`channels`** - _[array of string]_ - An array of Slack channel IDs to notify. 

    - **`webhooks`** - _[array of string]_ - An array of Slack webhook URLs to send notifications to. 

### `annotations`

_[object]_  