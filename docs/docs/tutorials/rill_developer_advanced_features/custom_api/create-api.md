---
title: "Custom APIs"
description:  "Creating the API"
sidebar_label: "Creating the API"
sidebar_position: 12
tags:
  - Rill Developer
  - Advanced Features
  - Tutorial
---

## Creating the custom API YAML in Developer

<img src = '/img/tutorials/api/create-api.gif' class='rounded-gif' />
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
  FROM advanced_commits_model 
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
    where author_date > '2024-07-01' 
    order by net_line_changes DESC 
    limit 10
```


:::tip 
Both of these SQL queries will return the same data, why? 

`Metric_sql` will implicity group by the aggregate metrics, while the SQL will not. Therefore we need to manually add the sum function and group by the author_name. 

:::

As discussed when creating the measure, we defined the name of the measure so we can use the name in the SQL query. You can test the API's output with the following syntax, `http://localhost:9009/v1/instances/default/api/<filename>`

Once you have confirmed that the local running APIs work as expected, we can select update [or push changes to GitHub repository](https://docs.rilldata.com/tutorials/rill_developer_advanced_features/advanced_developer/update-rill-cloud) to push the changes to your project.

:::note
If the `update` button is not available on your current UI, you can find this on the dashboard page!
:::

<img src = '/img/tutorials/api/api-status.png' class='rounded-gif' />
<br />

## Passing Arguments into the API
While the above example is an awesome way to retrieve data from an environment. APIs shine when allowing the user to customize the data that they retrieve. This is done via passing arguments into the SQL itself. Let's take another look at the metrics_sql example where instead of a set date, we can have the user define that.

```yaml
metrics_sql: |
  select 
    author_name, 
    net_line_changes 
    {{ if (.user.admin) }}
      , filename
    {{ end }}
  from advanced_metrics
    where author_date > '{{ .args.date }}' 
    order by net_line_changes DESC
  LIMIT {{ default 15 (get .args "limit") }}

```
You'll also see that if the user does not define a LIMIT, then we default to 15 rows. Also, if the requester is an admin or using a service token, it will also retrieve the filename column. This allows an end user to customize the data that is being retrieved and makes custom APIs much more useful.

If testing locally, you'd use:
```
http://localhost:9009/v1/instances/default/api/conditional_api?date=2024-10-01&limit=1
```

If deployed to Rill Cloud, see [testing api](test-api.md).

## Referencing an API in an API

Another use case for APIs is to create an API that references another that you can pass hard coded arguments into.

```yaml
type: api
name: api_reference_api_with_args

api: conditional_api
args: 
  date: '2025-01-01'
```

In the above example, the date argument is already defined, but we can still set the limit argument in the API. Some use cases for this is having a master API that allows for a `publisher_id`. Instead of recreating the master API several times, you can reference it using the above. 



For example:
```
http://localhost:9009/v1/instances/default/api/api_reference_api_with_args?limit=1
```