---
title: "Referencing Measures"
description: Tips & Tricks for Referencing Measure
sidebar_label: "Referencing Measures"
sidebar_position: 25
---

Within a metrics view, it is possible for a measure to reference another by using the `requires` array parameter. By doing this, you can easily aggregate already existing measures to simplify the expressions. For example, get a percentage of two already summed values. 


<img src = '/img/build/metrics-view/examples/requires-example.png' class='rounded-gif' />
<br />

Please review the reference documentation, [here.](/reference/project-files/metrics-views)

## Examples

### Simple Aggregation
In the following example, `percentage_reactive_to_active_measure` uses the already defined measures `total_global_active_power_measure` and `total_global_reactive_power_measure` to calculate the percentage without having to recalculate the sum of the respective columns.
<img src = '/img/build/metrics-view/examples/explore-percent.png' class='rounded-gif' />
<br />

```yaml
  - name: percentage_reactive_to_active_measure
    display_name: Percent Reactive to Active Power
    requires: [total_global_active_power_measure, total_global_reactive_power_measure]
    expression:  total_global_reactive_power_measure / total_global_active_power_measure
    format_preset: percentage
```


### Window Function
If you are using a [window function](./windows), you'll need to define the measure that you are building the window for. In this example, we are getting the rolling sum of average voltage measurements for all time— that's a lot of volts! You can modify the frame to include fewer rows based on the order column. 

<img src = '/img/build/metrics-view/examples/window-example.png' class='rounded-gif' />
<br /> 

```yaml
  - name: rolling_sum_avg_voltage_all_time
    display_name: Rolling Sum Windowed Voltage Average
    expression: SUM(average_voltage_measure)
    requires: [average_voltage_measure]
    window:
      order: Date
      frame: RANGE BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW
    treat_nulls_as: -1
```


## Demo
[See this project live in our demo!](https://ui.rilldata.com/demo/rill-kaggle-elec-consumption/explore/household_power_consumption_metrics_explore)