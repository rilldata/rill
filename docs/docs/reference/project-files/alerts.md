---
title: Alert YAML
sidebar_label: Alert YAML
sidebar_position: 60
---

Along with alertings at the dashboard level and can be created via the UI, there might be more extensive alerting that you might want to develop and can be done so the an alert.yaml. When creating an alert via a YAML file, you'll see this denoted in the UI as `Created through code`.

**`type`** — Refers to the resource type and must be `alert` _(required)_. 

**`title`** — Refers to the display name for the metrics view [deprecated, use `display_name`] _(required)_.

**`display_name`** - Refers to the display name for the metrics view _(required)_.

**`refresh`** - Specifies the refresh schedule that Rill should follow to re-ingest and update the underlying source data _(optional)_.
  - **`cron`** - a cron schedule expression, which should be encapsulated in single quotes, e.g. `'* * * * *'` _(optional)_.
  - **`every`** - a Go duration string, such as `24h` ([docs](https://pkg.go.dev/time#ParseDuration)) _(optional)_.
  - **`disable`** - boolean, completely disable the resource, without deleting it _(optional)_.
  - **`ref_update`** -: boolean, don't refresh when a dependency is refreshed _(optional)_.
```
refresh:
    cron: "0 8 * * *"
```

**`intervals`** - define the interval of the alert to check _(required)_.
  - **`duration`** - a valid ISO8601 duration to define the interval duration.  _(required)_.
  - **`limit`** -  maximum number of intervals to check for on invocation _(optional)_.
  - **`check_unclosed`** -  boolean, whether unclosed intervals should be checked  _(optional)_.

```yaml
intervals:
  duration: 'P3D'
#  limit: 5
#  check_unclosed: true
```

**`timeout`** - define the timeout of the alert in seconds _(optional)_.

**`data`** - define the alert constraints using various parameters  _(required)_.
  - **`connector`** - if running a SQL query or using `glob`, will need to define what connector to use _(optional)_.
  - **`sql`** - raw SQL query to run against existing tables and views in your project _(optional)_.
  - **`metrics_sql`** - a SQL query against a metrics view in your project _(optional)_.
  - **`api`** - name of existing custom API in your project _(optional)_.
  - **`args`** - used with `api` to define args to be passed to the API _(optional)_.
  - **`glob`** - define the path in your connector _(optional)_.
  - **`resource_status`** - 
	- **`where_error`** - boolean, if the returning data alert constraints returns true or false _(required)_.

```yaml
#Alert will trigger if any of the project's resources return with a reconile error.
data:
  resource_status:
    where_error: true
```

**`on_recover`** - boolean, send alert on recovery, defaults to false _(optional)_.

**`on_fail`** - boolean, send alert of failure, defaults to true _(optional)_.

**`on_error`** - boolean, send the alert on error, defaults to false _(optional)_.

**`renotify`** - boolean, enable to disable renotifcation of alert, defaults to false _(optional)_.

**`renotify_after`** - define the renotification of the alert in seconds, equiavalent to snooze duration in UI, defaults to 'Off' _(optional)_.

```yaml
on_recover: true
on_fail: true
on_error: true

renotify: true
# renotify_after: 360m
```

**`notify`** - define where to notify the user of the defined alert _(required)_.
  - **`email`** -  
	  - **`recipients`** - an array of emails to send the alert to _(optional)_.
  - **`slack`** -  
  	  - **`users`** -  an array of Slack users to send the alert notification to _(optional)_.
	  - **`channels`** -  an array of Slack channels to send the alert notification to _(optional)_.
	  - **`webhooks`** -  an array of webhooks to send the alert notification to _(optional)_.

```yaml
# Send notifications by email or slack 
notify:
  email:
    recipients: [email@domain.com]
  slack:
    users: []
    channels: []
    webhooks: []
```