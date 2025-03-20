---
title: "Window Functions"
description: Tips & Tricks for Metric Formatting
sidebar_label: "Window Functions"
sidebar_position: 05
---

In addition to standard metrics, it is possible to define running window calculations of your data whether you are looking to monitor a cumulative trend, smooth out fluctuations, etc. You'll need to navigate to the Code view in order to create a windowed measure.

Please review the reference documentation, [here.](/reference/project-files/metrics-view)


## Example
In the below example, bids is another measure defined in the metrics view and we are getting the previous week and current date's values and averaging them. This allows us to remove any short term trends to detect real patterms. You'll need to add the [`requires`](./referencing) array parameter to use reference another measure.

<img src = '/img/build/metrics-view/examples/explore-percent.png' class='rounded-gif' />
<br /> update this photo

```yaml
  - name: bids_7day_rolling_avg
    display_name: 7 Day Bid rolling avg
    expression: AVG(total_bids)
    requires: [total_bids]
    window:
      order: "__time"
      frame: RANGE BETWEEN INTERVAL 6 DAY PRECEDING AND CURRENT ROW
```


Another example is using a rolling sum with no bounding preceding rows, aka your whole data. This will be a cumilative sum of all of your measures data, in this case it is the average voltage measure.

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
```
## Demo
[See this project live in our demo!](https://ui.rilldata.com/demo/rill-kaggle-elec-consumption/explore/household_power_consumption_metrics_explore)