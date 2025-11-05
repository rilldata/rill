---
title: Rolling Average Example Metric
tags:
- metrics
- code
- snippets
- duckdb
docs: https://docs.rilldata.com/build/metrics-view/measures/windows
hash: 9f515215c05db393869885fd9665388d6c6d5a67f79d70e5675f758e7d7847fb
---

```YAML
  - display_name: "Bids"
    name: total_bids
    expression: "sum(bid_cnt)"
    description: "Total Bids"
    format_preset: humanize
  - name: bids_7day_rolling_avg
    display_name: 7 Day Bid rolling avg
    expression: AVG(total_bids)
    requires: [total_bids]
    window:
      order: "__time"
      frame: RANGE BETWEEN INTERVAL 6 DAY PRECEDING AND CURRENT ROW
```
