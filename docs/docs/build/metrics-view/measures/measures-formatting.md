---
title: "Formatting your Measures"
description: Tips & Tricks for Measure Formatting
sidebar_label: "Measure Formatting"
sidebar_position: 15
---

When creating your measures in Rill, you have the option to pick from a preset of formats that we provide to you or use the [d3-format](https://d3js.org/d3-format) parameter to format your data in any way you like. While the big number in the explore dashboard won't apply all the decimals changes (it will add currency or percentage if that is the type), you will be able to see the changes in the dimension leaderboard and pivot tables. 

<img src = '/img/build/metrics-view/metrics-editor.png' class='rounded-gif' />
<br />



Using `format_d3` to control the format of a measure in the metrics view allows for further customization. 

:::tip Invalid format Strings
If an invalid format string is supplied, measures will be formatted with `format_preset: humanize`. Measures cannot have both `format_preset` and `format_d3` entries. If neither `format_preset` nor `format_d3` is supplied, measures will be formatted with the `humanize` preset.

:::


Please review the reference documentation, [here.](/reference/project-files/metrics-views)

## Customization

For further customization of your measures, you can switch to the YAML view and, with our [metrics view reference documentation](/reference/project-files/metrics-views), use the [format_d3_locale](https://d3js.org/d3-format#formatLocale) parameter to create specific formatting. 

```yaml
 format_d3: 
  format_d3_locale: 
    grouping: 
    currency: 
```


## Examples

As explained in the introduction, you'll notice that in each of the screenshots the Big Number doesn't always follow the exact formatting, but will change based on percentage/currency formatting. This is as designed, as there is a fixed width that the number has to be displayed in. Instead, you'll see these values in the dimension leaderboard, TDD, and pivot tables.

If you have any questions, please review our [reference documentation.](/reference/project-files/metrics-views)

### Format a measure to include specific amount of decimals
<img src = '/img/build/metrics-view/examples/decimal-example.png' class='rounded-gif' />
<br />

In the case that you need to view more granular values of your data, you can set the decimal places to whatever value you need. In the above example, we are setting the average voltage measure to 4 decimal places to get a more accurate representation for each dimension.

```yaml
format_d3: ".4f"
```


### Format currency with different ',' locations. IE: Indian Rupee 
<img src = '/img/build/metrics-view/examples/currency-example.png' class='rounded-gif' />
<br />


```yaml
format_d3: "$,"
format_d3_locale: 
    grouping: [3, 2, 2]
    currency: ["₹", ""]
```
As Indian Rupees are formatted in a different way than USD and EUR, you'll need to use the `format_d3_locale` parameter to set the exact grouping and currency. Likewise, if the currency symbol is written after the numeric value, you can set the currency to `["", "$"]`. 

### Percentages 
<img src = '/img/build/metrics-view/examples/percent-example.png' class='rounded-gif' />
<br />

```yaml
format_d3: '.4%'
```
While our `format_preset: percentage` will automatically apply `.2%`, you can manually set the value in format_d3 if you are looking for a more specific measure format.



## Demo
[See this project live in our demo!](https://ui.rilldata.com/demo/rill-kaggle-elec-consumption/explore/household_power_consumption_metrics_explore)