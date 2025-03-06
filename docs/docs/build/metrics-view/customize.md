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

