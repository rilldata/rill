---
title: External OLAP Tables
description: Bringing in tables directly from your OLAP Store
sidebar_label: External OLAP Tables
sidebar_position: 30
---

import TwitterEmbed from '@site/src/components/TwitterEmbed';

## What are external OLAP tables?

Rill supports creating and powering dashboards using existing tables from alternative [OLAP engines](../../reference/olap-engines/olap-engines.md) that have been configured in a particular project. These tables are not managed by Rill, hence external, but allow users to bring in separate tables or datasets that might already exist in another preferred OLAP database of choice. This prevents the need of unnecessarily ingesting this data into Rill, especially if the table is already optimized for use by this other OLAP engine, and allowing Rill to connect to the data directly (and submit analytical queries).

<div className="center-content">
![Connecting to an external table](/img/build/connect/external-tables/external-olap-db.png)
</div>


## DuckDB

DuckDB is unique in that it can act as both a [source](../../reference/connectors/motherduck.md) and [OLAP engine](../../reference/olap-engines/duckdb.md) for Rill. If you wish to connect to existing tables in DuckDB though, rather than use an external table, you can simply use the [DuckDB connector](../../reference/connectors/motherduck.md#connecting-to-duckdb) to read the table and ingest the data into Rill. 

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