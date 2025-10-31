---
title: Using Multiple OLAP Engines
description: Using multiple OLAP Engines to power dashboards in the same project
sidebar_label: Using Multiple OLAP Engines
sidebar_position: 50
---

If you have access to another OLAP engine (such as [ClickHouse](/build/connectors/olap/clickhouse) or [Druid](/build/connectors/olap/druid), you have the option to either:
- Create dedicated projects that are powered by one specific OLAP engine (default)
- Use different OLAP engines _in the same project_ to power separate dashboards

On this page, we will walk through how to configure the latter.

### Why Multiple OLAP Engines?

There could be reasons why you wish to configure multiple OLAP engines within the same project:
- You have data sources that differ greatly in size but which you want to use within the same project. As a rule of thumb, DuckDB handles datasets _up to 50GB quite well_ and is performant. For much larger datasets, you may want a more enterprise-grade OLAP engine powering specific dashboards.
- You have existing datasets/tables from other OLAP stores that you wish to use in Rill, which may already be optimized, and which you do not want to separately ingest into Rill. Instead, you would like to create dashboards off these tables directly and have the OLAP engine power them.

:::info Don't see an OLAP engine?

If there's an OLAP engine you're interested in that isn't available, please don't hesitate to [contact us](/contact). We'd love to hear from you and learn more!

:::

## Enabling Multiple OLAP Engines

To configure multiple OLAP engines, you'll want to leave the <u>default</u> OLAP engine as [DuckDB](/build/connectors/olap/duckdb) in your project and configure dashboards that are powered by other OLAP engines individually (more on this below).

### Setting up your OLAP Engine connection string (DSN)

Before getting started, you'll need to first configure the appropriate connection string for each OLAP engine that you plan to use in Rill. Besides the built-in DuckDB OLAP engine, each OLAP engine should have its own `connector.<olap-engine>.dsn` variable that needs to be configured.

**For Rill Developer:**
- You can set these variables in your project's `.env` file or try pulling existing credentials locally using `rill env pull` if the project has already been deployed to Rill Cloud.

:::tip Getting DSN errors in dashboards after setting `.env`?

There might be instances where you've configured the project's `.env` file with the appropriate connection DSN strings but dashboards are still throwing errors. In these situations, try restarting Rill using the `rill start --reset` command.

:::

**For Rill Cloud:**
- You can pass in the appropriate DSN connection string for each required OLAP engine by using the `rill env configure` command.
- Alternatively, if the required `connector.<olap-engine>.dsn` parameters have been set in your project's `.env`, you can "push" these updated variables to your deployed project directly using `rill env push`.

:::info

Note that you must `cd` into the Git repository that your project was deployed from before running `rill env configure`.

:::

### Configuring DuckDB as the default OLAP engine

Not much needs to be done here as _DuckDB is the inherent default OLAP engine_ that is used by Rill. However, in case a different `olap_connector` is set in the project's `rill.yaml` file, this property should either be removed and/or set back to `duckdb`.

```yaml
olap_connector: duckdb
```

:::note rill.yaml

For more information about available configurations for `rill.yaml`, please see our [Project YAML](/reference/project-files/rill-yaml) reference documentation.

:::

### Setting the OLAP Engine in the metrics view YAML

For each metrics view that is using a separate OLAP engine (other than the default), you'll want to set the `connector` and `table` properties in the underlying [metrics view YAML](/reference/project-files/metrics-views) configuration to the OLAP engine and corresponding [external table](/build/connectors/olap#external-olap-tables) that exists in your OLAP store, respectively.

```yaml
type: metrics_view
title: <metrics_view_name>
connector: <olap_engine>
table: <external_table_in_olap>
...
```
