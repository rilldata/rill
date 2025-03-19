---
title: "Unnest Dimensions"
description: Tips & Tricks for Metric Formatting
sidebar_label: "Unnest Dimensions"
sidebar_position: 06
---
 For multi-value fields, you can set the unnest property within a dimension. If true, this property allows multi-valued dimension to be unnested (such as lists) and filters will automatically switch to "contains" instead of exact match.



## Example 
In this example, the data contains an array column that has the value `['deal_one', 'deal_two', 'deal_three']`. Setting the unnest property enables the user to filter on each value in the array. Metrics split by unnested values are non-additive, so that in this example the Total Impressions metric is applied equally across each value. Totals in Pivot Tables and the Time Dimension Detail view are calculated correctly, avoiding issues with "double counted" values when splitting multi-value dimensions.

<img src = '/img/build/metrics-view/examples/explore-percent.png' class='rounded-gif' />
<br /> update this photo

 ```yaml
  - label: "Deal Name"
    column: deal_name
    description: "Unnested Column"
    unnest: true
```
