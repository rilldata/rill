---
title: "Advanced Expressions"
description: Tips & Tricks for Defining Metrics & Dimensions
sidebar_label: "Advanced Expressions"
sidebar_position: 00
---

## Overview

Within the metrcs view yaml, you can apply aggregate sql expressions to create derived metrics or non-aggregate expressions to adjust dimension settings. SQL expressions are specific to the underlying OLAP engine so keep that in mind when editing directly in the yaml. 

We continually get questions about common metric definitions and other tricks so will update this page frequently. [Please let us know](/contact) if you have questions or are stuck on an expression so we can add more examples.

:::tip

Rill's modeling layer provides open-ended SQL compatibility for complex SQL queries. More details can be found in our [modeling section](/models/models).

:::

## Measure Expressions

Measure expressions can take any SQL numeric function, a set of aggregates and apply filters to create derived metrics. Reminder on basic expressions are available in the [create metrics view definition](../metrics-view.md#measures).






## Dimension Expressions

To utilize an expression, replace the `column` property with `expression` and apply a non-aggregate sql expression. Common use cases would be editing fields such as `string_split(domain, '.')` or combining values `concat(domain, child_url)`.

 ```yaml
  - label: "Example Column"
    expression: string_split(domain, '.')
    description: "Edited Column"
```


### Druid Lookups

For those looking to add id to name mappings with Druid (as an OLAP engine), you can utilize expressions in your **Dimension** settings. Simply use the lookup function and provide the name of the lookup and id, i.e. `lookup(city_id, 'cities')`. Be sure to include the lookup table name in single quotes.

 ```yaml
  - label: "Cities"
    expression: lookup(city_id, 'cities')
    description: Cities"
```