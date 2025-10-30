---
title: OLAP Engines
sidebar_label: OLAP Engines
sidebar_position: 00
---

## Overview

An online analytical processing (OLAP) database is optimized for fast, complex queries over large datasets. Since ingesting and transforming data can be time-consuming, Rill uses an OLAP engine to store and query data locally (or remotely) for dashboards. By default, Rill uses an embedded instance of [DuckDB](https://duckdb.org/), a fast in-process analytical database, allowing you to ingest, transform, and model GBs of data locally, in seconds. 

However, depending on your use case, you may want to use a different OLAP engine for querying your data (especially if you've already ingested the data into a separate OLAP-enabled data warehouse). Rill also supports reading from the following remote OLAP engines to power dashboards:

- [ClickHouse](https://clickhouse.com/docs/en/intro) - a column-oriented database for OLAP use cases (either self-hosted or via [ClickHouse Cloud](https://clickhouse.com/cloud))
- [Druid](https://druid.apache.org/docs/latest/design/) - a high-performance, real-time analytics database designed for fast ingestion and querying of large datasets
- [Pinot](https://docs.pinot.apache.org/) - a real-time distributed OLAP datastore designed for low-latency, high-throughput analytics

:::info

OLAP engines in Rill currently only work with [managed dashboards](../../deploy/deploy-dashboard/existing-project.md) on [Rill Cloud](../../deploy/rill-cloud.md).

:::

:::tip Switching back and forth between OLAP Engines

You can switch back and forth between different OLAP engine types for dashboards, but will need to run `rill deploy` again if you do make a change. There are more details about changing the OLAP engine in Rill Cloud [here](../../reference/project-files/rill-yaml.md#olap_connector).

:::

## DuckDB
## ClickHouse
### MotherDuck
### Druid
### Pinot
