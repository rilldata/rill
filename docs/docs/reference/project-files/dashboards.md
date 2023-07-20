---
title: Dashboard YAML
sidebar_label: Dashboard YAML
sidebar_position: 30
---

In your Rill project directory, create a `<dashboard_name>.yaml` file in the `dashboards` directory. Rill will ingest the dashboard definition next time you run `rill start`.

## Properties

_**`model`**_ — the model name powering the dashboard with no path _(required)_

_**`title`**_ — the display name for the dashboard _(required)_

_**`timeseries`**_ — the timestamp column from your model that will underlie x-axis data in the line charts _(optional)_. If not specified, the line charts will not appear.

_**`default_time_range`**_ — the default time range shown when a user initially loads the dashboard _(optional)_. The value must be either a valid [ISO 8601 duration](https://en.wikipedia.org/wiki/ISO_8601#Durations) (for example, `PT12H` for 12 hours, `P1M` for 1 month, or `P26W` for 26 weeks) or the constant value `inf` for all time (default). If not specified, defaults to the full time range of the `timeseries` column.

_**`smallest_time_grain`**_ — the smallest time granularity the user is allowed to view in the dashboard _(optional)_. The valid values are: `millisecond`, `second`, `minute`, `hour`, `day`, `week`, `month`, `quarter`, `year`. If not specified, the user will be 

<!-- UNCOMMENT WHEN RELEASED: -->
<!--
_**`available_time_zones`**_ — time zones that should be pinned to the top of the time zone selector _(optional)_. It should be a list of [IANA time zone identifiers](https://en.wikipedia.org/wiki/List_of_tz_database_time_zones).
-->

_**`dimensions`**_ — for exploring [segments](/develop/metrics-dashboard#dimensions) and filtering the dashboard _(required)_
  - _**`column`**_ — a categorical column _(required)_ 
  - _**`name`**_ — a stable identifier for the dimension _(optional)_
  - _**`label`**_ — a label for your dashboard dimension _(optional)_ 
  - _**`description`**_ — a freeform text description of the dimension for your dashboard _(optional)_ 
  - _**`ignore`**_ — hides the dimension _(optional)_ 

_**`measures`**_ — numeric [aggregates](/develop/metrics-dashboard#measures) of columns from your data model  _(required)_
  - _**`expression`**_ — a combination of operators and functions for aggregations _(required)_ 
  - _**`name`**_ — a stable identifier for the measure _(required)_
  - _**`label`**_ — a label for your dashboard measure _(optional)_ 
  - _**`description`**_ — a freeform text description of the dimension for your dashboard _(optional)_ 
  - _**`ignore`**_ — hides the measure _(optional)_ 
  - _**`valid_percent_of_total`**_ — a boolean indicating whether percent-of-total values should be rendered for this measure _(optional)_ 
  - _**`format_preset`**_ — one of a set of values that format dashboard measures. _(optional; default is humanize)_. Possible values include:
      - _`humanize`_ — round off numbers in an opinionated way to thousands (K), millions (M), billions B), etc
      - _`none`_ — raw output
      - _`currency_usd`_ —  output rounded to 2 decimal points prepended with a dollar sign
      - _`percentage`_ — output transformed from a rate to a percentage appended with a percentage sign
      - _`comma_separators`_ — output transformed to decimal formal with commas every 3 digits
