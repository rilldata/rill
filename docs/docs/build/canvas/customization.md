---
title: "Customization & Themes"
description: Alter dashboard look and feel
sidebar_label: "Customization & Themes"
sidebar_position: 30
---

Below are some common customizations and dashboard configurations available to end users. 

:::info Dashboard properties

For a full list of available dashboard properties and configurations, please see our [Dashboard YAML](/reference/project-files/canvas-dashboards.md) reference page.

:::


### Time Ranges

One of the more important configurations, available time ranges allow you to change the defaults in the time dropdown for periods to select. Updating this list allows users to quickly change between the most common analyses, like day over day, recent weeks, or period to date. Each range must be a valid [ISO 8601 duration](https://en.wikipedia.org/wiki/ISO_8601#Durations) or one of the [Rill ISO 8601 extensions](../../reference/rill-iso-extensions.md#extensions).

```yaml
time_ranges:
  - PT15M 
  - PT1H
  - P7D
  - P4W
  - rill-TD ## Today
  - rill-WTD ## Week-To-date
```

### Time Zones

Rill will automatically select several time zones that should be pinned to the top of the time zone selector. This should be a list of [IANA time zone identifiers](https://en.wikipedia.org/wiki/List_of_tz_database_time_zones). You can add or remove relevant time zones for your team from this list.

```yaml
time_zones:
  - America/Los_Angeles
  - America/Chicago
  - America/New_York
  - Europe/London
  - Europe/Paris
  - Asia/Jerusalem
  - Europe/Moscow
  - Asia/Kolkata
  - Asia/Shanghai
  - Asia/Tokyo
  - Australia/Sydney  
```

## Setting Default Views for Dashboards
:::tip
Starting from version 0.50, the default views have been consolidated into a single YAML struct, `defaults:`.
:::

### Default Time Range

Default time range controls the data analyzed on initial page load. Setting the default time range improves user experience by setting it to the most frequently used period— in particular, avoiding `all time` if you have a large data source but only analyze more recent data.

The value must be either a valid [ISO 8601 duration](https://en.wikipedia.org/wiki/ISO_8601#Durations) (for example, `PT12H` for 12 hours, `P1M` for 1 month, or `P26W` for 26 weeks) or one of the [Rill ISO 8601 extensions](../../reference/rill-iso-extensions.md#extensions).



### Default Comparison Modes

It is also possible to set up a default comparison mode for your dashboard. For Canvas Dashboards, we support [time comparisons](/explore/filters/#time-comparisons). 

```yaml
defaults:
  comparison_mode: time
```

## Row Access Policies
### Security

Defining security policies for your data is crucial. For more information, please refer to our [Data Access Policies](/build/metrics-view/security). Check our [examples](/build/metrics-view/security#examples) for frequently used patterns.

## Changing Themes & Colors

In your Rill project directory, create a `<theme_name>.yaml` file in any directory. Rill will automatically ingest the theme next time you run `rill start` or deploy to Rill Cloud and change the color scheme of your dashboard.

First, create the YAML file as below. In this example, the charts and hover in Rill will change to Plum while spinners will change to Violet.

```yaml
type: theme
colors:
  primary: plum
  secondary: violet 
```

Once you have created that file, update the `dashboard.yaml` with the following configuration (we typically add this at the top, along with time zones, time series, and other configurations):

`theme: <name of theme yaml file>` 

:::info Theme properties

For more details about configuring themes, you can refer to our [Theme YAML](/reference/project-files/themes.md) reference page.

:::
## Example

```yaml
type: canvas
title: "Canvas Dashboard"
defaults:
  time_range: PT24H
  comparison_mode: time
time_ranges:
  - PT6H
  - PT24H
  - P7D
  - P14D
  - P4W
  - P3M
  - P12M
  - rill-PDC
  - rill-PWC
  - rill-PMC
  - rill-PQC
  - rill-PYC
time_zones:
  - UTC
  - America/Los_Angeles
  - America/Chicago
  - America/New_York
  - Europe/London
  - Europe/Paris
  - Asia/Jerusalem
  - Europe/Moscow
  - Asia/Kolkata
  - Asia/Shanghai
  - Asia/Tokyo
  - Australia/Sydney

rows:


security:
    access: #only access can be set on dashboard level, see metric view for detailed access policies

```