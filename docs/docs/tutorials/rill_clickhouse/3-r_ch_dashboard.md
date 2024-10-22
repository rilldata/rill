---
title: "3. Create Metrics-view and Dashboard in Rill"
sidebar_label: "3. Create Metrics-view and Dashboard in Rill"
sidebar_position: 4
hide_table_of_contents: false
tags:
  - OLAP:ClickHouse
---

## Let's get started!

If you noticed in the previous screenshot, we had a table called `uk_price_paid`. This is a dataset that is used in ClickHouse's Learning portal so we thought it was fitting to go ahead and continue on this dataset.

:::note
In the case that you have not already added this table to your local or Cloud database, please follow the step on [ClickHouse's site](https://clickhouse.com/docs/en/getting-started/example-datasets/uk-price-paid) for the steps to do so!
:::

### Create metrics-view

Let's create a metrics-view based on the table via the `Generate metrics via AI`.

<img src = '/img/tutorials/ch/ai-generate.gif' class='rounded-gif' />
<br />

### What are we looking at?

This is our metrics-view, where we can define measures and dimensions to be used on dashboards.  

```yaml
# Metrics view YAML
# Reference documentation: https://docs.rilldata.com/reference/project-files/explore-dashboards
# This file was generated using AI.

version: 1
type: metrics_view

title: UK Price Paid Metrics
connector: clickhouse
table: uk_price_paid
timeseries: date

dimensions:

    ...

measures:
    ...
```

While we go into more details in our [Rill Basics course](/tutorials/rill_basics/dashboard) and [our documentation](https://docs.rilldata.com/build/dashboards/), let's go over it quickly.

---

`timeseries` - This is our time column that is used on as our x-axis for graphs.

`connector` - this is our manually defined ClickHouse connector

`dimensions` - These are our categorial columns that we can use on the dashboard to filter and slice;

`measures` - These are our numerical aggregates defined in the metrics layer. We can see functions such as MAX(), COUNT(), and AVG() used on the underlying table.

### Create the explore dashboard

When you're ready, you can create the visualization on top of the metric layer. Let's select `create explore`. This will create a simple explore-dashboards/uk_price_paid_metrics_explore.yaml file that reads in all the dimensions and measures. For more information on the available key-pairs, please refer to the [reference documentation.](https://docs.rilldata.com/reference/project-files/explores)

---

### What can we do in Rill?
In our case, as we have generated this with AI so we can look through the description of the populated measures for more information. Based on this, we can find some specific information on the UK properties dataset at a glance, such as:

1. In 2023, What was the minimum/maximum detached property sold in London? [46.5K, 65.0M]
2. In 2023, What was the average price of deteached properties sold in London? How many? [2.5M, 981]

![img](/img/tutorials/ch/2023-london.png)

If we wanted to go further into details, we can even compare detached vs flat vs terraced properties using our compare feature. Based on the x-axis, we can drill down futher from the 2023 year into a specific month, week or even day.

![img](/img/tutorials/ch/2023-london-compare.png)

Or, if you want to compare time periods 2022 to 2023's total transactions. In the below screenshot, we selected the Total Transactions metric and enable the time-compare feature to see the delta, delta percent of change from two time periods.
![img](/img/tutorials/ch/time-compare.png)

These are just a few examples of what we can do with Rill, the options expand further and are discussed further in [Rill Advanced](https://docs.rilldata.com/tutorials/rill_learn_200/201_0). If you're interested I recommended reviewing the contents after finishing up this course.