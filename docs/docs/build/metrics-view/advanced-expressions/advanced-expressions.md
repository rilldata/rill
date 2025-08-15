---
title: "Advanced Expressions"
description: Tips & Tricks for Defining Metrics & Dimensions
sidebar_label: "Advanced Expressions"
sidebar_position: 00
---

## Overview

Within the metrics view YAML, you can apply aggregate SQL expressions to create derived metrics or non-aggregate expressions to adjust dimension settings. SQL expressions are specific to the underlying OLAP engine, so keep that in mind when editing directly in the YAML. 

:::note Examples in Docs
For most of the examples here, DuckDB is being used. However, most if not all of the functionality is possible on different OLAP engines. You will need to refer to that specific OLAP's reference documentation. Don't hesitate to reach out to us if you have any questions! 
:::


:::tip

Rill's modeling layer provides open-ended SQL compatibility for complex SQL queries. More details can be found in our [modeling section](/build/models/models.md).

:::

## Measure Expressions

Measure expressions can take any SQL numeric function, a set of aggregates, and apply filters to create derived metrics. A reminder on basic expressions is available in the [create metrics view definition](../metrics-view.md#measures).

See our dedicated examples and pages for the following advanced measures!
- **[Metric Formatting](/build/metrics-view/advanced-expressions/metric-formatting)**
- **[Case Statements and Filters](/build/metrics-view/advanced-expressions/case-statements)**
- **[Referencing Measures](/build/metrics-view/advanced-expressions/referencing)**
- **[Quantiles](/build/metrics-view/advanced-expressions/quantiles)**
- **[Fixed Metrics](/build/metrics-view/advanced-expressions/fixed-metrics)**
- **[Window Functions](/build/metrics-view/advanced-expressions/windows)**

  



## Dimension Expressions

To utilize an expression, replace the `column` property with `expression` and apply a non-aggregate SQL expression. Common use cases include editing fields such as `string_split(email, '@')[2]` to extract the domain from an email or combining values `concat(domain, child_url)` to get the full URL.

### DuckSQL functions

 ```yaml
  - name: domain
    display_name: Domain Name
    expression: string_split(email, '@')[2]
    description: "Extract the domain from an email"
```

See our dedicated examples and pages for more use cases.

- **[Unnest Dimensions](/build/metrics-view/advanced-expressions/unnesting)**
- **[Time Functions](/build/metrics-view/advanced-expressions/time-functions)**

### Clickable Dimension Links 
Adding an additional parameter to your dimension in the [metrics view](/build/metrics-view) can allow for clickable links directly from the dashboard.

```yaml
dimensions:
  - label: Company Url
    column: Company URL
    uri: true #if already set to the URL, also accepts SQL expressions
```
 <img src = '/img/build/dashboard/clickable-dimension.png' class='rounded-gif' />
<br />


## Druid Lookups

For those looking to add ID to name mappings with Druid (as an OLAP engine), you can utilize expressions in your **Dimension** settings. Simply use the lookup function and provide the name of the lookup and ID, i.e. `lookup(city_id, 'cities')`. Be sure to include the lookup table name in single quotes.

 ```yaml
  - label: "Cities"
    expression: lookup(city_id, 'cities')
    description: "Cities"
```