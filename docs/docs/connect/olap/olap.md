---
title: OLAP Engines
sidebar_label: OLAP Engines
sidebar_position: 00
---

## What is OLAP?

OLAP stands for Online Analytical Processing. OLAP engines are database systems optimized for running fast, complex analytical queries on large datasets. Unlike traditional transactional databases (OLTP) that focus on handling many small read/write operations, OLAP systems are designed for read-heavy workloads involving aggregations, joins, and analytics across millions or billions of rows.

## OLAP in Rill

By default, Rill uses an embedded **DuckDB** database as its OLAP engine. DuckDB runs locally within Rill and provides extremely fast query performance for datasets up to hundreds of gigabytes. For most use cases, DuckDB's performance and ease of use make it the ideal choice.

However, Rill also supports connecting to external OLAP engines when you need:
- To query data that's already in a cloud data warehouse
- To handle datasets larger than what fits in local storage
- To leverage existing infrastructure and access controls
- To reduce data movement and egress costs

When using an external OLAP engine, Rill connects directly to your data warehouse and runs dashboard queries there, rather than ingesting data into a local DuckDB instance.

## Supported OLAP Engines

Rill supports the use of several different OLAP engines to power your dashboards and metrics layers, each offering unique advantages for different use cases and scale requirements:

### DuckDB
import DuckDB from '../../../static/img/reference/connectors/duckdb-logo.svg';

<DuckDB className="connector-icon"/>

[DuckDB](https://duckdb.org/) is an embedded analytical database designed for fast analytical queries and is the default OLAP engine used by Rill. It excels at running complex SQL queries on medium-sized datasets (up to several hundred gigabytes) with minimal setup and no external dependencies. For more information, please see our [DuckDB reference page](https://docs.rilldata.com/reference/olap-engines/duckdb).

### Clickhouse
import Clickhouse from '../../../static/img/reference/connectors/clickhouse.svg';

<Clickhouse className="connector-icon"/>

[ClickHouse](https://clickhouse.com/) is an open-source, column-oriented database management system designed for online analytical processing (OLAP) of queries and data. It is optimized for fast query performance on large datasets and is particularly well-suited for real-time analytics and business intelligence applications. For more information, please see our [ClickHouse reference page](https://docs.rilldata.com/reference/olap-engines/clickhouse).

### MotherDuck
### Druid
### Pinot
