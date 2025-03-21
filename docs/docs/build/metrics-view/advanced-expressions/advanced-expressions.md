---
title: "Advanced Expressions"
description: Tips & Tricks for Defining Metrics & Dimensions
sidebar_label: "Advanced Expressions"
sidebar_position: 00
---

## Overview

Within the metrcs view yaml, you can apply aggregate sql expressions to create derived metrics or non-aggregate expressions to adjust dimension settings. SQL expressions are specific to the underlying OLAP engine so keep that in mind when editing directly in the yaml. 

:::note Examples in Docs
For most of the examples here, DuckDB is being used. However most if not all the functionality is possible on different OLAP engines. You will need to refer to that specific OLAP's reference documentation. Don't hestitate to reach out to us if you have any questions! 
:::


:::tip

Rill's modeling layer provides open-ended SQL compatibility for complex SQL queries. More details can be found in our [modeling section](/build/models/models.md).

:::

## Measure Expressions

Measure expressions can take any SQL numeric function, a set of aggregates and apply filters to create derived metrics. Reminder on basic expressions are available in the [create metrics view definition](../metrics-view.md#measures).

See our dedicated examples and pages for the following advanced measures!
- **[Metric Formatting](./metric-formatting)**
- **[Case Statements and Filters](./case-statements)**
- **[Referencing Measures](./referencing)**
- **[Quantiles](./quantiles)**
- **[Fixed Metrics](./fixed-metrics)**
- **[Window Functions](./windows)**

  



## Dimension Expressions

To utilize an expression, replace the `column` property with `expression` and apply a non-aggregate sql expression. Common use cases would be editing fields such as `string_split(email, '@')[2]` to extract the domain from an email or combining values `concat(domain, child_url)` to get the full URL.

 ```yaml
  - name: domain
    display_name: Domain Name
    expression: string_split(email, '@')[2]
    description: "Extract the domain from an email"
```
See our dedicated examples and pages for the following advanced dimensions!

- **[Unnest Dimensions](./unnesting)**

## Druid Lookups

For those looking to add id to name mappings with Druid (as an OLAP engine), you can utilize expressions in your **Dimension** settings. Simply use the lookup function and provide the name of the lookup and id, i.e. `lookup(city_id, 'cities')`. Be sure to include the lookup table name in single quotes.

 ```yaml
  - label: "Cities"
    expression: lookup(city_id, 'cities')
    description: "Cities"
```