---
title: Powering your Metrics View
sidebar_label: Underlying Model/Table
sidebar_position: 05
---

Once you have finished [building your model](/build/models), you can start to design your metrics view to create measures and dimensions that you can visualize in your dashboard. To do this effectively, you need to understand how to specify the underlying data source in your metrics view YAML, which depends on your OLAP engine.

## One Big Table Approach

For optimal dashboard performance and flexibility, we recommend modeling your data sources into a "One Big Table" – a granular resource that contains as much information as possible and can be rolled up in meaningful ways. This flexible approach enables ad hoc slice-and-dice discovery through Rill's interactive dashboard.


For more details on the One Big Table approach, see our [Models 101 guide](/build/models/models-101#one-big-table-and-dashboarding).

## OLAP Engines 

Rill supports [multiple OLAP engines](/connect/olap) and this is the key component that determines the YAML key you will use for your metrics view.

### DuckDB
If you are using DuckDB (i.e., you haven't added any custom live connectors), you'll continue to use the term "model" that you're familiar with from building models. In your metrics view YAML, you'll use `model` to define the model that powers your metrics view.


```yaml
# Metrics View YAML
# Reference documentation: https://docs.rilldata.com/reference/project-files/metrics-views

version: 1
type: metrics_view

model: example_model # Choose a table to underpin your metrics view
```

### Live Connectors

If you're using live connectors such as ClickHouse, MotherDuck, or Druid, you'll need to modify the default YAML configuration. Instead of using `model`, you'll use `table`. You'll also need to specify the `connector` and `database_schema` fields.

```yaml
# Metrics View YAML
# Reference documentation: https://docs.rilldata.com/reference/project-files/metrics-views

version: 1
type: metrics_view

connector: clickhouse
database_schema: billing
table: events # Choose a table to underpin your metrics view
```

For more information, refer to our [metrics view YAML configuration](/reference/project-files/metrics-views).

## Visual Metrics Editor

If you're using the UI to select your table, choosing a live connector will automatically configure the YAML with the correct `table`, `connector`, and `database_schema` fields.

<img src='/img/build/metrics-view/clickhouse-metrics-view.png' class='rounded-gif' />
<br />