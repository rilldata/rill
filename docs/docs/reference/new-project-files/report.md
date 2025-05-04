---
note: GENERATED. DO NOT EDIT.
title: Report YAML
sidebar_position: 9
---



## Properties


**`type`**  - _[string]_ - Refers to the resource type and must be `report`  _(required)_

**`refs`**  - _[array]_ - List of resource references, each as a string or map. 

     *option 1* - _[string]_ - A string reference like 'resource-name' or 'Kind/resource-name'.

     *option 2* - _[object]_ - An object reference with at least a 'name' and 'type'.

    - **`type`**  - _[string]_ -  

    - **`name`**  - _[string]_ -   _(required)_

**`version`**  - _[integer]_ - Version of the parser to use for this file. Enables backwards compatibility for breaking changes. 

**`name`**  - _[string]_ - Name is usually inferred from the filename, but can be specified manually. 

**`namespace`**  - _[string]_ - Optional value to group resources by. Prepended to the resource name as `<namespace>/<name>`. 

**`annotations`**  - _[object]_ -  

**`display_name`**  - _[string]_ - the display name of your report. 

**`export`**  - _[object]_ - to define the export properties  _(required)_

  - **`format`**  - _[string]_ - Format for exported report: can be 'csv', 'xlsx', or 'parquet'.  _(required)_

  - **`limit`**  - _[integer]_ -  

**`intervals`**  - _[object]_ - define the interval of the report to check 

  - **`check_unclosed`**  - _[boolean]_ - boolean, whether unclosed intervals should be checked 

  - **`duration`**  - _[string]_ - a valid ISO8601 duration to define the interval duration 

  - **`limit`**  - _[integer]_ - maximum number of intervals to check for on invocation 

**`query`**  - _[object]_ -   _(required)_

  - **`args`**  - _[object]_ -  

  - **`args_json`**  - _[string]_ -  

  - **`name`**  - _[string]_ -   _(required)_

**`refresh`**  - _[object]_ - Specifies the refresh schedule that Rill should follow to re-ingest and update the underlying data 

  - **`run_in_dev`**  - _[boolean]_ - If true, allows the schedule to run in development mode. 

  - **`time_zone`**  - _[string]_ - Time zone to interpret the schedule in (e.g., 'UTC', 'America/Los_Angeles'). 

  - **`cron`**  - _[string]_ - A cron expression that defines the execution schedule 

  - **`disable`**  - _[boolean]_ - If true, disables the resource without deleting it. 

  - **`every`**  - _[string]_ - Run at a fixed interval using a Go duration string (e.g., '1h', '30m', '24h'). See: https://pkg.go.dev/time#ParseDuration 

  - **`ref_update`**  - _[boolean]_ - If true, allows the resource to run when a dependency updates. 

**`watermark`**  - _[string]_ - Specifies how the watermark is determined for incremental processing.
Use 'trigger_time' to set it at runtime or 'inherit' to use the upstream model's watermark. 

**`notify`**  - _[object]_ - Defines how and where to send notifications. At least one method (email or Slack) is required  _(required)_

**`timeout`**  - _[string]_ - define the timeout of the report in seconds (optional). 