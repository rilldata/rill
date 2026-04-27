---
title: DuckLake
description: Power Rill dashboards using DuckLake
sidebar_label: DuckLake
sidebar_position: 12
---

[DuckLake](https://ducklake.select/) is an open lakehouse format built on DuckDB. A DuckLake keeps table data as Parquet files in object storage (local, S3, GCS, or Azure) while the catalog (schemas, snapshots, statistics) lives in a separate database — DuckDB, SQLite, PostgreSQL, or MySQL. Rill connects to DuckLake through the DuckDB driver and uses it as a live OLAP engine, so no data is ingested into Rill and all queries are pushed down to DuckDB against your lake.

:::note DuckLake uses the DuckDB Driver
DuckLake connectors use `driver: duckdb` under the hood. The difference from a standard DuckDB connector is the `attach` clause, which points DuckDB at your DuckLake catalog and data path.
:::

## Configuring Rill Developer with DuckLake

Create the connector via **Add Data → DuckLake** in the UI. Rill will generate a `connectors/ducklake.yaml` file and set DuckLake as the default OLAP engine in `rill.yaml`.

```yaml
type: connector
driver: duckdb

attach: "'ducklake:metadata.ducklake' (DATA_PATH 'data/')"
```

The `attach` clause is passed directly to DuckDB's `ATTACH` statement. It must begin with the `ducklake:` metadata backend and should include a `DATA_PATH` pointing at the directory (local or object storage) that holds your Parquet files.

### Supported Metadata Backends

DuckLake can store its catalog in any of the following:

- **DuckDB** — `'ducklake:metadata.ducklake'`
- **SQLite** — `'ducklake:sqlite:metadata.sqlite'`
- **PostgreSQL** — `'ducklake:postgres:dbname=ducklake host=...'`
- **MySQL** — `'ducklake:mysql:host=... user=... database=ducklake'`

For cloud-hosted catalogs, credentials for the metadata database can be injected via environment variables, e.g. `host={{ .env.DUCKLAKE_PG_HOST }}`.

### Supported Data Paths

The `DATA_PATH` in your `attach` clause can point at either local storage or any object store supported by DuckDB:

- Local filesystem — `DATA_PATH 'data/'`
- S3 — `DATA_PATH 's3://my-bucket/ducklake/'`
- GCS — `DATA_PATH 'gs://my-bucket/ducklake/'`
- Azure Blob Storage — `DATA_PATH 'azure://my-container/ducklake/'`

See the [DuckLake docs](https://ducklake.select/docs/stable/duckdb/usage/connecting) for the full ATTACH syntax.

### Setting the Default OLAP Connection

Creating a DuckLake connector automatically sets `olap_connector` in your project's [rill.yaml](/reference/project-files/rill-yaml) to the new connector.

```yaml
olap_connector: ducklake
```

## Advanced Options

The `attach` clause is passed through to DuckDB and accepts the full set of DuckLake ATTACH options — see the [DuckLake ATTACH reference](https://ducklake.select/docs/stable/duckdb/usage/connecting) for the complete list.

Example with multiple options:

```yaml
type: connector
driver: duckdb

attach: "'ducklake:metadata.ducklake' (DATA_PATH 's3://my-bucket/ducklake/', OVERRIDE_DATA_PATH true, SNAPSHOT_VERSION '42')"
```

## Trying DuckLake Without Your Own Data

If you want to see DuckLake in action before pointing Rill at your own catalog, DuckDB hosts a public `lineitem` table from TPC-H (scale factor 3) as a read-only DuckLake. Load it in the [DuckDB browser visualizer](https://duckdb.org/visualizer/#resource_path=https%3A%2F%2Fblobs.duckdb.org%2Fdatalake%2Ftpch-sf3.ducklake&resource_type=ducklake&table_name=lineitem) to confirm the catalog is reachable, then point Rill at the same resource:

```yaml
type: connector
driver: duckdb

attach: "'ducklake:https://blobs.duckdb.org/datalake/tpch-sf3.ducklake'"
```

## Configuring Rill Cloud

When deploying a DuckLake-backed project to Rill Cloud:

1. Any secrets referenced in the `attach` clause (e.g. S3 credentials, Postgres passwords) should be set via `{{ .env.KEY_NAME }}` in your YAML and managed with the project `.env` file.
2. Use `rill env push` to sync local environment variables to your cloud deployment.
3. The `DATA_PATH` must be reachable from Rill Cloud — local filesystem paths will not deploy.

## Additional Notes

- DuckLake uses the same SQL dialect as DuckDB, so all standard DuckDB functions are available. [Measure definitions](/developers/build/metrics-view/#measures) should follow standard [DuckDB SQL](https://duckdb.org/docs/sql/introduction) syntax.
- Rill opens DuckLake in read-only mode by default. To allow Rill to create or modify tables in the lake, enable write mode in the connector advanced options.
- Combine DuckLake with [multiple OLAP engines](/developers/build/connectors/olap/multiple-olap) to power different dashboards from different catalogs in the same project.

:::info Need help connecting to DuckLake?

If you would like to connect Rill to DuckLake or need assistance with setup, please don't hesitate to [contact us](/contact). We'd love to help!

:::
