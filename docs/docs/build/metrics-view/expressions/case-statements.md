---
title: "Case Statements"
description: Tips & Tricks for Metric Formatting
sidebar_label: "Case Statements"
sidebar_position: 02
---

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