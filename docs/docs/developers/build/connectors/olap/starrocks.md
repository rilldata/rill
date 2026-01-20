---
title: StarRocks
description: Power Rill dashboards using StarRocks
sidebar_label: StarRocks
sidebar_position: 25
---

[StarRocks](https://www.starrocks.io/) is an open-source, high-performance analytical database designed for real-time, multi-dimensional analytics on large-scale data. It supports both primary key and aggregate data models, making it suitable for a variety of analytical workloads including real-time dashboards, ad-hoc queries, and complex analytical tasks.

Rill supports connecting to an existing StarRocks cluster via a "live connector" and using it as an OLAP engine built against [external tables](/developers/build/connectors/olap#external-olap-tables) to power Rill dashboards.

:::

## Connect to StarRocks

When using StarRocks for local development, you can connect via connection parameters or by using the DSN.

After selecting "Add Data", select StarRocks and fill in your connection parameters. This will automatically create the `starrocks.yaml` file in your `connectors` directory and populate the `.env` file with `connector.starrocks.password`.

### Connection Parameters

```yaml
type: connector
driver: starrocks

host: <HOSTNAME>
port: 9030
username: <USERNAME>
password: "{{ .env.connector.starrocks.password }}"
catalog: default_catalog
database: <DATABASE>
ssl: false
```

### Connection String (DSN)

Rill can also connect to StarRocks using a DSN connection string. StarRocks uses MySQL protocol, so the connection string follows the MySQL DSN format:

```yaml
type: connector
driver: starrocks

dsn: "{{ .env.connector.starrocks.dsn }}"
```

The DSN format is:
```
starrocks://user:password@host:port/database
```

Or using MySQL-style format:
```
user:password@tcp(host:port)/database?parseTime=true
```

## Configuration Properties

| Property   | Description                                                       | Default           |
| ---------- | ----------------------------------------------------------------- | ----------------- |
| `host`     | StarRocks FE (Frontend) server hostname                           | Required          |
| `port`     | MySQL protocol port of StarRocks FE                               | `9030`            |
| `username` | Username for authentication                                       | Required          |
| `password` | Password for authentication                                       | -                 |
| `catalog`  | StarRocks catalog name (for external catalogs like Iceberg, Hive) | `default_catalog` |
| `database` | StarRocks database name                                           | -                 |
| `ssl`      | Enable SSL/TLS encryption                                         | `false`           |
| `dsn`      | Full connection string (alternative to individual parameters)     | -                 |

## External Catalogs

StarRocks supports external catalogs for querying data in Hive, Iceberg, Delta Lake, and other external data sources. To use an external catalog:

1. Set the `catalog` property to your external catalog name (e.g., `iceberg_catalog`)
2. Set the `database` property to the database within that catalog

```yaml
type: connector
driver: starrocks

host: starrocks-fe.example.com
port: 9030
username: analyst
password: "{{ .env.connector.starrocks.password }}"
catalog: iceberg_catalog
database: my_database
```

## Naming Mapping

StarRocks uses a three-level hierarchy: Catalog > Database > Table. In Rill's API:

| Rill Parameter   | StarRocks Concept | Example                              |
| ---------------- | ----------------- | ------------------------------------ |
| `database`       | Catalog           | `default_catalog`, `iceberg_catalog` |
| `databaseSchema` | Database          | `my_database`                        |
| `table`          | Table             | `my_table`                           |

## Troubleshooting

### Connection Issues

If you encounter connection issues:

1. Verify the FE node hostname and port (default: 9030)
2. Check that your user has appropriate permissions
3. Ensure network connectivity to the StarRocks FE node
4. For SSL connections, verify SSL is enabled on the StarRocks server


## Known Limitations

- **Model execution**: Model creation and execution is not yet supported. This feature is under development.

:::info Need help connecting to StarRocks?

If you would like to connect Rill to an existing StarRocks instance, please don't hesitate to [contact us](/contact). We'd love to help!

:::
