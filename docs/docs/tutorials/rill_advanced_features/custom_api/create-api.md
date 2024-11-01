---
title: "Custom APIs"
description:  "Creating the API"
sidebar_label: "Creating the API"
sidebar_position: 12
---

## Creating the custom API YAML in Developer

<img src = '/img/tutorials/303/create-api.gif' class='rounded-gif' />
<br />

Now that the access has been created, we can make the actual custom APIs. For this example, we'll keep it fairly simple and create 2 files as outlined in [our documentation](https://docs.rilldata.com/integrate/custom-apis/):

- SQL_api.yaml
- metrics_view_api.yaml


For the `SQL_api`, we will retreive the author's with the most net line changes from the model.
```sql
sql: |
  SELECT 
      author_name, 
      sum(net_line_changes) as net_line_changes,
  FROM advanced_commits___model 
    where author_date > '2024-07-01 00:00:00' 
    group by author_name 
    order by net_line_changes DESC  
    limit 10 
```



For `metrics_view_api`, we will use `advanced_metrics_view` and run the following SQL query:
```sql
metrics_sql: |
  SELECT 
      author_name, 
      net_line_changes 
  FROM advanced_metrics_view
    where author_date > '2024-07-01 00:00:00' 
    order by net_line_changes DESC 
    limit 10
```


:::tip 
Both of these SQL queries will return the same data, why? 

`Metric_sql` will implicity group by the aggregate metrics, while the SQL will not. Therefore we need to manually add the sum function and group by the author_name. 

:::

As discussed when creating the measure, we defined the name of the measure so we can use the name in the SQL query. You can test the API's output with the following syntax, `http://localhost:9009/v1/instances/default/api/<filename>`

Once you have confirmed that the local running APIs work as expected, we can select update [or push changes to GitHub repository](https://docs.rilldata.com/tutorials/rill_advanced_features/advanced_developer/update-rill-cloud) to push the changes to your project.

:::note
If the `update` button is not available on your current UI, you can find this on the dashboard page!
:::

![my-rill-project](/img/tutorials/303/api-status.png)

