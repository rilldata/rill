---
title: Dashboard YAML
sidebar_label: Dashboard YAML
sidebar_position: 30
---

In your Rill project directory, create a `<dashboard_name>.yaml` file in the `dashboards` directory. Rill will ingest the dashboard definition next time you run `rill start`.

## Properties

_**`model`**_ — the model name powering the dashboard with no path _(required)_

_**`display_name`**_ — the display name for the dashboard _(required)_

_**`timeseries`**_ — column from your model that will underlie x-axis data in the line charts _(required)_

_**`dimensions:`**_ — for exploring [segments](/develop/metrics-dashboard#dimensions) and filtering the dashboard _(required)_
  - _**`property`**_ — a categorical column _(required)_ 
  - _**`label`**_ — a label for your dashboard dimension _(optional)_ 
  - _**`description`**_ — a freeform text description of the dimension for your dashboard _(optional)_ 

_**`measures:`**_ — numeric [aggregates](/develop/metrics-dashboard#measures) of columns from your data model  _(required)_
  - _**`expression`**_ — a combination of operators and functions for aggregations _(required)_ 
  - _**`label`**_ — a label for your dashboard measure _(optional)_ 
  - _**`description`**_ — a freeform text description of the dimension for your dashboard _(optional)_ 
  - _**`format_preset`**_ — one of a set of values that format dashboard measures. _(optional; default is humanize)_. Possible values include:
      - _`humanize`_ — round off numbers in an opinionated way to thousands (K), millions (M), billions B), etc
      - _`none`_ — raw output
      - _`currency_usd`_ —  output rounded to 2 decimal points prepended with a dollar sign
      - _`percentage`_ — output transformed from a rate to a percentage appended with a percentage sign
      - _`comma_separators`_ — output transformed to decimal formal with commas every 3 digits
