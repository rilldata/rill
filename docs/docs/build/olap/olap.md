---
title: "OLAP Engines"
description: Configure the OLAP engine used by Rill
sidebar_label: "OLAP Engines"
sidebar_position: 00
---


## What is OLAP?

OLAP (or Online Analytical Processing) is a computational approach designed to enable rapid, multidimensional analysis of large volumes of data. With OLAP, data is typically organized into cubes instead of traditional two-dimensional tables, which can facilitate complex queries and data analysis in a way that is significantly more efficient and user-friendly for analytical tasks. In particular, OLAP databases can be especially well suited for BI use cases that require deep, multi-dimensional analysis or real-time / user-facing analytics and applications. Additionally, many modern OLAP databases are optimized to ingest large volumes of data, execute low-latency queries with high throughput, and process billions of rows quickly with an emphasis on speed and efficiency in data retrieval. 

Unlike traditional relational databases or data warehouses that are optimized for transaction processing (with a focus on CRUD operations), OLAP databases are designed for query speed and complex analysis. Rather than storing data in a row-oriented manner, optimizing for transactional efficiency and operational queries, most OLAP databases are columnar and use pre-aggregated multidimensional cubes to speed up analytical queries. This allows a broad range of ad hoc queries and analysis to be performed without needing predefined schemas that are tailored to specific queries and it's this flexibility that enables the highly interactive slice-dicing and exploration of data that powers Rill dashboards. This paradigm allows OLAP to be particularly well-suited for organizations and teams that want to dive deep into and understand their data to support decision-making processes, where speed and flexibility in the actual data analysis are important. 

:::info Want to see OLAP in action?

Check [here](https://www.rilldata.com/case-studies) to see examples of use cases that can be powered by OLAP.

:::


## Available OLAP Engines

Rill supports the use of several different OLAP engines to power your dashboards in Rill, including:
- DuckDB
- Druid
- ClickHouse

:::note Additional OLAP Engines

Rill is continually evaluating additional OLAP engines that can be added. For a full list of OLAP engines that we support, you can refer to our [OLAP Engines](/reference/olap-engines) page. If you don't see an OLAP engine that you'd like to use, please don't hesitate to [reach out](contact.md)!

:::


## External OLAP tables

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