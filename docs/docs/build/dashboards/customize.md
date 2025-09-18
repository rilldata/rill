---
title: "Customization & Themes"
description: Alter dashboard look and feel
sidebar_label: "Customization & Themes"
sidebar_position: 30
---

Below are some common customizations and dashboard configurations available for end users. 

:::info Dashboard properties

For a full list of available dashboard properties and configurations, please see our [Dashboard YAML](/reference/project-files/explore-dashboards.md) reference page.

:::


### Time Ranges

One of the more important configurations, available time ranges allow you to change the defaults in the time dropdown for periods to select. Updating this list allows users to quickly change between the most common analyses, like day over day, recent weeks, or period to date. The range must be a valid [ISO 8601 duration](https://en.wikipedia.org/wiki/ISO_8601#Durations) or one of the [Rill ISO 8601 extensions](../../reference/rill-iso-extensions.md#extensions).

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

Rill will automatically select several time zones that should be pinned to the top of the time zone selector. It should be a list of [IANA time zone identifiers](https://en.wikipedia.org/wiki/List_of_tz_database_time_zones). You can add or remove relevant time zones for your team from this list.

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


### Default Dimensions

For dashboards with wide tables, setting default dimensions is a good way to make sure that users can focus on the primary analyses and ensure a positive first experience. Each dimension listed under the `dimensions` setting will appear on the screen, while the remainder of the dimensions will be hidden (and still available for selection under filters). Common use cases for setting default fields include simplifying dashboards on initial load, narrowing the dashboard to the most used fields, and avoiding high cardinality fields (that may take longer to load, but are used less often, so this improves performance). An example addition to the `dashboard.yaml` file is below.

```yaml
defaults:
  dimensions:
    - column1
    - column2
```

:::warning Column vs. Name Usage
The `column` property is used by default from the column name in your underlying source. If you decide to use the `name` property, you'd replace the column above with the field name.
:::

### Default Measures

A list of measures that should be visible by default. Operates the same as the `default_dimensions` configuration. When selecting measures, by default, consider hiding more computationally intensive measures like `count distinct` or other complicated expressions to improve performance.

```yaml
defaults:
  measures:
    - measure_1
    - measure_1
```

### Default Comparison Modes

It is also possible to set up a default comparison mode for your dashboard. In Rill, we support both [time comparison](/explore/time-series#time-comparisons) and [dimension comparison.](/explore/filters#filter-by-dimensions) Note that only one of these comparisons can be set as default. 

```yaml
defaults:
  comparison_mode: time
  # comparison_mode: dimension
  # comparison_dimension: action

```

## Row Access Policies
### Security

Defining security policies for your data is crucial for security. For more information on this, please refer to our [Data Access Policies](/build/metrics-view/security). Check our [examples](/build/metrics-view/security#examples) for frequently used patterns.

## Changing Themes & Colors

In your Rill project directory, create a `<theme_name>.yaml` file in any directory. Rill will automatically ingest the theme next time you run `rill start` or deploy to Rill Cloud and change the color scheme of your dashboard.

First, create the YAML file as below. In this example, the charts and hover in Rill will change to Plum while spinners will change to Violet.

```yaml
type: theme
colors:
  primary: plum
  secondary: violet 
```

Once you have created that file, update the `dashboard.yaml` with the following configuration (we typically add this at the top along with time zones, time series and other configurations):

`theme: <name of theme yaml file>` 

:::info Theme properties

For more details about configuring themes, you can refer to our [Theme YAML](/reference/project-files/themes.md) reference page.

:::

## Example

```yaml
type: explore

title: Title of your explore dashboard
description: a description
metrics_view: <your-metric-view-file-name>

dimensions: '*' #can use expressions
measures: '*' #can use expressions

theme: #your default theme

time_ranges: #was available_time_ranges
time_zones: #was available_time_zones

defaults: #define all the defaults within here
    dimensions:
  ...

security:
    access: #only access can be set on dashboard level, see metric view for detailed access policies
```