---
title: "Advanced Expressions"
description: Tips & Tricks for Defining Metrics & Dimensions
sidebar_label: "Advanced Expressions"
sidebar_position: 20
---

## Overview

Within the metrcs view yaml, you can apply aggregate sql expressions to create derived metrics or non-aggregate expressions to adjust dimension settings. SQL expressions are specific to the underlying OLAP engine so keep that in mind when editing directly in the yaml. 

We continually get questions about common metric definitions and other tricks so will update this page frequently. [Please let us know](../../contact.md) if you have questions or are stuck on an expression so we can add more examples.

:::tip

Rill's modeling layer provides open-ended SQL compatibility for complex SQL queries. More details can be found in our [modeling section](../models/models.md).

:::

## Measure Expressions

Measure expressions can take any SQL numeric function, a set of aggregates and apply filters to create derived metrics. Reminder on basic expressions are available in the [create metrics view definition](metrics-view.md#measures).

### Metric Formatting

In addition to standard presents, you can also use `format_d3` to control the formatting of a measure in the metrics view using a [d3-format string](https://d3js.org/d3-format). If an invalid format string is supplied, measures will be formatted with `format_preset: humanize`. Measures cannot have both `format_preset` and `format_d3` entries. _(optional; if neither `format_preset` nor `format_d3` is supplied, measures will be formatted with the `humanize` preset)_

    - **Example**: to show a measure using fixed point formatting with 2 digits after the decimal point, your measure specification would include: `format_d3: ".2f"`.
    - **Example**: to show a measure using grouped thousands with two significant digits, your measure specification would include: `format_d3: ",.2r"`.
    - **Example**: to increase decimal places on a currency metric would include: `format_d3: "$.3f"`.

### Case Statements

One of the most common advanced measure expressions are `case` statements used to filter or apply logic to part of the result set. Use cases for case statements include filtered sums (e.g. only sum if a flag is true) and bucketing data (e.g. if between threshold x and y the apply an aggregate). 

An example case statement to only sum cost when a record has an impression would look like:
```yaml
  - label: TOTAL COST
    expression: SUM(CASE WHEN imp_cnt = 1 THEN cost ELSE 0 END)
    name: total_cost
    description: Total Cost
    format_preset: currency_usd
    valid_percent_of_total: true
```

### Quantiles

In addition to common aggregates, you may wish to look at the value of a metric within a certain band or quantile. In the example below, we can measure the P95 query time as a benchmark.

```yaml
  - label: "P95 Query time"
    expression: QUANTILE_CONT(query_time, 0.95)
    format_preset: interval_ms
    description: "P95 time (in sec) of query time"
```

### Fixed Metrics / "Sum of Max"

Some metrics may be at a different level of granularity where a sum across the metric is no longer accurate. As an example, perhaps you have have a campaign with a daily budget of $5000 across five line items. Summing `daily_budget` column would give an inaccurate total of $25000 budget per day. For those familiar Tableau, this is referred to as a FIXED metric. 

To create the correct value, you can utilize DuckDB's unnest functionality. In the example below, you would be pulling a single value of `daily_budget` based on `campaign_id` to get the sum of budget for the day by campaign ids.

```
(select sum(a.val) as value from (select unnest(list(distinct {key: campaign_id, val: daily_budget })) a ))
```

:::note 

The syntax for fixed metrics is specific to DuckDB as an OLAP engine.

:::

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
## Dimension Expressions

To utilize an expression, replace the `column` property with `expression` and apply a non-aggregate sql expression. Common use cases would be editing fields such as `string_split(domain, '.')` or combining values `concat(domain, child_url)`.

 ```yaml
  - label: "Example Column"
    expression: string_split(domain, '.')
    description: "Edited Column"
```

### Unnest

 For multi-value fields, you can set the unnest property within a dimension. If true, this property allows multi-valued dimension to be unnested (such as lists) and filters will automatically switch to "contains" instead of exact match.

 ```yaml
  - label: "Example Column"
    column: multi_value_field
    description: "Unnested Column"
    unnest: true
```

### Druid Lookups

For those looking to add id to name mappings with Druid (as an OLAP engine), you can utilize expressions in your **Dimension** settings. Simply use the lookup function and provide the name of the lookup and id, i.e. `lookup(city_id, 'cities')`. Be sure to include the lookup table name in single quotes.

 ```yaml
  - label: "Cities"
    expression: lookup(city_id, 'cities')
    description: Cities"
```