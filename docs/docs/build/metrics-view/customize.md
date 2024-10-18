---
title: "Customization"
description: Alter metrics view features
sidebar_label: "Customization"
sidebar_position: 30
---

## Common Customizations

You will find below some common customizations and metrics view configurations that are available for end users. 

:::info Metric View properties

For a full list of available dashboard properties and configurations, please see our [Metrics View YAML](/reference/project-files/metrics-view.md) reference page.
:::


**`smallest_time_grain`**

Smallest time grain available for your users. Rill will try to infer the smallest time grain. One of the most common reasons to change this setting is your data has timestamps but is actually in hourly or daily increments. The valid values are: `millisecond`, `second`, `minute`, `hour`, `day`, `week`, `month`, `quarter`, `year`.

**`first_day_of_week`**

The start of the week defined as an integer. The valid values are 1 through 7 where Monday=`1` and Sunday=`7` _(optional)

**`first_month_of_year`**


The first month of the year for time grain aggregation. The valid values are 1 through 12 where January=`1` and December=`12` _(optional)_.


**`security`**

Defining security policies for your data is crucial for security. For more information on this, please refer to our [Dashboard Access Policies](/manage/security.md)

## Example

```yaml
version: 1 #defines version 
type: metrics_view # metrics_view

title: The title of your metrics view
display_name: The display_name
description: A description
model: refernce the model or table, 
database / database_schema: #not sure what this is used for.

timeseries: your timeseries column

smallest_time_grain: #defines the smallest time grain 

first_day_of_week: #defines first day of week
first_month_of_year: #defines first month of year

dimensions: #your dimensions, can be copied from dashboard.yaml
    - name:
      label:
      column/expression:
      property:
      description
      unnest:
      uri:

measures: #your measures, can be copied from dashboard.yaml
    - name:
      label:
      type:
      expressions:
      window:
      per:
      requires:
      description:
      format_preset / format_d3:
      valid_percent_of_total:

security: #your security policies can be copied from dashboard.yaml
```