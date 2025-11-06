---
title: "4. Create Explore Dashboard in Rill"
sidebar_label: "4. Create Explore Dashboard in Rill"
sidebar_position: 4
hide_table_of_contents: false
tags:
  - OLAP:ClickHouse
  - Tutorial
---


### Create the explore dashboard

When you're ready, you can create the visualization on top of the metric layer. Let's select `Create Explore dashboard`. This will create a simple explore-dashboards/uk_price_paid_metrics_explore.yaml file that reads in all the dimensions and measures. For more information on the available key-pairs, please refer to the [reference documentation.](https://docs.rilldata.com/reference/project-files/explore-dashboards)

---

### What can we do in Rill?
In our case, as we have generated this with AI, so we can look through the description of the populated measures for more information. Based on this, we can find some specific information on the UK properties dataset at a glance, such as:

1. In 2023, What was the minimum/maximum detached property sold in London? [46.5K, 65.0M]
2. In 2023, What was the average price of detached properties sold in London? How many? [2.5M, 981]


<img src = '/img/tutorials/ch/2023-london.png' class='rounded-gif' />
<br />

If we wanted to go further into details, we can even compare detached vs flat vs terraced properties using our compare feature. Based on the x-axis, we can drill down futher from the 2023 year into a specific month, week, or even day.


<img src = '/img/tutorials/ch/2023-london-compare.png' class='rounded-gif' />
<br />

Or, if you want to compare time periods 2022 to 2023's total transactions. In the below screenshot, we selected the Total Transactions metric and enable the time-compare feature to see the delta, delta percent of change from two time periods.

<img src = '/img/tutorials/ch/time-compare.png' class='rounded-gif' />
<br />
