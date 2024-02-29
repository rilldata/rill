---
title: DuckDB
description: Power Rill dashboards using DuckDB (default)
sidebar_label: DuckDB
sidebar_position: 1
---

## How to configure credentials in Rill

 A live connection enables users to discover existing tables and perform OLAP queries directly on the engine without transfering data to another OLAP engine.

OLAP drivers can be configured by passing additional `--db-driver` config in `rill start` CLI command. 


### Configure credentials for ClickHouse local development

Steps for configuring a Rill and ClickHouse connection

```bash
# Connecting to clickhouse local
rill start --db-driver clickhouse --db "clickhouse://localhost:9000"
```

### Configure credentials for deployments on Rill Cloud

Steps for configuring a Rill and ClickHouse cloud connection

```bash
# Connecting to ClickHouse cluster 
rill start --db-driver clickhouse --db "<clickhouse://<host>:<port>?username=<username>&password=<pass>>"
```
This would open up browser and shows all the existing ClickHouse tables in Rill. Dashboards can be then created on top of existing source.

Note: Data modeling is not supported for ClickHouse driver at the moment.


### Deploying a ClickHouse project to Rill cloud

The driver and the dsn can be passed in the `rill deploy` command as below 

```bash
rill deploy --prod-db-driver clickhouse --prod-db-dsn <clickhouse_dsn>
```