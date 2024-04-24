---
title: "Metrics SQL Syntax and Capabilities"
description: Learn how to utilize SQL for extracting data from metrics views effectively.
sidebar_label: "Metrics SQL"
sidebar_position: 60
---


<!-- WARNING: There are links to this page in source code. If you move it, find and replace the links and consider adding a redirect in docusaurus.config.js. -->

Metrics SQL is a specialized SQL dialect designed exclusively for querying data from metrics views.

## Querying Fundamentals
Metrics SQL transforms queries that reference `dimensions` and `measures` within a `metrics view` into their corresponding database columns or expressions. This transformation is based on the mappings defined in a metrics view YAML configuration, enabling reuse of dimension or measure defintions. Additionally any security policies defined in metrics_view is also inherited.

## Example: Crafting a Metrics SQL Query

Consider a metrics view configured as follows:
```
kind: metrics_view
title: Ad Bids
model: ad_bids
timeseries: timestamp
dimensions:
  - label: Publisher
    name: publisher
    column: toUpper(publisher)
    description: ""
  - label: Domain
    name: domain
    column: domain
    description: ""
measures:
  - name: total_records
    label: Total records
    expression: COUNT(*)
    description: ""
    format_preset: humanize
    valid_percent_of_total: true
```

To query this view, a user might write a Metrics SQL query like:
```
SELECT publisher, domain, AGGREGATE(total_records) FROM metrics_view GROUP BY publisher, domain
```
This Metrics SQL is internally translated to a standard SQL query as follows:
```
SELECT toUpper(publisher) AS publisher, domain AS domain, COUNT(*) AS total_records FROM ad_bids GROUP BY publisher, domain
```

## Supported SQL Features

- **SELECT** statements with plain `dimension` and `measure` references. The measures need to be accessed with a custom function `AGGREGATE(measure_name)`
- A single **FROM** clause referencing a `metrics view`.
- **WHERE** clause that can reference selected `dimensions` only.
- Operators in **WHERE** and **HAVING** clauses include `=`, `!=`, `>`, `>=`, `<`, `<=`, IN, LIKE, AND, OR, and parentheses for structuring the expression.
- **HAVING** clause for filtering on aggregated results, referencing selected dimension and measure names. Supports the same expression capabilities as the WHERE clause.
- **ORDER BY** clause for sorting the results.
- **LIMIT** and **OFFSET** clauses for controlling the result set size and pagination.


**Caution** : The Metrics SQL feature is currently evolving. We are dedicated to enhancing the synatx by introducing additional SQL features, while striving to maintain support for existing syntax. However, please be advised that backward compatibility cannot be guaranteed at all times. Additionally, users should be aware that there may be untested edge cases in the current implementation. We appreciate your understanding as we work to refine and improve this feature.

## Security and Compliance
Queries executed via Metrics SQL are subject to the security policies and access controls defined in the metrics view YAML configuration, ensuring data security and compliance.


## Limitations
Metrics SQL is specifically designed for querying metrics views and may not support all features found in standard SQL. Its primary focus is on providing an efficient and easy way to extract data within the constraints of metrics view configurations.