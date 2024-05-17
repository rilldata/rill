---
title: "Custom API"
description: Create your own API to pull data out in flexible manner 
sidebar_label: "Custom API"
sidebar_position: 30
---

Rill allows you to create custom APIs to pull data out in a flexible manner. You can write custom SQL queries and expose them as an API endpoint.

## Create a custom API

To create a custom API, create a new yaml file under `apis` directory in your Rill project. Currently, 
we support two types of custom APIs:

1. **SQL API**: You can write a SQL query and expose it as an API endpoint. This is useful when you want to directly 
    write queries against a [model](/build/models/models.md) that you have created. It should have the following structure:
    
    ```yaml
    type: api
    sql: SELECT abc FROM my_table
    ```
    where `my_table` is your model name. Read more details about [SQL apis](./sql-api.md).

2. **Metrics SQL API**: You can write a SQL query referring to metrics definition and dimensions defined in the [metrics view](/build/dashboards/dashboards.md). 
It should have the following structure:
    
    ```yaml
    type: api
    metrics_sql: SELECT dimension, AGGREGATE(measure) FROM my_metrics GROUP BY dimension
    ```
    where `my_metrics` is your metrics view name, `measure` is a custom metrics that you have defined. 
    Read more details about [Metrics SQL API](./metrics-sql-api.md).

## How to use custom APIs
Refer to the integration docs [here](/integrate/custom-api.md) to learn how to use custom APIs in your application.