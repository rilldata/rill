---
title: Databricks
description: Power Rill dashboards using Databricks
sidebar_label: Databricks
sidebar_position: 02
---

[Databricks](https://docs.databricks.com) is a unified data and AI platform built on top of Apache Spark and the lakehouse architecture, with Unity Catalog for governance and Databricks SQL for analytics. Rill supports connecting to Databricks as a live connector, allowing you to build metrics views and dashboards directly on top of existing Databricks tables — no data movement required.

:::tip Databricks as a Live Connector vs. Data Source Connector
Rill supports Databricks in two distinct modes:

- **OLAP (Live Connector)** — Rill queries Databricks directly at dashboard load time. No data is ingested into Rill. Use this when your data is already modeled and optimized in Databricks and you want Rill as a visual layer on top. Set `olap_connector: databricks` in `rill.yaml`.

- **Data Source Connector** — Rill extracts data from Databricks and ingests it into its embedded engine (DuckDB). Use this when you want Rill to manage the data pipeline, apply transformations via SQL models, or combine Databricks data with other sources. See the [Databricks data source docs](/developers/build/connectors/data-source/databricks).

In general, use the live connector if your Databricks tables are already production-ready and large. Use data source ingestion if you need to transform, join, or enrich the data before building dashboards.
:::

## Connect to Databricks

After selecting "Add Data", select Databricks and fill in your connection parameters. This will automatically create the `databricks.yaml` file in your `connectors` directory and populate the `.env` file with your credentials.

For more information on supported parameters, see our [Databricks connector YAML reference docs](/reference/project-files/connectors#databricks).

```yaml
type: connector
driver: databricks

host: "dbc-xxxxxxxx-xxxx.cloud.databricks.com"
http_path: "/sql/1.0/warehouses/xxxxxxxxxxxxxxxx"
token: "{{ .env.DATABRICKS_TOKEN }}"
catalog: "main"                                # optional
schema: "default"                              # optional
```

:::tip Finding your connection parameters
In the Databricks workspace, navigate to **SQL Warehouses**, select the warehouse you want to use, and open the **Connection details** tab. The **Server hostname** is your `host`, the **HTTP path** is your `http_path`, and you can generate a personal access token under **User Settings → Developer → Access tokens**.
:::

### Connection String (DSN)

For advanced configuration, you can specify a single DSN instead of the individual fields above. The DSN cannot be combined with `host`, `http_path`, `token`, `catalog`, or `schema`.

```yaml
type: connector
driver: databricks

dsn: "{{ .env.DATABRICKS_DSN }}"
```

```bash
# .env
DATABRICKS_DSN=token:dapi...@dbc-xxxxxxxx-xxxx.cloud.databricks.com:443/sql/1.0/warehouses/xxxxxxxxxxxxxxxx?catalog=main&schema=default
```

See the [Databricks SQL Go driver](https://github.com/databricks/databricks-sql-go) documentation for the full list of supported DSN parameters.

## Build a Metrics View on Databricks

Once connected, set Databricks as the OLAP connector in your `rill.yaml`:

```yaml
olap_connector: databricks
```

Then create a metrics view that references a Databricks table:

```yaml
type: metrics_view

connector: databricks
database: main                # Unity Catalog (optional; defaults to the connector's catalog)
database_schema: my_schema    # Schema within the catalog
model: my_table

timeseries: created_at
dimensions:
  - column: region
  - column: category
measures:
  - name: total_revenue
    expression: SUM(revenue)
```

In Databricks terminology, `database` maps to the **catalog**, `database_schema` maps to the **schema**, and `model` maps to the **table**. Measure expressions must use [Databricks SQL](https://docs.databricks.com/aws/en/sql/language-manual/) syntax.

:::note
Rill supports metrics views directly on Databricks as a live connector. Incremental models and partitioned ingestion are not supported in live connector mode.
:::
