---
title: "Common Expressions"
description: Tips & Tricks for Defining Metrics & Dimensions
sidebar_label: "Common Expressions"
sidebar_position: 20
---

## Overview

Within the dashboard yaml, you can apply aggregate sql expressions to create derived metrics or non-aggregate expressions to adjust dimension settings. SQL expressions are specific to the underlying OLAP engine so keep that in mind when editing directly in the yaml. 

We continually get questions about common metric definitions and other tricks so will update this page frequently. [Please let us know](../../contact.md) if you have questions or are stuck on an expression so we can add more examples.

## Metric Expressions

### Fixed Metrics / "Sum of Max"

Some metrics may be at a different level of granularity where a sum across the metric is no longer accurate. As an example, perhaps you have have a campaign with a daily budget of $5000 across five line items. Summing `daily_budget` column would give an inaccurate total of $25000 budget per day. For those familiar Tableau, this is referred to as a FIXED metric. 

To create the correct value, you can utilize DuckDB's unnest functionality. In the example below, you would be pulling a single value of `daily_budget` based on `campaign_id` to get the sum of budget for the day by campaign ids.

```
(select sum(a.val) as value from (select unnest(list(distinct {key: concat(campaign_id), val: daily_budget })) a ))
```

Note: the syntax for fixed metrics is specific to DuckDB as an OLAP engine.

## Dimensions Expressions

For those looking to add id to name mappings, you can utilize expressions in your Dimension settings.  To utilize an expression, replace the `column` property with `expression` and apply a non-aggregate sql expression. Common use cases would be editing fields such as `string_split(domain, '.')` or combining values `concat(domain, child_url)`.

### Druid Lookups

For those looking to add id to name mappings with Druid as an OLAP engine, you can utilize expressions in your Dimension settings. Simply use the lookup function and provide the name of the lookup and the id like - `lookup(city_id, 'cities')`. Be sure to include the lookup table name in single quotes.
