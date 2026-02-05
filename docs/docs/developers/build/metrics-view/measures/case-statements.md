---
title: "Case Statements and Filters"
description: Tips & Tricks for Case Statements
sidebar_label: "Case Statements and Filters"
sidebar_position: 20
---

One of the most common advanced measure expressions is the use of [`case`](https://duckdb.org/docs/stable/sql/expressions/case.html) statements and [`filters`](https://duckdb.org/docs/stable/sql/query_syntax/filter.html) to filter or apply logic to part of the result set. Use cases for case statements include filtered sums (e.g., only sum if a flag is true) and bucketing data (e.g., if between threshold x and y, then apply an aggregate). While similar, case statements give you a bit more flexibility as they allow you to set a custom value depending on the case. See below for some examples!

<img src = '/img/build/metrics-view/examples/case-example.png' class='rounded-gif' />
<br />

Please review the reference documentation, [here.](/reference/project-files/metrics-views)

## Examples

### Case Statements
The following expression sums of the values of Global_active_power only when considered to be a lower value.

```yaml
  - name: total_low_active_power_measure
    display_name: Total Low Global Active Power
    description: Total sum of Global Active Power where considered Low
    expression: SUM(CASE WHEN GAP_category = 'Low' THEN Global_active_power ELSE 0 END)
    format_preset: humanize
    valid_percent_of_total: true
```

The following expression only considers the total value of users who are identified.

```yaml
  - name: total_value_for_identified_users
    display_name: Total Value for Identified Users
    description: Total Sum of Value for Identified Users
    expression: SUM(CASE WHEN user_id != '' OR user_id IS NOT NULL THEN value ELSE 0 END)
    format_preset: humanize
    valid_percent_of_total: true
```

The following expression modifies the value of the column based on the value of column XX

```yaml
  - name: modify_value
    display_name: Arithmetic on Value 
    description: Arithmetic on Value based on XX
    expression: | 
              SUM(
                CASE 
                   WHEN XX = 'multiply_10' THEN Value * 10
                   WHEN XX = 'multiply_2' THEN Value * 2
                   WHEN XX = 'divide_5' THEN Value / 5
                END
                )
    format_preset: humanize
    valid_percent_of_total: true
```


### Filters
Similar to the above case statements, you can use the filter expression to filter the data on a specific column's value. However, in the example where we are explicitly changing the value in the CASE statement, this is not possible using only a filter.

```yaml
  - name: total_low_active_power_measure
    display_name: Total Low Global Active Power
    description: Total sum of Global Active Power where considered Low
    expression: sum(Global_active_power) FILTER (WHERE GAP_category = 'Low')
    format_preset: humanize
    valid_percent_of_total: true
```

```yaml
  - name: total_value_for_identified_users
    display_name: Total Value for Identified Users
    description: Total Sum of Value for Identified Users
    expression: SUM(value) FILTER (WHERE user_id != '' OR user_id IS NOT NULL)
    format_preset: humanize
    valid_percent_of_total: true
```
## Demo
[See this project live in our demo!](https://ui.rilldata.com/demo/rill-kaggle-elec-consumption/explore/household_power_consumption_metrics_explore)