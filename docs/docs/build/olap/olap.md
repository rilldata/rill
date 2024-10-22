---
title: "OLAP Engines"
description: Configure the OLAP engine used by Rill
sidebar_label: "Connect OLAP Engines"
sidebar_position: 00
---

## How to connect to your OLAP Engine?

There are two ways to define an OLAP engine within Rill.

1. Set the [default OLAP engine](../../reference/project-files/rill-yaml#configuring-the-default-olap-engine) via the rill.yaml file.
2. Set the [OLAP engine](../../reference/project-files/explore-dashboards.md) for a specific dashboard.

The OLAP engine set on the dashboard will take precedence over the project-level defined OLAP engine.

## Available OLAP Engines

Rill supports the use of several different OLAP engines to power your dashboards in Rill, including:
- [DuckDB](/reference/olap-engines/duckdb.md)
- [Druid](/reference/olap-engines/druid.md)
- [ClickHouse](/reference/olap-engines/clickhouse.md)
- [Pinot](/reference/olap-engines/pinot.md)

:::note Additional OLAP Engines

Rill is continually evaluating additional OLAP engines that can be added. For a full list of OLAP engines that we support, you can refer to our [OLAP Engines](/reference/olap-engines) page. If you don't see an OLAP engine that you'd like to use, please don't hesitate to [reach out](contact.md)!

:::


## DuckDB

DuckDB is unique in that it can act as both a [source](../../reference/connectors/motherduck.md) and [OLAP engine](../../reference/olap-engines/duckdb.md) for Rill. If you wish to connect to existing tables in DuckDB though, you can simply use the [DuckDB connector](../../reference/connectors/motherduck.md#connecting-to-external-duckdb-as-a-source) to read a specific table or view from an accessible DuckDB database file and ingest the data into Rill.

:::warning

Rill will use DuckDB by default as an embedded OLAP engine but it is **not** currently possible to "bring your own DuckDB database" to be used as an alternative OLAP Engine. The internal DuckDB that Rill uses is hardcoded and not configurable (necessary for Rill Cloud consistency).

:::


## Druid

When Druid has been configured as the [default OLAP engine](../../reference/project-files/rill-yaml.md#configuring-the-default-olap-engine) for your project, any existing external tables that Rill can read and query should be shown through the Rill Developer UI. You can then create dashboards using these external Druid tables.

<div className="center-content">
![External Druid tables](/img/build/connect/external-tables/external-druid-table.png)
</div>

## ClickHouse

When ClickHouse has been configured as the [default OLAP engine](../../reference/project-files/rill-yaml.md#configuring-the-default-olap-engine) for your project, any existing external tables that Rill can read and query should be shown through the Rill Developer UI. You can then create dashboards using these external ClickHouse tables.

<div className="center-content">
![External ClickHouse tables](/img/build/connect/external-tables/external-clickhouse-table.png)
</div>

## Pinot

When Pinot has been configured as the [default OLAP engine](../../reference/project-files/rill-yaml.md#configuring-the-default-olap-engine) for your project, any existing external tables that Rill can read and query should be shown through the Rill Developer UI under `Tables` section in left pane. You can then create dashboards using these external Pinot tables.

<div className="center-content">
![External Pinot tables](/img/build/connect/external-tables/external-pinot-table.png)
</div>
