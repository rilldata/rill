---
title: "Customization & Themes"
description: Alter dashboard look and feel
sidebar_label: "Customization & Themes"
sidebar_position: 30
---

## Common Customizations

You will find below some common customizations and dashboard configurations that are available for end users. 

:::info Dashboard properties

For a full list of available dashboard properties and configurations, please see our [Dashboard YAML](/reference/project-files/explore-dashboards.md) reference page.

:::


**`time_ranges`**

One of the more important configurations, available time ranges allow you to change the defaults in the time dropdown for periods to select. Updating this list allows users to quickly change between the most common analyses like day over day, recent weeks, or period to date. The range must be a valid [ISO 8601 duration](https://en.wikipedia.org/wiki/ISO_8601#Durations) or one of the [Rill ISO 8601 extensions](../../reference/rill-iso-extensions.md#extensions).

```yaml
  available_time_ranges:
  - PT15M 
  - PT1H
  - P7D
  - P4W
  - rill-TD ## Today
  - rill-WTD ## Week-To-date
```

**`time_zones`**

Rill will automatically select several time zones that should be pinned to the top of the time zone selector. It should be a list of [IANA time zone identifiers](https://en.wikipedia.org/wiki/List_of_tz_database_time_zones). You can add or remove from this list for the relevant time zones for your team.

```yaml
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

### Setting Default Views for Dashboards
:::tip
Starting from version 0.50, the default views have been consolidated into a single YAML struct, `defaults:`.
:::

**`time_range`**

Default time range controls the data analyzed on initial page load. Setting the default time range improves user experience by setting to most frequently used period - in particular, avoiding `all time` if you have a large datasource but only analyze more recent data.

The value must be either a valid [ISO 8601 duration](https://en.wikipedia.org/wiki/ISO_8601#Durations) (for example, `PT12H` for 12 hours, `P1M` for 1 month, or `P26W` for 26 weeks) or one of the [Rill ISO 8601 extensions](../../reference/rill-iso-extensions.md#extensions).


**`dimensions`**

For dashboards with wide tables, setting default dimensions is a good way to make sure that users can focus on the primary analyses and ensure a positive first experience. Each dimension listed under the `dimensions` setting would appear on the screen, while the remainder of the dimensions would be hidden (and still available for selection under filters). Common use cases for setting default fields would be to simplify dashboards on initial load, to narrow the dashboard to the most used fields, and to avoid high cardinality fields (that may take longer to load, but are used less often so improve performance). An example addition to the `dashboard.yaml` file is below.

```yaml
defaults:
  dimensions:
    - column1
    - column2
```

:::warning Column vs. Name Usage
The `column` property is used by default from the column name in your underlying source. If you decide to use the `name` property, you'd replace the column above with the field name.
:::

**`measures`** 

A list of measures that should be visible by default. Operates the same as the `default_dimensions` configuration. When selecting measures, by default, consider hiding more computationally intensive measures like `count distinct` or other complicated expressions to improve performance.

```yaml
defaults:
  measures:
    - measure_1
    - measure_1
```

**`security`**

Defining security policies for your data is crucial for security. For more information on this, please refer to our [Dashboard Access Policies](/manage/security.md)

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