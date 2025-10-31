---
title: "Unnest Dimensions"
description: Tips & Tricks for Measure Formatting
sidebar_label: "Unnest Dimensions"
sidebar_position: 50
---
 For multi-value fields, you can set the `unnest` property within the dimension. If `true`, this property allows a multi-valued dimension to be unnested (such as lists), and filters will automatically switch to "contains" instead of exact match.



## Example 
In this example, the data contains an array column that has the value `['deal_one', 'deal_two', 'deal_three']`. Setting the unnest property enables the user to filter on each value in the array. Measures split by unnested values are non-additive, so that in this example the “Total Impressions” measure is applied equally across each value. Totals in Pivot Tables and the Time Dimension Detail view are calculated correctly, avoiding issues with double counted values when splitting multi-value dimensions.

<img src = '/img/build/metrics-view/examples/unnested-dimension.png' class='rounded-gif' />
<br /> 

 ```yaml
  - display_name: "Deal Name"
    column: deal_name
    description: "Unnested Column"
    unnest: true
```

In another example, we are provided with a directory_path column that gives us information on which path was edited. Using DuckDB's `regexp_split_to_array`, we have converted the string into an array. Using `unnest`, we can see which top directories are being updated more than others. 
<img src = '/img/build/metrics-view/examples/tutorial-unnest.png' class='rounded-gif' />
<br /> 

```yaml
  - expression: regexp_split_to_array(directory_path, '/')
    display_name: "The directory unnested"
    description: "The directory path"
    name: directory_path_unnested
    unnest: true
```

## Demo
[See this project live in our demo!](https://ui.rilldata.com/demo/my-rill-tutorial/explore/advanced_explore?f=directory_path_unnested+IN+%28%27docs%27%29)