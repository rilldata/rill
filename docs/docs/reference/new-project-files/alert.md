---
note: GENERATED. DO NOT EDIT.
title: Alert YAML
sidebar_position: 1
---
## Alert YAML

Along with alertings at the dashboard level and can be created via the UI, there might be more extensive alerting that you might want to develop and can be done so the an alert.yaml. When creating an alert via a YAML file, you'll see this denoted in the UI as `Created through code`.

Type: `object`

## Properties:
#### All of the following:
- Part 1:
  ## type

  Type: `object`

  ## Properties:

  - **type** _(required)_:
    Type: `%!s(<nil>)`

- Part 2:
  ## common_properties

  Type: `object`

  ## Properties:

  - **namespace**:
    Optional value to group resources by. Prepended to the resource name as `<namespace>/<name>`.

    Type: `string`


  - **refs**:
    List of resource references, each as a string or map.

    Type: `array`

    #### Array Items:
      Type: `%!s(<nil>)`

      #### One of the following:
      - Option 1:
        A string reference like 'resource-name' or 'Kind/resource-name'.

        Type: `string`

      - Option 2:
        An object reference with at least a 'name' and 'type'.

        Type: `object`

        ## Properties:

        - **name** _(required)_:
          Type: `string`


        - **type**:
          Type: `string`


  - **version**:
    Version of the parser to use for this file. Enables backwards compatibility for breaking changes.

    Type: `integer`


  - **name**:
    Name is usually inferred from the filename, but can be specified manually.

    Type: `string`

- Part 3:
  ## alert_properties

  Type: `object`

  ## Properties:

  - **for**:
    Type: `object`

    ## Properties:
    #### One of the following:
    - Option 1:
      Type: `%!s(<nil>)`

    - Option 2:
      Type: `%!s(<nil>)`

    - Option 3:
      Type: `%!s(<nil>)`


  - **on_error**:
    Send an alert when an error occurs during evaluation. Defaults to false.

    Type: `boolean`


  - **on_fail**:
    Send an alert when a failure occurs. Defaults to true.

    Type: `boolean`


  - **renotify**:
    Enable repeated notifications for unresolved alerts. Defaults to false.

    Type: `boolean`


  - **renotify_after**:
    Defines the re-notification interval for the alert (e.g., '10m', '24h'), equivalent to snooze duration in UI, defaults to 'Off'

    Type: `string`


  - **data** _(required)_:
    Type: `object`

    ## Properties:
    #### One of the following:
    - Option 1:
      ## sql

      Type: `%!s(<nil>)`

    - Option 2:
      ## metrics_sql

      Type: `%!s(<nil>)`

    - Option 3:
      ## api

      Type: `%!s(<nil>)`

    - Option 4:
      ## glob

      Type: `%!s(<nil>)`

    - Option 5:
      ## resource_status

      Type: `%!s(<nil>)`


  - **intervals**:
    define the interval of the alert to check

    Type: `object`

    ## Properties:

    - **check_unclosed**:
      boolean, whether unclosed intervals should be checked

      Type: `boolean`


    - **duration**:
      a valid ISO8601 duration to define the interval duration

      Type: `string`


    - **limit**:
      maximum number of intervals to check for on invocation

      Type: `integer`


  - **notify** _(required)_:
    ## notify_properties

    Defines how and where to send notifications. At least one method (email or Slack) is required.

    Type: `object`

    ## Properties:
    #### Any of the following:
    - Option 1:
      ## email_properties

      Type: `%!s(<nil>)`

    - Option 2:
      ## slack_properties

      Type: `%!s(<nil>)`


  - **on_recover**:
    Send an alert when a previously failing alert recovers. Defaults to false.

    Type: `boolean`


  - **refresh** _(required)_:
    Specifies the refresh schedule that Rill should follow to re-ingest and update the underlying data

    Type: `object`

    ## Properties:

    - **cron**:
      A cron expression that defines the execution schedule

      Type: `string`


    - **disable**:
      If true, disables the resource without deleting it.

      Type: `boolean`


    - **every**:
      Run at a fixed interval using a Go duration string (e.g., '1h', '30m', '24h'). See: https://pkg.go.dev/time#ParseDuration

      Type: `string`


    - **ref_update**:
      If true, allows the resource to run when a dependency updates.

      Type: `boolean`


    - **run_in_dev**:
      If true, allows the schedule to run in development mode.

      Type: `boolean`


    - **time_zone**:
      Time zone to interpret the schedule in (e.g., 'UTC', 'America/Los_Angeles').

      Type: `string`


  - **timeout**:
    define the timeout of the alert in seconds (optional).

    Type: `string`


  - **watermark**:
    Specifies how the watermark is determined for incremental processing.
Use 'trigger_time' to set it at runtime or 'inherit' to use the upstream model's watermark.

    Type: `string`

    Enum: `[trigger_time inherit]`


  - **annotations**:
    Type: `object`

    ## Properties:

  - **display_name**:
    Refers to the display name for the alert

    Type: `string`

- Part 4:
  ## environment_overrides

  Type: `%!s(<nil>)`

