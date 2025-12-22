---
title: Powering your Metrics View
sidebar_label: Underlying Model/Table
sidebar_position: 05
---

Once you have finished [building your model](/build/models), you can create a metrics view to define measures and dimensions for your dashboard. The way you specify the underlying data source depends on your OLAP engine.

## Choosing Your Data Source

Rill supports [multiple OLAP engines](/build/connectors/olap), and the engine you're using determines which YAML property you'll use in your metrics view:

- **Use `model`** for DuckDB and Rill-managed ClickHouse
- **Use `table`** for self-managed live connectors

## DuckDB and Rill-Managed ClickHouse

For DuckDB (the default engine) and Rill-managed ClickHouse, use the `model` property to reference your data model:


```yaml
# Metrics View YAML
# Reference documentation: https://docs.rilldata.com/reference/project-files/metrics-views

version: 1
type: metrics_view

model: example_model # Choose a model to underpin your metrics view
```

## Self-Managed Live Connectors

For self-managed live connectors (like your own ClickHouse, MotherDuck, or Druid instance), use the `table` property and specify connection details:

```yaml
# Metrics View YAML
# Reference documentation: https://docs.rilldata.com/reference/project-files/metrics-views

version: 1
type: metrics_view


database: default
connector: clickhouse
database_schema: billing
table: events # Choose a table to underpin your metrics view
```

For more information, refer to our [metrics view YAML configuration](/reference/project-files/metrics-views).

## Visual Metrics Editor

If you're using the UI to select your table, choosing a live connector will automatically configure the YAML with the correct `table`, `connector`, and `database_schema` fields.

<img src='/img/build/metrics-view/clickhouse-metrics-view.png' class='rounded-gif' />
<br />