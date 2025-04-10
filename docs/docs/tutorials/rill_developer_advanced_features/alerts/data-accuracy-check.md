---
title: "Data Accuracy"
description: Project Maintanence
sidebar_label: " Data Accuracy Alerting"
tags:
  - CLI
  - Administration
---

Another important type of alert is a custom SQL alert that allows you to check for data accuracy. You will want to check for a historical data point to use as your reference point. While it is the responsibilty of the user who is making model changes to ensure data accuracy, there might be occasions that this is not done properly, and faulty models are pushed to Rill Cloud. 

## SQL Alerts
Alerts in YAML allow for a custom SQL to be created. When the result set is **not empty**, the alert will trigger. In our example, let's use the ClickHouse dashboard to find a historic point. In the below screenshow, we are using the following filters:
- author_date=2024-10-18
- author_name="Alexey Milovidov"
  
  With a resulting 196 total number of added lines or SUM(added_lines)


<img src = '/img/tutorials/alert/alert-sql.png' class='rounded-gif' />
<br />

Using this as a SQL alert, we can do something like:
```SQL
      SELECT 
        'Expected 196 added lines, got ' || SUM(added_lines) AS error_message
      FROM CH_incremental_joined_model
      WHERE CAST(author_date as DATE) = '2024-10-18'
        AND author_name = 'Alexey Milovidov'
      GROUP BY author_name
      HAVING SUM(added_lines) != 196
```

This query will return nothing when the value equals 196, however when it doesn't, the alert will return a non-empty result and trigger the alert. Depending on how you set up the recipients, this can be sent via email or to a slack channel to alert your admins that the current data is **incorrect** and needs to be rolled back or updated.

For the full alert file, see the Rill Tutorial GitHub repository, [here.](https://github.com/rilldata/rill-examples/tree/demo/my-rill-tutorial)