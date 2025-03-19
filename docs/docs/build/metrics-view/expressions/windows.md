---
title: "Window Functions"
description: Tips & Tricks for Metric Formatting
sidebar_label: "Window Functions"
sidebar_position: 05
---

### Window Functions

In addition to standard metrics, it is possible to define running window calculations of your data whether you are looking to monitor a cumulative trend, smooth out fluctuations, etc.
In the below example, bids is another measure defined in the metrics view and we are getting the previous and current date's values and averaging them. 
```yaml
  - display_name: bids_1day_rolling_avg
    expression: AVG(bids)
    requires: [bids]
    window:
      order: timestamp
      frame: RANGE BETWEEN INTERVAL 1 DAY PRECEDING AND CURRENT ROW
```

all time historical sum
