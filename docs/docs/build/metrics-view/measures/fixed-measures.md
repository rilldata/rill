---
title: "Fixed Measures"
description: Tips & Tricks for Fixed Measures
sidebar_label: "Fixed Measures"
sidebar_position: 35
---

Some measures may be at a different level of granularity where a sum across the measure is no longer accurate. As an example, perhaps you have a campaign with a daily budget of $5000 across five line items. Summing the `daily_budget` column would give an inaccurate total of $25,000 budget per day. For those familiar with Tableau, this is referred to as a `FIXED metric`. 

<img src = '/img/build/metrics-view/examples/incorrect-sum.png' class='rounded-gif' />
<br />

To create the correct value, you can utilize DuckDB's unnest functionality. In the example below, you would be pulling a single value of `daily_budget` based on `campaign_id` to get the sum of budget for the day by campaign ids. Note that you can use multiple keys if your granularity is defined by multiple dimensions.

```yaml 
expression: |
    select
        sum(a.val) as value
    from
        ( select unnest(
                list(distinct {
                    key: campaign_id,
                    val: daily_budget 
                    })
                ) a
            )
```

:::note 

The syntax for fixed metrics is specific to DuckDB as an OLAP engine as it requires DuckDB-specific commands. However, you can create a similar SQL expression using a different OLAP engine, too!

:::
Please review the reference documentation, [here.](/reference/project-files/metrics-views)


## Example

In the following example, each publishing company has a monthly minimum guarantee. As you'll see in the measure, `incorrect_sum_of_guarantee`, you'll get an incorrect value as this will sum multiple values as there are multiple shows and days per publisher. Another workaround would be to use MIN, MAX, or AVG, but when selecting multiple publishers, the values will not be accurate. 

<img src = '/img/build/metrics-view/examples/selecting-publishers.png' class='rounded-gif' />
<br />

```yaml
  - name: incorrect_sum_of_guarantee
    expression: sum(min_guarantee_usd)
    format_preset: currency_usd
    valid_percent_of_total: false
    
  - name: guarantee_usd_measure_monthly
    display_name: Monthly Minimum Guarantee USD
    description: Total minimum guarantee in USD recorded in the dataset.
    expression: > 
              SELECT SUM(a.val) AS value 
                FROM (
                    SELECT unnest(
                        list(distinct {
                            key: publisher_id,
                            month: date_trunc('month', date),
                            val: min_guarantee_usd
                             })
                    ) a
                )
    format_preset: currency_usd
    valid_percent_of_total: false
```

## Demo
[See this project live in our demo!](https://ui.rilldata.com/demo/sample-podcast-project/explore/podcast_explore)