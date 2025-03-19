---
title: "Fixed Metrics"
description: Tips & Tricks for Metric Formatting
sidebar_label: "Fixed Metrics"
sidebar_position: 04
---


### Fixed Metrics / "Sum of Max"

Some metrics may be at a different level of granularity where a sum across the metric is no longer accurate. As an example, perhaps you have have a campaign with a daily budget of $5000 across five line items. Summing `daily_budget` column would give an inaccurate total of $25000 budget per day. For those familiar Tableau, this is referred to as a FIXED metric. 

To create the correct value, you can utilize DuckDB's unnest functionality. In the example below, you would be pulling a single value of `daily_budget` based on `campaign_id` to get the sum of budget for the day by campaign ids.

```
(select sum(a.val) as value from (select unnest(list(distinct {key: campaign_id, val: daily_budget })) a ))
```

:::note 

The syntax for fixed metrics is specific to DuckDB as an OLAP engine.

:::